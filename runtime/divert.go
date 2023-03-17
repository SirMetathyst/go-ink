package runtime

import (
	"fmt"
	"strings"
)

type Divert struct {
	ObjectImpl

	// Private
	_targetPath    *Path
	_targetPointer *Pointer

	// Public
	VariableDivertName string
	PushesToStack      bool
	StackPushType      PushPopType
	IsExternal         bool
	ExternalArgs       int
	IsConditional      bool
}

func (s *Divert) TargetPath() *Path {

	if s._targetPath != nil && s._targetPath.IsRelative() {
		targetObj := s._targetPointer.Resolve()
		if targetObj != nil {
			s._targetPath = targetObj.Path()
		}
	}
	return s._targetPath
}

func (s *Divert) SetTargetPath(value *Path) {
	s._targetPath = value
	s._targetPointer = NullPointer()
}

func (s *Divert) TargetPointer() *Pointer {
	if s._targetPointer.IsNull() {
		targetObj := s.ResolvePath(s._targetPath).Obj
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
	return s.CompactPathString(s.TargetPath())
}

func (s *Divert) SetTargetPathString(value string) {
	if value == "" {
		s.SetTargetPath(nil)
	} else {
		s.SetTargetPath(NewPathFromComponentString(value))
	}
}

func (s *Divert) HasVariableTarget() bool {
	return s.VariableDivertName != ""
}

func NewDivert() *Divert {

	newDivert := new(Divert)
	newDivert.this = newDivert
	newDivert.PushesToStack = true

	return newDivert
}

func NewDivertWithPushPopType(stackPushType PushPopType) *Divert {

	newDivert := new(Divert)
	newDivert.this = newDivert
	newDivert.PushesToStack = true
	newDivert.StackPushType = stackPushType

	return newDivert
}

func (s *Divert) Equals(obj interface{}) bool {

	otherDivert, _ := obj.(*Divert)
	if otherDivert != nil {

		if s.HasVariableTarget() == otherDivert.HasVariableTarget() {
			if s.HasVariableTarget() {
				return s.VariableDivertName == otherDivert.VariableDivertName
			} else {
				return s.TargetPath().Equals(otherDivert.TargetPath())
			}
		}
	}

	return false
}

func (s *Divert) HashCode() string {

	if s.HasVariableTarget() {
		return fmt.Sprintf("HasVariableTarget:%s", s.VariableDivertName)
	} else {
		return fmt.Sprintf("DoesNotHaveVariableTarget:%s", s.VariableDivertName)
	}
}

func (s *Divert) String() string {
	if s.HasVariableTarget() {
		return "Divert(variable: " + s.VariableDivertName + ")"
	} else if s.TargetPath() == nil {
		return "Divert(null)"
	} else {

		var sb strings.Builder

		targetStr := s.TargetPath().String()
		if targetLineNum, ok := s.DebugLineNumberOfPath(s.TargetPath()); ok {
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
}
