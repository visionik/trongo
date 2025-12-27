package tron

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string][]string
		wantErr bool
	}{
		{
			name:  "single class with two properties",
			input: "class A: name,age\n\n",
			want:  map[string][]string{"A": {"name", "age"}},
		},
		{
			name:  "single class with one property",
			input: "class B: id\n\n",
			want:  map[string][]string{"B": {"id"}},
		},
		{
			name:  "multiple classes",
			input: "class A: x,y\nclass B: name\n\n",
			want: map[string][]string{
				"A": {"x", "y"},
				"B": {"name"},
			},
		},
		{
			name:  "quoted properties",
			input: `class A: "first name","last-name"` + "\n\n",
			want:  map[string][]string{"A": {"first name", "last-name"}},
		},
		{
			name:  "mixed quoted and unquoted",
			input: `class A: name,"user-id",age` + "\n\n",
			want:  map[string][]string{"A": {"name", "user-id", "age"}},
		},
		{
			name:  "empty header",
			input: "\n",
			want:  map[string][]string{},
		},
		{
			name:    "missing colon",
			input:   "class A name,age\n",
			wantErr: true,
		},
		{
			name:    "missing class name",
			input:   "class : name,age\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenize(tt.input)
			require.NoError(t, err)

			p := newParser(tokens)
			err = p.parseHeader()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, p.classes)
			}
		})
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		header  string
		want    interface{}
		wantErr bool
	}{
		{
			name:  "boolean true",
			input: "true",
			want:  true,
		},
		{
			name:  "boolean false",
			input: "false",
			want:  false,
		},
		{
			name:  "null",
			input: "null",
			want:  nil,
		},
		{
			name:  "positive integer",
			input: "42",
			want:  float64(42),
		},
		{
			name:  "negative integer",
			input: "-17",
			want:  float64(-17),
		},
		{
			name:  "float",
			input: "3.14",
			want:  float64(3.14),
		},
		{
			name:  "string",
			input: `"hello"`,
			want:  "hello",
		},
		{
			name:  "empty string",
			input: `""`,
			want:  "",
		},
		{
			name:  "string with escapes",
			input: `"hello\nworld"`,
			want:  "hello\nworld",
		},
		{
			name:  "empty array",
			input: "[]",
			want:  []interface{}{},
		},
		{
			name:  "array of numbers",
			input: "[1,2,3]",
			want:  []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:  "array of strings",
			input: `["a","b","c"]`,
			want:  []interface{}{"a", "b", "c"},
		},
		{
			name:  "array of mixed types",
			input: `[1,"two",true,null]`,
			want:  []interface{}{float64(1), "two", true, nil},
		},
		{
			name:  "nested arrays",
			input: "[[1,2],[3,4]]",
			want:  []interface{}{[]interface{}{float64(1), float64(2)}, []interface{}{float64(3), float64(4)}},
		},
		{
			name:  "empty object",
			input: "{}",
			want:  map[string]interface{}{},
		},
		{
			name:  "simple object",
			input: `{"name":"Alice","age":30}`,
			want: map[string]interface{}{
				"name": "Alice",
				"age":  float64(30),
			},
		},
		{
			name:  "object with identifier keys",
			input: `{name:"Alice",age:30}`,
			want: map[string]interface{}{
				"name": "Alice",
				"age":  float64(30),
			},
		},
		{
			name:  "nested object",
			input: `{"user":{"name":"Bob"},"active":true}`,
			want: map[string]interface{}{
				"user":   map[string]interface{}{"name": "Bob"},
				"active": true,
			},
		},
		{
			name:   "class instantiation",
			header: "class A: name,age\n\n",
			input:  `A("Bob",25)`,
			want: map[string]interface{}{
				"name": "Bob",
				"age":  float64(25),
			},
		},
		{
			name:   "class instantiation with various types",
			header: "class B: name,active,score\n\n",
			input:  `B("Alice",true,98.5)`,
			want: map[string]interface{}{
				"name":   "Alice",
				"active": true,
				"score":  float64(98.5),
			},
		},
		{
			name:   "empty class instantiation",
			header: "class C: \n\n",
			input:  `C()`,
			want:   map[string]interface{}{},
		},
		{
			name:    "undefined class",
			input:   `A("Bob",25)`,
			wantErr: true,
		},
		{
			name:    "class argument count mismatch",
			header:  "class A: name,age\n\n",
			input:   `A("Bob")`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullInput := tt.header + tt.input
			tokens, err := tokenize(fullInput)
			require.NoError(t, err)

			p := newParser(tokens)
			got, err := p.parse()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestParseComplexExamples(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  interface{}
	}{
		{
			name: "array of class instantiations",
			input: `class A: Name,Age

[A("Alice",30),A("Bob",25)]`,
			want: []interface{}{
				map[string]interface{}{"Name": "Alice", "Age": float64(30)},
				map[string]interface{}{"Name": "Bob", "Age": float64(25)},
			},
		},
		{
			name: "nested structures with classes",
			input: `class Person: name,age
class Team: leader,members

Team("Alice",[Person("Bob",25),Person("Charlie",30)])`,
			want: map[string]interface{}{
				"leader": "Alice",
				"members": []interface{}{
					map[string]interface{}{"name": "Bob", "age": float64(25)},
					map[string]interface{}{"name": "Charlie", "age": float64(30)},
				},
			},
		},
		{
			name: "mixed objects and classes",
			input: `class A: x,y

{"point":A(1,2),"name":"origin"}`,
			want: map[string]interface{}{
				"point": map[string]interface{}{"x": float64(1), "y": float64(2)},
				"name":  "origin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenize(tt.input)
			require.NoError(t, err)

			p := newParser(tokens)
			got, err := p.parse()

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParserTokenNavigation(t *testing.T) {
	input := "class A: x,y\n\ntrue"
	tokens, err := tokenize(input)
	require.NoError(t, err)

	p := newParser(tokens)

	// Test current()
	assert.Equal(t, TokenClass, p.current().Type)

	// Test advance()
	tok := p.advance()
	assert.Equal(t, TokenClass, tok.Type)
	assert.Equal(t, TokenIdentifier, p.current().Type)

	// Test peek()
	assert.Equal(t, TokenColon, p.peek(1).Type)
	assert.Equal(t, TokenIdentifier, p.current().Type) // Should not advance

	// Test skipNewlines()
	p.pos = 6 // Position at newline
	p.skipNewlines()
	assert.Equal(t, TokenTrue, p.current().Type)
}
