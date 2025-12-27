package tron

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// ClassDef represents a class definition with name and property keys.
type ClassDef struct {
	Name string
	Keys []string
}

// marshal is the internal implementation of Marshal and MarshalIndent.
func marshal(v interface{}, prefix, indent string) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}

	// Create encoder state
	e := &encoder{
		classes:       make([]ClassDef, 0),
		schemaToClass: make(map[string]ClassDef),
		schemaCounts:  make(map[string]int),
		visited:       make(map[uintptr]bool),
		prefix:        prefix,
		indent:        indent,
	}

	// Phase 1: Discover classes through DFS
	if err := e.discoverClasses(reflect.ValueOf(v), 0); err != nil {
		return nil, err
	}

	// Phase 2: Filter classes based on property count and occurrence
	e.filterClasses()

	// Phase 3: Generate output
	var output strings.Builder

	// Generate header (class definitions)
	for _, cls := range e.filteredClasses {
		output.WriteString("class ")
		output.WriteString(cls.Name)
		output.WriteString(": ")

		for i, key := range cls.Keys {
			if i > 0 {
				output.WriteString(",")
			}
			if isValidIdentifier(key) {
				output.WriteString(key)
			} else {
				// Quote keys with special characters
				quoted, _ := json.Marshal(key)
				output.Write(quoted)
			}
		}
		output.WriteString("\n")
	}

	if len(e.filteredClasses) > 0 {
		output.WriteString("\n")
	}

	// Generate data
	data, err := e.serialize(reflect.ValueOf(v), make(map[uintptr]bool), 0)
	if err != nil {
		return nil, err
	}
	output.WriteString(data)

	return []byte(output.String()), nil
}

// encoder holds the state for marshaling.
type encoder struct {
	classes           []ClassDef
	schemaToClass     map[string]ClassDef
	schemaCounts      map[string]int
	filteredClasses   []ClassDef
	filteredSchemaMap map[string]ClassDef
	visited           map[uintptr]bool
	prefix            string
	indent            string
	classCounter      int

	structCache sync.Map // map[reflect.Type]*structTypeInfo
}

