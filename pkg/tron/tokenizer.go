package tron

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
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
	cursor := 0
	line := 1
	column := 1

	for cursor < len(input) {
		char := rune(input[cursor])

		// Handle whitespace (except newlines)
		if char == ' ' || char == '\t' || char == '\r' {
			cursor++
			column++
			continue
		}

		// Handle newlines
		if char == '\n' {
			tokens = append(tokens, Token{
				Type:   TokenNewline,
				Value:  "\n",
				Line:   line,
				Column: column,
			})
			cursor++
			line++
			column = 1
			continue
		}

		// Handle comments
		if char == '#' {
			for cursor < len(input) && input[cursor] != '\n' {
				cursor++
			}
			// Don't consume newline here, let the next iteration handle it
			continue
		}

		// Handle single-character tokens
		switch char {
		case '(':
			tokens = append(tokens, Token{TokenLParen, "(", line, column})
			cursor++
			column++
			continue
		case ')':
			tokens = append(tokens, Token{TokenRParen, ")", line, column})
			cursor++
			column++
			continue
		case '[':
			tokens = append(tokens, Token{TokenLBracket, "[", line, column})
			cursor++
			column++
			continue
		case ']':
			tokens = append(tokens, Token{TokenRBracket, "]", line, column})
			cursor++
			column++
			continue
		case '{':
			tokens = append(tokens, Token{TokenLBrace, "{", line, column})
			cursor++
			column++
			continue
		case '}':
			tokens = append(tokens, Token{TokenRBrace, "}", line, column})
			cursor++
			column++
			continue
		case ',':
			tokens = append(tokens, Token{TokenComma, ",", line, column})
			cursor++
			column++
			continue
		case ':':
			tokens = append(tokens, Token{TokenColon, ":", line, column})
			cursor++
			column++
			continue
		case ';':
			tokens = append(tokens, Token{TokenSemicolon, ";", line, column})
			cursor++
			column++
			continue
		case '=':
			tokens = append(tokens, Token{TokenEquals, "=", line, column})
			cursor++
			column++
			continue
		}

		// Handle strings
		if char == '"' {
			value, newCursor, newColumn, err := parseString(input, cursor, line, column)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, Token{
				Type:   TokenString,
				Value:  value,
				Line:   line,
				Column: column,
			})
			cursor = newCursor
			column = newColumn
			continue
		}

		// Handle numbers
		if char == '-' || unicode.IsDigit(char) {
			value, newCursor, newColumn := parseNumber(input, cursor, column)
			tokens = append(tokens, Token{
				Type:   TokenNumber,
				Value:  value,
				Line:   line,
				Column: column,
			})
			cursor = newCursor
			column = newColumn
			continue
		}

		// Handle identifiers and keywords
		if unicode.IsLetter(char) || char == '_' {
			value, newCursor, newColumn := parseIdentifier(input, cursor, column)
			tokenType := getKeywordType(value)
			tokens = append(tokens, Token{
				Type:   tokenType,
				Value:  value,
				Line:   line,
				Column: column,
			})
			cursor = newCursor
			column = newColumn
			continue
		}

		return nil, &SyntaxError{
			msg:    fmt.Sprintf("Unexpected character '%c' at %d:%d", char, line, column),
			Offset: int64(cursor),
		}
	}

	tokens = append(tokens, Token{
		Type:   TokenEOF,
		Value:  "",
		Line:   line,
		Column: column,
	})

	return tokens, nil
}

// parseString parses a quoted string literal starting at the given cursor position.
func parseString(input string, cursor, line, column int) (string, int, int, error) {
	var value strings.Builder
	cursor++ // skip opening quote
	column++

	for cursor < len(input) {
		char := input[cursor]
		if char == '"' {
			cursor++
			column++
			break
		}
		if char == '\\' {
			cursor++
			column++
			if cursor >= len(input) {
				return "", 0, 0, &SyntaxError{
					msg:    fmt.Sprintf("Unexpected end of input in string at %d:%d", line, column),
					Offset: int64(cursor),
				}
			}
			escaped := input[cursor]
			switch escaped {
			case '"':
				value.WriteByte('"')
			case '\\':
				value.WriteByte('\\')
			case '/':
				value.WriteByte('/')
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
				// Handle unicode escape sequences
				if cursor+4 >= len(input) {
					value.WriteByte('u')
				} else {
					hex := input[cursor+1 : cursor+5]
					if isValidHex(hex) {
						if codePoint, err := strconv.ParseInt(hex, 16, 32); err == nil {
							value.WriteRune(rune(codePoint))
							cursor += 4
							column += 4
						} else {
							value.WriteByte('u')
						}
					} else {
						value.WriteByte('u')
					}
				}
			default:
				value.WriteByte(escaped)
			}
			cursor++
			column++
		} else {
			value.WriteByte(char)
			cursor++
			column++
		}
	}

	return value.String(), cursor, column, nil
}

// parseNumber parses a numeric literal starting at the given cursor position.
func parseNumber(input string, cursor, column int) (string, int, int) {
	start := cursor

	// Handle negative sign
	if cursor < len(input) && input[cursor] == '-' {
		cursor++
		column++
	}

	// Parse digits and decimal point
	for cursor < len(input) {
		char := input[cursor]
		if unicode.IsDigit(rune(char)) || char == '.' || char == 'e' || char == 'E' || char == '+' || char == '-' {
			cursor++
			column++
		} else {
			break
		}
	}

	return input[start:cursor], cursor, column
}

// parseIdentifier parses an identifier starting at the given cursor position.
func parseIdentifier(input string, cursor, column int) (string, int, int) {
	start := cursor

	for cursor < len(input) {
		char := rune(input[cursor])
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '_' {
			cursor++
			column++
		} else {
			break
		}
	}

	return input[start:cursor], cursor, column
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
