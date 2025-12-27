// Package tron provides encoding and decoding of TRON (Token Reduced Object Notation) format.
//
// This package provides an API compatible with the standard encoding/json package,
// allowing for easy migration from JSON to the more token-efficient TRON format.
//
// The TRON format reduces redundancy by defining reusable class structures for
// objects with the same schema, making it particularly efficient for arrays of
// similar objects.
//
// Basic usage:
//
//	data, err := tron.Marshal(v)
//	if err != nil {
//		// handle error
//	}
//
//	var v interface{}
//	err = tron.Unmarshal(data, &v)
//	if err != nil {
//		// handle error
//	}
package tron

import (
	"reflect"
)

// Marshal returns the TRON encoding of v.
//
// Marshal traverses the value v recursively. If an encountered value implements
// the Marshaler interface and is not a nil pointer, Marshal calls its MarshalTRON
// method to produce TRON. If no MarshalTRON method is present but the value
// implements encoding.TextMarshaler, Marshal calls its MarshalText method and
// encodes the result as a TRON string.
//
// Otherwise, Marshal uses the following type-dependent default encodings:
//
// Boolean values encode as TRON booleans.
//
// Floating point, integer, and Number values encode as TRON numbers.
//
// String values encode as TRON strings coerced to valid UTF-8,
// replacing invalid bytes with the Unicode replacement rune.
//
// Array and slice values encode as TRON arrays, except that
// []byte encodes as a base64-encoded string, and a nil slice
// encodes as the null TRON value.
//
// Struct values encode as TRON objects. Each exported struct field
// becomes a member of the object, using the field name as the object
// key, unless the field is omitted for one of the reasons given below.
//
// The encoding of each struct field can be customized by the format string
// stored under the "json" key in the struct field's tag. The format string
// gives the name of the field, possibly followed by a comma-separated
// list of options. The name may be empty in order to specify options
// without overriding the default field name.
//
// The "omitempty" option specifies that the field should be omitted
// from the encoding if the field has an empty value, defined as
// false, 0, a nil pointer, a nil interface value, and any empty array,
// slice, map, or string.
//
// As a special case, if the field tag is "-", the field is always omitted.
// Note that a field with name "-" can still be generated using the tag "-,".
//
// Map values encode as TRON objects. The map's key type must either be a
// string, an integer type, or implement encoding.TextMarshaler. The map keys
// are sorted and used as TRON object keys by applying the following rules,
// subject to the UTF-8 coercion described for string values above:
//   - keys of string type are used directly
//   - encoding.TextMarshalers are marshaled
//   - integer keys are converted to strings
//
// Pointer values encode as the value pointed to.
// A nil pointer encodes as the null TRON value.
//
// Interface values encode as the value contained in the interface.
// A nil interface value encodes as the null TRON value.
//
// Channel, complex, and function values cannot be encoded in TRON.
// Attempting to encode such a value causes Marshal to return
// an UnsupportedTypeError.
//
// TRON cannot represent cyclic data structures and Marshal does not
// handle them. Passing cyclic structures to Marshal will result in
// an error.
func Marshal(v interface{}) ([]byte, error) {
	return marshal(v, "", "")
}

// MarshalIndent is like Marshal but applies Indent to format the output.
// Each TRON element in the output will begin on a new line beginning with prefix
// followed by one or more copies of indent according to the indentation nesting.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return marshal(v, prefix, indent)
}

