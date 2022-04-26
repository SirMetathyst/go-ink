package runtime

import (
	"errors"
)

var _ Value = (*ListValue)(nil)

type ListValue struct {
	*objectImpl
	Value *InkList
}

func (s *ListValue) ValueType() ValueType {
	return ValueTypeList
}

func (s *ListValue) IsTruthy() bool {
	return s.Value.Length() > 0
}

func (s *ListValue) ValueObject() interface{} {
	return s.Value
}

func (s *ListValue) Copy() interface{} {
	return CreateValue(s.ValueObject())
}

func (s *ListValue) Cast(newType ValueType) (Value, error) {

	switch newType {
	case s.ValueType():
		return s, nil
	case ValueTypeInt:

		key, value := s.Value.MaxItem()
		if key.IsNil() {
			return NewIntValue(0), nil
		}

		return NewIntValue(value), nil

	case ValueTypeFloat:

		key, value := s.Value.MaxItem()
		if key.IsNil() {
			return NewFloatValue(0.0), nil
		}
		return NewFloatValue(float64(value)), nil

	case ValueTypeString:

		key, value := s.Value.MaxItem()
		if key.IsNil() {
			return NewStringValue(""), nil
		}
		return NewStringValue(key.String()), nil
	}

	return nil, errors.New("cast to specified value type is not valid for this value")
}

func RetainListOriginsForAssignment(oldValue Object, newValue Object) {

	oldList, _ := oldValue.(*ListValue)
	newList, _ := newValue.(*ListValue)

	if oldList != nil && newList != nil && newList.Value.Length() == 0 {
		newList.Value.SetInitialOriginNames(oldList.Value.OriginNames())
	}
}

func NewListValue() *ListValue {
	s := &ListValue{}
	s.objectImpl = &objectImpl{instance: s}
	return s
}

func NewListValueFromInkList(inkList *InkList) *ListValue {
	s := &ListValue{Value: inkList}
	s.objectImpl = &objectImpl{instance: s}
	return s
}

func NewListValueFromSingleInkListItem(singleItem InkListItem, singleValue int) *ListValue {
	s := &ListValue{Value: NewInkListFromSingleItem(singleItem, singleValue)}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
