package runtime

import (
	"fmt"
	"strconv"
)

// ValueType
// Order is significant for type coersion.
// If types aren't directly compatible for an operation,
// they're coerced to the same type, downward.
// Higher value types "infect" an operation.
// (This may not be the most sensible thing to do, but it's worked so far!)
type ValueType int

const (
	// ValueTypeBool is new addition, keep enum values the same, with Int==0, Float==1 etc,
	// but for coersion rules, we want to keep bool with a lower value than Int
	// so that it converts in the right direction
	ValueTypeBool ValueType = iota - 1
	// Used in coersion
	ValueTypeInt
	ValueTypeFloat
	ValueTypeList
	ValueTypeString
	// Not used for coersion described above
	ValueTypeDivertTarget
	ValueTypeVariablePointer
)

type Value interface {
	Object
	ValueType() ValueType
	IsTruthy() bool
	Cast(newType ValueType) Value
	ValueObject() interface{}
	Copy() Object
	String() string
}

type ValueT[T any] interface {
	Value
	Value() T
}

// TODO:

func CreateValue(val interface{}) Value {

	// Implicitly lose precision from any doubles we get passed in
	//if (val is double) {
	//	double doub = (double)val;
	//	val = (float)doub;
	//}
	//
	//else if (val is long) {
	//	return new IntValue ((int)(long)val);
	//}
	//

	// else if (val is double) {
	//		return new FloatValue ((float)(double)val);
	//	}

	if v, ok := val.(bool); ok {
		return NewBoolValueFromBool(v)
	}

	if v, ok := val.(int); ok {
		return NewIntValueFromInt(v)
	}

	if v, ok := val.(float64); ok {
		return NewFloatValueFromFloat(v)
	}

	if v, ok := val.(string); ok {
		return NewStringValueFromString(v)
	}

	if v, ok := val.(*Path); ok {
		return NewDivertTargetValueFromPath(v)
	}

	if v, ok := val.(*InkList); ok {
		return NewListValueFromList(v)
	}

	return nil
}

func BadCastException(this Value, targetType ValueType) string {

	return fmt.Sprintf("Can't cast %v from %v to %v", fmt.Sprint(this.ValueObject()), fmt.Sprint(this.ValueType()), targetType)
}

var _ ValueT[bool] = (*BoolValue)(nil)
var _ Object = (*BoolValue)(nil)

type BoolValue struct {
	ObjectImpl
	_value bool
}

func NewBoolValueFromBool(boolVal bool) *BoolValue {

	newBoolValue := new(BoolValue)
	newBoolValue._value = boolVal

	return newBoolValue
}

func (s *BoolValue) ValueType() ValueType {
	return ValueTypeBool
}

func (s *BoolValue) IsTruthy() bool {
	return s._value
}

func (s *BoolValue) ValueObject() interface{} {
	return s._value
}

func (s *BoolValue) Value() bool {
	return s._value
}

func (s *BoolValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	if newType == ValueTypeInt {
		i := 0
		if s._value {
			i = 1
		}

		return NewIntValueFromInt(i)
	}

	if newType == ValueTypeFloat {
		i := 0.0
		if s._value {
			i = 1.0
		}

		return NewFloatValueFromFloat(i)
	}

	if newType == ValueTypeString {
		str := "false"
		if s._value {
			str = "true"
		}

		return NewStringValueFromString(str)
	}

	panic(BadCastException(s, newType))
}

func (s *BoolValue) Copy() Object {
	return CreateValue(s._value)
}

func (s *BoolValue) String() string {

	if s._value {
		return "true"
	} else {
		return "false"
	}
}

var _ ValueT[int] = (*IntValue)(nil)
var _ Object = (*IntValue)(nil)

type IntValue struct {
	ObjectImpl
	_value int
}

func NewIntValueFromInt(intVal int) *IntValue {

	newIntValue := new(IntValue)
	newIntValue._value = intVal

	return newIntValue
}

func (s *IntValue) ValueType() ValueType {
	return ValueTypeInt
}

func (s *IntValue) IsTruthy() bool {
	return s._value != 0
}

func (s *IntValue) ValueObject() interface{} {
	return s._value
}

func (s *IntValue) Value() int {
	return s._value
}

