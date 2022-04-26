package runtime

import (
	"errors"
	"fmt"
)

var _ Value = (*IntValue)(nil)

type IntValue struct {
	*objectImpl
	Value int
}

func (s *IntValue) ValueType() ValueType {
	return ValueTypeInt
}

func (s *IntValue) IsTruthy() bool {
	return s.Value != 0
}

func (s *IntValue) ValueObject() interface{} {
	return s.Value
}

func (s *IntValue) Copy() interface{} {
	return CreateValue(s.ValueObject())
}

func (s *IntValue) Cast(newType ValueType) (Value, error) {
	switch newType {
	case s.ValueType():
		return s, nil
	case ValueTypeBool:
		x := true
		if s.Value == 0 {
			x = false
		}
		return NewBoolValue(x), nil
	case ValueTypeFloat:
		return NewFloatValue(float64(s.Value)), nil
	case ValueTypeString:
		return NewStringValue(fmt.Sprint(s.Value)), nil
	}

	return nil, errors.New("cast to specified value type is not valid for this value")
}

func NewIntValue(value int) *IntValue {
	s := &IntValue{Value: value}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
