package tron

import (
	"reflect"
	"testing"
)

type ifaceBool interface{ M() }

type ifaceBoolImpl struct{}

func (ifaceBoolImpl) M() {}

func TestTokenizer_CoversManyTokenTypes(t *testing.T) {
	input := "class A: a,b\n" +
		"a=1;\n" +
		"{a:1,b:[true,false,null],c:(x)}\n" +
		"# comment\n" +
		"A(1.25e+2,\"s\")\n"

	toks, err := tokenize(input)
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}

	seen := map[TokenType]bool{}
	for _, tok := range toks {
		seen[tok.Type] = true
	}

	want := []TokenType{
		TokenClass, TokenIdentifier, TokenColon, TokenComma, TokenNewline,
		TokenEquals, TokenSemicolon,
		TokenLBrace, TokenRBrace, TokenLBracket, TokenRBracket,
		TokenLParen, TokenRParen,
		TokenTrue, TokenFalse, TokenNull,
		TokenNumber, TokenString,
		TokenEOF,
	}
	for _, tt := range want {
		if !seen[tt] {
			t.Fatalf("expected to see token type %v", tt)
		}
	}
}

func TestParseImplicitObjectDepth_CommaSeparatorBranch(t *testing.T) {
	toks, err := tokenize("a: 1, b: 2")
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	p := newParser(toks)
	m, err := p.parseImplicitObjectDepth(1)
	if err != nil {
		t.Fatalf("parseImplicitObjectDepth: %v", err)
	}
	if m["a"].(float64) != 1 || m["b"].(float64) != 2 {
		t.Fatalf("unexpected: %#v", m)
	}
}

func TestDecodeBool_InterfaceWithMethods_ErrorBranch(t *testing.T) {
	d := &decoder{}
	var x ifaceBool
	dst := reflect.ValueOf(&x).Elem()
	if err := d.decodeBool(true, dst); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDecodeSlice_ElementErrorBranch(t *testing.T) {
	d := &decoder{}
	var dst []int
	rv := reflect.ValueOf(&dst).Elem()
	if err := d.decodeSlice([]interface{}{"x"}, rv); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDecodeArrayFixed_ErrorBranch(t *testing.T) {
	d := &decoder{}
	var arr [2]int
	rv := reflect.ValueOf(&arr).Elem()
	if err := d.decodeArrayFixed([]interface{}{float64(1), "x"}, rv); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDecodeMap_KeyAndValueErrorBranches(t *testing.T) {
	d := &decoder{}

	// key parse error for int map
	{
		m := map[int]int{}
		rv := reflect.ValueOf(&m).Elem()
		if err := d.decodeMap(map[string]interface{}{"x": float64(1)}, rv); err == nil {
			t.Fatalf("expected error")
		}
	}

	// value decode error
	{
		m := map[string]int{}
		rv := reflect.ValueOf(&m).Elem()
		if err := d.decodeMap(map[string]interface{}{"a": "x"}, rv); err == nil {
			t.Fatalf("expected error")
		}
	}
}

func TestMinMax_DefaultBranches(t *testing.T) {
	// Kinds not explicitly handled should hit default.
	if got := minInt(reflect.TypeOf(complex64(0))); got != minInt(reflect.TypeOf(int64(0))) {
		t.Fatalf("unexpected minInt default")
	}
	if got := maxInt(reflect.TypeOf(complex64(0))); got != maxInt(reflect.TypeOf(int64(0))) {
		t.Fatalf("unexpected maxInt default")
	}
	if got := maxUint(reflect.TypeOf(complex64(0))); got != maxUint(reflect.TypeOf(uint64(0))) {
		t.Fatalf("unexpected maxUint default")
	}
}
