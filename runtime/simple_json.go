package runtime

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
			return int(i)
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

type State int

const (
	StateNone State = iota
	StateObject
	StateArray
	StateProperty
	StatePropertyName
	StateString
)

type StateElement struct {
	_type      State
	childCount int
}

type Writer struct {

	// Private
	stateStack Stack[StateElement]
	writer     strings.Builder
}

func NewWriter() *Writer {
	return new(Writer)
}

func (s *Writer) WriteObjectFunc(inner func(w *Writer)) {
	s.WriteObjectStart()
	inner(s)
	s.WriteObjectEnd()
}

func (s *Writer) WriteObjectStart() {
	s.StartNewObject(true)
	s.stateStack.Push(StateElement{_type: StateObject})
	s.writer.WriteString("{")
}

func (s *Writer) WriteObjectEnd() {
	// Assert(state == State.Object);
	if s.State() != StateObject {
		panic("state != StateObject")
	}
	s.writer.WriteString("}")
	s.stateStack.Pop()
}

func (s *Writer) WriteStringPropertyFunc(name string, inner func(w *Writer)) {
	s.WritePropertyStart(name)
	inner(s)
	s.WritePropertyEnd()
}

func (s *Writer) WriteIntPropertyFunc(id int, inner func(w *Writer)) {
	s.WritePropertyStart(id)
	inner(s)
	s.WritePropertyEnd()
}

func (s *Writer) WriteStringProperty(name string, content string) {
	s.WritePropertyStart(name)
	s.WriteString(content, true)
	s.WritePropertyEnd()
}

func (s *Writer) WriteIntProperty(name string, content int) {
	s.WritePropertyStart(name)
	s.WriteInt(content)
	s.WritePropertyEnd()
}

func (s *Writer) WriteBoolProperty(name string, content bool) {
	s.WritePropertyStart(name)
	s.WriteBool(content)
	s.WritePropertyEnd()
}

func (s *Writer) WritePropertyStart(name interface{}) {

	// Assert(state == State.Object)
	if s.State() != StateObject {
		panic("state != StateObject")
	}

	if s.ChildCount() > 0 {
		s.writer.WriteString(",")
	}

	s.writer.WriteString("\"")
	s.writer.WriteString(fmt.Sprint(name))
	s.writer.WriteString("\":")

	s.IncrementChildCount()

	s.stateStack.Push(StateElement{_type: StateProperty})
}

func (s *Writer) WritePropertyFunc(name interface{}, inner func(w *Writer)) {
	s.WritePropertyStart(name)
	inner(s)
	s.WritePropertyEnd()
}

func (s *Writer) WritePropertyEnd() {
	// Assert(state == State.Property)
	if s.State() != StateProperty {
		panic("state != StateProperty")
	}

	// Assert(childCount == 1)
	if s.ChildCount() != 1 {
		panic("child count != 1")
	}

	s.stateStack.Pop()
}

func (s *Writer) WritePropertyNameStart() {
	//Assert(state == State.Object);

	if s.ChildCount() > 0 {
		s.writer.WriteString(",")
	}

	s.writer.WriteString("\"")

	s.IncrementChildCount()

	s.stateStack.Push(StateElement{_type: StateProperty})
	s.stateStack.Push(StateElement{_type: StatePropertyName})
}

func (s *Writer) WritePropertyNameEnd() {
	//Assert(state == State.PropertyName);
	s.writer.WriteString("\":")
	// Pop PropertyName, leaving Property state
	s.stateStack.Pop()
}

func (s *Writer) WritePropertyNameInner(str string) {
	//Assert(state == State.PropertyName);
	s.writer.WriteString(str)
}

func (s *Writer) WriteArrayStart() {
	s.StartNewObject(true)
	s.stateStack.Push(StateElement{_type: StateArray})
	s.writer.WriteString("[")
}

func (s *Writer) WriteArrayEnd() {
	//Assert(state == State.Array);
	s.writer.WriteString("]")
	s.stateStack.Pop()
}

func (s *Writer) WriteInt(i int) {
	s.StartNewObject(false)
	s.writer.WriteString(fmt.Sprint(i))
}

