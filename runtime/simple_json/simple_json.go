package simple_json

import (
	"fmt"
	"strconv"
	"strings"
)

func TextToDictionary(text string) map[string]interface{} {
	return NewReader(text).ToDictionary()
}

func TextToArray(text string) []interface{} {
	return NewReader(text).ToArray()
}

type Reader struct {
	text       string
	offset     int
	rootObject interface{}
}

func NewReader(text string) *Reader {
	r := &Reader{text: text, offset: 0}
	r.skipWhitespace()
	r.rootObject = r.readObject()
	return r
}

func (s *Reader) ToDictionary() map[string]interface{} {
	return s.rootObject.(map[string]interface{})
}

func (s *Reader) ToArray() []interface{} {
	return s.rootObject.([]interface{})
}

func IsNumberChar(c uint8) bool {
	return c >= '0' && c <= '9' || c == '.' || c == '-' || c == '+' || c == 'E' || c == 'e'
}

func IsFirstNumberChar(c uint8) bool {
	return c >= '0' && c <= '9' || c == '-' || c == '+'
}

func (s *Reader) readObject() interface{} {

	currentChar := s.text[s.offset]

	if currentChar == '{' {
		return s.ReadDictionary()
	}

	if currentChar == '[' {
		return s.ReadArray()
	}

	if currentChar == '"' {
		return s.readString()
	}

	if IsFirstNumberChar(currentChar) {
		return s.readNumber()
	}

	if s.tryRead("true") {
		return true
	}

	if s.tryRead("false") {
		return false
	}

	if s.tryRead("null") {
		return nil
	}

	panic(fmt.Sprintf("Unhandled object type in JSON: %s", s.text[s.offset:s.offset+30]))
}

func (s *Reader) ReadDictionary() map[string]interface{} {

	dict := make(map[string]interface{}, 0)

	s.expect("{")
	s.skipWhitespace()

	// Empty dictionary?
	if s.tryRead("}") {
		return dict
	}

	for do := true; do; do = s.tryRead(",") {

		s.skipWhitespace()

		// Key
		key := s.readString()
		s.expectCondition(key != "", "dictionary key")

		s.skipWhitespace()

		// :
		s.expect(":")

		s.skipWhitespace()

		// Value
		val := s.readObject()
		s.expectCondition(val != nil, "dictionary value")

		// Add to dictionary
		dict[key] = val

		s.skipWhitespace()
	}

	s.expect("}")

	return dict
}

func (s *Reader) ReadArray() []interface{} {

	var list []interface{}

	s.expect("[")
	s.skipWhitespace()

	// Empty list?
	if s.tryRead("]") {
		return list
	}

	for do := true; do; do = s.tryRead(",") {

		s.skipWhitespace()

		// Value
		var val = s.readObject()

		// Add to array
		list = append(list, val)

		s.skipWhitespace()
	}

	s.expect("]")
	return list
}

func (s *Reader) readString() string {

	s.expect("\"")

	sb := strings.Builder{}

	for ; s.offset < len(s.text); s.offset++ {
		c := s.text[s.offset]

		if c == '\\' {
			// Escaped character
			s.offset++
			if s.offset >= len(s.text) {
				panic("Unexpected EOF while reading string")
			}
			c = s.text[s.offset]
			switch c {
			case '"':
				fallthrough
			case '\\':
				fallthrough
			case '/': // Yes, JSON allows this to be escaped
				sb.WriteByte(c)
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('t')
			case 'r':
				fallthrough
			case 'b':
				fallthrough
			case 'f':
			// Ignore other control characters
			case 'u':
				// 4-digit Unicode
				if s.offset+4 >= len(s.text) {
					panic("Unexpected EOF while reading string")
				}
				digits := s.text[s.offset+1 : s.offset+5]
				// remove 0x suffix if found in the input string
				cleaned := strings.Replace(digits, "0x", "", -1)
				// base 16 for hexadecimal
				uchar, err := strconv.ParseUint(cleaned, 16, 64)
				if err == nil {
					sb.WriteByte(uint8(uchar))
					s.offset += 4
				} else {
					panic(fmt.Sprintf("Invalid Unicode escape character at offset %d", s.offset-1))
				}
			default:
				panic(fmt.Sprintf("Invalid Unicode escape character at offset %d", s.offset-1))
			}
		} else if c == '"' {
			break
		} else {
			sb.WriteByte(c)
		}
	}

	s.expect("\"")
	return sb.String()
}

func (s *Reader) readNumber() interface{} {

	startOffset := s.offset

	isFloat := false
	for ; s.offset < len(s.text); s.offset++ {
		var c = s.text[s.offset]
		if c == '.' || c == 'e' || c == 'E' {
			isFloat = true
		}
		if IsNumberChar(c) {
			continue
		} else {
			break
		}
	}

	numStr := s.text[startOffset:s.offset]

	if isFloat {
		f, err := strconv.ParseFloat(numStr, 32)
		if err == nil {
			return f
		}
	} else {
		i, err := strconv.ParseInt(numStr, 10, 32)
		if err == nil {
			return i
		}
	}

	panic("Failed to parse number value: " + numStr)
}

func (s *Reader) tryRead(textToRead string) bool {

	if s.offset+len(textToRead) > len(s.text) {
		return false
	}

	for i := 0; i < len(textToRead); i++ {
		if textToRead[i] != s.text[s.offset+i] {
			return false
		}
	}

	s.offset += len(textToRead)

	return true
}

func (s *Reader) expect(expectedStr string) {

	if !s.tryRead(expectedStr) {
		s.expectCondition(false, expectedStr)
	}

}

func (s *Reader) expectCondition(condition bool, message string) {

	if !condition {
		if message == "" {
			message = "Unexpected token"
		} else {
			message = "Expected " + message
		}
		message += fmt.Sprintf(" at offset %d", s.offset)

		panic(message)
	}
}

func (s *Reader) skipWhitespace() {

	for s.offset < len(s.text) {
		c := s.text[s.offset]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			s.offset++
		} else {
			break
		}
	}
}
