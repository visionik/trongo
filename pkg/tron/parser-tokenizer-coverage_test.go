package tron

import "testing"

func TestTokenize_InvalidUTF8Branch(t *testing.T) {
	// Unmarshal rejects invalid UTF-8 before tokenization, so call tokenize directly.
	s := string([]byte{0xff})
	_, err := tokenize(s)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParser_CurrentBeyondEOFBranch(t *testing.T) {
	p := newParser([]Token{{Type: TokenEOF}})
	p.pos = 999
	_ = p.current()
}

func TestParser_ImplicitObjectDepthLimitBranch(t *testing.T) {
	toks, err := tokenize("a: 1")
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	p := newParser(toks)
	_, err = p.parseImplicitObjectDepth(maxParseDepth + 1)
	if err == nil {
		t.Fatalf("expected error")
	}
}