// Unmarshal parses the TRON-encoded data and stores the result
// in the value pointed to by v. If v is nil or not a pointer,
// Unmarshal returns an InvalidUnmarshalError.
//
// Unmarshal uses the inverse of the encodings that
// Marshal uses, allocating maps, slices, and pointers as necessary,
// with the following additional rules:
//
// To unmarshal TRON into a pointer, Unmarshal first handles the case of
// the TRON being the TRON literal null. In that case, Unmarshal sets
// the pointer to nil. Otherwise, Unmarshal unmarshals the TRON into
// the value pointed at by the pointer. If the pointer is nil, Unmarshal
// allocates a new value for it to point to.
//
// To unmarshal TRON into a value implementing the Unmarshaler interface,
// Unmarshal calls that value's UnmarshalTRON method, including
// when the input is a TRON null.
//
// To unmarshal TRON into a struct, Unmarshal matches incoming object
// keys to the keys used by Marshal (either the struct field name or its tag),
// preferring an exact match but also accepting a case-insensitive match. By
// default, object keys which don't have a corresponding struct field are
// ignored (see Decoder.DisallowUnknownFields for an alternative).
//
// To unmarshal TRON into an interface{} value,
// Unmarshal stores one of these in the interface{} value:
//
//	bool, for TRON booleans
//	float64, for TRON numbers
//	string, for TRON strings
//	[]interface{}, for TRON arrays
//	map[string]interface{}, for TRON objects
//	nil for TRON null
//
// To unmarshal a TRON array into a slice, Unmarshal resets the slice length
// to zero and then appends each element to the slice.
// As a special case, to unmarshal an empty TRON array into a slice,
// Unmarshal replaces the slice with a new empty slice.
//
// To unmarshal a TRON array into a Go array, Unmarshal decodes
// TRON array elements into corresponding Go array elements.
// If the Go array is smaller than the TRON array,
// the additional TRON array elements are discarded.
// If the TRON array is smaller than the Go array,
// the additional Go array elements are set to zero values.
//
// To unmarshal a TRON object into a map, Unmarshal first establishes a map to
// use. If the map is nil, Unmarshal allocates a new map. Otherwise Unmarshal
// reuses the existing map, keeping existing entries. Unmarshal then stores
// key-value pairs from the TRON object into the map. The map's key type must
// either be any string type, any integer type, any unsigned integer type, or
// an implementation of encoding.TextUnmarshaler.
//
// If the TRON-encoded data contain a syntax error, Unmarshal returns a SyntaxError.
//
// If a TRON value is not appropriate for a given target type,
// or if a TRON number overflows the target type, Unmarshal
// skips that field and completes the unmarshaling as best it can.
// If no more serious errors are encountered, Unmarshal returns
// an UnmarshalTypeError describing the earliest such error. In any
// case, it's not guaranteed that all the remaining fields following
// the problematic one will be unmarshaled into the target object.
//
// The TRON null value unmarshals into an interface{}, map, pointer, or slice
// by setting that Go value to nil. Because null is often used in TRON to mean
// "not present," unmarshaling a TRON null into any other Go type has no effect
// on the value and produces no error.
//
// When unmarshaling quoted strings, invalid UTF-8 or
// invalid UTF-16 surrogate pairs are not treated as an error.
// Instead, they are replaced by the Unicode replacement
// character U+FFFD.
func Unmarshal(data []byte, v interface{}) error {
	return unmarshal(data, v)
}

// Marshaler is the interface implemented by types that
// can marshal themselves into valid TRON.
type Marshaler interface {
	MarshalTRON() ([]byte, error)
}

// Unmarshaler is the interface implemented by types
// that can unmarshal a TRON description of themselves.
// The input can be assumed to be a valid encoding of
// a TRON value. UnmarshalTRON must copy the TRON data
// if it wishes to retain the data after returning.
//
// By convention, to approximate the behavior of Unmarshal itself,
// Unmarshalers implement UnmarshalTRON([]byte("null")) as a no-op.
type Unmarshaler interface {
	UnmarshalTRON([]byte) error
}

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "tron: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "tron: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "tron: Unmarshal(nil " + e.Type.String() + ")"
}

// A SyntaxError is a description of a TRON syntax error.
// Unmarshal will return a SyntaxError if the TRON can't be parsed.
type SyntaxError struct {
	msg    string // description of error
	Offset int64  // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }

// An UnmarshalTypeError describes a TRON value that was
// not appropriate for a value of a specific Go type.
type UnmarshalTypeError struct {
	Value  string       // description of TRON value - "bool", "array", "number -5"
	Type   reflect.Type // type of Go value it could not be assigned to
	Offset int64        // error occurred after reading Offset bytes
	Struct string       // name of the struct type containing the field
	Field  string       // the full path from root node to the field
}

func (e *UnmarshalTypeError) Error() string {
	if e.Struct != "" || e.Field != "" {
		return "tron: cannot unmarshal " + e.Value + " into Go struct field " + e.Struct + "." + e.Field + " of type " + e.Type.String()
	}
	return "tron: cannot unmarshal " + e.Value + " into Go value of type " + e.Type.String()
}

// An UnsupportedTypeError is returned by Marshal when attempting
// to encode an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "tron: unsupported type: " + e.Type.String()
}

// An UnsupportedValueError is returned by Marshal when attempting
// to encode an unsupported value.
type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "tron: unsupported value: " + e.Str
}
