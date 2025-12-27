package tron

import (
	"encoding/json"
	"reflect"
	"testing"
)

// normalizeJSONValue makes sure values decoded into interface{} can be compared
// deterministically between JSON and TRON.
//
// Both json.Unmarshal and tron.Unmarshal into interface{} should yield the same
// JSON-like shapes:
//   - bool
//   - float64
//   - string
//   - []interface{}
//   - map[string]interface{}
//   - nil
func normalizeJSONValue(v interface{}) interface{} {
	switch vv := v.(type) {
	case []interface{}:
		out := make([]interface{}, len(vv))
		for i := range vv {
			out[i] = normalizeJSONValue(vv[i])
		}
		return out
	case map[string]interface{}:
		out := make(map[string]interface{}, len(vv))
		for k, val := range vv {
			out[k] = normalizeJSONValue(val)
		}
		return out
	default:
		return v
	}
}

func TestJSONCompatibilityFixtures(t *testing.T) {
	fixtures := []string{
		`null`,
		`true`,
		`false`,
		`0`,
		`-0`,
		`1`,
		`-1`,
		`123.456`,
		`1e10`,
		`1.5e-5`,
		`""`,
		`"simple"`,
		`"with\\nnewline"`,
		`"with\\ttab"`,
		`"with\\\"quote\\\""`,
		`"unicode: \u00f1 \u00e9 \u00fc"`,
		`[]`,
		`[1,2,3]`,
		`[1,"two",true,null,3.14]`,
		`{}`,
		`{"name":"Alice","age":30}`,
		`{"nested":{"value":42,"arr":[1,2,3]}}`,
		`[{"a":1},{"a":2}]`,
		`{"a":1,"b":2,"c":{"d":4}}`,
		`{"emptyArr":[],"emptyObj":{},"emptyStr":""}`,
		`{"esc":"line1\\nline2\\r\\nwindows"}`,
		// Large integers (JSON decodes to float64; precision may be lossy but should match TRON behavior)
		`9223372036854775807`,
	}

	for _, input := range fixtures {
		t.Run(input, func(t *testing.T) {
			var jsonVal interface{}
			if err := json.Unmarshal([]byte(input), &jsonVal); err != nil {
				t.Fatalf("fixture is not valid JSON: %v", err)
			}

			var tronVal interface{}
			if err := Unmarshal([]byte(input), &tronVal); err != nil {
				t.Fatalf("TRON failed to unmarshal valid JSON: %v", err)
			}

			jsonNorm := normalizeJSONValue(jsonVal)
			tronNorm := normalizeJSONValue(tronVal)
			if !reflect.DeepEqual(jsonNorm, tronNorm) {
				t.Fatalf("JSON and TRON results differ\ninput: %s\njson: %#v\ntron: %#v", input, jsonNorm, tronNorm)
			}
		})
	}
}
