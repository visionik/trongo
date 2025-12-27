package tron

import (
	"errors"
	"strings"
	"testing"
)

type ptrMarshalerOK struct{}

type ptrMarshalerErr struct{}

func (*ptrMarshalerOK) MarshalTRON() ([]byte, error)  { return []byte("null"), nil }
func (*ptrMarshalerErr) MarshalTRON() ([]byte, error) { return nil, errors.New("boom") }

func TestMarshal_PointerReceiverMarshaler(t *testing.T) {
	{
		b, err := Marshal(&ptrMarshalerOK{})
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if string(b) != "null" {
			t.Fatalf("unexpected: %q", string(b))
		}
	}
	{
		_, err := Marshal(&ptrMarshalerErr{})
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("unexpected: %v", err)
		}
	}
}

func TestMarshal_ByteSliceBranch(t *testing.T) {
	b, err := Marshal([]byte("abc"))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != "\"abc\"" {
		t.Fatalf("unexpected: %q", string(b))
	}
}

func TestMarshal_ClassInstantiationBranch(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	// Repeat schema to force class generation.
	v := []person{{Name: "a", Age: 1}, {Name: "b", Age: 2}}
	out, err := Marshal(v)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "class") || !strings.Contains(s, "(") {
		t.Fatalf("expected class instantiation output, got: %q", s)
	}
}
