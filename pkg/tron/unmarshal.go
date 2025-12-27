package tron

import (
	"encoding"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// decoder handles type conversion from parsed values to Go types.
type decoder struct {
	classes map[string][]string
}

// unmarshal is the internal implementation of Unmarshal.
func unmarshal(data []byte, v interface{}) error {
	// Validate input
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}

	// Tokenize
	tokens, err := tokenize(string(data))
	if err != nil {
		return err
	}

	// Parse
	parser := newParser(tokens)
	parsedValue, err := parser.parse()
	if err != nil {
		return err
	}

	// Decode into target
	d := &decoder{
		classes: parser.classes,
	}

	return d.decode(parsedValue, rv.Elem())
}

// decode assigns a parsed value to a reflect.Value.
func (d *decoder) decode(src interface{}, dst reflect.Value) error {
	// Handle nil
	if src == nil {
		return d.decodeNull(dst)
	}

	// Handle custom unmarshalers
	if dst.CanAddr() {
		addr := dst.Addr()
		if addr.Type().Implements(unmarshalerType) {
			// For custom unmarshalers, we would need to re-marshal the value
			// For now, we'll just let it fall through to standard decoding
		}

		if addr.Type().Implements(textUnmarshalerType) {
			if str, ok := src.(string); ok {
				return addr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(str))
			}
		}
	}

	// Type-based decoding
	switch srcVal := src.(type) {
	case bool:
		return d.decodeBool(srcVal, dst)
	case float64:
		return d.decodeNumber(srcVal, dst)
	case string:
		return d.decodeString(srcVal, dst)
	case []interface{}:
		return d.decodeArray(srcVal, dst)
	case map[string]interface{}:
		return d.decodeObject(srcVal, dst)
	default:
		return fmt.Errorf("unknown parsed type: %T", src)
	}
}

// decodeNull handles null values.
func (d *decoder) decodeNull(dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	default:
		// Null into other types is a no-op (JSON compatibility)
		return nil
	}
}

// decodeBool decodes a boolean value.
func (d *decoder) decodeBool(src bool, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Bool:
		dst.SetBool(src)
		return nil
	case reflect.Interface:
		if dst.NumMethod() == 0 {
			dst.Set(reflect.ValueOf(src))
			return nil
		}
	}
	return &UnmarshalTypeError{
		Value: "bool",
		Type:  dst.Type(),
	}
}

// decodeNumber decodes a numeric value.
func (d *decoder) decodeNumber(src float64, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		min := minInt(dst.Type())
		max := maxInt(dst.Type())
		if src < float64(min) || src > float64(max) {
			return &UnmarshalTypeError{Value: fmt.Sprintf("number %v", src), Type: dst.Type()}
		}
		dst.SetInt(int64(src))
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if src < 0 || src > float64(maxUint(dst.Type())) {
			return &UnmarshalTypeError{Value: fmt.Sprintf("number %v", src), Type: dst.Type()}
		}
		dst.SetUint(uint64(src))
		return nil

	case reflect.Float32, reflect.Float64:
		dst.SetFloat(src)
		return nil

	case reflect.Interface:
		if dst.NumMethod() == 0 {
			dst.Set(reflect.ValueOf(src))
			return nil
		}
	}
	return &UnmarshalTypeError{Value: "number", Type: dst.Type()}
}

// decodeString decodes a string value.
func (d *decoder) decodeString(src string, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(src)
		return nil
	case reflect.Interface:
		if dst.NumMethod() == 0 {
			dst.Set(reflect.ValueOf(src))
			return nil
		}
	case reflect.Slice:
		if dst.Type().Elem().Kind() == reflect.Uint8 {
			// []byte - store string as bytes
			dst.SetBytes([]byte(src))
			return nil
		}
	}
	return &UnmarshalTypeError{Value: "string", Type: dst.Type()}
}

// decodeArray decodes an array value.
func (d *decoder) decodeArray(src []interface{}, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Slice:
		return d.decodeSlice(src, dst)
	case reflect.Array:
		return d.decodeArrayFixed(src, dst)
	case reflect.Interface:
		if dst.NumMethod() == 0 {
			// Create []interface{} with decoded elements
			result := make([]interface{}, len(src))
			copy(result, src)
			dst.Set(reflect.ValueOf(result))
			return nil
		}
	}
	return &UnmarshalTypeError{Value: "array", Type: dst.Type()}
}

// decodeSlice decodes into a slice.
func (d *decoder) decodeSlice(src []interface{}, dst reflect.Value) error {
	// Create new slice
	slice := reflect.MakeSlice(dst.Type(), len(src), len(src))

	// Decode each element
	for i, item := range src {
		if err := d.decode(item, slice.Index(i)); err != nil {
			return err
		}
	}

	dst.Set(slice)
	return nil
}

