package runtime

import (
	"errors"
)

var _ Value = (*DivertTargetValue)(nil)

type DivertTargetValue struct {
	*objectImpl
	Value *Path
}

func (s *DivertTargetValue) ValueType() ValueType {
	return ValueTypeDivertTarget
}

func (s *DivertTargetValue) IsTruthy() bool {
	panic("Shouldn't be checking the truthiness of a divert target")
}

func (s *DivertTargetValue) ValueObject() interface{} {
	return s.Value
}

func (s *DivertTargetValue) String() string {
	return "DivertTargetValue(" + s.Value.String() + ")"
}

func (s *DivertTargetValue) Copy() interface{} {
	return CreateValue(s.ValueObject())
}

func (s *DivertTargetValue) Cast(newType ValueType) (Value, error) {

	if newType == s.ValueType() {
		return s, nil
	}

	return nil, errors.New("cast to specified value type is not valid for this value")
}

func NewDivertTargetValue(path *Path) *DivertTargetValue {
	s := &DivertTargetValue{Value: path}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
