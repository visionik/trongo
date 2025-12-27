package tron

import (
	"encoding"
	"reflect"
	"testing"
)

type ifaceWithMethod interface{ M() }

type textKey struct{ S string }

func (k *textKey) UnmarshalText(b []byte) error {
	k.S = string(b)
	return nil
}

var _ encoding.TextUnmarshaler = (*textKey)(nil)

func TestDecodeString_MoreBranches(t *testing.T) {
	d := &decoder{}

	// string -> []byte
	{
		var b []byte
		dst := reflect.ValueOf(&b).Elem()
		if err := d.decodeString("hi", dst); err != nil {
			t.Fatalf("decodeString: %v", err)
		}
		if string(b) != "hi" {
			t.Fatalf("expected hi, got %q", string(b))
		}
	}

	// string -> interface with methods should error
	{
		var x ifaceWithMethod
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeString("hi", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// string -> non-byte slice should error
	{
		var s []string
		dst := reflect.ValueOf(&s).Elem()
		if err := d.decodeString("hi", dst); err == nil {
			t.Fatalf("expected error")
		}
	}
}

func TestDecodeMapKey_MoreBranches(t *testing.T) {
	d := &decoder{}

	// bad int key
	{
		var k int
		dst := reflect.ValueOf(&k).Elem()
		if err := d.decodeMapKey("not-an-int", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// bad uint key
	{
		var k uint
		dst := reflect.ValueOf(&k).Elem()
		if err := d.decodeMapKey("-1", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// TextUnmarshaler key
	{
		k := &textKey{}
		dst := reflect.ValueOf(k).Elem()
		if err := d.decodeMapKey("abc", dst); err != nil {
			t.Fatalf("decodeMapKey: %v", err)
		}
		if k.S != "abc" {
			t.Fatalf("expected abc, got %q", k.S)
		}
	}

	// unsupported kind
	{
		var k bool
		dst := reflect.ValueOf(&k).Elem()
		if err := d.decodeMapKey("x", dst); err == nil {
			t.Fatalf("expected error")
		}
	}
}

func TestDecoderDecode_TypedInterfaceFallback(t *testing.T) {
	// Ensure decode() returns UnmarshalTypeError when interface has methods.
	d := &decoder{}
	var x ifaceWithMethod
	dst := reflect.ValueOf(&x).Elem()
	if err := d.decode("hi", dst); err == nil {
		t.Fatalf("expected error")
	}
}
