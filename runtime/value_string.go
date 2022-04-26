package runtime

import (
	"errors"
	"strconv"
)

var _ Value = (*StringValue)(nil)

type StringValue struct {
	*objectImpl
	Value              string
	isNewLine          bool
	isInlineWhitespace bool
}

func (s *StringValue) IsNonWhitespace() bool {
	return !s.isNewLine && !s.isInlineWhitespace
}

func (s *StringValue) ValueType() ValueType {
	return ValueTypeString
}

func (s *StringValue) IsTruthy() bool {
	return len(s.Value) > 0
}

func (s *StringValue) ValueObject() interface{} {
	return s.Value
}

func (s *StringValue) Copy() interface{} {
	return CreateValue(s.ValueObject())
}

func (s *StringValue) Cast(newType ValueType) (Value, error) {
	switch newType {
	case s.ValueType():
		return s, nil
	case ValueTypeInt:

		x, err := strconv.Atoi(s.Value)
		if err != nil {
			return nil, nil
		}

		return NewIntValue(x), nil
	case ValueTypeFloat:

		x, err := strconv.ParseFloat(s.Value, 64)
		if err != nil {
			return nil, nil
		}

		return NewFloatValue(x), nil
	}

	return nil, errors.New("cast to specified value type is not valid for this value")
}

func NewStringValue(value string) *StringValue {

	isNewLine := value == "\n"
	isInlineWhitespace := true

	for _, c := range value {
		if c != ' ' && c != '\t' {
			isInlineWhitespace = false
			break
		}
	}

	s := &StringValue{Value: value, isNewLine: isNewLine, isInlineWhitespace: isInlineWhitespace}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
