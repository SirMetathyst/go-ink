package runtime

import (
	"errors"
)

var _ Value = (*VariablePointerValue)(nil)

type VariablePointerValue struct {
	*objectImpl
	Value        string
	ContextIndex int
}

func (s *VariablePointerValue) ValueType() ValueType {
	return ValueTypeVariablePointer
}

func (s *VariablePointerValue) IsTruthy() bool {
	panic("Shouldn't be checking the truthiness of a variable pointer")
}

func (s *VariablePointerValue) ValueObject() interface{} {
	return s.Value
}

func (s *VariablePointerValue) String() string {
	return "VariablePointerValue(" + s.Value + ")"
}

func (s *VariablePointerValue) Copy() interface{} {
	return CreateValue(s.ValueObject())
}

func (s *VariablePointerValue) Cast(newType ValueType) (Value, error) {

	if newType == s.ValueType() {
		return s, nil
	}

	return nil, errors.New("cast to specified value type is not valid for this value")
}

func NewVariablePointerValueWithContextIndex(variableName string, contextIndex int) *VariablePointerValue {
	s := &VariablePointerValue{Value: variableName, ContextIndex: contextIndex}
	s.objectImpl = &objectImpl{instance: s}
	return s
}

func NewVariablePointerValue(variableName string) *VariablePointerValue {
	s := &VariablePointerValue{Value: variableName}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
