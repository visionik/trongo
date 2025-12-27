package tron

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalPrimitives(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		target  interface{}
		check   func(t *testing.T, target interface{})
		wantErr bool
	}{
		{
			name:   "bool true",
			input:  "true",
			target: new(bool),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, true, *target.(*bool))
			},
		},
		{
			name:   "bool false",
			input:  "false",
			target: new(bool),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, false, *target.(*bool))
			},
		},
		{
			name:   "int",
			input:  "42",
			target: new(int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, 42, *target.(*int))
			},
		},
		{
			name:   "negative int",
			input:  "-17",
			target: new(int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, -17, *target.(*int))
			},
		},
		{
			name:   "float64",
			input:  "3.14",
			target: new(float64),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, 3.14, *target.(*float64))
			},
		},
		{
			name:   "string",
			input:  `"hello"`,
			target: new(string),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, "hello", *target.(*string))
			},
		},
		{
			name:   "null into pointer",
			input:  "null",
			target: new(*string),
			check: func(t *testing.T, target interface{}) {
				assert.Nil(t, *target.(**string))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.check(t, tt.target)
			}
		})
	}
}

func TestUnmarshalArrays(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		target  interface{}
		check   func(t *testing.T, target interface{})
		wantErr bool
	}{
		{
			name:   "int slice",
			input:  "[1,2,3]",
			target: new([]int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, []int{1, 2, 3}, *target.(*[]int))
			},
		},
		{
			name:   "string slice",
			input:  `["a","b","c"]`,
			target: new([]string),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, []string{"a", "b", "c"}, *target.(*[]string))
			},
		},
		{
			name:   "empty slice",
			input:  "[]",
			target: new([]int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, []int{}, *target.(*[]int))
			},
		},
		{
			name:   "fixed array",
			input:  "[1,2,3]",
			target: new([3]int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, [3]int{1, 2, 3}, *target.(*[3]int))
			},
		},
		{
			name:   "interface slice",
			input:  `[1,"two",true]`,
			target: new([]interface{}),
			check: func(t *testing.T, target interface{}) {
				expected := []interface{}{float64(1), "two", true}
				assert.Equal(t, expected, *target.(*[]interface{}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.check(t, tt.target)
			}
		})
	}
}

func TestUnmarshalMaps(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		target  interface{}
		check   func(t *testing.T, target interface{})
		wantErr bool
	}{
		{
			name:   "string map",
			input:  `{"name":"Alice","city":"NYC"}`,
			target: new(map[string]string),
			check: func(t *testing.T, target interface{}) {
				expected := map[string]string{"name": "Alice", "city": "NYC"}
				assert.Equal(t, expected, *target.(*map[string]string))
			},
		},
		{
			name:   "interface map",
			input:  `{"name":"Alice","age":30,"active":true}`,
			target: new(map[string]interface{}),
			check: func(t *testing.T, target interface{}) {
				expected := map[string]interface{}{"name": "Alice", "age": float64(30), "active": true}
				assert.Equal(t, expected, *target.(*map[string]interface{}))
			},
		},
		{
			name:   "empty map",
			input:  `{}`,
			target: new(map[string]string),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, map[string]string{}, *target.(*map[string]string))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.check(t, tt.target)
			}
		})
	}
}

func TestUnmarshalStructs(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	type Tagged struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type WithOmit struct {
		Required string `json:"required"`
		Optional string `json:"optional,omitempty"`
		Ignored  string `json:"-"`
	}

	tests := []struct {
		name    string
		input   string
		target  interface{}
		check   func(t *testing.T, target interface{})
		wantErr bool
	}{
		{
			name:   "simple struct from object",
			input:  `{"Name":"Alice","Age":30}`,
			target: new(Person),
			check: func(t *testing.T, target interface{}) {
				expected := Person{Name: "Alice", Age: 30}
				assert.Equal(t, expected, *target.(*Person))
			},
		},
		{
			name: "struct from class instantiation",
			input: `class A: Name,Age

A("Bob",25)`,
			target: new(Person),
			check: func(t *testing.T, target interface{}) {
				expected := Person{Name: "Bob", Age: 25}
				assert.Equal(t, expected, *target.(*Person))
			},
		},
		{
			name:   "struct with json tags",
			input:  `{"name":"Alice","value":42}`,
			target: new(Tagged),
			check: func(t *testing.T, target interface{}) {
				expected := Tagged{Name: "Alice", Value: 42}
				assert.Equal(t, expected, *target.(*Tagged))
			},
		},
		{
			name:   "struct with ignored field",
			input:  `{"required":"yes","optional":"maybe","Ignored":"never"}`,
			target: new(WithOmit),
			check: func(t *testing.T, target interface{}) {
				expected := WithOmit{Required: "yes", Optional: "maybe", Ignored: ""}
				assert.Equal(t, expected, *target.(*WithOmit))
			},
		},
		{
			name:   "struct with unknown fields",
			input:  `{"Name":"Alice","Age":30,"Unknown":"ignored"}`,
			target: new(Person),
			check: func(t *testing.T, target interface{}) {
				expected := Person{Name: "Alice", Age: 30}
				assert.Equal(t, expected, *target.(*Person))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.check(t, tt.target)
			}
		})
	}
}

