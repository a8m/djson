package djson

import "errors"

// A SyntaxError is a description of a JSON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int    // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }

// Predefined errors
var (
	ErrUnexpectedEOF    = &SyntaxError{"unexpected end of JSON input", -1}
	ErrInvalidHexEscape = &SyntaxError{"invalid hexadecimal escape sequence", -1}
	ErrStringEscape     = &SyntaxError{"encountered an invalid escape sequence in a string", -1}
)

// ValueType identifies the type of a parsed value.
type ValueType int

func (v ValueType) String() string {
	return types[v]
}

const (
	Null ValueType = iota
	Bool
	String
	Number
	Object
	Array
	Unknown
)

var types = map[ValueType]string{
	Null:    "null",
	Bool:    "boolean",
	String:  "string",
	Number:  "number",
	Object:  "object",
	Array:   "array",
	Unknown: "unknown",
}

// Type returns the JSON-type of the given value
func Type(v interface{}) ValueType {
	t := Unknown
	switch v.(type) {
	case nil:
		t = Null
	case bool:
		t = Bool
	case string:
		t = String
	case float64:
		t = Number
	case []interface{}:
		t = Array
	case map[string]interface{}:
		t = Object
	}
	return t
}

// Decode parses the JSON-encoded data and returns an interface value.
// The interface value could be one of these:
//
//	bool, for JSON booleans
//	float64, for JSON numbers
//	string, for JSON strings
//	[]interface{}, for JSON arrays
//	map[string]interface{}, for JSON objects
//	nil for JSON null
//
// Note that the Decode is compatible with the the following
// insructions:
//
//	var v interface{}
//	err := json.Unmarshal(data, &v)
//
func Decode(data []byte) (interface{}, error) {
	d := newDecoder(data)
	val, err := d.any()
	if err != nil {
		return nil, err
	}
	if c := d.skipSpaces(); d.pos < d.end {
		return nil, d.error(c, "after top-level value")
	}
	return val, nil
}

// DecodeObject is the same as Decode but it returns map[string]interface{}.
// You should use it to parse JSON objects.
func DecodeObject(data []byte) (map[string]interface{}, error) {
	d := newDecoder(data)
	if c := d.skipSpaces(); c != '{' {
		return nil, d.error(c, "looking for beginning of object")
	}
	val, err := d.object()
	if err != nil {
		return nil, err
	}
	if c := d.skipSpaces(); d.pos < d.end {
		return nil, d.error(c, "after top-level value")
	}
	return val, nil
}

// DecodeArray is the same as Decode but it returns []interface{}.
// You should use it to parse JSON arrays.
func DecodeArray(data []byte) ([]interface{}, error) {
	d := newDecoder(data)
	if c := d.skipSpaces(); c != '[' {
		return nil, d.error(c, "looking for beginning of array")
	}
	val, err := d.array()
	if err != nil {
		return nil, err
	}
	if c := d.skipSpaces(); d.pos < d.end {
		return nil, d.error(c, "after top-level value")
	}
	return val, nil
}
