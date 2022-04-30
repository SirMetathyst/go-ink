package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

type Reader struct {
	text       string
	offset     int
	rootObject interface{}
}

func (s *Reader) ToMap() map[string]interface{} {
	return s.rootObject.(map[string]interface{})
}

func (s *Reader) ToArray() []interface{} {
	return s.rootObject.([]interface{})
}

func (s *Reader) isNumberChar(c rune) bool {
	return c >= '0' && c <= '9' || c == '.' || c == '-' || c == '+' || c == 'E' || c == 'e'
}

func (s *Reader) isFirstNumberChar(c rune) bool {
	return c >= '0' && c <= '9' || c == '-' || c == '+'
}

func (s *Reader) readObject() interface{} {

	var currentChar = rune(s.text[s.offset])

	if currentChar == '{' {
		return s.readMap()
	}

	if currentChar == '[' {
		return s.readArray
	}

	if currentChar == '"' {
		return s.readString()
	}

	if s.isFirstNumberChar(currentChar) {
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

	panic("Unhandled object type in JSON: " + s.text[s.offset:s.offset+30])
}

func (s *Reader) readMap() map[string]interface{} {

	m := make(map[string]interface{})
	s.expect("{")
	s.skipWhitespace()

	if s.tryRead("}") {
		return m
	}

	for {

		s.skipWhitespace()

		// Key
		key := s.readString()
		s.expectCondition(len(key) != 0, "dictionary key")
		s.skipWhitespace()

		// :
		s.expect(":")
		s.skipWhitespace()

		// Value
		val := s.readObject()
		s.expectCondition(val != nil, "dictionary value")

		// Add to map
		m[key] = val
		s.skipWhitespace()

		if s.tryRead(",") == false {
			break
		}
	}

	s.expect("}")
	return m
}

func (s *Reader) readArray() []interface{} {

	var list []interface{}

	s.expect("[")
	s.skipWhitespace()

	if s.tryRead("]") {
		return list
	}

	for {

		s.skipWhitespace()

		// Value
		val := s.readObject()
		list = append(list, val)

		s.skipWhitespace()

		if s.tryRead(",") == false {
			break
		}
	}

	s.expect("]")
	return list
}

func (s *Reader) readString() string {

	s.expect("\"")

	sb := strings.Builder{}

	for ; s.offset < len(s.text); s.offset++ {

		c := rune(s.text[s.offset])

		if c == '\\' {
			// Escaped character
			s.offset++
			if s.offset >= len(s.text) {
				panic("Unexpected EOF while reading string")
			}
			c = rune(s.text[s.offset])
			switch c {
			case '"':
			case '\\':
			case '/': // Yes, JSON allows this to be escaped
				sb.WriteRune(c)
				break
			case 'n':
				sb.WriteRune('\n')
				break
			case 't':
				sb.WriteRune('\t')
				break
			case 'r':
			case 'b':
			case 'f':
				// Ignore other control characters
				break
			case 'u':
				if s.offset+4 >= len(s.text) {
					panic("Unexpected EOF while reading string")
				}
				digits := s.text[s.offset+1 : s.offset+5]
				v, err := strconv.Atoi(digits)
				if err != nil {
					panic("Invalid Unicode escape character at offset " + fmt.Sprint(s.offset-1))
				}
				sb.WriteRune(rune(v))
				s.offset += 4
				break
			default:
				panic("Invalid Unicode escape character at offset " + fmt.Sprint(s.offset-1))
			}
		} else if c == '"' {
			break
		} else {
			sb.WriteRune(c)
		}
	}

	s.expect("\"")
	return sb.String()
}

func (s *Reader) readNumber() interface{} {

	startOffset := s.offset
	isFloat := false

	for ; s.offset < len(s.text); s.offset++ {
		c := rune(s.text[s.offset])
		if c == '.' || c == 'e' || c == 'E' {
			isFloat = true
		}
		if s.isNumberChar(c) {
			continue
		} else {
			break
		}
	}

	numStr := s.text[startOffset : startOffset+(s.offset-startOffset)]

	if isFloat {
		if f, err := strconv.ParseFloat(numStr, 64); err == nil {
			return f
		}
	} else {
		if i, err := strconv.Atoi(numStr); err == nil {
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
		if len(message) == 0 {
			message = "Unexpected token"
		} else {
			message = "Expected " + message
		}
		message += " at offset " + fmt.Sprint(s.offset)

		panic(message)
	}
}

func (s *Reader) skipWhitespace() {

	for s.offset < len(s.text) {
		c := rune(s.text[s.offset])
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			s.offset++
		} else {
			break
		}
	}
}

func NewReader(text string) *Reader {
	s := &Reader{text: text, offset: 0}
	s.skipWhitespace()
	s.rootObject = s.readObject()
	return s
}
