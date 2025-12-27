package tron

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// TokenType represents the type of a token in TRON format.
type TokenType int

const (
	// TokenClass represents the "class" keyword
	TokenClass TokenType = iota
	// TokenIdentifier represents an identifier (class name, property name)
	TokenIdentifier
	// TokenString represents a quoted string literal
	TokenString
	// TokenNumber represents a numeric literal
	TokenNumber
	// TokenTrue represents the "true" keyword
	TokenTrue
	// TokenFalse represents the "false" keyword
	TokenFalse
	// TokenNull represents the "null" keyword
	TokenNull
	// TokenLParen represents "("
	TokenLParen
	// TokenRParen represents ")"
	TokenRParen
	// TokenLBracket represents "["
	TokenLBracket
	// TokenRBracket represents "]"
	TokenRBracket
	// TokenLBrace represents "{"
	TokenLBrace
	// TokenRBrace represents "}"
	TokenRBrace
	// TokenComma represents ","
	TokenComma
	// TokenColon represents ":"
	TokenColon
	// TokenSemicolon represents ";"
	TokenSemicolon
	// TokenEquals represents "="
	TokenEquals
	// TokenNewline represents a newline character
	TokenNewline
	// TokenEOF represents end of input
	TokenEOF
)

// String returns a string representation of the token type.
func (t TokenType) String() string {
	switch t {
	case TokenClass:
		return "CLASS"
	case TokenIdentifier:
		return "IDENTIFIER"
	case TokenString:
		return "STRING"
	case TokenNumber:
		return "NUMBER"
	case TokenTrue:
		return "TRUE"
	case TokenFalse:
		return "FALSE"
	case TokenNull:
		return "NULL"
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenLBracket:
		return "LBRACKET"
	case TokenRBracket:
		return "RBRACKET"
	case TokenLBrace:
		return "LBRACE"
	case TokenRBrace:
		return "RBRACE"
	case TokenComma:
		return "COMMA"
	case TokenColon:
		return "COLON"
	case TokenSemicolon:
		return "SEMICOLON"
	case TokenEquals:
		return "EQUALS"
	case TokenNewline:
		return "NEWLINE"
	case TokenEOF:
		return "EOF"
	default:
		return "UNKNOWN"
	}
}

// Token represents a single token in TRON format.
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// String returns a string representation of the token.
func (t Token) String() string {
	return fmt.Sprintf("%s(%q) at %d:%d", t.Type, t.Value, t.Line, t.Column)
}