func (s *IntValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	if newType == ValueTypeBool {
		return NewBoolValueFromBool(s._value != 0)
	}

	if newType == ValueTypeFloat {
		return NewFloatValueFromFloat(float64(s._value))
	}

	if newType == ValueTypeString {
		return NewStringValueFromString(fmt.Sprint(s._value))
	}

	panic(BadCastException(s, newType))
}

func (s *IntValue) Copy() Object {
	return CreateValue(s._value)
}

func (s *IntValue) String() string {
	return fmt.Sprint(s._value)
}

var _ ValueT[float64] = (*FloatValue)(nil)
var _ Object = (*FloatValue)(nil)

type FloatValue struct {
	ObjectImpl
	_value float64
}

func NewFloatValueFromFloat(floatVal float64) *FloatValue {

	newFloatValue := new(FloatValue)
	newFloatValue._value = floatVal

	return newFloatValue
}

func (s *FloatValue) ValueType() ValueType {
	return ValueTypeFloat
}

func (s *FloatValue) IsTruthy() bool {
	return s._value != 0.0
}

func (s *FloatValue) ValueObject() interface{} {
	return s._value
}

func (s *FloatValue) Value() float64 {
	return s._value
}

func (s *FloatValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	if newType == ValueTypeBool {
		return NewBoolValueFromBool(s._value != 0.0)
	}

	if newType == ValueTypeInt {
		return NewIntValueFromInt(int(s._value))
	}

	if newType == ValueTypeString {
		return NewStringValueFromString(fmt.Sprint(s._value))
	}

	panic(BadCastException(s, newType))
}

func (s *FloatValue) Copy() Object {
	return CreateValue(s._value)
}

func (s *FloatValue) String() string {
	return fmt.Sprint(s._value)
}

var _ ValueT[string] = (*StringValue)(nil)
var _ Object = (*StringValue)(nil)

type StringValue struct {
	ObjectImpl

	// Private
	_value              string
	_isNewline          bool
	_isInlineWhitespace bool
}

func NewStringValueFromString(str string) *StringValue {

	newStringValue := new(StringValue)
	newStringValue._value = str
	newStringValue._isNewline = str == "\n"
	newStringValue._isInlineWhitespace = true

	for _, c := range str {
		if c != ' ' && c != '\t' {
			newStringValue._isInlineWhitespace = false
			break
		}
	}

	return newStringValue
}

func (s *StringValue) IsNewline() bool {
	return s._isNewline
}

func (s *StringValue) IsInlineWhitespace() bool {
	return s._isInlineWhitespace
}

func (s *StringValue) IsNonWhitespace() bool {
	return !s._isNewline && !s._isInlineWhitespace
}

func (s *StringValue) ValueType() ValueType {
	return ValueTypeString
}

func (s *StringValue) IsTruthy() bool {
	return len(s._value) > 0
}

func (s *StringValue) ValueObject() interface{} {
	return s._value
}

func (s *StringValue) Value() string {
	return s._value
}

func (s *StringValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	if newType == ValueTypeInt {

		if parsedInt, err := strconv.Atoi(s._value); err == nil {
			return NewIntValueFromInt(parsedInt)
		}

		return nil
	}

	if newType == ValueTypeFloat {

		if parsedFloat, err := strconv.ParseFloat(s._value, 32); err == nil {
			return NewFloatValueFromFloat(parsedFloat)
		}

		return nil
	}

	panic(BadCastException(s, newType))
}

func (s *StringValue) Copy() Object {
	return CreateValue(s._value)
}

func (s *StringValue) String() string {
	return s._value
}

var _ ValueT[*Path] = (*DivertTargetValue)(nil)
var _ Object = (*DivertTargetValue)(nil)

type DivertTargetValue struct {
	ObjectImpl

	// Private
	_value *Path
}

func NewDivertTargetValueFromPath(targetPath *Path) *DivertTargetValue {

	newDivertTargetValue := new(DivertTargetValue)
	newDivertTargetValue._value = targetPath

	return newDivertTargetValue
}

func (s *DivertTargetValue) TargetPath() *Path {
	return s._value
}

func (s *DivertTargetValue) SetTargetPath(value *Path) {
	s._value = value
}

func (s *DivertTargetValue) ValueType() ValueType {
	return ValueTypeDivertTarget
}

