package tron

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

// FuzzTokenize tests the tokenizer with random input
func FuzzTokenize(f *testing.F) {
	// Seed corpus with valid TRON inputs
	seeds := []string{
		`class A: x,y`,
		`[1,2,3]`,
		`{"name":"Alice"}`,
		`A("test",123)`,
		`true`,
		`false`,
		`null`,
		`"hello world"`,
		`-123.456`,
		`class B: a,b,c

[B(1,2,3)]`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Tokenize should never panic
		tokens, err := tokenize(input)

		// If no error, tokens should be valid
		if err == nil {
			if len(tokens) == 0 {
				t.Errorf("tokenize returned no tokens and no error for input: %q", input)
			}
			// Last token should always be EOF
			if tokens[len(tokens)-1].Type != TokenEOF {
				t.Errorf("last token is not EOF for input: %q", input)
			}
		}
	})
}

// FuzzParser tests the parser with random input
func FuzzParser(f *testing.F) {
	// Seed corpus
	seeds := []string{
		`42`,
		`"test"`,
		`[1,2,3]`,
		`{"key":"value"}`,
		`class A: x

A(1)`,
		`true`,
		`null`,
		`[1,"two",true,null]`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Parser should never panic
		tokens, err := tokenize(input)
		if err != nil {
			return // Skip if tokenization fails
		}

		p := newParser(tokens)
		result, err := p.parse()

		// If parsing succeeds, result should be non-nil or error should be set
		if err == nil && result == nil {
			// Empty input is ok
			if input != "" && input != "\n" {
				t.Errorf("parser returned nil result with no error for: %q", input)
			}
		}
	})
}

// FuzzUnmarshal tests unmarshaling with random input
func FuzzUnmarshal(f *testing.F) {
	// Seed corpus
	seeds := []string{
		`42`,
		`"hello"`,
		`true`,
		`null`,
		`[1,2,3]`,
		`{"name":"Alice","age":30}`,
		`class A: x,y

A(1,2)`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Unmarshal should never panic
		var result interface{}
		err := Unmarshal([]byte(input), &result)

		// We don't care if it errors (invalid input), just that it doesn't panic
		_ = err
	})
}

// FuzzRoundTrip tests that Marshal → Unmarshal preserves data
func FuzzRoundTrip(f *testing.F) {
	// Seed corpus with various Go values serialized
	type SimpleStruct struct {
		Name  string
		Value int
	}

	seeds := []interface{}{
		42,
		"hello",
		true,
		[]int{1, 2, 3},
		map[string]int{"a": 1, "b": 2},
		SimpleStruct{Name: "test", Value: 123},
		[]SimpleStruct{
			{Name: "a", Value: 1},
			{Name: "b", Value: 2},
		},
	}

	for _, seed := range seeds {
		data, err := Marshal(seed)
		if err == nil {
			f.Add(data)
		}
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Try to unmarshal into interface{}
		var result interface{}
		err := Unmarshal(data, &result)
		if err != nil {
			return // Invalid TRON is ok
		}

		// Try to marshal the result back
		data2, err := Marshal(result)
		if err != nil {
			t.Errorf("failed to re-marshal: %v", err)
			return
		}

		// Unmarshal again
		var result2 interface{}
		err = Unmarshal(data2, &result2)
		if err != nil {
			t.Errorf("failed to re-unmarshal: %v", err)
			return
		}

		// Results should be deeply equal
		if !reflect.DeepEqual(result, result2) {
			t.Errorf("round-trip not equal:\noriginal: %#v\nfinal: %#v", result, result2)
		}
	})
}

// FuzzMarshal tests marshaling doesn't panic on various inputs
func FuzzMarshalString(f *testing.F) {
	// Seed with various strings
	seeds := []string{
		"",
		"hello",
		"with\nnewline",
		"with\ttab",
		"unicode: ñ é ü",
		`with"quotes`,
		"very long string: " + string(make([]byte, 1000)),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, s string) {
		// Marshal should never panic
		data, err := Marshal(s)
		if err != nil {
			t.Errorf("failed to marshal string: %v", err)
			return
		}

		// Should be able to unmarshal back
		var result string
		err = Unmarshal(data, &result)
		if err != nil {
			t.Errorf("failed to unmarshal string: %v", err)
			return
		}

		// Should match original
		if result != s {
			t.Errorf("string round-trip failed:\noriginal: %q\nresult: %q", s, result)
		}
	})
}

