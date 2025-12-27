package tron

import (
	"reflect"
	"testing"
)

func TestParseImplicitObjectDepth_ErrorBranches(t *testing.T) {
	// Unexpected token after a value (neither comma, key, nor EOF).
	{
		toks, err := tokenize("a: 1 2")
		if err != nil {
			t.Fatalf("tokenize: %v", err)
		}
		p := newParser(toks)
		_, err = p.parseImplicitObjectDepth(1)
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	// Invalid key token.
	{
		toks, err := tokenize("1: 2")
		if err != nil {
			t.Fatalf("tokenize: %v", err)
		}
		p := newParser(toks)
		_, err = p.parseImplicitObjectDepth(1)
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	// No newline/comma separator, but next token looks like a key:colon => allowed.
	{
		toks, err := tokenize("a: 1 b: 2")
		if err != nil {
			t.Fatalf("tokenize: %v", err)
		}
		p := newParser(toks)
		v, err := p.parseImplicitObjectDepth(1)
		if err != nil {
			t.Fatalf("expected ok, got %v", err)
		}
		if v["a"].(float64) != 1 || v["b"].(float64) != 2 {
			t.Fatalf("unexpected result: %#v", v)
		}
	}
}

func TestGetStructFieldValue_MissBranch(t *testing.T) {
	type s struct {
		A int `json:"a"`
	}
	e := &encoder{}
	v := reflect.ValueOf(s{A: 1})

	got := e.getStructFieldValue(v, "a")
	if !got.IsValid() || got.Int() != 1 {
		t.Fatalf("expected a=1")
	}

	miss := e.getStructFieldValue(v, "missing")
	if miss.IsValid() {
		t.Fatalf("expected invalid reflect.Value")
	}
}

func TestInvalidUnmarshalError_ErrorBranches(t *testing.T) {
	// nil type
	{
		e := &InvalidUnmarshalError{Type: nil}
		_ = e.Error()
	}
	// non-pointer type
	{
		e := &InvalidUnmarshalError{Type: reflect.TypeOf(0)}
		_ = e.Error()
	}
	// nil pointer type
	{
		var p *int
		e := &InvalidUnmarshalError{Type: reflect.TypeOf(p)}
		_ = e.Error()
	}
}
