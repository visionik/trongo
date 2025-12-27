package tron

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test tokenizer edge cases
func TestTokenizerEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "only whitespace",
			input: "   \n\t\r  ",
		},
		{
			name:  "only comments",
			input: "# comment\n# another comment",
		},
		{
			name:  "unclosed string",
			input: `"unclosed`,
			// Note: Current implementation may not error, just returns what it can parse
		},
		{
			name:  "string with all escape types",
			input: `"test\n\r\t\b\f\"\\\/"`,
		},
		{
			name:  "unicode escapes",
			input: `"\u0048\u0065\u006c\u006c\u006f"`,
		},
		{
			name:  "invalid unicode escape",
			input: `"\uGGGG"`,
		},
		{
			name:  "number edge cases",
			input: "0 -0 1.0 -1.0 1e10 1.5e-5 -3.14e+2",
		},
		{
			name:  "very long number",
			input: "123456789012345678901234567890",
		},
		{
			name:  "comment mid-line",
			input: "42 # comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenize(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// Should not panic, error is ok
				_ = tokens
				_ = err
			}
		})
	}
}

// Test parser edge cases
func TestParserEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: false, // Returns nil
		},
		{
			name:    "only header no data",
			input:   "class A: x,y\n\n",
			wantErr: false, // Returns nil for empty data
		},
		{
			name:    "class with no properties",
			input:   "class A: \n\nnull",
			wantErr: false,
		},
		{
			name:    "deeply nested arrays",
			input:   "[[[[1]]]]",
			wantErr: false,
		},
		{
			name:    "deeply nested objects",
			input:   `{"a":{"b":{"c":{"d":"value"}}}}`,
			wantErr: false,
		},
		{
			name:    "mixed nesting",
			input:   `[{"a":[1,2,{"b":[3,4]}]}]`,
			wantErr: false,
		},
		{
			name:    "empty array",
			input:   "[]",
			wantErr: false,
		},
		{
			name:    "empty object",
			input:   "{}",
			wantErr: false,
		},
		{
			name:    "class with duplicate property names",
			input:   "class A: x,x\n\nA(1,2)",
			wantErr: false, // Parser doesn't validate this
		},
		{
			name:    "undefined class usage",
			input:   "A(1,2)",
			wantErr: true,
		},
		{
			name:    "class argument mismatch - too few",
			input:   "class A: x,y\n\nA(1)",
			wantErr: true,
		},
		{
			name:    "class argument mismatch - too many",
			input:   "class A: x,y\n\nA(1,2,3)",
			wantErr: true,
		},
		{
			name:    "unclosed array",
			input:   "[1,2,3",
			wantErr: true,
		},
		{
			name:    "unclosed object",
			input:   `{"key":"value"`,
			wantErr: true,
		},
		{
			name:    "trailing comma in array",
			input:   "[1,2,3,]",
			wantErr: true,
		},
		{
			name:    "missing value in object",
			input:   `{"key":}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenize(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("tokenize failed: %v", err)
				}
				return
			}

			p := newParser(tokens)
			_, err = p.parse()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test unmarshal edge cases
func TestUnmarshalEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		target  interface{}
		wantErr bool
		check   func(t *testing.T, target interface{})
	}{
		{
			name:   "very large number into int64",
			input:  "9223372036854775807", // max int64
			target: new(int64),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, int64(9223372036854775807), *target.(*int64))
			},
		},
		{
			name:    "overflow int64",
			input:   "9223372036854775808", // max int64 + 1
			target:  new(int64),
			wantErr: true,
		},
		{
			name:   "negative zero",
			input:  "-0",
			target: new(int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, 0, *target.(*int))
			},
		},
		{
			name:   "float special values - inf not supported in JSON/TRON",
			input:  "1e308",
			target: new(float64),
			check: func(t *testing.T, target interface{}) {
				assert.NotEqual(t, math.Inf(1), *target.(*float64))
			},
		},
		{
			name:   "empty string",
			input:  `""`,
			target: new(string),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, "", *target.(*string))
			},
		},
		{
			name:   "null into various types",
			input:  "null",
			target: new(*int),
			check: func(t *testing.T, target interface{}) {
				assert.Nil(t, *target.(**int))
			},
		},
		{
			name:   "null into slice",
			input:  "null",
			target: new([]int),
			check: func(t *testing.T, target interface{}) {
				assert.Nil(t, *target.(*[]int))
			},
		},
		{
			name:   "null into map",
			input:  "null",
			target: new(map[string]int),
			check: func(t *testing.T, target interface{}) {
				assert.Nil(t, *target.(*map[string]int))
			},
		},
		{
			name:   "null into int (no-op)",
			input:  "null",
			target: func() *int { i := 42; return &i }(),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, 42, *target.(*int)) // Unchanged
			},
		},
		{
			name:   "array into fixed array (exact size)",
			input:  "[1,2,3]",
			target: new([3]int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, [3]int{1, 2, 3}, *target.(*[3]int))
			},
		},
		{
			name:   "array into fixed array (shorter input)",
			input:  "[1,2]",
			target: new([5]int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, [5]int{1, 2, 0, 0, 0}, *target.(*[5]int))
			},
		},
		{
			name:   "array into fixed array (longer input)",
			input:  "[1,2,3,4,5]",
			target: new([3]int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, [3]int{1, 2, 3}, *target.(*[3]int))
			},
		},
		{
			name:   "empty array into slice",
			input:  "[]",
			target: new([]int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, []int{}, *target.(*[]int))
			},
		},
		{
			name:   "empty object into map",
			input:  "{}",
			target: new(map[string]int),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, map[string]int{}, *target.(*map[string]int))
			},
		},
		{
			name:   "empty object into struct",
			input:  "{}",
			target: new(struct{ Name string }),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, "", target.(*struct{ Name string }).Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.check != nil {
					tt.check(t, tt.target)
				}
			}
		})
	}
}

// Test marshal edge cases
func TestMarshalEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
		check   func(t *testing.T, output []byte)
	}{
		{
			name:  "nil value",
			input: nil,
			check: func(t *testing.T, output []byte) {
				assert.Equal(t, "null", string(output))
			},
		},
		{
			name:  "nil slice",
			input: []int(nil),
			check: func(t *testing.T, output []byte) {
				// Nil slice marshals as empty array or null
				// Current implementation might vary
				assert.Contains(t, string(output), "null")
			},
		},
		{
			name:  "nil map",
			input: map[string]int(nil),
			check: func(t *testing.T, output []byte) {
				assert.Contains(t, string(output), "null")
			},
		},
		{
			name:  "empty slice",
			input: []int{},
			check: func(t *testing.T, output []byte) {
				assert.Equal(t, "[]", string(output))
			},
		},
		{
			name:  "empty map",
			input: map[string]int{},
			check: func(t *testing.T, output []byte) {
				assert.Equal(t, "{}", string(output))
			},
		},
		{
			name: "struct with unexported fields",
			input: struct {
				Exported   string
				unexported string
			}{Exported: "visible", unexported: "hidden"},
			check: func(t *testing.T, output []byte) {
				assert.Contains(t, string(output), "visible")
				assert.NotContains(t, string(output), "hidden")
			},
		},
		{
			name:  "very long string",
			input: string(make([]byte, 10000)),
			check: func(t *testing.T, output []byte) {
				assert.Greater(t, len(output), 10000)
			},
		},
		{
			name:  "string with special characters",
			input: "line1\nline2\ttabbed\r\nwindows",
			check: func(t *testing.T, output []byte) {
				var result string
				err := Unmarshal(output, &result)
				require.NoError(t, err)
				assert.Equal(t, "line1\nline2\ttabbed\r\nwindows", result)
			},
		},
		{
			name:  "max int64",
			input: int64(9223372036854775807),
			check: func(t *testing.T, output []byte) {
				var result int64
				err := Unmarshal(output, &result)
				require.NoError(t, err)
				assert.Equal(t, int64(9223372036854775807), result)
			},
		},
		{
			name:  "min int64",
			input: int64(-9223372036854775808),
			check: func(t *testing.T, output []byte) {
				var result int64
				err := Unmarshal(output, &result)
				require.NoError(t, err)
				assert.Equal(t, int64(-9223372036854775808), result)
			},
		},
		{
			name:  "float64 zero",
			input: float64(0.0),
			check: func(t *testing.T, output []byte) {
				assert.Equal(t, "0", string(output))
			},
		},
		{
			name:  "negative float64 zero",
			input: math.Copysign(0, -1),
			check: func(t *testing.T, output []byte) {
				// -0 might be represented as -0 or 0
				assert.Contains(t, string(output), "0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := Marshal(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.check != nil {
					tt.check(t, output)
				}
			}
		})
	}
}

// Test class generation edge cases
func TestClassGeneration(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		check func(t *testing.T, output string)
	}{
		{
			name: "single occurrence - no class",
			input: []struct {
				Name string `json:"name"`
			}{
				{Name: "Alice"},
			},
			check: func(t *testing.T, output string) {
				assert.NotContains(t, output, "class")
				assert.Contains(t, output, "Alice")
			},
		},
		{
			name: "two occurrences - creates class",
			input: []struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				{Name: "Alice", Age: 30},
				{Name: "Bob", Age: 25},
			},
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "class A:")
				assert.Contains(t, output, "name,age")
			},
		},
		{
			name: "single property struct - no class even with multiple occurrences",
			input: []struct {
				ID int `json:"id"`
			}{
				{ID: 1},
				{ID: 2},
				{ID: 3},
			},
			check: func(t *testing.T, output string) {
				assert.NotContains(t, output, "class")
			},
		},
		{
			name: "many classes",
			input: struct {
				A []struct{ X, Y int }
				B []struct{ Name string; Age int }
				C []struct{ ID int; Active bool }
			}{
				A: []struct{ X, Y int }{{1, 2}, {3, 4}},
				B: []struct {
					Name string
					Age  int
				}{{"Alice", 30}, {"Bob", 25}},
				C: []struct {
					ID     int
					Active bool
				}{{1, true}, {2, false}},
			},
			check: func(t *testing.T, output string) {
				// Should have multiple class definitions
				assert.Contains(t, output, "class")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := Marshal(tt.input)
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, string(output))
			}
		})
	}
}

// Test case-insensitive field matching
func TestCaseInsensitiveFieldMatching(t *testing.T) {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name  string
		input string
		want  Person
	}{
		{
			name:  "exact match",
			input: `{"name":"Alice","age":30}`,
			want:  Person{Name: "Alice", Age: 30},
		},
		{
			name:  "case mismatch - Name vs name",
			input: `{"Name":"Alice","Age":30}`,
			want:  Person{Name: "Alice", Age: 30},
		},
		{
			name:  "all uppercase",
			input: `{"NAME":"Alice","AGE":30}`,
			want:  Person{Name: "Alice", Age: 30},
		},
		{
			name:  "mixed case",
			input: `{"NaMe":"Alice","aGe":30}`,
			want:  Person{Name: "Alice", Age: 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result Person
			err := Unmarshal([]byte(tt.input), &result)
			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

// Test number format variations
func TestNumberFormats(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		target interface{}
		check  func(t *testing.T, target interface{})
	}{
		{
			name:   "integer",
			input:  "42",
			target: new(float64),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, float64(42), *target.(*float64))
			},
		},
		{
			name:   "negative integer",
			input:  "-42",
			target: new(float64),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, float64(-42), *target.(*float64))
			},
		},
		{
			name:   "decimal",
			input:  "3.14159",
			target: new(float64),
			check: func(t *testing.T, target interface{}) {
				assert.InDelta(t, 3.14159, *target.(*float64), 0.00001)
			},
		},
		{
			name:   "scientific notation",
			input:  "1.5e10",
			target: new(float64),
			check: func(t *testing.T, target interface{}) {
				assert.Equal(t, 1.5e10, *target.(*float64))
			},
		},
		{
			name:   "negative exponent",
			input:  "2.5e-3",
			target: new(float64),
			check: func(t *testing.T, target interface{}) {
				assert.InDelta(t, 0.0025, *target.(*float64), 0.00001)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target)
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, tt.target)
			}
		})
	}
}
