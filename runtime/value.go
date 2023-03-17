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
	// Bool is new addition, keep enum values the same, with Int==0, Float==1 etc,
	// but for coersion rules, we want to keep bool with a lower value than Int
	// so that it converts in the right direction
	Bool ValueType = iota - 1
	// Used in coersion
	Int
	Float
	List
	String
	// Not used for coersion described above
	DivertTarget
	VariablePointer
)

type Value interface {
	Object
	ValueType() ValueType
	IsTruthy() bool
	Cast(newType ValueType) Value
	ValueObject() interface{}
	Copy() Object
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
	} else if v, ok := val.(int); ok {
		return NewIntValueFromInt(v)
	} else if v, ok := val.(float64); ok {
		return NewFloatValueFromFloat(v)
	} else if v, ok := val.(string); ok {
		return NewStringValueFromString(v)
	} else if v, ok := val.(*Path); ok {
		return NewDivertTargetValueFromPath(v)
	} else if v, ok := val.(*InkList); ok {
		return NewListValueFromList(v)
	}

	return nil
}

type ValueImpl[T any] struct {
	ObjectImpl
	Value T
}

func (s *ValueImpl[T]) ValueObject() interface{} {
	return s.Value
}

func (s *ValueImpl[T]) String() string {
	return fmt.Sprintln(s.Value)
}

func BadCastException(this Value, targetType ValueType) string {
	return fmt.Sprintf("Can't cast %v from %v to %v", this.ValueObject(), this.ValueType(), targetType)
}

var _ Value = (*BoolValue)(nil)
var _ Object = (*BoolValue)(nil)

type BoolValue struct {
	ValueImpl[bool]
}

func (s *BoolValue) ValueType() ValueType {
	return Bool
}

func (s *BoolValue) IsTruthy() bool {
	return s.Value
}

func NewBoolValueFromBool(boolVal bool) *BoolValue {

	newBoolValue := new(BoolValue)
	newBoolValue.Value = boolVal
	newBoolValue.this = newBoolValue

	return newBoolValue
}

func NewBoolValue() *BoolValue {

	newBoolValue := new(BoolValue)
	newBoolValue.Value = false
	newBoolValue.this = newBoolValue

	return newBoolValue
}

func (s *BoolValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	if newType == Int {
		i := 0
		if s.Value {
			i = 1
		}

		return NewIntValueFromInt(i)
	}

	if newType == Float {
		i := 0.0
		if s.Value {
			i = 1.0
		}

		return NewFloatValueFromFloat(i)
	}

	if newType == String {
		str := "false"
		if s.Value {
			str = "true"
		}

		return NewStringValueFromString(str)
	}

	panic(BadCastException(s, newType))
}

func (s *BoolValue) String() string {
	if s.Value {
		return "true"
	} else {
		return "false"
	}
}

func (s *BoolValue) Copy() Object {

	return CreateValue(s.ValueObject())
}

var _ Value = (*IntValue)(nil)
var _ Object = (*IntValue)(nil)

type IntValue struct {
	ValueImpl[int]
}

func (s *IntValue) ValueType() ValueType {
	return Bool
}

func (s *IntValue) IsTruthy() bool {
	return s.Value != 0
}

func NewIntValueFromInt(intVal int) *IntValue {

	newIntValue := new(IntValue)
	newIntValue.Value = intVal
	newIntValue.this = newIntValue

	return newIntValue
}

func NewIntValue() *IntValue {

	newIntValue := new(IntValue)
	newIntValue.Value = 0
	newIntValue.this = newIntValue

	return newIntValue
}

func (s *IntValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	if newType == Bool {
		b := false
		if s.Value == 1 {
			b = true
		}

		return NewBoolValueFromBool(b)
	}

	if newType == Float {
		return NewFloatValueFromFloat(float64(s.Value))
	}

	if newType == String {
		return NewStringValueFromString(fmt.Sprint(s.Value))
	}

	panic(BadCastException(s, newType))
}

var _ Value = (*FloatValue)(nil)
var _ Object = (*FloatValue)(nil)

type FloatValue struct {
	ValueImpl[float64]
}

func (s *FloatValue) ValueType() ValueType {
	return Float
}

func (s *FloatValue) IsTruthy() bool {
	return s.Value != 0.0
}

func NewFloatValueFromFloat(floatVal float64) *FloatValue {

	newFloatValue := new(FloatValue)
	newFloatValue.Value = floatVal
	newFloatValue.this = newFloatValue

	return newFloatValue
}

func NewFloatValue() *FloatValue {

	newFloatValue := new(FloatValue)
	newFloatValue.Value = 0
	newFloatValue.this = newFloatValue

	return newFloatValue
}

func (s *FloatValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	if newType == Bool {
		b := false
		if s.Value == 1 {
			b = true
		}

		return NewBoolValueFromBool(b)
	}

	if newType == Int {
		return NewIntValueFromInt(int(s.Value))
	}

	if newType == String {
		return NewStringValueFromString(fmt.Sprint(s.Value))
	}

	panic(BadCastException(s, newType))
}

var _ Value = (*StringValue)(nil)
var _ Object = (*StringValue)(nil)

type StringValue struct {
	ValueImpl[string]

	// Private
	isNewline          bool
	isInlineWhitespace bool
}

func (s *StringValue) ValueType() ValueType {
	return String
}

func (s *StringValue) IsTruthy() bool {
	return len(s.Value) > 0
}

func (s *StringValue) IsNewline() bool {
	return s.isNewline
}

func (s *StringValue) IsInlineWhitespace() bool {
	return s.isInlineWhitespace
}

func (s *StringValue) IsNonWhitespace() bool {
	return !s.isNewline && !s.isInlineWhitespace
}

