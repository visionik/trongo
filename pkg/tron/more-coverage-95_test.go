package tron

import (
	"encoding"
	"reflect"
	"testing"
)

type ifaceObj interface{ M() }

type textKeyOK struct{ S string }

func (k textKeyOK) MarshalText() ([]byte, error) { return []byte(k.S), nil }

var _ encoding.TextMarshaler = textKeyOK{}

func TestSerializeMapKey_AllKinds(t *testing.T) {
	e := &encoder{}

	// string key
	{
		out, err := e.serializeMapKey(reflect.ValueOf("k"))
		if err != nil {
			t.Fatalf("string: %v", err)
		}
		if out != "\"k\"" {
			t.Fatalf("unexpected: %q", out)
		}
	}

	// int key
	{
		out, err := e.serializeMapKey(reflect.ValueOf(int64(1)))
		if err != nil {
			t.Fatalf("int: %v", err)
		}
		if out != "\"1\"" {
			t.Fatalf("unexpected: %q", out)
		}
	}

	// uint key
	{
		out, err := e.serializeMapKey(reflect.ValueOf(uint64(2)))
		if err != nil {
			t.Fatalf("uint: %v", err)
		}
		if out != "\"2\"" {
			t.Fatalf("unexpected: %q", out)
		}
	}

	// TextMarshaler success
	{
		out, err := e.serializeMapKey(reflect.ValueOf(textKeyOK{S: "txt"}))
		if err != nil {
			t.Fatalf("text: %v", err)
		}
		if out != "\"txt\"" {
			t.Fatalf("unexpected: %q", out)
		}
	}
}

func TestDecodeObject_InterfaceWithMethods_ErrorBranch(t *testing.T) {
	d := &decoder{}
	var x ifaceObj
	dst := reflect.ValueOf(&x).Elem()
	if err := d.decodeObject(map[string]interface{}{"a": float64(1)}, dst); err == nil {
		t.Fatalf("expected error")
	}
}
