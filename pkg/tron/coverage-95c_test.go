package tron

import (
	"reflect"
	"testing"
)

type namedStructForDecode struct {
	A int `json:"a"`
}

func TestDecodeStruct_ErrorAndIgnoreBranches(t *testing.T) {
	d := &decoder{}

	// Unknown field should be ignored.
	{
		var s namedStructForDecode
		dst := reflect.ValueOf(&s).Elem()
		err := d.decodeStruct(map[string]interface{}{"unknown": float64(1)}, dst)
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	}

	// Type mismatch should return UnmarshalTypeError with Struct/Field.
	{
		var s namedStructForDecode
		dst := reflect.ValueOf(&s).Elem()
		err := d.decodeStruct(map[string]interface{}{"a": "x"}, dst)
		if err == nil {
			t.Fatalf("expected error")
		}
		ute, ok := err.(*UnmarshalTypeError)
		if !ok {
			t.Fatalf("expected *UnmarshalTypeError, got %T", err)
		}
		if ute.Struct != "namedStructForDecode" || ute.Field != "A" {
			t.Fatalf("unexpected struct/field: %q.%q", ute.Struct, ute.Field)
		}
	}
}

func TestParseString_MoreErrorBranches(t *testing.T) {
	// invalid escape too short
	if _, err := tokenize("\"\\u\""); err == nil {
		t.Fatalf("expected error")
	}
	// invalid UTF-8 inside string
	if _, err := tokenize(string([]byte{'"', 0xff, '"'})); err == nil {
		t.Fatalf("expected error")
	}
	// invalid UTF-8 right after backslash
	if _, err := tokenize(string([]byte{'"', '\\', 0xff, '"'})); err == nil {
		t.Fatalf("expected error")
	}
}
