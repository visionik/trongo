package tron

import (
	"errors"
	"reflect"
	"testing"
)

type testTextKeyErr struct{ S string }

func (k testTextKeyErr) MarshalText() ([]byte, error) {
	return nil, errors.New("boom")
}

func TestIsValidIdentifier(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"a", true},
		{"_", true},
		{"a1", true},
		{"a_b", true},
		{"名", true},
		{"名2", true},
		{"a\u0001", false}, // control
		{"1abc", false},    // starts with digit
		{"\u001fa", false}, // starts with control
		{"\u0301a", false}, // combining mark as first (U+0301)
		{"a\u0301", true},  // combining mark after letter
		{"a-1", false},     // hyphen not allowed
	}

	for _, tc := range cases {
		if got := isValidIdentifier(tc.in); got != tc.want {
			t.Fatalf("isValidIdentifier(%q)=%v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestIsEmptyValue_AllKinds(t *testing.T) {
	type s struct{ A int }
	var (
		nilIface interface{}
		nilPtr   *int
	)

	cases := []struct {
		name string
		v    interface{}
		want bool
	}{
		{"empty-string", "", true},
		{"nonempty-string", "x", false},
		{"empty-slice", []int{}, true},
		{"nonempty-slice", []int{1}, false},
		{"nil-slice", ([]int)(nil), true},
		{"empty-map", map[string]int{}, true},
		{"nonempty-map", map[string]int{"a": 1}, false},
		{"nil-map", (map[string]int)(nil), true},
		{"bool-false", false, true},
		{"bool-true", true, false},
		{"int-zero", int(0), true},
		{"int-nonzero", int(1), false},
		{"uint-zero", uint(0), true},
		{"uint-nonzero", uint(1), false},
		{"float-zero", float64(0), true},
		{"float-nonzero", float64(0.5), false},
		{"nil-interface", nilIface, true},
		{"nil-ptr", nilPtr, true},
		{"struct", s{}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rv := reflect.ValueOf(tc.v)
			// Special-case typed nils: create a typed interface value with nil contents.
			if tc.name == "nil-interface" {
				rv = reflect.New(reflect.TypeOf((*interface{})(nil)).Elem()).Elem()
			}
			if tc.name == "nil-ptr" {
				rv = reflect.ValueOf((*int)(nil))
			}
			if got := isEmptyValue(rv); got != tc.want {
				t.Fatalf("isEmptyValue(%s)=%v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	if contains([]string{"a", "b"}, "b") != true {
		t.Fatalf("expected true")
	}
	if contains([]string{"a", "b"}, "c") != false {
		t.Fatalf("expected false")
	}
}

func TestSerializeMapKey_ErrorBranches(t *testing.T) {
	e := &encoder{}

	// TextMarshaler error
	{
		_, err := e.serializeMapKey(reflect.ValueOf(testTextKeyErr{S: "x"}))
		if err == nil {
			t.Fatalf("expected error")
		}
	}

	// Unsupported key type
	{
		_, err := e.serializeMapKey(reflect.ValueOf([]int{1}))
		if err == nil {
			t.Fatalf("expected error")
		}
	}
}
