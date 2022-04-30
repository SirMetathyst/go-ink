package runtime

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

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
	Type       State
	ChildCount int
}

type Writer struct {
	stateStack Stack[StateElement]
	writer     io.Writer
}

func (s *Writer) State() State {

	if len(s.stateStack) > 0 {
		return s.stateStack.Peek().Type
	}

	return StateNone
}

func (s *Writer) ChildCount() int {

	if len(s.stateStack) > 0 {
		return s.stateStack.Peek().ChildCount
	}

	return 0
}

func (s *Writer) WriteObject(inner func(writer *Writer)) {
	s.WriteObjectStart()
	inner(s)
	s.WriteObjectEnd()
}

func (s *Writer) WriteObjectStart() {
	s.StartNewObject(true)
	s.stateStack.Push(StateElement{Type: StateObject})
	s.writer.Write([]byte("{"))
}

func (s *Writer) WriteObjectEnd() {
	s.writer.Write([]byte("}"))
	s.stateStack.Pop()
}

func (s *Writer) WriteStringProperty(name string, content string) {
	s.WritePropertyStart(name)
	s.WriteString(content)
	s.WritePropertyEnd()
}

func (s *Writer) WriteStringPropertyFunc(name string, inner func(writer *Writer)) {
	s.WritePropertyFunc(name, inner)
}

func (s *Writer) WriteIntProperty(name string, content int) {
	s.WritePropertyStart(name)
	s.WriteInt(content)
	s.WritePropertyEnd()
}

func (s *Writer) WriteIntPropertyFunc(name string, inner func(writer *Writer)) {
	s.WritePropertyFunc(name, inner)
}

func (s *Writer) WriteBoolProperty(name string, content bool) {
	s.WritePropertyStart(name)
	s.WriteBool(content)
	s.WritePropertyEnd()
}

func (s *Writer) WriteBoolPropertyFunc(name string, inner func(writer *Writer)) {
	s.WritePropertyFunc(name, inner)
}

func (s *Writer) WritePropertyStart(name interface{}) {

	if s.ChildCount() > 0 {
		s.writer.Write([]byte(","))
	}

	sName := ""

	switch v := name.(type) {
	case string:
		sName = v
	case int:
		sName = strconv.Itoa(v)
	}

	s.writer.Write([]byte("\""))
	s.writer.Write([]byte(sName))
	s.writer.Write([]byte("\":"))

	s.IncrementChildCount()

	s.stateStack.Push(StateElement{Type: StateProperty})
}

func (s *Writer) WritePropertyEnd() {
	s.stateStack.Pop()
}

func (s *Writer) WritePropertyFunc(name interface{}, inner func(writer *Writer)) {
	s.WritePropertyStart(name)
	inner(s)
	s.WritePropertyEnd()
}

func (s *Writer) WriteString(str string) {
	s.WriteEscapedString(str)
}

func (s *Writer) WriteInt(i int) {
	s.StartNewObject(false)
	s.writer.Write([]byte(strconv.Itoa(i)))
}

func (s *Writer) WriteFloat(f float64) {
	s.StartNewObject(false)
	s.writer.Write([]byte(fmt.Sprintf("%f", f)))
}

func (s *Writer) WriteBool(b bool) {
	bs := "false"
	if b == true {
		bs = "true"
	}
	s.StartNewObject(false)
	s.writer.Write([]byte(bs))
}

func (s *Writer) WriteNull() {
	s.StartNewObject(false)
	s.writer.Write([]byte("null"))
}

func (s *Writer) WriteStringStart() {
	s.StartNewObject(false)
	s.stateStack.Push(StateElement{Type: StateString})
	s.writer.Write([]byte("\""))
}

func (s *Writer) WriteStringEnd() {
	s.writer.Write([]byte("\""))
	s.stateStack.Pop()
}

func (s *Writer) WritePropertyNameStart() {

	if s.ChildCount() > 0 {
		s.writer.Write([]byte(","))
	}

	s.writer.Write([]byte("\""))

	s.IncrementChildCount()

	s.stateStack.Push(StateElement{Type: StateProperty})
	s.stateStack.Push(StateElement{Type: StatePropertyName})
}

func (s *Writer) WritePropertyNameEnd() {

	s.writer.Write([]byte("\":"))
	s.stateStack.Pop()
}

func (s *Writer) WritePropertyNameInner(str string) {
	s.writer.Write([]byte(str))
}

func (s *Writer) WriteEscapedString(str string) {
	for _, c := range str {

		if c < ' ' {

			switch c {
			case '\n':
				s.writer.Write([]byte("\\n"))
				break
			case '\t':
				s.writer.Write([]byte("\\t"))
				break
			}

		} else {

			switch c {
			case '\\':
			case '"':
				s.writer.Write([]byte("\\"))
				s.writer.Write([]byte(string(c)))
				break
			}
		}
	}
}

func (s *Writer) StartNewObject(container bool) {

	if s.State() == StateArray && s.ChildCount() > 0 {
		s.writer.Write([]byte(","))
	}

	if s.State() == StateArray || s.State() == StateProperty {
		s.IncrementChildCount()
	}
}

func (s *Writer) WriteArrayStart() {
	s.StartNewObject(true)
	s.stateStack.Push(StateElement{Type: StateArray})
	s.writer.Write([]byte("["))
}

func (s *Writer) WriteArrayEnd() {
	s.writer.Write([]byte("]"))
	s.stateStack.Pop()
}

func (s *Writer) IncrementChildCount() {
	if currEl, ok := s.stateStack.Pop(); ok {
		currEl.ChildCount++
		s.stateStack.Push(currEl)
	}
}

func NewWriter() *Writer {
	s := &Writer{writer: &strings.Builder{}}
	return s
}

func NewWriterWith(writer io.Writer) *Writer {
	s := &Writer{writer: writer}
	return s
}
