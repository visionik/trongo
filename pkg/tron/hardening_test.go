package tron

import (
	"strings"
	"sync"
	"testing"
)

var limitsTestMu sync.Mutex

func withLimits(t *testing.T, inputBytes, tokens, parseDepth, walkDepth int) {
	t.Helper()
	limitsTestMu.Lock()
	oldInputBytes := maxInputBytes
	oldTokens := maxTokens
	oldParseDepth := maxParseDepth
	oldWalkDepth := maxWalkDepth

	maxInputBytes = inputBytes
	maxTokens = tokens
	maxParseDepth = parseDepth
	maxWalkDepth = walkDepth

	t.Cleanup(func() {
		maxInputBytes = oldInputBytes
		maxTokens = oldTokens
		maxParseDepth = oldParseDepth
		maxWalkDepth = oldWalkDepth
		limitsTestMu.Unlock()
	})
}

func TestStringUnicodeEscapes_Surrogates(t *testing.T) {
	var v interface{}
	// 😀 as surrogate pair
	input := "\"\\uD83D\\uDE00\""
	if err := Unmarshal([]byte(input), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("expected string, got %T", v)
	}
	if s != "😀" {
		t.Fatalf("expected 😀, got %q", s)
	}
}

func TestStringUnicodeEscapes_ExtraHexDigitIsLiteral(t *testing.T) {
	var v interface{}
	input := "\"\\u12345\""
	if err := Unmarshal([]byte(input), &v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("expected string, got %T", v)
	}
	if s != string(rune(0x1234))+"5" {
		t.Fatalf("unexpected value: %q", s)
	}
}

func TestStringUnicodeEscapes_Invalid(t *testing.T) {
	cases := []string{
		"\"\\u12G4\"",        // bad hex
		"\"\\uD83D\"",        // lone high surrogate
		"\"\\uDE00\"",        // lone low surrogate
		"\"\\uD83D\\u0041\"", // high surrogate not followed by low surrogate
		"\"\\uD83D\\uD83D\"", // two highs
		"\"\\uDE00\\uDE00\"", // two lows
		"\"\\u\"",            // too short
		"\"\\u123\"",         // too short
		"\"\\uD83D\\uDE0\"",  // too short second
	}

	for _, input := range cases {
		input := input
		t.Run(input, func(t *testing.T) {
			var v interface{}
			if err := Unmarshal([]byte(input), &v); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestNumberGrammar_Valid(t *testing.T) {
	cases := []string{
		"0",
		"-0",
		"1",
		"-1",
		"10",
		"0.1",
		"1.0",
		"-1.25",
		"1e0",
		"1E0",
		"1e+9",
		"1e-9",
		"-1E-9",
	}
	for _, input := range cases {
		input := input
		t.Run(input, func(t *testing.T) {
			var v interface{}
			if err := Unmarshal([]byte(input), &v); err != nil {
				t.Fatalf("expected ok, got err: %v", err)
			}
		})
	}
}

func TestNumberGrammar_Invalid(t *testing.T) {
	cases := []string{
		"+1",
		"-",
		".",
		"-.1",
		"01",
		"00",
		"1.",
		"1e",
		"1e+",
		"1e-",
		"1E+",
		"1-2",
		"NaN",
		"Infinity",
		"-Infinity",
	}
	for _, input := range cases {
		input := input
		t.Run(input, func(t *testing.T) {
			var v interface{}
			if err := Unmarshal([]byte(input), &v); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestTrailingTokensRejected(t *testing.T) {
	cases := []string{
		"true false",
		"0 1",
		"{} {}",
		"[] []",
		"\"a\" \"b\"",
	}

	for _, input := range cases {
		input := input
		t.Run(input, func(t *testing.T) {
			var v interface{}
			if err := Unmarshal([]byte(input), &v); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestTokenLimitEnforced(t *testing.T) {
	withLimits(t, maxInputBytes, 20, maxParseDepth, maxWalkDepth)

	// Each line: identifier ':' identifier '\n' => ~4 tokens.
	// Build enough lines to exceed maxTokens.
	var b strings.Builder
	for i := 0; i < 10; i++ {
		b.WriteString("a:b\n")
	}

	var v interface{}
	if err := Unmarshal([]byte(b.String()), &v); err == nil {
		t.Fatalf("expected error")
	}
}

func TestInputSizeLimitEnforced(t *testing.T) {
	withLimits(t, 64, maxTokens, maxParseDepth, maxWalkDepth)

	var v interface{}
	big := strings.Repeat("a", 128)
	if err := Unmarshal([]byte("\""+big+"\""), &v); err == nil {
		t.Fatalf("expected error")
	}
}

func TestMarshalWalkDepthLimitEnforced(t *testing.T) {
	withLimits(t, maxInputBytes, maxTokens, maxParseDepth, 4)

	// Create depth > 4: [[[[[0]]]]]
	v := []interface{}{[]interface{}{[]interface{}{[]interface{}{[]interface{}{0}}}}}
	if _, err := Marshal(v); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSyntaxErrorOffsetCountsBytesWithMultibyteUTF8(t *testing.T) {
	var v interface{}
	input := "名: 1\n$" // '$' is unexpected
	err := Unmarshal([]byte(input), &v)
	if err == nil {
		t.Fatalf("expected error")
	}
	syn, ok := err.(*SyntaxError)
	if !ok {
		t.Fatalf("expected *SyntaxError, got %T (%v)", err, err)
	}
	// "名" is 3 bytes. Up to '$': len("名: 1\n") == 7 bytes.
	if syn.Offset != 7 {
		t.Fatalf("expected Offset=7, got %d", syn.Offset)
	}
	if !strings.Contains(syn.Error(), "Unexpected character") {
		t.Fatalf("unexpected error message: %q", syn.Error())
	}
}
