package djson

import (
	"strconv"
	"unicode"
)

// decoder is the object that holds the state of the scaning
type decoder struct {
	data []byte
	pos  int
	end  int
}

func newDecoder(data []byte) *decoder {
	return &decoder{
		data: data,
		end:  len(data),
	}
}

// any used to decode any valid JSON value, and returns an
// interface{} that holds the actual data
func (d *decoder) any() (interface{}, error) {
	var (
		err error
		lit *literal
		val interface{}
	)

	switch c := d.skipSpaces(); c {
	case 'f':
		lit = litFalse
	case 't':
		lit = litTrue
	case 'n':
		lit = litNull
	case '[':
		val, err = d.array()
	case '{':
		val, err = d.object()
	case '"':
		val, err = d.string()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		val, err = d.number(c == '-')
	default:
		return nil, d.error(c, "looking for beginning of value")
	}

	// if we encounter a start of literal, we consume the literal string,
	// while expect it to be equal to the literal variables above, and then
	// decode it into the value v.
	if lit != nil {
		d.pos++
		for i, c := range lit.bytes {
			if nc := d.data[d.pos+i]; nc != c {
				err = d.error(nc, "in literal "+lit.name+"(expecting '"+string(c)+"')")
				break
			}
		}
		d.pos += len(lit.bytes)
		val = lit.val
	}

	if err != nil {
		return nil, err
	}

	return val, nil
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
				s = string(d.data[start:d.pos])
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
		n     float64
		c     byte
		start int

		hasE       bool
		hasDot     bool
		wantNumber bool
	)

	if neg {
		d.pos++
		wantNumber = true
	}

	start = d.pos

scan:
	for d.pos < d.end {
		c = d.data[d.pos]
		switch {
		case '0' <= c && c <= '9':
			n = 10*n + float64(c-'0')
			wantNumber = false
		case (c == 'E' || c == 'e') && !hasE && !wantNumber:
			hasE = true
			if c = d.peek(); c == '+' || c == '-' {
				d.pos++
			}
			fallthrough
		case c == '.' && !hasDot && !wantNumber:
			hasDot = true
			wantNumber = true
		default:
			// if we're done
			if !wantNumber {
				break scan
			}
			return 0, &SyntaxError{"invalid number literal, trying to decode " + string(d.data[start:d.pos]) + " into Number", d.pos}
		}
		d.pos++
	}

	if hasDot {
		v, err := strconv.ParseFloat(string(d.data[start:d.pos]), 64)
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

// peek returns but does not consume the next byte in the input.
func (d *decoder) peek() byte {
	if d.pos < d.end-1 {
		return d.data[d.pos+1]
	}
	return 0
}

// next return the next byte in the input
func (d *decoder) next() byte {
	c := d.peek()
	if d.pos < d.end {
		d.pos++
	}
	return c
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
	return &SyntaxError{"invalid character " + quoteChar(c) + " " + context, d.pos + 1}
}

// literal
type literal struct {
	name  string
	bytes []byte
	val   interface{}
}

var (
	litNull  = &literal{"null", []byte("ull"), nil}
	litTrue  = &literal{"true", []byte("rue"), true}
	litFalse = &literal{"false", []byte("alse"), false}
)