func TestUnmarshalComplexStructures(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type Team struct {
		Name    string   `json:"name"`
		Members []Person `json:"members"`
	}

	tests := []struct {
		name    string
		input   string
		target  interface{}
		check   func(t *testing.T, target interface{})
		wantErr bool
	}{
		{
			name: "array of structs with class",
			input: `class A: name,age

[A("Alice",30),A("Bob",25)]`,
			target: new([]Person),
			check: func(t *testing.T, target interface{}) {
				expected := []Person{
					{Name: "Alice", Age: 30},
					{Name: "Bob", Age: 25},
				}
				assert.Equal(t, expected, *target.(*[]Person))
			},
		},
		{
			name: "nested structs",
			input: `class Person: name,age
class Team: name,members

Team("Engineering",[Person("Alice",30),Person("Bob",25)])`,
			target: new(Team),
			check: func(t *testing.T, target interface{}) {
				expected := Team{
					Name: "Engineering",
					Members: []Person{
						{Name: "Alice", Age: 30},
						{Name: "Bob", Age: 25},
					},
				}
				assert.Equal(t, expected, *target.(*Team))
			},
		},
		{
			name:   "array of maps",
			input:  `[{"name":"Alice"},{"name":"Bob"}]`,
			target: new([]map[string]string),
			check: func(t *testing.T, target interface{}) {
				expected := []map[string]string{
					{"name": "Alice"},
					{"name": "Bob"},
				}
				assert.Equal(t, expected, *target.(*[]map[string]string))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.check(t, tt.target)
			}
		})
	}
}

func TestUnmarshalErrors(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		target    interface{}
		wantError string
	}{
		{
			name:      "nil target",
			input:     "42",
			target:    nil,
			wantError: "Unmarshal(nil)",
		},
		{
			name:      "non-pointer target",
			input:     "42",
			target:    42,
			wantError: "non-pointer",
		},
		{
			name:      "syntax error",
			input:     "[1,2,",
			target:    new([]int),
			wantError: "expected",
		},
		{
			name:      "type mismatch - string into int",
			input:     `"hello"`,
			target:    new(int),
			wantError: "cannot unmarshal",
		},
		{
			name:      "type mismatch - array into string",
			input:     `[1,2,3]`,
			target:    new(string),
			wantError: "cannot unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)
			assert.Error(t, err)
			if tt.wantError != "" {
				assert.Contains(t, err.Error(), tt.wantError)
			}
		})
	}
}

func TestUnmarshalInterface(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  interface{}
	}{
		{
			name:  "bool into interface",
			input: "true",
			want:  true,
		},
		{
			name:  "number into interface",
			input: "42.5",
			want:  float64(42.5),
		},
		{
			name:  "string into interface",
			input: `"hello"`,
			want:  "hello",
		},
		{
			name:  "array into interface",
			input: "[1,2,3]",
			want:  []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:  "object into interface",
			input: `{"name":"Alice"}`,
			want:  map[string]interface{}{"name": "Alice"},
		},
		{
			name:  "null into interface",
			input: "null",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal([]byte(tt.input), &result)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestUnmarshalREADMEExample(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	input := `class A: name,age

[A("Alice",30),A("Bob",25)]`

	var result []Person
	err := Unmarshal([]byte(input), &result)
	require.NoError(t, err)

	expected := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}

	assert.Equal(t, expected, result)
}

func TestRoundTrip(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	type Team struct {
		Name    string   `json:"name"`
		Members []Person `json:"members"`
	}

	tests := []struct {
		name string
		data interface{}
	}{
		{
			name: "simple struct",
			data: Person{Name: "Alice", Age: 30},
		},
		{
			name: "array of structs",
			data: []Person{
				{Name: "Alice", Age: 30},
				{Name: "Bob", Age: 25},
				{Name: "Charlie", Age: 35},
			},
		},
		{
			name: "nested structs",
			data: Team{
				Name: "Engineering",
				Members: []Person{
					{Name: "Alice", Age: 30},
					{Name: "Bob", Age: 25},
				},
			},
		},
		{
			name: "primitives",
			data: 42,
		},
		{
			name: "string",
			data: "hello world",
		},
		{
			name: "bool",
			data: true,
		},
		{
			name: "array of ints",
			data: []int{1, 2, 3, 4, 5},
		},
		{
			name: "map",
			data: map[string]interface{}{
				"name":   "Alice",
				"age":    float64(30),
				"active": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			encoded, err := Marshal(tt.data)
			require.NoError(t, err)

			// Unmarshal
			result := newSameType(tt.data)
			err = Unmarshal(encoded, result)
			require.NoError(t, err)

			// Compare
			resultVal := getDeref(result)
			assert.Equal(t, tt.data, resultVal)
		})
	}
}

// Helper function to create a new pointer of the same type
func newSameType(v interface{}) interface{} {
	t := reflect.TypeOf(v)
	ptr := reflect.New(t)
	return ptr.Interface()
}

// Helper function to dereference a pointer
func getDeref(v interface{}) interface{} {
	return reflect.ValueOf(v).Elem().Interface()
}
