package tron

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Additional tests to increase coverage

func TestMarshalIndent(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := Person{Name: "Alice", Age: 30}
	result, err := MarshalIndent(data, "", "  ")
	require.NoError(t, err)
	assert.Contains(t, string(result), "Alice")
	assert.Contains(t, string(result), "30")
}

func TestMarshalWithOmitempty(t *testing.T) {
	type OptionalFields struct {
		Required string `json:"required"`
		Optional string `json:"optional,omitempty"`
		Empty    int    `json:"empty,omitempty"`
	}

	tests := []struct {
		name  string
		data  OptionalFields
		check func(t *testing.T, result string)
	}{
		{
			name: "with all fields",
			data: OptionalFields{Required: "yes", Optional: "maybe", Empty: 42},
			check: func(t *testing.T, result string) {
				assert.Contains(t, result, "yes")
				assert.Contains(t, result, "maybe")
				assert.Contains(t, result, "42")
			},
		},
		{
			name: "with omitted fields",
			data: OptionalFields{Required: "yes", Optional: "", Empty: 0},
			check: func(t *testing.T, result string) {
				assert.Contains(t, result, "yes")
				assert.NotContains(t, result, "optional")
				assert.NotContains(t, result, "empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Marshal(tt.data)
			require.NoError(t, err)
			tt.check(t, string(result))
		})
	}
}

func TestUnmarshalNumberOverflow(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		target  interface{}
		wantErr bool
	}{
		{
			name:    "int8 overflow",
			input:   "200",
			target:  new(int8),
			wantErr: true,
		},
		{
			name:    "int8 underflow",
			input:   "-200",
			target:  new(int8),
			wantErr: true,
		},
		{
			name:    "uint overflow negative",
			input:   "-1",
			target:  new(uint),
			wantErr: true,
		},
		{
			name:   "uint8 valid",
			input:  "255",
			target: new(uint8),
		},
		{
			name:   "int16 valid",
			input:  "32767",
			target: new(int16),
		},
		{
			name:   "int32 valid",
			input:  "2147483647",
			target: new(int32),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnmarshalMapWithIntKeys(t *testing.T) {
	input := `{"1":"one","2":"two","3":"three"}`
	var result map[int]string
	err := Unmarshal([]byte(input), &result)
	require.NoError(t, err)

	expected := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}
	assert.Equal(t, expected, result)
}

func TestUnmarshalMapWithUintKeys(t *testing.T) {
	input := `{"1":"one","2":"two"}`
	var result map[uint]string
	err := Unmarshal([]byte(input), &result)
	require.NoError(t, err)

	expected := map[uint]string{
		1: "one",
		2: "two",
	}
	assert.Equal(t, expected, result)
}

func TestMarshalMapWithIntKeys(t *testing.T) {
	data := map[int]string{
		1: "one",
		2: "two",
	}
	result, err := Marshal(data)
	require.NoError(t, err)
	assert.Contains(t, string(result), "one")
	assert.Contains(t, string(result), "two")
}

func TestErrorMessages(t *testing.T) {
	// Test InvalidUnmarshalError
	err := &InvalidUnmarshalError{Type: nil}
	assert.Contains(t, err.Error(), "Unmarshal(nil)")

	// Test SyntaxError
	syntaxErr := &SyntaxError{msg: "test error", Offset: 10}
	assert.Equal(t, "test error", syntaxErr.Error())

	// Test UnmarshalTypeError - need to use reflect to create a valid Type
	var intVal int
	typeErr := &UnmarshalTypeError{
		Value:  "string",
		Type:   reflect.TypeOf(intVal),
		Struct: "Person",
		Field:  "Age",
	}
	assert.Contains(t, typeErr.Error(), "cannot unmarshal")
	assert.Contains(t, typeErr.Error(), "Person")
	assert.Contains(t, typeErr.Error(), "Age")

	// Test UnsupportedTypeError
	unsupportedErr := &UnsupportedTypeError{Type: reflect.TypeOf(intVal)}
	assert.Contains(t, unsupportedErr.Error(), "unsupported type")

	// Test UnsupportedValueError
	unsupportedValErr := &UnsupportedValueError{Str: "test value"}
	assert.Contains(t, unsupportedValErr.Error(), "unsupported value")
}

func TestUnmarshalStringWithEscapes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "unicode escape",
			input: `"\u0048\u0065\u006c\u006c\u006f"`,
			want:  "Hello",
		},
		{
			name:  "mixed escapes",
			input: `"Line1\nLine2\tTabbed"`,
			want:  "Line1\nLine2\tTabbed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			err := Unmarshal([]byte(tt.input), &result)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}