func (s *Writer) WriteFloat(f float64) {
	s.StartNewObject(false)

	// TODO: Find an heap-allocation-free way to do this please!
	// _writer.Write(formatStr, obj (the float)) requires boxing
	// Following implementation seems to work ok but requires creating temporary garbage string.
	floatStr := fmt.Sprint(f)
	//if (floatStr == "Infinity") {
	//	_writer.Write("3.4E+38"); // JSON doesn't support, do our best alternative
	//} else if (floatStr == "-Infinity") {
	//	_writer.Write("-3.4E+38"); // JSON doesn't support, do our best alternative
	//} else if (floatStr == "NaN") {
	//	_writer.Write("0.0"); // JSON doesn't support, not much we can do
	//} else {
	s.writer.WriteString(floatStr)
	//if (!floatStr.Contains(".") && !floatStr.Contains("E"))
	//	_writer.Write(".0"); // ensure it gets read back in as a floating point value
	//}
}

// (default) escape: true
func (s *Writer) WriteString(str string, escape bool) {
	s.StartNewObject(false)

	s.writer.WriteString("\"")
	if escape {
		s.WriteEscapedString(str)
	} else {
		s.writer.WriteString(str)
	}
	s.writer.WriteString("\"")
}

func (s *Writer) WriteBool(b bool) {
	s.StartNewObject(false)

	str := "false"
	if b == true {
		str = "true"
	}

	s.writer.WriteString(str)
}

func (s *Writer) WriteNull() {
	s.StartNewObject(false)
	s.writer.WriteString("null")
}

func (s *Writer) WriteStringStart() {
	s.StartNewObject(false)
	s.stateStack.Push(StateElement{_type: StateString})
	s.writer.WriteString("\"")
}

func (s *Writer) WriteStringEnd() {
	//Assert(state == State.String);
	s.writer.WriteString("\"")
	s.stateStack.Pop()
}

func (s *Writer) WriteStringInner(str string, escape bool) {
	//Assert(state == State.String);
	if escape {
		s.WriteEscapedString(str)
	} else {
		s.writer.WriteString(str)
	}
}

func (s *Writer) WriteEscapedString(str string) {

	for _, c := range str {
		if c < ' ' {
			switch c {
			case '\n':
				s.writer.WriteString("\\n")
			case '\t':
				s.writer.WriteString("\\t")
			}
		} else {
			switch c {
			case '\\':
				fallthrough
			case '"':
				s.writer.WriteString("\\")
				s.writer.WriteRune(c)
			default:
				s.writer.WriteRune(c)
			}
		}
	}
}

//func (s *Writer) WriteProperty(name interface{}, content interface{}) {
//	s.WritePropertyStart(name)
//	s.WriteString(fmt.Sprint(content), true)
//	s.WritePropertyEnd()
//}

func (s *Writer) String() string {
	return s.writer.String()
}

func (s *Writer) IncrementChildCount() {

	// Assert(_stateStack.Count > 0);
	if s.stateStack.Len() <= 0 {
		panic("state stack length <= 0")
	}

	currEL, _ := s.stateStack.Pop()
	currEL.childCount++
	s.stateStack.Push(currEL)
}

func (s *Writer) StartNewObject(container bool) {
	//
	//if (container) {
	//	Assert(state == State.None || state == State.Property || state == State.Array);
	//} else {
	//	Assert(state == State.Property || state == State.Array);
	//}

	if s.State() == StateArray && s.ChildCount() > 0 {
		s.writer.WriteString(",")
	}

	//if (state == State.Property)
	//	Assert(childCount == 0);

	if s.State() == StateArray || s.State() == StateProperty {
		s.IncrementChildCount()
	}
}

func (s *Writer) State() State {

	if s.stateStack.Len() > 0 {
		if elm, ok := s.stateStack.Peek(); ok {
			return elm._type
		}
	}

	return StateNone
}

func (s *Writer) ChildCount() int {

	if s.stateStack.Len() > 0 {
		if elm, ok := s.stateStack.Peek(); ok {
			return elm.childCount
		}
	}

	return 0
}
