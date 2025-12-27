package tron

import (
	"encoding"
	"reflect"
	"testing"
)

type textValue struct{ S string }

func (t *textValue) UnmarshalText(b []byte) error {
	t.S = string(b)
	return nil
}

var _ encoding.TextUnmarshaler = (*textValue)(nil)

func TestParserPeek_OutOfRangeBranch(t *testing.T) {
	p := newParser([]Token{{Type: TokenEOF}})
	if tok := p.peek(999); tok.Type != TokenEOF {
		t.Fatalf("expected EOF, got %v", tok.Type)
	}
}

func TestParseNumberValue_ErrorBranch(t *testing.T) {
	p := newParser([]Token{{Type: TokenEOF}})
	if _, err := p.parseNumberValue("not-a-number"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDecoderDecode_TextUnmarshalerPath(t *testing.T) {
	d := &decoder{}
	v := &textValue{}
	dst := reflect.ValueOf(v).Elem()
	if err := d.decode("abc", dst); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if v.S != "abc" {
		t.Fatalf("expected abc, got %q", v.S)
	}
}