// tokenize parses the input string and returns a slice of tokens.
func tokenize(input string) ([]Token, error) {
	var tokens []Token
	cursor := 0 // byte index
	line := 1
	column := 1 // rune column within line

	appendToken := func(tok Token) error {
		if len(tokens) >= maxTokens {
			return &SyntaxError{msg: "too many tokens", Offset: int64(cursor)}
		}
		tokens = append(tokens, tok)
		return nil
	}

	for cursor < len(input) {
		r, size := utf8.DecodeRuneInString(input[cursor:])
		if r == utf8.RuneError && size == 1 {
			return nil, &SyntaxError{msg: "invalid UTF-8", Offset: int64(cursor)}
		}

		// Handle whitespace (except newlines)
		if r == ' ' || r == '\t' || r == '\r' {
			cursor += size
			column++
			continue
		}

		// Handle newlines
		if r == '\n' {
			if err := appendToken(Token{Type: TokenNewline, Value: "\n", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			line++
			column = 1
			continue
		}

		// Handle comments
		if r == '#' {
			// Consume until newline or EOF
			cursor += size
			column++
			for cursor < len(input) {
				r2, s2 := utf8.DecodeRuneInString(input[cursor:])
				if r2 == utf8.RuneError && s2 == 1 {
					return nil, &SyntaxError{msg: "invalid UTF-8", Offset: int64(cursor)}
				}
				if r2 == '\n' {
					break
				}
				cursor += s2
				column++
			}
			continue
		}

		// Handle single-character tokens
		switch r {
		case '(':
			if err := appendToken(Token{Type: TokenLParen, Value: "(", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		case ')':
			if err := appendToken(Token{Type: TokenRParen, Value: ")", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		case '[':
			if err := appendToken(Token{Type: TokenLBracket, Value: "[", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		case ']':
			if err := appendToken(Token{Type: TokenRBracket, Value: "]", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		case '{':
			if err := appendToken(Token{Type: TokenLBrace, Value: "{", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		case '}':
			if err := appendToken(Token{Type: TokenRBrace, Value: "}", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		case ',':
			if err := appendToken(Token{Type: TokenComma, Value: ",", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		case ':':
			if err := appendToken(Token{Type: TokenColon, Value: ":", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		case ';':
			if err := appendToken(Token{Type: TokenSemicolon, Value: ";", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		case '=':
			if err := appendToken(Token{Type: TokenEquals, Value: "=", Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor += size
			column++
			continue
		}

		// Handle strings
		if r == '"' {
			value, newCursor, newColumn, err := parseString(input, cursor, line, column)
			if err != nil {
				return nil, err
			}
			if err := appendToken(Token{Type: TokenString, Value: value, Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor = newCursor
			column = newColumn
			continue
		}

		// Handle numbers (JSON-style)
		if r == '-' || (r >= '0' && r <= '9') {
			value, newCursor, newColumn, ok := parseNumberJSON(input, cursor, column)
			if !ok {
				return nil, &SyntaxError{msg: "invalid number", Offset: int64(cursor)}
			}
			if err := appendToken(Token{Type: TokenNumber, Value: value, Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor = newCursor
			column = newColumn
			continue
		}

		// Handle identifiers and keywords
		if unicode.IsLetter(r) || r == '_' {
			value, newCursor, newColumn := parseIdentifierUTF8(input, cursor, column)
			tokenType := getKeywordType(value)
			if err := appendToken(Token{Type: tokenType, Value: value, Line: line, Column: column}); err != nil {
				return nil, err
			}
			cursor = newCursor
			column = newColumn
			continue
		}

		return nil, &SyntaxError{msg: fmt.Sprintf("Unexpected character '%c' at %d:%d", r, line, column), Offset: int64(cursor)}
	}

	if err := appendToken(Token{Type: TokenEOF, Value: "", Line: line, Column: column}); err != nil {
		return nil, err
	}
	return tokens, nil
}

// parseString parses a quoted string literal starting at the given cursor position.
func parseString(input string, cursor, line, column int) (string, int, int, error) {
	var value strings.Builder

	// Consume opening quote
	r, size := utf8.DecodeRuneInString(input[cursor:])
	if r != '"' {
		return "", 0, 0, &SyntaxError{msg: "expected string", Offset: int64(cursor)}
	}
	cursor += size
	column++

	closed := false
	for cursor < len(input) {
		r, size := utf8.DecodeRuneInString(input[cursor:])
		if r == utf8.RuneError && size == 1 {
			return "", 0, 0, &SyntaxError{msg: "invalid UTF-8", Offset: int64(cursor)}
		}
		if r == '"' {
			cursor += size
			column++
			closed = true
			break
		}
		if r == '\\' {
			// Backslash
			cursor += size
			column++
			if cursor >= len(input) {
				return "", 0, 0, &SyntaxError{msg: fmt.Sprintf("Unexpected end of input in string at %d:%d", line, column), Offset: int64(cursor)}
			}
			r2, s2 := utf8.DecodeRuneInString(input[cursor:])
			if r2 == utf8.RuneError && s2 == 1 {
				return "", 0, 0, &SyntaxError{msg: "invalid UTF-8", Offset: int64(cursor)}
			}
			cursor += s2
			column++
			switch r2 {
			case '"', '\\', '/':
				value.WriteRune(r2)
			case 'b':
				value.WriteByte('\b')
			case 'f':
				value.WriteByte('\f')
			case 'n':
				value.WriteByte('\n')
			case 'r':
				value.WriteByte('\r')
			case 't':
				value.WriteByte('\t')
			case 'u':
				// \uXXXX (optionally surrogate pairs)
				if cursor+4 > len(input) {
					return "", 0, 0, &SyntaxError{msg: "invalid unicode escape", Offset: int64(cursor)}
				}
				hex := input[cursor : cursor+4]
				if !isValidHex(hex) {
					return "", 0, 0, &SyntaxError{msg: "invalid unicode escape", Offset: int64(cursor)}
				}
				cp, err := strconv.ParseInt(hex, 16, 32)
				if err != nil {
					return "", 0, 0, &SyntaxError{msg: "invalid unicode escape", Offset: int64(cursor)}
				}
				cursor += 4
				column += 4
				runeVal := rune(cp)

				// Handle surrogate pairs. Unpaired surrogates are invalid.
				if utf16.IsSurrogate(runeVal) {
					// Must be a high surrogate followed by a low surrogate.
					if runeVal < 0xD800 || runeVal > 0xDBFF {
						return "", 0, 0, &SyntaxError{msg: "invalid unicode escape", Offset: int64(cursor)}
					}
					if !(cursor+6 <= len(input) && input[cursor] == '\\' && input[cursor+1] == 'u') {
						return "", 0, 0, &SyntaxError{msg: "invalid unicode escape", Offset: int64(cursor)}
					}
					hex2 := input[cursor+2 : cursor+6]
					if !isValidHex(hex2) {
						return "", 0, 0, &SyntaxError{msg: "invalid unicode escape", Offset: int64(cursor)}
					}
					cp2, err2 := strconv.ParseInt(hex2, 16, 32)
					if err2 != nil {
						return "", 0, 0, &SyntaxError{msg: "invalid unicode escape", Offset: int64(cursor)}
					}
					r2v := rune(cp2)
					if r2v < 0xDC00 || r2v > 0xDFFF {
						return "", 0, 0, &SyntaxError{msg: "invalid unicode escape", Offset: int64(cursor)}
					}
					runeVal = utf16.DecodeRune(runeVal, r2v)
					// consume \\uXXXX
					cursor += 6
					column += 6
				}
				value.WriteRune(runeVal)
			default:
				// Non-standard escapes are kept as-is
				value.WriteRune(r2)
			}
			continue
		}

		// Regular rune
		value.WriteRune(r)
		cursor += size
		column++
	}

	if !closed {
		return "", 0, 0, &SyntaxError{msg: "unterminated string", Offset: int64(cursor)}
	}
	return value.String(), cursor, column, nil
}

// parseNumberJSON scans a JSON-compatible number literal.
// Returns ok=false if the prefix does not match the JSON number grammar.
func parseNumberJSON(input string, cursor, column int) (string, int, int, bool) {
	start := cursor
	i := cursor

	if i < len(input) && input[i] == '-' {
		i++
	}
	if i >= len(input) {
		return "", cursor, column, false
	}

	// int
	if input[i] == '0' {
		i++
	} else if input[i] >= '1' && input[i] <= '9' {
		i++
		for i < len(input) && input[i] >= '0' && input[i] <= '9' {
			i++
		}
	} else {
		return "", cursor, column, false
	}

	// frac
	if i < len(input) && input[i] == '.' {
		i++
		if i >= len(input) || input[i] < '0' || input[i] > '9' {
			return "", cursor, column, false
		}
		for i < len(input) && input[i] >= '0' && input[i] <= '9' {
			i++
		}
	}

	// exp
	if i < len(input) && (input[i] == 'e' || input[i] == 'E') {
		i++
		if i < len(input) && (input[i] == '+' || input[i] == '-') {
			i++
		}
		if i >= len(input) || input[i] < '0' || input[i] > '9' {
			return "", cursor, column, false
		}
		for i < len(input) && input[i] >= '0' && input[i] <= '9' {
			i++
		}
	}

	// Column counts ASCII runes in the number.
	newColumn := column + (i - start)
	return input[start:i], i, newColumn, true
}

// parseIdentifierUTF8 parses an identifier starting at the given cursor position.
// Identifiers support Unicode letters/digits and underscore.
func parseIdentifierUTF8(input string, cursor, column int) (string, int, int) {
	start := cursor
	i := cursor
	col := column
	first := true

	for i < len(input) {
		r, size := utf8.DecodeRuneInString(input[i:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		ok := false
		if first {
			ok = unicode.IsLetter(r) || r == '_'
			first = false
		} else {
			ok = unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r) || r == '_'
		}
		if !ok {
			break
		}
		i += size
		col++
	}

	return input[start:i], i, col
}

// getKeywordType returns the appropriate token type for a keyword, or TokenIdentifier for non-keywords.
func getKeywordType(value string) TokenType {
	switch value {
	case "class":
		return TokenClass
	case "true":
		return TokenTrue
	case "false":
		return TokenFalse
	case "null":
		return TokenNull
	default:
		return TokenIdentifier
	}
}

// isValidHex checks if a string contains exactly 4 hexadecimal characters.
func isValidHex(s string) bool {
	if len(s) != 4 {
		return false
	}
	for _, char := range s {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	return true
}
