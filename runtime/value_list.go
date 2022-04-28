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

		pair := s.Value.MaxItem()
		if pair.Key.IsNil() {
			return NewIntValue(0), nil
		}

		return NewIntValue(pair.Value), nil

	case ValueTypeFloat:

		pair := s.Value.MaxItem()
		if pair.Key.IsNil() {
			return NewFloatValue(0.0), nil
		}
		return NewFloatValue(float64(pair.Value)), nil

	case ValueTypeString:

		pair := s.Value.MaxItem()
		if pair.Key.IsNil() {
			return NewStringValue(""), nil
		}
		return NewStringValue(pair.Key.String()), nil
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
	s := &ListValue{Value: NewInkListFromKeyValuePair(KeyValuePair[InkListItem, int]{singleItem, singleValue})}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
