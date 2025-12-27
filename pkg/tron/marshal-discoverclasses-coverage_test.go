package tron

import (
	"strings"
	"testing"
)

func TestMarshal_DiscoverClasses_MapAndSliceTraversal(t *testing.T) {
	type child struct {
		X int `json:"x"`
	}
	type root struct {
		Arr   [2]child         `json:"arr"`
		Slice []*child         `json:"slice"`
		Map   map[string]child `json:"map"`
		Any   interface{}      `json:"any"`
		Nil   *child           `json:"nil"`
	}

	c1 := &child{X: 1}
	r := &root{
		Arr:   [2]child{{X: 2}, {X: 3}},
		Slice: []*child{c1, nil, c1},
		Map:   map[string]child{"a": {X: 4}, "b": {X: 5}},
		Any:   &child{X: 6},
		Nil:   nil,
	}

	out, err := Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	s := string(out)
	// Basic sanity: output includes object/map/array syntax.
	if !strings.Contains(s, "{") || !strings.Contains(s, "[") {
		t.Fatalf("unexpected output: %q", s)
	}
}
