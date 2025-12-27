package tron

import (
	"reflect"
	"testing"
)

func TestGenerateClassName_Branches(t *testing.T) {
	if got := generateClassName(0); got != "A" {
		t.Fatalf("expected A, got %q", got)
	}
	if got := generateClassName(25); got != "Z" {
		t.Fatalf("expected Z, got %q", got)
	}
	// cycle > 0
	if got := generateClassName(26); got != "A1" {
		t.Fatalf("expected A1, got %q", got)
	}
}

func TestIsValidHex_Branches(t *testing.T) {
	if isValidHex("") {
		t.Fatalf("expected false")
	}
	if isValidHex("123") {
		t.Fatalf("expected false")
	}
	if isValidHex("12345") {
		t.Fatalf("expected false")
	}
	if isValidHex("12G4") {
		t.Fatalf("expected false")
	}
	if !isValidHex("12aF") {
		t.Fatalf("expected true")
	}
}

func TestDecodeNumberLiteral_ErrorBranches(t *testing.T) {
	d := &decoder{}

	// int overflow
	{
		var x int8
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeNumberLiteral("128", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// int: non-plain int falls back to float parse but still errors for int targets
	{
		var x int64
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeNumberLiteral("1e3", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// uint: negative
	{
		var x uint64
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeNumberLiteral("-1", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// float: parse error
	{
		var x float64
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeNumberLiteral("not-a-number", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// interface{}: parse error
	{
		var x interface{}
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeNumberLiteral("not-a-number", dst); err == nil {
			t.Fatalf("expected error")
		}
	}
}

func TestDecode_UnknownParsedTypeBranch(t *testing.T) {
	d := &decoder{}
	var out interface{}
	dst := reflect.ValueOf(&out).Elem()
	if err := d.decode(struct{}{}, dst); err == nil {
		t.Fatalf("expected error")
	}
}

func TestTokenize_StringErrorBranches(t *testing.T) {
	// Unterminated string
	if _, err := tokenize("\""); err == nil {
		t.Fatalf("expected error")
	}

	// Invalid unicode escape (bad hex)
	if _, err := tokenize("\"\\u12G4\""); err == nil {
		t.Fatalf("expected error")
	}

	// Unpaired surrogate should error
	if _, err := tokenize("\"\\uD800\""); err == nil {
		t.Fatalf("expected error")
	}

	// Invalid UTF-8 inside comment scanning should error
	if _, err := tokenize(string([]byte{'#', 0xff})); err == nil {
		t.Fatalf("expected error")
	}
}
