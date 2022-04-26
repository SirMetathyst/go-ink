package runtime

import (
	"errors"
	"fmt"
)

var _ Value = (*FloatValue)(nil)

type FloatValue struct {
	*objectImpl
	Value float64
}

func (s *FloatValue) ValueType() ValueType {
	return ValueTypeFloat
}

func (s *FloatValue) IsTruthy() bool {
	return s.Value != 0.0
}

func (s *FloatValue) ValueObject() interface{} {
	return s.Value
}

func (s *FloatValue) Copy() interface{} {
	return CreateValue(s.ValueObject())
}

func (s *FloatValue) Cast(newType ValueType) (Value, error) {
	switch newType {
	case s.ValueType():
		return s, nil
	case ValueTypeBool:
		x := true
		if s.Value == 0.0 {
			x = false
		}
		return NewBoolValue(x), nil
	case ValueTypeInt:
		return NewIntValue(int(s.Value)), nil
	case ValueTypeString:
		return NewStringValue(fmt.Sprint(s.Value)), nil
	}

	return nil, errors.New("cast to specified value type is not valid for this value")
}

func NewFloatValue(value float64) *FloatValue {
	s := &FloatValue{Value: value}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