func NewStringValueFromString(stringVal string) *StringValue {

	newStringValue := new(StringValue)
	newStringValue.Value = stringVal
	newStringValue.this = newStringValue
	newStringValue.isNewline = stringVal == "\n"
	newStringValue.isInlineWhitespace = true

	for _, c := range stringVal {
		if c != ' ' && c != '\t' {
			newStringValue.isInlineWhitespace = false
			break
		}
	}

	return newStringValue
}

func NewStringValue() *StringValue {

	newStringValue := new(StringValue)
	newStringValue.Value = ""
	newStringValue.this = newStringValue

	return newStringValue
}

func (s *StringValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	if newType == Int {

		if parsedInt, err := strconv.Atoi(s.Value); err == nil {
			return NewIntValueFromInt(parsedInt)
		}

		return nil
	}

	if newType == Float {

		if parsedFloat, err := strconv.ParseFloat(s.Value, 32); err == nil {
			return NewFloatValueFromFloat(parsedFloat)
		}

		return nil
	}

	panic(BadCastException(s, newType))
}

func (s *StringValue) String() string {
	return s.Value
}

var _ Value = (*DivertTargetValue)(nil)
var _ Object = (*DivertTargetValue)(nil)

type DivertTargetValue struct {
	ValueImpl[*Path]

	// Public
	TargetPath *Path
}

func (s *DivertTargetValue) ValueType() ValueType {
	return DivertTarget
}

func (s *DivertTargetValue) IsTruthy() bool {
	panic("Shouldn't be checking the truthiness of a divert target")
}

func NewDivertTargetValueFromPath(targetPath *Path) *DivertTargetValue {

	newDivertTargetValue := new(DivertTargetValue)
	newDivertTargetValue.Value = targetPath
	newDivertTargetValue.this = newDivertTargetValue

	return newDivertTargetValue
}

func NewDivertTargetValue() *DivertTargetValue {

	newDivertTargetValue := new(DivertTargetValue)
	newDivertTargetValue.Value = nil
	newDivertTargetValue.this = newDivertTargetValue

	return newDivertTargetValue
}

func (s *DivertTargetValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	panic(BadCastException(s, newType))
}

func (s *DivertTargetValue) String() string {
	return fmt.Sprintf("DivertTargetValue(%s)", s.TargetPath.String())
}

var _ Value = (*VariablePointerValue)(nil)
var _ Object = (*VariablePointerValue)(nil)

type VariablePointerValue struct {
	ValueImpl[string]

	// Public
	VariableName string
	ContextIndex int
}

func (s *VariablePointerValue) ValueType() ValueType {
	return VariablePointer
}

func (s *VariablePointerValue) IsTruthy() bool {
	panic("Shouldn't be checking the truthiness of a variable pointer")
}

// NewVariablePointerValueFromValue
// (default) contextIndex = -1
func NewVariablePointerValueFromValue(variableName string, contextIndex int) *VariablePointerValue {

	newVariablePointerValue := new(VariablePointerValue)
	newVariablePointerValue.Value = variableName
	newVariablePointerValue.ContextIndex = contextIndex
	newVariablePointerValue.this = newVariablePointerValue

	return newVariablePointerValue
}

func NewVariablePointerValue() *VariablePointerValue {

	newVariablePointerValue := new(VariablePointerValue)
	newVariablePointerValue.Value = ""
	newVariablePointerValue.this = newVariablePointerValue

	return newVariablePointerValue
}

func (s *VariablePointerValue) Cast(newType ValueType) Value {

	if newType == s.ValueType() {
		return s
	}

	panic(BadCastException(s, newType))
}

func (s *VariablePointerValue) Copy() Object {

	return NewVariablePointerValueFromValue(s.VariableName, s.ContextIndex)
}

var _ Value = (*ListValue)(nil)
var _ Object = (*ListValue)(nil)

type ListValue struct {
	ValueImpl[*InkList]
}

func (s *ListValue) ValueType() ValueType {
	return List
}

func (s *ListValue) IsTruthy() bool {
	return s.Value.Length() > 0
}

func (s *ListValue) Cast(newType ValueType) Value {

	if newType == Int {
		k, v := s.Value.MaxItem()
		if k.IsNull() {
			return NewIntValueFromInt(0)
		} else {
			return NewIntValueFromInt(v)
		}
	}

	if newType == Float {
		k, v := s.Value.MaxItem()
		if k.IsNull() {
			return NewFloatValueFromFloat(0.0)
		} else {
			return NewFloatValueFromFloat(float64(v))
		}
	}

	if newType == String {
		k, v := s.Value.MaxItem()
		if k.IsNull() {
			return NewStringValueFromString("")
		} else {
			return NewStringValueFromString(fmt.Sprintf("%d", v))
		}
	}

	if newType == s.ValueType() {
		return s
	}

	panic(BadCastException(s, newType))
}

func NewListValueFromList(list *InkList) *ListValue {

	newListValue := new(ListValue)
	newListValue.Value = list
	newListValue.this = newListValue

	return newListValue
}

func NewListValue() *ListValue {

	newListValue := new(ListValue)
	newListValue.Value = NewInkList()
	newListValue.this = newListValue

	return newListValue
}

func NewListValueFromItem(singleItem *InkListItem, singleValue int) *ListValue {

	newListValue := new(ListValue)
	newListValue.Value = NewInkList()
	newListValue.this = newListValue
	newListValue.Value.Dict[singleItem] = singleValue

	return newListValue
}

func RetainListOriginsForAssignment(oldValue Object, newValue Object) {

	oldList, _ := oldValue.(*ListValue)
	newList, _ := newValue.(*ListValue)

	if oldList != nil && newList != nil && newList.Value.Length() == 0 {
		newList.Value.SetInitialOriginNames(oldList.Value.originNames)
	}
}
