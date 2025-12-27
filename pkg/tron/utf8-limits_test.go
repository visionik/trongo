package tron

import (
	"bytes"
	"strings"
	"testing"
)

func TestUnmarshalRejectsInvalidUTF8(t *testing.T) {
	var v interface{}
	data := []byte{0xff}
	if err := Unmarshal(data, &v); err == nil {
		t.Fatalf("expected error")
	}
}

func TestUnicodeIdentifiersImplicitRootObject(t *testing.T) {
	var v interface{}
	input := "ключ: 1\n名: \"v\"\n"
	if err := Unmarshal([]byte(input), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", v)
	}
	if m["ключ"].(float64) != 1 {
		t.Fatalf("expected ключ=1, got %#v", m["ключ"])
	}
	if m["名"].(string) != "v" {
		t.Fatalf("expected 名=\"v\", got %#v", m["名"])
	}
}

func TestUnicodeIdentifiersInClassDefinition(t *testing.T) {
	var v interface{}
	input := strings.Join([]string{
		"class 名: 値,説明",
		"",
		"名(1, \"ok\")",
	}, "\n")
	if err := Unmarshal([]byte(input), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", v)
	}
	if m["値"].(float64) != 1 {
		t.Fatalf("expected 値=1, got %#v", m["値"])
	}
	if m["説明"].(string) != "ok" {
		t.Fatalf("expected 説明=\"ok\", got %#v", m["説明"])
	}
}

func TestInvalidNumberIsRejected(t *testing.T) {
	var v interface{}
	if err := Unmarshal([]byte("1-2"), &v); err == nil {
		t.Fatalf("expected error")
	}
	if err := Unmarshal([]byte("01"), &v); err == nil {
		t.Fatalf("expected error")
	}
	if err := Unmarshal([]byte("1."), &v); err == nil {
		t.Fatalf("expected error")
	}
	if err := Unmarshal([]byte("1e"), &v); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseDepthLimit(t *testing.T) {
	// Build deeply nested arrays beyond maxParseDepth.
	depth := maxParseDepth + 5
	var b bytes.Buffer
	for i := 0; i < depth; i++ {
		b.WriteByte('[')
	}
	b.WriteByte('0')
	for i := 0; i < depth; i++ {
		b.WriteByte(']')
	}

	var v interface{}
	if err := Unmarshal(b.Bytes(), &v); err == nil {
		t.Fatalf("expected error")
	}
}
