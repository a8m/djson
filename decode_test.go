package djson

import (
	"encoding/json"
	"reflect"
	"testing"
)

type decodeTest struct {
	in       string
	err      error
	expected interface{}
}

var decodeTests = []decodeTest{
	// basic types
	{in: "null", expected: nil},
	{in: `true`, expected: true},
	{in: `false`, expected: false},
	{in: `5`, expected: 5.0},
	{in: `-5`, expected: -5.0},
	{in: `5.5`, expected: 5.5},
	{in: `"a\u1234"`, expected: "a\u1234"},
	{in: `"http:\/\/"`, expected: "http://"},
	{in: `"g-clef: \uD834\uDD1E"`, expected: "g-clef: \U0001D11E"},
	{in: `"invalid: \uD834x\uDD1E"`, expected: "invalid: \uFFFDx\uFFFD"},
	{in: `{"X": [1], "Y": 4}`, expected: map[string]interface{}{
		"X": []interface{}{1.0},
		"Y": 4.0,
	}},
	{in: `{"k1":1e-3,"k2":"s","k3":[1,2.0,3e-3],"k4":{"kk1":"s","kk2":2}}`, expected: map[string]interface{}{
		"k1": 1e-3,
		"k2": "s",
		"k3": []interface{}{1.0, 2.0, 3e-3},
		"k4": map[string]interface{}{
			"kk1": "s",
			"kk2": 2.0,
		},
	}},

	// raw values with whitespace
	{in: "\n true ", expected: true},
	{in: "\n false ", expected: false},
	{in: "\t 1 ", expected: float64(1)},
	{in: "\r 1.2 ", expected: 1.2},
	{in: "\t -5 \n", expected: float64(-5)},
	{in: "\t \"a\\u1234\" \n", expected: "a\u1234"},

	// syntax errors
	{in: `{"X": "foo", "Y"}`, err: &SyntaxError{"invalid character '}' after object key", 17}},
	{in: `[1, 2, 3+]`, err: &SyntaxError{"invalid character '+' after array element", 9}},
	{in: `{"X":12x}`, err: &SyntaxError{"invalid character 'x' after object key:value pair", 8}},

	// raw value errors
	{in: "\x01 42", err: &SyntaxError{"invalid character '\\x01' looking for beginning of value", 1}},
	{in: " 42 \x01", err: &SyntaxError{"invalid character '\\x01' after top-level value", 5}},
	{in: "\x01 true", err: &SyntaxError{"invalid character '\\x01' looking for beginning of value", 1}},
	{in: " false \x01", err: &SyntaxError{"invalid character '\\x01' after top-level value", 8}},
	{in: "\x01 1.2", err: &SyntaxError{"invalid character '\\x01' looking for beginning of value", 1}},
	{in: " 3.4 \x01", err: &SyntaxError{"invalid character '\\x01' after top-level value", 6}},
	{in: "\x01 \"string\"", err: &SyntaxError{"invalid character '\\x01' looking for beginning of value", 1}},
	{in: " \"string\" \x01", err: &SyntaxError{"invalid character '\\x01' after top-level value", 11}},

	// array tests
	{in: `[1, 2, 3]`, expected: []interface{}{1.0, 2.0, 3.0}},
	{in: `{"T":[1]}`, expected: map[string]interface{}{
		"T": []interface{}{1.0},
	}},
	{in: `{"T":null}`, expected: map[string]interface{}{
		"T": nil,
	}},

	// invalid UTF-8 is coerced to valid UTF-8.
	{in: "\"hello\xffworld\"", expected: "hello\ufffdworld"},
	{in: "\"hello\xc2\xc2world\"", expected: "hello\ufffd\ufffdworld"},
	{in: "\"hello\xc2\xffworld\"", expected: "hello\ufffd\ufffdworld"},
	{in: "\"hello\\ud800world\"", expected: "hello\ufffdworld"},
	{in: "\"hello\\ud800\\ud800world\"", expected: "hello\ufffd\ufffdworld"},
	{in: "\"hello\\ud800\\ud800world\"", expected: "hello\ufffd\ufffdworld"},
	{in: "\"hello\xed\xa0\x80\xed\xb0\x80world\"", expected: "hello\ufffd\ufffd\ufffd\ufffd\ufffd\ufffdworld"},
}

func TestDecode(t *testing.T) {
	for i, tt := range decodeTests {
		out, err := Decode([]byte(tt.in))
		if !reflect.DeepEqual(err, tt.err) {
			t.Errorf("#%d: %v, want %v", i, err, tt.err)
		}
		if out != nil {
			if !reflect.DeepEqual(out, tt.expected) {
				t.Errorf("#%d: %v, want %v", i, out, tt.expected)
			}
		}
	}
}

