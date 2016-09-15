package djson

import (
	"errors"
	"strconv"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// A SyntaxError is a description of a JSON syntax error.
type SyntaxError struct {
	msg    string // description of error
	Offset int    // error occurred after reading Offset bytes
}

func (e *SyntaxError) Error() string { return e.msg }

// Predefined errors
var (
	// Syntax errors
	ErrUnexpectedEOF    = &SyntaxError{"unexpected end of JSON input", -1}
	ErrInvalidHexEscape = &SyntaxError{"invalid hexadecimal escape sequence", -1}
	ErrStringEscape     = errors.New("djson: encountered an invalid escape sequence in a string")
)

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

type decoder struct {
	data []byte
	pos  int
	end  int
}

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

	// if is literal
	if lit != nil {
		d.pos++
		for i, c := range lit.bytes {
			if nc := d.data[d.pos+i]; nc != c {
				err = d.error(nc, "in literal null (expecting '"+string(c)+"')")
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

func (d *decoder) string() (string, error) {
	d.pos++

	var (
		start   = d.pos
		unquote bool
		// This idea comes from jsonparser.
		// stack-allocated array for allocation-free unescaping of small strings
		// if a string longer than this needs to be escaped, it will result in a heap allocation
		stackbuf [64]byte
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
		return 0, &SyntaxError{"invalid number literal, trying to decode " + string(d.data[start:d.pos]) + " into Number", d.pos}
	}
	d.pos++
	goto scan
}

func (d *decoder) array() ([]interface{}, error) {
	// TODO:
	// should test if the array contains 1 element
	// if is, we append and we're done, else, use the
	// growable ogix
	// do benchmark before goes to implemtation
	var (
		c     byte
		v     interface{}
		err   error
		array = make([]interface{}, 0)
	)

	d.pos++

scan:
	c = d.skipSpaces()
	if c == ']' {
		d.pos++
		goto exit
	}

	// read value
	if v, err = d.any(); err != nil {
		goto exit
	}

	array = append(array, v)

	c = d.skipSpaces()
	if c == ',' {
		d.pos++
		goto scan
	} else if c == ']' {
		d.pos++
	} else {
		err = d.error(c, "after array element")
	}

exit:
	return array, err
}

func (d *decoder) object() (map[string]interface{}, error) {
	var (
		c   byte
		k   string
		v   interface{}
		err error
		obj = make(map[string]interface{})
	)
	// '{' already scanned
	d.pos++
	for {
		c = d.skipSpaces()
		if c == '}' {
			d.pos++
			return obj, nil
		}

		// expecting for key
		if c != '"' {
			err = d.error(c, "looking for beginning of object key string")
			break
		}
		if k, err = d.string(); err != nil {
			break
		}

		// expecting for colon
		c = d.skipSpaces()
		if c != ':' {
			err = d.error(c, "after object key")
			break
		}
		d.pos++

		// read value
		if v, err = d.any(); err != nil {
			break
		}

		obj[k] = v

		c = d.skipSpaces()
		// comma or object close
		if c == '}' {
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

// TODO: maybe is better to return -1 as end; or (c, ok)
// and do conversion
func (d *decoder) peek() byte {
	if d.pos < d.end-1 {
		return d.data[d.pos+1]
	}
	return 0
}

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

// quoteChar formats c as a quoted character literal
func quoteChar(c byte) string {
	// special cases - different from quoted strings
	if c == '\'' {
		return `'\''`
	}
	if c == '"' {
		return `'"'`
	}

	// use quoted string with different quotation marks
	s := strconv.Quote(string(c))
	return "'" + s[1:len(s)-1] + "'"
}

func unquoteBytes(s, b []byte) (t []byte, ok bool) {
	if len(s) == 0 {
		return t, true
	}
	// Check for unusual characters. If there are none,
	// then no unquoting is needed, so return a slice of the
	// original bytes.
	r := 0
	for r < len(s) {
		c := s[r]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			r++
			continue
		}
		rr, size := utf8.DecodeRune(s[r:])
		if rr == utf8.RuneError && size == 1 {
			break
		}
		r += size
	}
	if r == len(s) {
		return s, true
	}

	if cap(b) < len(s) {
		b = make([]byte, len(s)+2*utf8.UTFMax)
	}
	w := copy(b, s[0:r])
	for r < len(s) {
		// Out of room?  Can only happen if s is full of
		// malformed UTF-8 and we're replacing each
		// byte with RuneError.
		if w >= len(b)-2*utf8.UTFMax {
			nb := make([]byte, (len(b)+utf8.UTFMax)*2)
			copy(nb, b[0:w])
			b = nb
		}
		switch c := s[r]; {
		case c == '\\':
			r++
			if r >= len(s) {
				return
			}
			switch s[r] {
			default:
				return
			case '"', '\\', '/', '\'':
				b[w] = s[r]
				r++
				w++
			case 'b':
				b[w] = '\b'
				r++
				w++
			case 'f':
				b[w] = '\f'
				r++
				w++
			case 'n':
				b[w] = '\n'
				r++
				w++
			case 'r':
				b[w] = '\r'
				r++
				w++
			case 't':
				b[w] = '\t'
				r++
				w++
			case 'u':
				r--
				rr := getu4(s[r:])
				if rr < 0 {
					return
				}
				r += 6
				if utf16.IsSurrogate(rr) {
					rr1 := getu4(s[r:])
					if dec := utf16.DecodeRune(rr, rr1); dec != unicode.ReplacementChar {
						// A valid pair; consume.
						r += 6
						w += utf8.EncodeRune(b[w:], dec)
						break
					}
					// Invalid surrogate; fall back to replacement rune.
					rr = unicode.ReplacementChar
				}
				w += utf8.EncodeRune(b[w:], rr)
			}

		// Quote, control characters are invalid.
		case c == '"', c < ' ':
			return

		// ASCII
		case c < utf8.RuneSelf:
			b[w] = c
			r++
			w++

		// Coerce to well-formed UTF-8.
		default:
			rr, size := utf8.DecodeRune(s[r:])
			r += size
			w += utf8.EncodeRune(b[w:], rr)
		}
	}
	return b[0:w], true
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}

	// logic taken from:
	// github.com/buger/jsonparser/blob/master/escape.go#L20
	var h [4]int
	for i := range h {
		c := s[2+i]
		switch {
		case c >= '0' && c <= '9':
			h[i] = int(c - '0')
		case c >= 'A' && c <= 'F':
			h[i] = int(c - 'A' + 10)
		case c >= 'a' && c <= 'f':
			h[i] = int(c - 'a' + 10)
		default:
			return -1
		}
	}
	return rune(h[0]<<12 + h[1]<<8 + h[2]<<4 + h[3])
}
