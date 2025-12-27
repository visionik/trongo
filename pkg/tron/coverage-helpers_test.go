package tron

import (
	"math"
	"reflect"
	"testing"
)

func TestDecodeNumber_DeprecatedPath(t *testing.T) {
	d := &decoder{}

	{
		var v int64
		dst := reflect.ValueOf(&v).Elem()
		if err := d.decodeNumber(123, dst); err != nil {
			t.Fatalf("decodeNumber int64: %v", err)
		}
		if v != 123 {
			t.Fatalf("expected 123, got %d", v)
		}
	}

	{
		var v float64
		dst := reflect.ValueOf(&v).Elem()
		if err := d.decodeNumber(1.25, dst); err != nil {
			t.Fatalf("decodeNumber float64: %v", err)
		}
		if v != 1.25 {
			t.Fatalf("expected 1.25, got %v", v)
		}
	}

	{
		var v interface{}
		dst := reflect.ValueOf(&v).Elem()
		if err := d.decodeNumber(2.5, dst); err != nil {
			t.Fatalf("decodeNumber interface{}: %v", err)
		}
		if f, ok := v.(float64); !ok || f != 2.5 {
			t.Fatalf("expected float64(2.5), got %#v", v)
		}
	}
}

func TestMinMaxHelpers(t *testing.T) {
	intCases := []struct {
		kind reflect.Kind
		min  int64
		max  int64
	}{
		{reflect.Int8, math.MinInt8, math.MaxInt8},
		{reflect.Int16, math.MinInt16, math.MaxInt16},
		{reflect.Int32, math.MinInt32, math.MaxInt32},
		{reflect.Int64, math.MinInt64, math.MaxInt64},
		{reflect.Int, math.MinInt64, math.MaxInt64},
	}

	for _, tc := range intCases {
		t.Run(tc.kind.String(), func(t *testing.T) {
			typ := reflect.TypeOf(int64(0))
			switch tc.kind {
			case reflect.Int8:
				typ = reflect.TypeOf(int8(0))
			case reflect.Int16:
				typ = reflect.TypeOf(int16(0))
			case reflect.Int32:
				typ = reflect.TypeOf(int32(0))
			case reflect.Int64:
				typ = reflect.TypeOf(int64(0))
			case reflect.Int:
				typ = reflect.TypeOf(int(0))
			}
			if got := minInt(typ); got != tc.min {
				t.Fatalf("minInt: expected %d, got %d", tc.min, got)
			}
			if got := maxInt(typ); got != tc.max {
				t.Fatalf("maxInt: expected %d, got %d", tc.max, got)
			}
		})
	}

	uintCases := []struct {
		kind reflect.Kind
		max  uint64
	}{
		{reflect.Uint8, math.MaxUint8},
		{reflect.Uint16, math.MaxUint16},
		{reflect.Uint32, math.MaxUint32},
		{reflect.Uint64, math.MaxUint64},
		{reflect.Uint, math.MaxUint64},
	}

	for _, tc := range uintCases {
		t.Run(tc.kind.String(), func(t *testing.T) {
			typ := reflect.TypeOf(uint64(0))
			switch tc.kind {
			case reflect.Uint8:
				typ = reflect.TypeOf(uint8(0))
			case reflect.Uint16:
				typ = reflect.TypeOf(uint16(0))
			case reflect.Uint32:
				typ = reflect.TypeOf(uint32(0))
			case reflect.Uint64:
				typ = reflect.TypeOf(uint64(0))
			case reflect.Uint:
				typ = reflect.TypeOf(uint(0))
			}
			if got := maxUint(typ); got != tc.max {
				t.Fatalf("maxUint: expected %d, got %d", tc.max, got)
			}
		})
	}
}