// discoverClasses performs DFS to discover all object schemas.
func (e *encoder) discoverClasses(v reflect.Value, depth int) error {
	if depth > maxWalkDepth {
		return fmt.Errorf("maximum walk depth exceeded")
	}
	if !v.IsValid() {
		return nil
	}

	// Handle pointers and interfaces
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Check for cycles
	if v.CanAddr() {
		addr := v.UnsafeAddr()
		if e.visited[addr] {
			return nil
		}
		e.visited[addr] = true
		defer func() { delete(e.visited, addr) }()
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := e.discoverClasses(v.Index(i), depth+1); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range v.MapKeys() {
			if err := e.discoverClasses(v.MapIndex(key), depth+1); err != nil {
				return err
			}
		}

	case reflect.Struct:
		// Get field information
		keys, err := e.getStructKeys(v)
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			// Create schema signature (sorted keys for consistency)
			sortedKeys := make([]string, len(keys))
			copy(sortedKeys, keys)
			sort.Strings(sortedKeys)
			schemaSignature := strings.Join(sortedKeys, ",")

			// Track occurrence count
			e.schemaCounts[schemaSignature]++

			if _, exists := e.schemaToClass[schemaSignature]; !exists {
				className := generateClassName(e.classCounter)
				e.classCounter++
				classDef := ClassDef{Name: className, Keys: keys}
				e.classes = append(e.classes, classDef)
				e.schemaToClass[schemaSignature] = classDef
			}

			// Recursively visit struct fields
			for _, key := range keys {
				fieldValue := e.getStructFieldValue(v, key)
				if err := e.discoverClasses(fieldValue, depth+1); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// filterClasses filters classes based on property count and occurrence.
func (e *encoder) filterClasses() {
	e.filteredClasses = make([]ClassDef, 0)
	e.filteredSchemaMap = make(map[string]ClassDef)
	filteredClassCounter := 0

	for schemaSignature, classDef := range e.schemaToClass {
		propertyCount := len(classDef.Keys)
		occurrenceCount := e.schemaCounts[schemaSignature]

		// Define class if: 2+ properties AND 2+ occurrences
		shouldDefineClass := propertyCount > 1 && occurrenceCount > 1
		if shouldDefineClass {
			newClassName := generateClassName(filteredClassCounter)
			filteredClassCounter++
			newClassDef := ClassDef{Name: newClassName, Keys: classDef.Keys}
			e.filteredClasses = append(e.filteredClasses, newClassDef)
			e.filteredSchemaMap[schemaSignature] = newClassDef
		}
	}
}

// serialize converts a Go value to TRON format string.
func (e *encoder) serialize(v reflect.Value, stack map[uintptr]bool, depth int) (string, error) {
	if depth > maxWalkDepth {
		return "", fmt.Errorf("maximum walk depth exceeded")
	}
	if !v.IsValid() {
		return "null", nil
	}

	// Check for cycles in pointers BEFORE dereferencing
	// Note: Only pointers can create cycles in Go value structures
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "null", nil
		}
		if v.CanAddr() {
			addr := v.UnsafeAddr()
			if stack[addr] {
				return "", fmt.Errorf("converting circular structure to TRON")
			}
			stack[addr] = true
			defer func() { delete(stack, addr) }()
		}
		v = v.Elem()
	}

	// Handle interfaces
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return "null", nil
		}
		v = v.Elem()
	}

	// Check for custom marshaler
	if v.Type().Implements(reflect.TypeOf((*Marshaler)(nil)).Elem()) {
		marshaler := v.Interface().(Marshaler)
		data, err := marshaler.MarshalTRON()
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// Check for text marshaler
	if v.Type().Implements(reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()) {
		marshaler := v.Interface().(encoding.TextMarshaler)
		text, err := marshaler.MarshalText()
		if err != nil {
			return "", err
		}
		quoted, _ := json.Marshal(string(text))
		return string(quoted), nil
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			return "true", nil
		}
		return "false", nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil

	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, v.Type().Bits()), nil

	case reflect.String:
		quoted, _ := json.Marshal(v.String())
		return string(quoted), nil

	case reflect.Array, reflect.Slice:
		// Check for nil slice
		if v.Kind() == reflect.Slice && v.IsNil() {
			return "null", nil
		}

		if v.Type().Elem().Kind() == reflect.Uint8 {
			// Handle []byte as base64 string
			bytes := v.Bytes()
			quoted, _ := json.Marshal(string(bytes))
			return string(quoted), nil
		}

		var items []string
		for i := 0; i < v.Len(); i++ {
			item, err := e.serialize(v.Index(i), stack, depth+1)
			if err != nil {
				return "", err
			}
			items = append(items, item)
		}
		return "[" + strings.Join(items, ",") + "]", nil

	case reflect.Map:
		// Check for nil map
		if v.IsNil() {
			return "null", nil
		}
		if v.Len() == 0 {
			return "{}", nil
		}

		// Convert map to object notation
		var pairs []string
		keys := v.MapKeys()

		// Sort keys for consistent output
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
		})

		for _, key := range keys {
			keyStr, err := e.serializeMapKey(key)
			if err != nil {
				return "", err
			}
			value, err := e.serialize(v.MapIndex(key), stack, depth+1)
			if err != nil {
				return "", err
			}
			pairs = append(pairs, keyStr+":"+value)
		}
		return "{" + strings.Join(pairs, ",") + "}", nil

	case reflect.Struct:
		keys, err := e.getStructKeys(v)
		if err != nil {
			return "", err
		}

		if len(keys) == 0 {
			return "{}", nil
		}

		// Check if we should use class instantiation
		sortedKeys := make([]string, len(keys))
		copy(sortedKeys, keys)
		sort.Strings(sortedKeys)
		schemaSignature := strings.Join(sortedKeys, ",")

		if classDef, exists := e.filteredSchemaMap[schemaSignature]; exists {
			// Use class instantiation
			var args []string
			for _, key := range classDef.Keys {
				fieldValue := e.getStructFieldValue(v, key)
				arg, err := e.serialize(fieldValue, stack, depth+1)
				if err != nil {
					return "", err
				}
				args = append(args, arg)
			}
			return classDef.Name + "(" + strings.Join(args, ",") + ")", nil
		} else {
			// Use JSON object syntax
			var pairs []string
			for _, key := range keys {
				fieldValue := e.getStructFieldValue(v, key)
				value, err := e.serialize(fieldValue, stack, depth+1)
				if err != nil {
					return "", err
				}
				keyStr, _ := json.Marshal(key)
				pairs = append(pairs, string(keyStr)+":"+value)
			}
			return "{" + strings.Join(pairs, ",") + "}", nil
		}

	default:
		return "", &UnsupportedTypeError{Type: v.Type()}
	}
}

