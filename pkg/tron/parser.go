package tron

import (
	"fmt"
	"strconv"
)

// parser parses TRON format into Go native types.
type parser struct {
	tokens  []Token
	pos     int
	classes map[string][]string // className -> propertyNames
}

// newParser creates a new parser from tokens.
func newParser(tokens []Token) *parser {
	return &parser{
		tokens:  tokens,
		pos:     0,
		classes: make(map[string][]string),
	}
}

// current returns the current token without advancing.
func (p *parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

// advance moves to the next token and returns the current token.
func (p *parser) advance() Token {
	tok := p.current()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

// peek looks ahead n tokens without advancing.
func (p *parser) peek(n int) Token {
	pos := p.pos + n
	if pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[pos]
}

// expect consumes a token of the specified type or returns an error.
func (p *parser) expect(tokenType TokenType) (Token, error) {
	tok := p.current()
	if tok.Type != tokenType {
		return tok, p.syntaxError(fmt.Sprintf("expected %s, got %s", tokenType, tok.Type))
	}
	p.advance()
	return tok, nil
}

// skipNewlines skips all consecutive newline tokens.
func (p *parser) skipNewlines() {
	for p.current().Type == TokenNewline {
		p.advance()
	}
}

// syntaxError creates a SyntaxError with the current position.
func (p *parser) syntaxError(msg string) error {
	return &SyntaxError{
		msg:    msg,
		Offset: int64(p.pos),
	}
}

// parse is the main entry point that parses TRON format.
func (p *parser) parse() (interface{}, error) {
	// Parse header (class definitions)
	if err := p.parseHeader(); err != nil {
		return nil, err
	}

	// Skip blank lines between header and data
	p.skipNewlines()

	// Parse data
	if p.current().Type == TokenEOF {
		return nil, nil
	}

	return p.parseValue()
}

// parseHeader parses all class definitions from the header.
func (p *parser) parseHeader() error {
	p.skipNewlines()

	for p.current().Type == TokenClass {
		if err := p.parseClassDefinition(); err != nil {
			return err
		}
		p.skipNewlines()
	}

	return nil
}

// parseClassDefinition parses a single class definition: class A: prop1,prop2
func (p *parser) parseClassDefinition() error {
	// Consume "class" keyword
	if _, err := p.expect(TokenClass); err != nil {
		return err
	}

	// Get class name
	className, err := p.expect(TokenIdentifier)
	if err != nil {
		return p.syntaxError("expected class name")
	}

	// Consume colon
	if _, err := p.expect(TokenColon); err != nil {
		return err
	}

	// Parse property list
	properties := []string{}
	for {
		prop := p.current()
		if prop.Type == TokenIdentifier {
			properties = append(properties, prop.Value)
			p.advance()
		} else if prop.Type == TokenString {
			properties = append(properties, prop.Value)
			p.advance()
		} else {
			break
		}

		// Check for comma
		if p.current().Type == TokenComma {
			p.advance()
		} else {
			break
		}
	}

	// Store class definition
	p.classes[className.Value] = properties

	// Expect newline or EOF after class definition
	tok := p.current()
	if tok.Type != TokenNewline && tok.Type != TokenEOF {
		return p.syntaxError("expected newline after class definition")
	}

	return nil
}

// parseValue is the main recursive parser for all TRON values.
func (p *parser) parseValue() (interface{}, error) {
	tok := p.current()

	switch tok.Type {
	case TokenTrue:
		p.advance()
		return true, nil

	case TokenFalse:
		p.advance()
		return false, nil

	case TokenNull:
		p.advance()
		return nil, nil

	case TokenNumber:
		p.advance()
		return p.parseNumberValue(tok.Value)

	case TokenString:
		p.advance()
		return tok.Value, nil

	case TokenLBracket:
		return p.parseArray()

	case TokenLBrace:
		return p.parseObject()

	case TokenIdentifier:
		// Could be class instantiation A(...)
		return p.parseClassInstantiation()

	default:
		return nil, p.syntaxError(fmt.Sprintf("unexpected token: %s", tok.Type))
	}
}

// parseNumberValue parses a number string into float64.
func (p *parser) parseNumberValue(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, p.syntaxError(fmt.Sprintf("invalid number: %s", s))
	}
	return f, nil
}

// parseArray parses an array: [item1,item2,...]
func (p *parser) parseArray() ([]interface{}, error) {
	if _, err := p.expect(TokenLBracket); err != nil {
		return nil, err
	}

	items := []interface{}{}

	// Handle empty array
	if p.current().Type == TokenRBracket {
		p.advance()
		return items, nil
	}

	// Parse array elements
	for {
		item, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		items = append(items, item)

		// Check for comma
		if p.current().Type != TokenComma {
			break
		}
		p.advance() // consume comma
	}

	// Expect closing bracket
	if _, err := p.expect(TokenRBracket); err != nil {
		return nil, err
	}

	return items, nil
}

// parseObject parses an object: {"key":value,"key2":value2}
func (p *parser) parseObject() (map[string]interface{}, error) {
	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	obj := make(map[string]interface{})

	// Handle empty object
	if p.current().Type == TokenRBrace {
		p.advance()
		return obj, nil
	}

	// Parse key-value pairs
	for {
		// Parse key (must be string or identifier)
		key := ""
		tok := p.current()
		if tok.Type == TokenString {
			key = tok.Value
			p.advance()
		} else if tok.Type == TokenIdentifier {
			key = tok.Value
			p.advance()
		} else {
			return nil, p.syntaxError("expected object key")
		}

		// Expect colon
		if _, err := p.expect(TokenColon); err != nil {
			return nil, err
		}

		// Parse value
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		obj[key] = value

		// Check for comma
		if p.current().Type != TokenComma {
			break
		}
		p.advance() // consume comma
	}

	// Expect closing brace
	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return obj, nil
}

// parseClassInstantiation parses class instantiation: A(arg1,arg2,...)
func (p *parser) parseClassInstantiation() (map[string]interface{}, error) {
	// Get class name
	className := p.current().Value
	p.advance()

	// Expect opening paren
	if _, err := p.expect(TokenLParen); err != nil {
		return nil, p.syntaxError("expected ( for class instantiation")
	}

	// Look up class definition
	properties, exists := p.classes[className]
	if !exists {
		return nil, p.syntaxError(fmt.Sprintf("undefined class: %s", className))
	}

	args := []interface{}{}

	// Handle empty argument list
	if p.current().Type == TokenRParen {
		p.advance()
		if len(properties) != 0 {
			return nil, p.syntaxError(fmt.Sprintf("class %s expects %d arguments, got 0", className, len(properties)))
		}
		return make(map[string]interface{}), nil
	}

	// Parse arguments
	for {
		arg, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)

		// Check for comma
		if p.current().Type != TokenComma {
			break
		}
		p.advance() // consume comma
	}

	// Expect closing paren
	if _, err := p.expect(TokenRParen); err != nil {
		return nil, err
	}

	// Validate argument count
	if len(args) != len(properties) {
		return nil, p.syntaxError(
			fmt.Sprintf("class %s expects %d arguments, got %d",
				className, len(properties), len(args)),
		)
	}

	// Convert to object using property names as keys
	obj := make(map[string]interface{})
	for i, prop := range properties {
		obj[prop] = args[i]
	}

	return obj, nil
}