var allValueIndent = []byte(`{
	"Bool": true,
	"Int": 2,
	"Int8": 3,
	"Int16": 4,
	"Int32": 5,
	"Int64": 6,
	"Uint": 7,
	"Uint8": 8,
	"Uint16": 9,
	"Uint32": 10,
	"Uint64": 11,
	"Uintptr": 12,
	"Float32": 14.1,
	"Float64": 15.1,
	"bar": "foo",
	"bar2": "foo2",
	"IntStr": "42",
	"PBool": null,
	"PInt": null,
	"PInt8": null,
	"PInt16": null,
	"PInt32": null,
	"PInt64": null,
	"PUint": null,
	"PUint8": null,
	"PUint16": null,
	"PUint32": null,
	"PUint64": null,
	"PUintptr": null,
	"PFloat32": null,
	"PFloat64": null,
	"String": "16",
	"PString": null,
	"Map": {
		"17": {
			"Tag": "tag17"
		},
		"18": {
			"Tag": "tag18"
		}
	},
	"MapP": {
		"19": {
			"Tag": "tag19"
		},
		"20": null
	},
	"PMap": null,
	"PMapP": null,
	"EmptyMap": {},
	"NilMap": null,
	"Slice": [
		{
			"Tag": "tag20"
		},
		{
			"Tag": "tag21"
		}
	],
	"SliceP": [
		{
			"Tag": "tag22"
		},
		null,
		{
			"Tag": "tag23"
		}
	],
	"PSlice": null,
	"PSliceP": null,
	"EmptySlice": [],
	"NilSlice": null,
	"StringSlice": [
		"str24",
		"str25",
		"str26"
	],
	"ByteSlice": "Gxwd",
	"Small": {
		"Tag": "tag30"
	},
	"PSmall": {
		"Tag": "tag31"
	},
	"PPSmall": null,
	"Interface": 5.2,
	"PInterface": null
}`)

func TestWithStdDecoder(t *testing.T) {
	expected := make(map[string]interface{})
	json.Unmarshal(allValueIndent, &expected)
	out, _ := Decode(allValueIndent)
	if actual := out.(map[string]interface{}); !reflect.DeepEqual(actual, expected) {
		t.Errorf("compare to std unmarshaler \n\tactual: %v\n\twant: %v", actual, expected)

	}
}

func TestDecodeArray(t *testing.T) {
	for i, tt := range []struct {
		err      error
		in       string
		expected []interface{}
	}{
		{in: `["a"]`, expected: []interface{}{"a"}},
		{in: `["a"]   `, expected: []interface{}{"a"}},
		{in: `   ["a"]`, expected: []interface{}{"a"}},
		{in: `   [     "a"]`, expected: []interface{}{"a"}},
		{in: `   ["a"      ]`, expected: []interface{}{"a"}},
		{in: `["a"      ]1`, err: &SyntaxError{"invalid character '1' after top-level value", 12}},
	} {
		out, err := DecodeArray([]byte(tt.in))
		if !reflect.DeepEqual(err, tt.err) {
			t.Errorf("#%d: %v, want %v", i, err, tt.err)
		}
		if !reflect.DeepEqual(out, tt.expected) {
			t.Errorf("#%d: %v, want %v", i, out, tt.expected)
		}
	}
}

func TestDecodeObject(t *testing.T) {
	for i, tt := range []struct {
		err      error
		in       string
		expected map[string]interface{}
	}{
		{in: `{"a":"a"}`, expected: map[string]interface{}{"a": "a"}},
		{in: `   {"a":"a"}`, expected: map[string]interface{}{"a": "a"}},
		{in: `{"a":"1"}   `, expected: map[string]interface{}{"a": "1"}},
		{in: `{   "a":"1"}`, expected: map[string]interface{}{"a": "1"}},
		{in: `{"a"   :1  }`, expected: map[string]interface{}{"a": float64(1)}},
		{in: `{"a":1}   1`, err: &SyntaxError{"invalid character '1' after top-level value", 11}},
	} {
		out, err := DecodeObject([]byte(tt.in))
		if !reflect.DeepEqual(err, tt.err) {
			t.Errorf("#%d: %v, want %v", i, err, tt.err)
		}
		if !reflect.DeepEqual(out, tt.expected) {
			t.Errorf("#%d: %v, want %v", i, out, tt.expected)
		}
	}
}
