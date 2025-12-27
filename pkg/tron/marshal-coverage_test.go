package tron

import (
	"encoding"
	"errors"
	"strings"
	"testing"
)

type testTRONMarshalerOK struct{}

func (testTRONMarshalerOK) MarshalTRON() ([]byte, error) {
	return []byte("{\"ok\":true}"), nil
}

type testTRONMarshalerErr struct{}

func (testTRONMarshalerErr) MarshalTRON() ([]byte, error) {
	return nil, errors.New("boom")
}

type testTextMarshalerOK string

func (t testTextMarshalerOK) MarshalText() ([]byte, error) {
	return []byte(string(t)), nil
}

type testTextMarshalerErr struct{}

func (testTextMarshalerErr) MarshalText() ([]byte, error) {
	return nil, errors.New("boom")
}

type testTextKey string

func (k testTextKey) MarshalText() ([]byte, error) {
	return []byte(string(k)), nil
}

var _ Marshaler = testTRONMarshalerOK{}
var _ Marshaler = testTRONMarshalerErr{}
var _ encoding.TextMarshaler = testTextMarshalerOK("")
var _ encoding.TextMarshaler = testTextMarshalerErr{}
var _ encoding.TextMarshaler = testTextKey("")

func TestMarshal_CustomMarshalerPaths(t *testing.T) {
	{
		b, err := Marshal(testTRONMarshalerOK{})
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if string(b) != "{\"ok\":true}" {
			t.Fatalf("unexpected output: %q", string(b))
		}
	}

	{
		_, err := Marshal(testTRONMarshalerErr{})
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestMarshal_TextMarshalerPaths(t *testing.T) {
	{
		b, err := Marshal(testTextMarshalerOK("hello"))
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if string(b) != "\"hello\"" {
			t.Fatalf("unexpected output: %q", string(b))
		}
	}

	{
		_, err := Marshal(testTextMarshalerErr{})
		if err == nil {
			t.Fatalf("expected error")
		}
		if !strings.Contains(err.Error(), "boom") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

func TestMarshal_UnsupportedType(t *testing.T) {
	ch := make(chan int)
	_, err := Marshal(ch)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestMarshal_CycleDetection(t *testing.T) {
	type node struct {
		Next *node `json:"next"`
	}

	n := &node{}
	n.Next = n

	_, err := Marshal(n)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMarshal_MapKeyPaths(t *testing.T) {
	// int and uint keys
	{
		b, err := Marshal(map[int]interface{}{1: true})
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if !strings.Contains(string(b), "\"1\"") {
			t.Fatalf("expected stringified key, got: %q", string(b))
		}
	}
	{
		b, err := Marshal(map[uint]interface{}{2: true})
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if !strings.Contains(string(b), "\"2\"") {
			t.Fatalf("expected stringified key, got: %q", string(b))
		}
	}

	// TextMarshaler key
	{
		b, err := Marshal(map[testTextKey]interface{}{testTextKey("k"): 1})
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if !strings.Contains(string(b), "\"k\"") {
			t.Fatalf("expected marshaled key, got: %q", string(b))
		}
	}

	// Unsupported key
	{
		_, err := Marshal(map[struct{}]interface{}{{}: 1})
		if err == nil {
			t.Fatalf("expected error")
		}
	}
}

func TestMarshal_OmitemptyCoverage(t *testing.T) {
	type s struct {
		A string              `json:"a,omitempty"`
		B []int               `json:"b,omitempty"`
		C map[string]int      `json:"c,omitempty"`
		D int                 `json:"d,omitempty"`
		E bool                `json:"e,omitempty"`
		F *int                `json:"f,omitempty"`
		G interface{}         `json:"g,omitempty"`
		H testTextMarshalerOK `json:"h,omitempty"`
	}

	// All empty => {}
	{
		b, err := Marshal(s{})
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		if string(b) != "{}" {
			t.Fatalf("expected {}, got %q", string(b))
		}
	}

	// One non-empty field => only that key present
	{
		b, err := Marshal(s{D: 1})
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		out := string(b)
		if !strings.Contains(out, "\"d\"") {
			t.Fatalf("expected d key, got %q", out)
		}
		if strings.Contains(out, "\"a\"") || strings.Contains(out, "\"b\"") || strings.Contains(out, "\"c\"") {
			t.Fatalf("unexpected omitted keys present: %q", out)
		}
	}
}
