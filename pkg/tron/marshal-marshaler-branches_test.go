package tron

import (
	"encoding"
	"errors"
	"strings"
	"testing"
)

type valueMarshalerOK struct{}

type valueMarshalerErr struct{}

func (valueMarshalerOK) MarshalTRON() ([]byte, error)  { return []byte("true"), nil }
func (valueMarshalerErr) MarshalTRON() ([]byte, error) { return nil, errors.New("boom") }

type addrMarshalerOnly struct{}

func (*addrMarshalerOnly) MarshalTRON() ([]byte, error) { return []byte("null"), nil }

type addrMarshalerErr struct{}

func (*addrMarshalerErr) MarshalTRON() ([]byte, error) { return nil, errors.New("boom") }

type valueTextMarshalerOK string

type valueTextMarshalerErr struct{}

func (v valueTextMarshalerOK) MarshalText() ([]byte, error) { return []byte(string(v)), nil }
func (valueTextMarshalerErr) MarshalText() ([]byte, error)  { return nil, errors.New("boom") }

type addrTextMarshalerOnly struct{ S string }

type addrTextMarshalerErr struct{ S string }

func (*addrTextMarshalerOnly) MarshalText() ([]byte, error) { return []byte("addr"), nil }
func (*addrTextMarshalerErr) MarshalText() ([]byte, error)  { return nil, errors.New("boom") }

var _ Marshaler = valueMarshalerOK{}
var _ Marshaler = valueMarshalerErr{}
var _ Marshaler = (*addrMarshalerOnly)(nil)
var _ Marshaler = (*addrMarshalerErr)(nil)
var _ encoding.TextMarshaler = valueTextMarshalerOK("")
var _ encoding.TextMarshaler = valueTextMarshalerErr{}
var _ encoding.TextMarshaler = (*addrTextMarshalerOnly)(nil)
var _ encoding.TextMarshaler = (*addrTextMarshalerErr)(nil)

func TestMarshal_Marshaler_ValueReceiver(t *testing.T) {
	b, err := Marshal(valueMarshalerOK{})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != "true" {
		t.Fatalf("unexpected: %q", string(b))
	}

	// Stored inside interface{}
	{
		var x interface{} = valueMarshalerOK{}
		b2, err2 := Marshal(x)
		if err2 != nil {
			t.Fatalf("Marshal(interface{}): %v", err2)
		}
		if string(b2) != "true" {
			t.Fatalf("unexpected: %q", string(b2))
		}
	}

	_, err = Marshal(valueMarshalerErr{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarshal_Marshaler_InterfaceHoldsPointerReceiver(t *testing.T) {
	// Pointer receiver inside interface{} should still be honored.
	var x interface{} = &ptrMarshalerOK{}
	b, err := Marshal(x)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != "null" {
		t.Fatalf("unexpected: %q", string(b))
	}
}

func TestMarshal_Marshaler_AddrBranch(t *testing.T) {
	// Marshal pointer to struct so fields are addressable; field type only has pointer-receiver Marshaler.
	type holder struct {
		V addrMarshalerOnly `json:"v"`
	}
	b, err := Marshal(&holder{V: addrMarshalerOnly{}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != "{\"v\":null}" {
		t.Fatalf("unexpected: %q", string(b))
	}

	// Error path
	type holderErr struct {
		V addrMarshalerErr `json:"v"`
	}
	_, err = Marshal(&holderErr{V: addrMarshalerErr{}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarshal_TextMarshaler_ValueReceiver(t *testing.T) {
	b, err := Marshal(valueTextMarshalerOK("hello"))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != "\"hello\"" {
		t.Fatalf("unexpected: %q", string(b))
	}

	_, err = Marshal(valueTextMarshalerErr{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarshal_TextMarshaler_AddrBranch(t *testing.T) {
	type holder struct {
		V addrTextMarshalerOnly `json:"v"`
	}
	b, err := Marshal(&holder{V: addrTextMarshalerOnly{S: "x"}})
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(b) != "{\"v\":\"addr\"}" {
		t.Fatalf("unexpected: %q", string(b))
	}

	type holderErr struct {
		V addrTextMarshalerErr `json:"v"`
	}
	_, err = Marshal(&holderErr{V: addrTextMarshalerErr{S: "x"}})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected error: %v", err)
	}
}
