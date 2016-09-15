package djson

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

// Decode is the exported method to decode arbitrary data into Value object.
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

// Decode is the exported method to decode arbitrary data into Value object.
func DecodeObject(data []byte) (map[string]interface{}, error) {
	d := newDecoder(data)
	d.skipSpaces()
	val, err := d.object()
	if err != nil {
		return nil, err
	}
	if c := d.skipSpaces(); d.pos < d.end {
		return nil, d.error(c, "after top-level value")
	}
	return val, nil
}

// Decode is the exported method to decode arbitrary data into Value object.
func DecodeArray(data []byte) ([]interface{}, error) {
	d := newDecoder(data)
	d.skipSpaces()
	val, err := d.array()
	if err != nil {
		return nil, err
	}
	if c := d.skipSpaces(); d.pos < d.end {
		return nil, d.error(c, "after top-level value")
	}
	return val, nil
}

// Decode is the exported method to decode arbitrary data into Value object.
func DecodeString(data []byte) (string, error) {
	d := newDecoder(data)
	d.skipSpaces()
	val, err := d.string()
	if err != nil {
		return "", err
	}
	if c := d.skipSpaces(); d.pos < d.end {
		return "", d.error(c, "after top-level value")
	}
	return val, nil
}

// Decode is the exported method to decode arbitrary data into Value object.
func DecodeNumber(data []byte) (float64, error) {
	d := newDecoder(data)
	c := d.skipSpaces()
	val, err := d.number(c == '-')
	if err != nil {
		return 0, err
	}
	if c := d.skipSpaces(); d.pos < d.end {
		return 0, d.error(c, "after top-level value")
	}
	return val, nil
}
