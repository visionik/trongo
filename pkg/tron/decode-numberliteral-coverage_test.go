package tron

import (
	"reflect"
	"testing"
)

type ifaceNum interface{ M() }

type ifaceNumImpl struct{}

func (ifaceNumImpl) M() {}

func TestDecodeNumberLiteral_MoreBranches(t *testing.T) {
	d := &decoder{}

	// int target: fractional should error
	{
		var x int64
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeNumberLiteral("1.5", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// int target: NaN should error (non-integral semantics)
	{
		var x int64
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeNumberLiteral("NaN", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// uint target: exponent form should error even if integral
	{
		var x uint64
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeNumberLiteral("1e3", dst); err == nil {
			t.Fatalf("expected error")
		}
	}

	// interface with methods: should not set
	{
		var x ifaceNum
		dst := reflect.ValueOf(&x).Elem()
		if err := d.decodeNumberLiteral("1", dst); err == nil {
			t.Fatalf("expected error")
		}
	}
}

func TestNormalizeInterfaceValue_NumberLiteralParseFailBranch(t *testing.T) {
	d := &decoder{}
	out := d.normalizeInterfaceValue(numberLiteral("not-a-number"))
	if s, ok := out.(string); !ok || s != "not-a-number" {
		t.Fatalf("expected string 'not-a-number', got %#v", out)
	}
}
