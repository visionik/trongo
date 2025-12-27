package tron

import (
	"strings"
	"testing"
)

func TestTokenTypeStringAndTokenString(t *testing.T) {
	// Known token types should produce non-empty names.
	for tt := TokenClass; tt <= TokenEOF; tt++ {
		s := tt.String()
		if s == "" {
			t.Fatalf("empty TokenType string for %d", tt)
		}
	}

	// Unknown token type should map to UNKNOWN.
	if got := TokenType(999).String(); got != "UNKNOWN" {
		t.Fatalf("expected UNKNOWN, got %q", got)
	}

	tok := Token{Type: TokenIdentifier, Value: "x", Line: 3, Column: 5}
	out := tok.String()
	if !strings.Contains(out, "IDENTIFIER") || !strings.Contains(out, "3:5") {
		t.Fatalf("unexpected Token.String(): %q", out)
	}
}
