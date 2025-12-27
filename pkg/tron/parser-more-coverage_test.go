package tron

import "testing"

func TestParser_ClassDefinition_NewlineRequirement(t *testing.T) {
	var v interface{}
	// Missing newline/EOF after class definition (extra token on same line).
	input := "class A: a {}"
	if err := Unmarshal([]byte(input), &v); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParser_ClassInstantiation_Errors(t *testing.T) {
	var v interface{}

	// Undefined class
	if err := Unmarshal([]byte("A(1)"), &v); err == nil {
		t.Fatalf("expected error")
	}

	// Wrong arg count
	input := "class A: a,b\n\nA(1)"
	if err := Unmarshal([]byte(input), &v); err == nil {
		t.Fatalf("expected error")
	}

	// Empty arg list but class expects args
	input = "class A: a\n\nA()"
	if err := Unmarshal([]byte(input), &v); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParser_Object_Errors(t *testing.T) {
	var v interface{}

	// Missing key
	if err := Unmarshal([]byte("{:1}"), &v); err == nil {
		t.Fatalf("expected error")
	}

	// Missing closing brace
	if err := Unmarshal([]byte("{a:1"), &v); err == nil {
		t.Fatalf("expected error")
	}
}