// decodeArrayFixed decodes into a fixed-size array.
func (d *decoder) decodeArrayFixed(src []interface{}, dst reflect.Value) error {
	length := dst.Len()

	// Decode elements up to array length
	for i := 0; i < length && i < len(src); i++ {
		if err := d.decode(src[i], dst.Index(i)); err != nil {
			return err
		}
	}

	// Zero out remaining elements if src is shorter
	for i := len(src); i < length; i++ {
		dst.Index(i).Set(reflect.Zero(dst.Type().Elem()))
	}

	return nil
}

// decodeObject decodes an object (map or struct).
func (d *decoder) decodeObject(src map[string]interface{}, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.Map:
		return d.decodeMap(src, dst)
	case reflect.Struct:
		return d.decodeStruct(src, dst)
	case reflect.Interface:
		if dst.NumMethod() == 0 {
			// Create map[string]interface{} with decoded values
			result := make(map[string]interface{})
			for k, v := range src {
				result[k] = v
			}
			dst.Set(reflect.ValueOf(result))
			return nil
		}
	}
	return &UnmarshalTypeError{Value: "object", Type: dst.Type()}
}

// decodeMap decodes into a map.
func (d *decoder) decodeMap(src map[string]interface{}, dst reflect.Value) error {
	keyType := dst.Type().Key()
	elemType := dst.Type().Elem()

	// Create map if nil
	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}

	for k, v := range src {
		// Convert key
		keyVal := reflect.New(keyType).Elem()
		if err := d.decodeMapKey(k, keyVal); err != nil {
			return err
		}

		// Convert value
		elemVal := reflect.New(elemType).Elem()
		if err := d.decode(v, elemVal); err != nil {
			return err
		}

		dst.SetMapIndex(keyVal, elemVal)
	}

	return nil
}

// decodeMapKey decodes a string key into the appropriate map key type.
func (d *decoder) decodeMapKey(src string, dst reflect.Value) error {
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(src)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(src, 10, 64)
		if err != nil {
			return err
		}
		dst.SetInt(i)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(src, 10, 64)
		if err != nil {
			return err
		}
		dst.SetUint(u)
		return nil
	}

	// Try TextUnmarshaler
	if dst.CanAddr() && dst.Addr().Type().Implements(textUnmarshalerType) {
		return dst.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(src))
	}

	return &UnmarshalTypeError{Value: "string (as map key)", Type: dst.Type()}
}

// structField holds information about a struct field.
type structField struct {
	index int
	name  string
	typ   reflect.Type
}

// decodeStruct decodes into a struct.
func (d *decoder) decodeStruct(src map[string]interface{}, dst reflect.Value) error {
	t := dst.Type()

	// Build field map (json tag name -> field info)
	fields := make(map[string]structField)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				name = parts[0]
			}
		}

		sf := structField{
			index: i,
			name:  field.Name,
			typ:   field.Type,
		}

		fields[name] = sf
		// Also support case-insensitive matching
		fields[strings.ToLower(name)] = sf
	}

	// Decode each source field
	for key, value := range src {
		// Try exact match first
		field, ok := fields[key]
		if !ok {
			// Try case-insensitive
			field, ok = fields[strings.ToLower(key)]
		}

		if !ok {
			// Unknown field - ignore (JSON behavior)
			continue
		}

		fieldVal := dst.Field(field.index)
		if err := d.decode(value, fieldVal); err != nil {
			return &UnmarshalTypeError{
				Value:  fmt.Sprintf("%T", value),
				Type:   field.typ,
				Struct: t.Name(),
				Field:  field.name,
			}
		}
	}

	return nil
}

// Helper variables for interface types.
var (
	unmarshalerType     = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// minInt returns the minimum value for an integer type.
func minInt(t reflect.Type) int64 {
	switch t.Kind() {
	case reflect.Int8:
		return math.MinInt8
	case reflect.Int16:
		return math.MinInt16
	case reflect.Int32:
		return math.MinInt32
	case reflect.Int64:
		return math.MinInt64
	case reflect.Int:
		// Use Int64 limits for generic int (conservative)
		return math.MinInt64
	default:
		return math.MinInt64
	}
}

// maxInt returns the maximum value for an integer type.
func maxInt(t reflect.Type) int64 {
	switch t.Kind() {
	case reflect.Int8:
		return math.MaxInt8
	case reflect.Int16:
		return math.MaxInt16
	case reflect.Int32:
		return math.MaxInt32
	case reflect.Int64:
		return math.MaxInt64
	case reflect.Int:
		// Use Int64 limits for generic int (conservative)
		return math.MaxInt64
	default:
		return math.MaxInt64
	}
}

// maxUint returns the maximum value for an unsigned integer type.
func maxUint(t reflect.Type) uint64 {
	switch t.Kind() {
	case reflect.Uint8:
		return math.MaxUint8
	case reflect.Uint16:
		return math.MaxUint16
	case reflect.Uint32:
		return math.MaxUint32
	case reflect.Uint64:
		return math.MaxUint64
	case reflect.Uint:
		// Use Uint64 limits for generic uint (conservative)
		return math.MaxUint64
	default:
		return math.MaxUint64
	}
}
