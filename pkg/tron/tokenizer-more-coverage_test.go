package tron

import "testing"

func TestTokenize_InvalidNumberBranch(t *testing.T) {
	// '-' alone should be rejected by parseNumberJSON.
	if _, err := tokenize("-"); err == nil {
		t.Fatalf("expected error")
	}
	// exponent with missing digits should be rejected.
	if _, err := tokenize("1e"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestTokenize_WhitespaceBranch(t *testing.T) {
	toks, err := tokenize("\t\r \n")
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	if len(toks) < 2 {
		t.Fatalf("expected at least NEWLINE and EOF")
	}
	if toks[0].Type != TokenNewline {
		t.Fatalf("expected first token NEWLINE, got %v", toks[0].Type)
	}
	if toks[len(toks)-1].Type != TokenEOF {
		t.Fatalf("expected last token EOF")
	}
}

func TestTokenize_CommentStopsAtNewline(t *testing.T) {
	toks, err := tokenize("# hi\ntrue")
	if err != nil {
		t.Fatalf("tokenize: %v", err)
	}
	foundNewline := false
	foundTrue := false
	for _, tok := range toks {
		if tok.Type == TokenNewline {
			foundNewline = true
		}
		if tok.Type == TokenTrue {
			foundTrue = true
		}
	}
	if !foundNewline || !foundTrue {
		t.Fatalf("expected NEWLINE and TRUE tokens")
	}
}
