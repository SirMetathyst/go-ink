package runtime

import "errors"

var _ Value = (*BoolValue)(nil)

type BoolValue struct {
	*objectImpl
	Value bool
}

func (s *BoolValue) ValueType() ValueType {
	return ValueTypeBool
}

func (s *BoolValue) IsTruthy() bool {
	return s.Value
}

func (s *BoolValue) ValueObject() interface{} {
	return s.Value
}

func (s *BoolValue) String() string {
	if s.Value {
		return "true"
	}
	return "false"
}

func (s *BoolValue) Copy() interface{} {
	return CreateValue(s.ValueObject())
}

func (s *BoolValue) Cast(newType ValueType) (Value, error) {
	switch newType {
	case s.ValueType():
		return s, nil
	case ValueTypeInt:
		x := 0
		if s.Value == true {
			x = 1
		}
		return NewIntValue(x), nil
	case ValueTypeFloat:
		x := 0.0
		if s.Value == true {
			x = 1.0
		}
		return NewFloatValue(x), nil
	case ValueTypeString:
		x := "false"
		if s.Value == true {
			x = "true"
		}
		return NewStringValue(x), nil
	}

	return nil, errors.New("cast to specified value type is not valid for this value")
}

func NewBoolValue(value bool) *BoolValue {
	s := &BoolValue{Value: value}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