type structTypeInfo struct {
	fields []structFieldInfo
	byName map[string]int // json name -> field index
}

type structFieldInfo struct {
	name      string
	index     int
	omitempty bool
}

// getStructKeys returns the field names for a struct, respecting json tags.
func (e *encoder) getStructKeys(v reflect.Value) ([]string, error) {
	ti := e.getStructTypeInfo(v.Type())
	keys := make([]string, 0, len(ti.fields))
	for _, f := range ti.fields {
		fv := v.Field(f.index)
		if f.omitempty && isEmptyValue(fv) {
			continue
		}
		keys = append(keys, f.name)
	}
	return keys, nil
}

func (e *encoder) getStructTypeInfo(t reflect.Type) *structTypeInfo {
	if v, ok := e.structCache.Load(t); ok {
		return v.(*structTypeInfo)
	}

	info := &structTypeInfo{
		fields: make([]structFieldInfo, 0, t.NumField()),
		byName: make(map[string]int),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := field.Name
		omitempty := false
		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] == "-" {
				continue
			}
			if parts[0] != "" {
				name = parts[0]
			}
			if len(parts) > 1 && contains(parts[1:], "omitempty") {
				omitempty = true
			}
		}

		info.fields = append(info.fields, structFieldInfo{name: name, index: i, omitempty: omitempty})
		// First field wins for name collisions (matches encoding/json behavior).
		if _, exists := info.byName[name]; !exists {
			info.byName[name] = i
		}
	}

	// Publish
	e.structCache.Store(t, info)
	return info
}

// getStructFieldValue returns the value of a struct field by name, respecting json tags.
func (e *encoder) getStructFieldValue(v reflect.Value, name string) reflect.Value {
	ti := e.getStructTypeInfo(v.Type())
	idx, ok := ti.byName[name]
	if !ok {
		return reflect.Value{}
	}
	return v.Field(idx)
}

// serializeMapKey converts a map key to a string for TRON object notation.
func (e *encoder) serializeMapKey(key reflect.Value) (string, error) {
	switch key.Kind() {
	case reflect.String:
		quoted, _ := json.Marshal(key.String())
		return string(quoted), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		quoted, _ := json.Marshal(strconv.FormatInt(key.Int(), 10))
		return string(quoted), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		quoted, _ := json.Marshal(strconv.FormatUint(key.Uint(), 10))
		return string(quoted), nil
	default:
		if key.Type().Implements(reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()) {
			marshaler := key.Interface().(encoding.TextMarshaler)
			text, err := marshaler.MarshalText()
			if err != nil {
				return "", err
			}
			quoted, _ := json.Marshal(string(text))
			return string(quoted), nil
		}
		return "", &UnsupportedTypeError{Type: key.Type()}
	}
}

// generateClassName generates a class name from an index (A, B, ..., Z, A1, B1, ...).
func generateClassName(index int) string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	cycle := index / 26
	position := index % 26

	if cycle == 0 {
		return string(letters[position])
	}
	return string(letters[position]) + strconv.Itoa(cycle)
}

// isValidIdentifier checks if a string is a valid identifier (no need to quote).
// Must match the tokenizer's identifier rules.
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !(unicode.IsLetter(r) || r == '_') {
				return false
			}
			continue
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || r == '_') {
			return false
		}
	}
	return true
}

// isEmptyValue checks if a value is considered empty for omitempty.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// contains checks if a slice contains a string.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