func (s *DivertTargetValue) IsTruthy() bool {
	panic("Shouldn't be checking the truthiness of a divert target")
}

func (s *DivertTargetValue) ValueObject() interface{} {
	return s._value
}

func (s *DivertTargetValue) Value() *Path {
	return s._value
}

func (s *DivertTargetValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	panic(BadCastException(s, newType))
}

func (s *DivertTargetValue) Copy() Object {
	return CreateValue(s._value)
}

func (s *DivertTargetValue) String() string {
	return fmt.Sprintf("DivertTargetValue(%s)", s.TargetPath().String())
}

var _ ValueT[string] = (*VariablePointerValue)(nil)
var _ Object = (*VariablePointerValue)(nil)

type VariablePointerValue struct {
	ObjectImpl

	// Public
	_value        string
	_contextIndex int
}

func NewVariablePointerValueFromValue(variableName string, contextIndex int) *VariablePointerValue {

	newVariablePointerValue := new(VariablePointerValue)
	newVariablePointerValue._value = variableName
	newVariablePointerValue.SetContextIndex(contextIndex)

	return newVariablePointerValue
}

func (s *VariablePointerValue) ContextIndex() int {
	return s._contextIndex
}

func (s *VariablePointerValue) SetContextIndex(value int) {
	s._contextIndex = value
}

func (s *VariablePointerValue) ValueType() ValueType {
	return ValueTypeVariablePointer
}

func (s *VariablePointerValue) IsTruthy() bool {
	panic("Shouldn't be checking the truthiness of a variable pointer")
}

func (s *VariablePointerValue) ValueObject() interface{} {
	return s._value
}

func (s *VariablePointerValue) Value() string {
	return s._value
}

func (s *VariablePointerValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	panic(BadCastException(s, newType))
}

func (s *VariablePointerValue) Copy() Object {

	return NewVariablePointerValueFromValue(s._value, s._contextIndex)
}

func (s *VariablePointerValue) String() string {
	return "VariablePointerValue(" + s._value + ")"
}

var _ ValueT[*InkList] = (*ListValue)(nil)
var _ Object = (*ListValue)(nil)

type ListValue struct {
	ObjectImpl

	// Private
	_value *InkList
}

func NewListValueFromInkListItem(singleItem InkListItem, singleValue int) *ListValue {

	newListValue := new(ListValue)
	newListValue._value = NewInkListFromSingleElement(KeyValuePair[InkListItem, int]{singleItem, singleValue})

	return newListValue
}

func NewListValueFromList(list *InkList) *ListValue {

	newListValue := new(ListValue)
	newListValue._value = list

	return newListValue
}

func NewListValue() *ListValue {

	newListValue := new(ListValue)
	newListValue._value = NewInkList()

	return newListValue
}

func (s *ListValue) ValueType() ValueType {
	return ValueTypeList
}

func (s *ListValue) IsTruthy() bool {
	return s._value.Count() > 0
}

func (s *ListValue) ValueObject() interface{} {
	return s._value
}

func (s *ListValue) Value() *InkList {
	return s._value
}

func (s *ListValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	if newType == ValueTypeInt {

		max := s._value.MaxItem()
		if max.Key.IsNull() {
			return NewIntValueFromInt(0)
		}

		return NewIntValueFromInt(max.Value)
	}

	if newType == ValueTypeFloat {

		max := s._value.MaxItem()
		if max.Key.IsNull() {
			return NewFloatValueFromFloat(0.0)
		}

		return NewFloatValueFromFloat(float64(max.Value))
	}

	if newType == ValueTypeString {

		max := s._value.MaxItem()
		if max.Key.IsNull() {
			return NewStringValueFromString("")
		}

		return NewStringValueFromString(fmt.Sprintf("%d", max.Value))
	}

	panic(BadCastException(s, newType))
}

func (s *ListValue) Copy() Object {
	return CreateValue(s._value)
}

func (s *ListValue) String() string {
	return s._value.String()
}

func RetainListOriginsForAssignment(oldValue Object, newValue Object) {

	oldList, _ := oldValue.(*ListValue)
	newList, _ := newValue.(*ListValue)

	if oldList != nil && newList != nil && newList._value.Count() == 0 {
		newList._value.SetInitialOriginNames(oldList._value.OriginNames())
	}
}
