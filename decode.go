package djson

import (
	"strconv"
	"unicode"
)

// decoder is the object that holds the state of the scaning
type decoder struct {
	data  []byte
	sdata string
	pos   int
	end   int
}

func newDecoder(data []byte) *decoder {
	return &decoder{
		data: data,
		// Add a string version of the data. it is good because we do one allocation
		// operation for string conversion(from bytes to string) and then
		// use "slicing" to create strings in the "decoder.string" method.
		// However, string is a read-only slice, and since the slice references the
		// original array, as long as the slice is kept around the garbage collector
		// can't release the array.
		//
		// Here is the improvements:
		// small payload  - 0.13~ time faster, does 0.45~ less memory allocations but
		//                  the total number of bytes that allocated is 0.03~ bigger
		// medium payload - 0.16~ time faster, does 0.5~ less memory allocations but
		//                  the total number of bytes that allocated is 0.05~ bigger
		// large payload  - 0.13~ time faster, does 0.50~ less memory allocations but
		//                  the total number of bytes that allocated is 0.02~ bigger
		//
		// I don't know if it's worth it, let's wait for the community feedbacks and
		// then I'll see where I go from there.
		sdata: string(data),
		end:   len(data),
	}
}

// any used to decode any valid JSON value, and returns an
// interface{} that holds the actual data
func (d *decoder) any() (interface{}, error) {
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
func (d *decoder) string() (string, error) {
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
				s = d.sdata[start:d.pos]
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
func (d *decoder) number(neg bool) (float64, error) {
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
		v, err := strconv.ParseFloat(d.sdata[start:d.pos], 64)
		if err != nil {
			return 0, err
		}
		n = v
	}
	if neg {
		return -n, nil
	}
	return n, nil
}

// array accept valid JSON array value
func (d *decoder) array() ([]interface{}, error) {
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
func (d *decoder) object() (map[string]interface{}, error) {
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
func (d *decoder) next() byte {
	d.pos++
	if d.pos < d.end {
		return d.data[d.pos]
	}
	return 0
}

// returns the next char after white spaces
func (d *decoder) skipSpaces() byte {
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
func (d *decoder) error(c byte, context string) error {
	if d.pos < d.end {
		return &SyntaxError{"invalid character " + quoteChar(c) + " " + context, d.pos + 1}
	}
	return ErrUnexpectedEOF
}
