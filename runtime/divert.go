package runtime

import (
	"fmt"
	"strings"
)

type Divert struct {
	ObjectImpl

	// Private
	_targetPath    *Path
	_targetPointer Pointer

	// Public
	VariableDivertName string
	IsExternal         bool
	ExternalArgs       int
	IsConditional      bool
	PushesToStack      bool
	StackPushType      PushPopType
}

func (s *Divert) TargetPath() *Path {

	if s._targetPath != nil && s._targetPath.IsRelative() {

		targetObj := s.TargetPointer().Resolve()
		if targetObj != nil {
			s._targetPath = targetObj.Path(targetObj)
		}
	}

	return s._targetPath
}

func (s *Divert) SetTargetPath(value *Path) {

	s._targetPath = value
	s._targetPointer = NullPointer

}

func (s *Divert) TargetPointer() Pointer {

	if s._targetPointer.IsNull() {

		targetObj := ResolvePath(s, s._targetPath).Obj

		if s._targetPath.LastComponent().IsIndex() {

			s._targetPointer.Container, _ = targetObj.Parent().(*Container)
			s._targetPointer.Index = s._targetPath.LastComponent().Index()
		} else {

			container, _ := targetObj.(*Container)
			s._targetPointer = StartOfPointer(container)
		}
	}

	return s._targetPointer
}

func (s *Divert) TargetPathString() string {

	if s.TargetPath() == nil {
		return ""
	}

	return CompactPathString(s, s.TargetPath())
}

func (s *Divert) SetTargetPathString(value string) {

	if value == "" {
		s.SetTargetPath(nil)
		return
	}

	s.SetTargetPath(NewPathFromString(value))
}

func (s *Divert) HasVariableTarget() bool {

	return s.VariableDivertName != ""
}

func NewDivert() *Divert {

	newDivert := new(Divert)
	newDivert.PushesToStack = false

	return newDivert
}

func (s *Divert) String() string {

	if s.HasVariableTarget() {
		return "Divert(variable: " + s.VariableDivertName + ")"
	}

	if s.TargetPath() == nil {
		return "Divert(null)"
	}

	var sb strings.Builder

	targetStr := s.TargetPath().String()
	if targetLineNum, ok := s.DebugLineNumberOfPath(s, s.TargetPath()); ok {
		targetStr = fmt.Sprintf("line %d", targetLineNum)
	}

	sb.WriteString("Divert")

	if s.IsConditional {
		sb.WriteString("?")
	}
	if s.PushesToStack {
		if s.StackPushType == Function {
			sb.WriteString(" function")
		} else {
			sb.WriteString(" tunnel")
		}
	}

	sb.WriteString(" -> ")
	sb.WriteString(s.TargetPathString())

	sb.WriteString(" (")
	sb.WriteString(targetStr)
	sb.WriteString(")")

	return sb.String()
}