// FuzzMarshalInt tests integer marshaling
func FuzzMarshalInt(f *testing.F) {
	// Seed with edge cases
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(-1))
	f.Add(int64(9223372036854775807))  // max int64
	f.Add(int64(-9223372036854775808)) // min int64

	f.Fuzz(func(t *testing.T, i int64) {
		// Marshal
		data, err := Marshal(i)
		if err != nil {
			t.Errorf("failed to marshal int64: %v", err)
			return
		}

		// Unmarshal
		var result int64
		err = Unmarshal(data, &result)
		if err != nil {
			t.Errorf("failed to unmarshal int64: %v", err)
			return
		}

		// Should match
		if result != i {
			t.Errorf("int64 round-trip failed: %d != %d", i, result)
		}
	})
}

// FuzzMarshalSlice tests slice marshaling with random data
func FuzzMarshalSlice(f *testing.F) {
	// Seed with various slice patterns
	f.Add([]byte{1, 2, 3})
	f.Add([]byte{})
	f.Add([]byte{255, 0, 128})

	f.Fuzz(func(t *testing.T, data []byte) {
		// Convert to []int for testing
		ints := make([]int, len(data))
		for i, b := range data {
			ints[i] = int(b)
		}

		// Marshal
		tronData, err := Marshal(ints)
		if err != nil {
			t.Errorf("failed to marshal slice: %v", err)
			return
		}

		// Unmarshal
		var result []int
		err = Unmarshal(tronData, &result)
		if err != nil {
			t.Errorf("failed to unmarshal slice: %v", err)
			return
		}

		// Should match
		if !reflect.DeepEqual(ints, result) {
			t.Errorf("slice round-trip failed")
		}
	})
}

// FuzzJSONCompatibility tests that valid JSON can be unmarshaled as TRON
func FuzzJSONCompatibility(f *testing.F) {
	// Seed with JSON-compatible values
	seeds := []string{
		`{"name":"Alice","age":30}`,
		`[1,2,3,4,5]`,
		`true`,
		`false`,
		`null`,
		`"string"`,
		`123.456`,
		`{"nested":{"value":42}}`,
		`[{"a":1},{"a":2}]`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, jsonInput string) {
		// Try to unmarshal with JSON
		var jsonResult interface{}
		err := json.Unmarshal([]byte(jsonInput), &jsonResult)
		if err != nil {
			return // Not valid JSON, skip
		}

		// Try to unmarshal with TRON (should work for JSON-compatible syntax)
		var tronResult interface{}
		err = Unmarshal([]byte(jsonInput), &tronResult)

		// If it's valid JSON object notation, TRON should handle it
		// (Note: TRON also allows unquoted identifiers in objects, but JSON doesn't)
		if err != nil {
			// Some JSON might not be valid TRON (e.g., numbers with leading zeros)
			// That's ok, we just want to ensure no panics
			return
		}

		// Both should produce equivalent results (though types might differ slightly)
		// We mainly care that parsing succeeds without panic
	})
}

// FuzzClassDefinition tests class definition parsing
func FuzzClassDefinition(f *testing.F) {
	seeds := []string{
		"class A: x",
		"class B: x,y",
		"class C: a,b,c,d,e",
		`class D: "prop with space"`,
		"class E: _underscore",
		"class F: camelCase",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, classDef string) {
		// Add a simple value to make it complete TRON
		input := classDef + "\n\nnull"

		tokens, err := tokenize(input)
		if err != nil {
			return
		}

		p := newParser(tokens)
		_, err = p.parse()

		// Don't care about errors, just that it doesn't panic
		_ = err
	})
}

// FuzzNumberParsing tests various number formats
func FuzzNumberParsing(f *testing.F) {
	seeds := []string{
		"0",
		"1",
		"-1",
		"123.456",
		"-123.456",
		"1e10",
		"1.5e-5",
		"0.0",
		"-0.0",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, numStr string) {
		var result float64
		err := Unmarshal([]byte(numStr), &result)

		// Don't care if it fails (invalid number), just that it doesn't panic
		_ = err
	})
}

// FuzzStringEscaping tests string escape handling
func FuzzStringEscaping(f *testing.F) {
	seeds := []string{
		`"simple"`,
		`"with\nnewline"`,
		`"with\ttab"`,
		`"with\"quote"`,
		`"with\\backslash"`,
		`"with\u0048unicode"`,
		`""`,
		`"emoji: 😀"`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, str string) {
		var result string
		err := Unmarshal([]byte(str), &result)

		// Don't care about errors, just no panics
		_ = err
	})
}
