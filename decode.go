package djson

import (
	"strconv"
	"unicode"
)

// Decoder is the object that holds the state of the decoding
type Decoder struct {
	pos       int
	end       int
	data      []byte
	sdata     string
	usestring bool
}

// NewDecoder creates new Decoder from the JSON-encoded data
func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data: data,
		end:  len(data),
	}
}

// AllocString pre-allocate a string version of the data before starting
// decoding.
// It is used to make the decode operation more fast(see below) by doing one
// allocation operation for string conversion(from bytes), and then use
// "slicing" to create non-escaped strings in the "Decoder.string" method.
// However, string is a read-only slice, and since the slice references the
// original array, as long as the slice is kept around, the garbage collector
// can't release the array.
// For this reason, you want to use this method only when the Decoder's result
// is a "read-only" or you are adding more elements to it. see example below.
//
// Here are the improvements:
//
//	small payload  - 0.13~ time faster, does 0.45~ less memory allocations but
// 			 the total number of bytes that are allocated is 0.03~ bigger
//
// 	medium payload - 0.16~ time faster, does 0.5~ less memory allocations but
// 			 the total number of bytes that are allocated is 0.05~ bigger
//
// 	large payload  - 0.13~ time faster, does 0.50~ less memory allocations but
// 			 the total number of bytes that are allocated is 0.02~ bigger
//
// Here is an example to illustrate when you don't want to use this method
//
// 	str := fmt.Sprintf(`{"foo": "bar", "baz": "%s"}`, strings.Repeat("#", 1024 * 1024))
//	dec := djson.NewDecoder([]byte(str))
// 	dec.AllocString()
// 	ev, err := dec.DecodeObject()
//
// 	// inpect memory stats here; MemStats.Alloc ~= 1M
//
// 	delete(ev, "baz") // or ev["baz"] = "qux"
//
// 	// inpect memory stats again; MemStats.Alloc ~= 1M
// 	// it means that the chunk that sat in the "baz" value is not freed
//
func (d *Decoder) AllocString() {
	d.sdata = string(d.data)
	d.usestring = true
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
func (d *Decoder) Decode() (interface{}, error) {
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
func (d *Decoder) DecodeObject() (map[string]interface{}, error) {
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
func (d *Decoder) DecodeArray() ([]interface{}, error) {
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

// any used to decode any valid JSON value, and returns an
// interface{} that holds the actual data
func (d *Decoder) any() (interface{}, error) {
	switch c := d.skipSpaces(); c {
	case '"':
		return d.string()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return d.number(false)
	case '-':
		return d.number(true)
	case 'f':
		d.pos++
		if d.end-d.pos < 4 {
			return nil, ErrUnexpectedEOF
		}
		if d.data[d.pos] == 'a' && d.data[d.pos+1] == 'l' && d.data[d.pos+2] == 's' && d.data[d.pos+3] == 'e' {
			d.pos += 4
			return false, nil
		}
		return nil, d.error(d.data[d.pos], "in literal false")
	case 't':
		d.pos++
		if d.end-d.pos < 3 {
			return nil, ErrUnexpectedEOF
		}
		if d.data[d.pos] == 'r' && d.data[d.pos+1] == 'u' && d.data[d.pos+2] == 'e' {
			d.pos += 3
			return true, nil
		}
		return nil, d.error(d.data[d.pos], "in literal true")
	case 'n':
		d.pos++
		if d.end-d.pos < 3 {
			return nil, ErrUnexpectedEOF
		}
		if d.data[d.pos] == 'u' && d.data[d.pos+1] == 'l' && d.data[d.pos+2] == 'l' {
			d.pos += 3
			return nil, nil
		}
		return nil, d.error(d.data[d.pos], "in literal null")
	case '[':
		return d.array()
	case '{':
		return d.object()
	default:
		return nil, d.error(c, "looking for beginning of value")
	}
}

// string called by `any` or `object`(for map keys) after reading `"`
func (d *Decoder) string() (string, error) {
	d.pos++

	var (
		unquote bool
		start   = d.pos
	)

scan:
	for {
		if d.pos >= d.end {
			return "", ErrUnexpectedEOF
		}

		c := d.data[d.pos]
		switch {
		case c == '"':
			var s string
			if unquote {
				// stack-allocated array for allocation-free unescaping of small strings
				// if a string longer than this needs to be escaped, it will result in a
				// heap allocation; idea comes from github.com/burger/jsonparser
				var stackbuf [64]byte
				data, ok := unquoteBytes(d.data[start:d.pos], stackbuf[:])
				if !ok {
					return "", ErrStringEscape
				}
				s = string(data)
			} else {
				if d.usestring {
					s = d.sdata[start:d.pos]
				} else {

					s = string(d.data[start:d.pos])
				}
			}
			d.pos++
			return s, nil
		case c == '\\':
			d.pos++
			unquote = true
			switch c := d.data[d.pos]; c {
			case 'u':
				goto escape_u
			case 'b', 'f', 'n', 'r', 't', '\\', '/', '"':
				d.pos++
			default:
				return "", d.error(c, "in string escape code")
			}
		case c < 0x20:
			return "", d.error(c, "in string literal")
		default:
			d.pos++
			if c > unicode.MaxASCII {
				unquote = true
			}
		}
	}

escape_u:
	d.pos++
	for i := 0; i < 3; i++ {
		if d.pos < d.end {
			c := d.data[d.pos+i]
			if '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F' {
				continue
			}
			return "", d.error(c, "in \\u hexadecimal character escape")
		}
		return "", ErrInvalidHexEscape
	}
	d.pos += 3
	goto scan
}

// number called by `any` after reading `-` or number between 0 to 9
func (d *Decoder) number(neg bool) (float64, error) {
	var (
		n       float64
		c       byte
		start   int
		isFloat bool
	)

	if neg {
		d.pos++
		if c = d.data[d.pos]; c < '0' && c > '9' {
			return 0, d.error(c, "in negative numeric literal")
		}
	}

	start = d.pos
	c = d.data[d.pos]

	// digits first
	switch {
	case c == '0':
		c = d.next()
	case '1' <= c && c <= '9':
		for ; c >= '0' && c <= '9'; c = d.next() {
			n = 10*n + float64(c-'0')
		}
	}

	// . followed by 1 or more digits
	if c == '.' {
		d.pos++
		isFloat = true
		if c = d.data[d.pos]; c < '0' && c > '9' {
			return 0, d.error(c, "after decimal point in numeric literal")
		}
		for c = d.next(); '0' <= c && c <= '9'; {
			c = d.next()
		}
	}

	// e or E followed by an optional - or + and
	// 1 or more digits.
	if c == 'e' || c == 'E' {
		isFloat = true
		if c = d.next(); c == '+' || c == '-' {
			if c = d.next(); c < '0' || c > '9' {
				return 0, d.error(c, "in exponent of numeric literal")
			}
		}
		for c = d.next(); '0' <= c && c <= '9'; {
			c = d.next()
		}
	}

	if isFloat {
		var (
			err error
			sn  string
		)
		if d.usestring {
			sn = d.sdata[start:d.pos]
		} else {
			sn = string(d.data[start:d.pos])
		}
		if n, err = strconv.ParseFloat(sn, 64); err != nil {
			return 0, err
		}
	}
	if neg {
		return -n, nil
	}
	return n, nil
}

// array accept valid JSON array value
func (d *Decoder) array() ([]interface{}, error) {
	// the '[' token already scanned
	d.pos++

	var (
		c     byte
		v     interface{}
		err   error
		array = make([]interface{}, 0)
	)

	// look ahead for ] - if the array is empty.
	if c = d.skipSpaces(); c == ']' {
		d.pos++
		goto out
	}

scan:
	if v, err = d.any(); err != nil {
		goto out
	}

	array = append(array, v)

	// next token must be ',' or ']'
	if c = d.skipSpaces(); c == ',' {
		d.pos++
		goto scan
	} else if c == ']' {
		d.pos++
	} else {
		err = d.error(c, "after array element")
	}

out:
	return array, err
}

// object accept valid JSON array value
func (d *Decoder) object() (map[string]interface{}, error) {
	// the '{' token already scanned
	d.pos++

	var (
		c   byte
		k   string
		v   interface{}
		err error
		obj = make(map[string]interface{})
	)

	// look ahead for } - if the object has no keys.
	if c = d.skipSpaces(); c == '}' {
		d.pos++
		return obj, nil
	}

	for {
		// read string key
		if c = d.skipSpaces(); c != '"' {
			err = d.error(c, "looking for beginning of object key string")
			break
		}
		if k, err = d.string(); err != nil {
			break
		}

		// read colon before value
		c = d.skipSpaces()
		if c != ':' {
			err = d.error(c, "after object key")
			break
		}
		d.pos++

		// read and assign value
		if v, err = d.any(); err != nil {
			break
		}

		obj[k] = v

		// next token must be ',' or '}'
		if c = d.skipSpaces(); c == '}' {
			d.pos++
			break
		} else if c == ',' {
			d.pos++
		} else {
			err = d.error(c, "after object key:value pair")
			break
		}
	}

	return obj, err
}

// next return the next byte in the input
func (d *Decoder) next() byte {
	d.pos++
	if d.pos < d.end {
		return d.data[d.pos]
	}
	return 0
}

// returns the next char after white spaces
func (d *Decoder) skipSpaces() byte {
loop:
	if d.pos == d.end {
		return 0
	}
	switch c := d.data[d.pos]; c {
	case ' ', '\t', '\n', '\r':
		d.pos++
		goto loop
	default:
		return c
	}
}

// emit sytax errors
func (d *Decoder) error(c byte, context string) error {
	if d.pos < d.end {
		return &SyntaxError{"invalid character " + quoteChar(c) + " " + context, d.pos + 1}
	}
	return ErrUnexpectedEOF
}
