package runtime

import (
	"fmt"
	"math"
	"strings"
)

const (
	Add      = "+"
	Subtract = "-"
	Divide   = "/"
	Multiply = "*"
	Mod      = "%"
	Negate   = "_" // distinguish from "-" for subtraction

	Equal               = "=="
	Greater             = ">"
	Less                = "<"
	GreaterThanOrEquals = ">="
	LessThanOrEquals    = "<="
	NotEquals           = "!="
	Not                 = "!"

	And = "&&"
	Or  = "||"

	Min = "MIN"
	Max = "MAX"

	Pow     = "POW"
	Floor   = "FLOOR"
	Ceiling = "CEILING"
	Int     = "INT"
	Float   = "FLOAT"

	Has       = "?"
	Hasnt     = "!?"
	Intersect = "^"

	ListMin     = "LIST_MIN"
	ListMax     = "LIST_MAX"
	All         = "LIST_ALL"
	Count       = "LIST_COUNT"
	ValueOfList = "LIST_VALUE"
	Invert      = "LIST_INVERT"
)

type NativeFunctionCall struct {
	ObjectImpl

	// Private

	//
	_name               string
	_isPrototype        bool
	_prototype          *NativeFunctionCall
	_operationFuncs     map[ValueType]interface{}
	_numberOfParameters int
}

func CallExistsWithName(functionName string) bool {

	GenerateNativeFunctionsIfNecessary()

	fmt.Println("FUNCTION_NAME", functionName)

	_, containsKey := _nativeFunctions[functionName]

	fmt.Println(len(_nativeFunctions))

	return containsKey
}

func (s *NativeFunctionCall) Name() string {
	return s._name
}

func (s *NativeFunctionCall) SetName(value string) {

	s._name = value

	if s._isPrototype == false {
		s._prototype = _nativeFunctions[s._name]
	}
}

func (s *NativeFunctionCall) NumberOfParameters() int {

	if s._prototype != nil {

		return s._prototype.NumberOfParameters()
	}

	return s._numberOfParameters
}

func (s *NativeFunctionCall) SetNumberOfParameters(value int) {

	s._numberOfParameters = value
}

func (s *NativeFunctionCall) Call(parameters []Object) Object {

	if s._prototype != nil {

		return s._prototype.Call(parameters)
	}

	if s.NumberOfParameters() != len(parameters) {
		panic("Unexpected number of parameters")
	}

	hasList := false
	for _, p := range parameters {
		switch p.(type) {
		case *Void:
			panic("Attempting to perform operation on a void value. Did you forget to 'return' a value from a function you called here?")
		case *ListValue:
			hasList = true
		}
	}

	// Binary operations on lists are treated outside of the standard coerscion rules
	if len(parameters) == 2 && hasList {

		return s.CallBinaryListOperation(parameters)
	}

	coercedParams := s.CoerceValuesToSingleType(parameters)

	switch coercedParams[0].ValueType() {
	case ValueTypeInt:
		return Call[int](s, coercedParams)
	case ValueTypeFloat:
		return Call[float64](s, coercedParams)
	case ValueTypeString:
		return Call[string](s, coercedParams)
	case ValueTypeDivertTarget:
		return Call[*Path](s, coercedParams)
	case ValueTypeList:
		return Call[*InkList](s, coercedParams)
	}

	return nil
}

func Call[T any](nativeFunctionCall *NativeFunctionCall, parametersOfSingleType []Value) Value {

	param1 := parametersOfSingleType[0].(Value)
	valType := param1.ValueType()

	val1 := param1.(ValueT[T])
	paramCount := len(parametersOfSingleType)

	if paramCount == 2 || paramCount == 1 {

		opForTypeObj, ok := nativeFunctionCall._operationFuncs[valType]
		if !ok {
			// TODO: don't panic
			// Story Exception
			panic("Cannot perform operation '" + nativeFunctionCall.Name() + "' on " + fmt.Sprint(valType))
		}

		// Binary
		if paramCount == 2 {
			param2 := parametersOfSingleType[1]
			val2 := param2.(ValueT[T])
			opForType := opForTypeObj.(BinaryOp[T])
			resultVal := opForType(val1.Value(), val2.Value())
			return CreateValue(resultVal)
		} else {
			// Unary
			opForType := opForTypeObj.(UnaryOp[T])
			resultVal := opForType(val1.Value())
			return CreateValue(resultVal)
		}
	}

	panic("Unexpected number of parameters to NativeFunctionCall: " + fmt.Sprint(len(parametersOfSingleType)))
}

func (s *NativeFunctionCall) CallBinaryListOperation(parameters []Object) Value {

	if s.Name() == "+" || s.Name() == "-" {
		if _, isListValue := parameters[0].(*ListValue); isListValue {
			if _, isIntValue := parameters[1].(*IntValue); isIntValue {
				s.CallListIncrementOperation(parameters)
			}
		}
	}

	v1, _ := parameters[0].(Value) // C# as
	v2, _ := parameters[1].(Value) // C# as

	// And/or with any other type requires coerscion to bool (int)
	if (s.Name() == "&&" || s.Name() == "||") && (v1.ValueType() != ValueTypeList || v2.ValueType() != ValueTypeList) {
		op, _ := s._operationFuncs[ValueTypeInt].(BinaryOp[int]) // C# as

		v1Truthy := 0
		if v1.IsTruthy() {
			v1Truthy = 1
		}

		v2Truthy := 0
		if v2.IsTruthy() {
			v2Truthy = 1
		}

		result := op(v1Truthy, v2Truthy).(bool)
		return NewBoolValueFromBool(result)
	}

	// Normal (list • list) operation
	if v1.ValueType() == ValueTypeList && v2.ValueType() == ValueTypeList {
		return Call[*InkList](s, []Value{v1, v2})
	}

	// TODO: don't panic
	// Story Exception
	panic("Can not call use '" + s.Name() + "' operation on " + fmt.Sprint(v1.ValueType()) + " and " + fmt.Sprint(v2.ValueType()))
}

func (s *NativeFunctionCall) CallListIncrementOperation(listIntParams []Object) Value {

	listVal := listIntParams[0].(*ListValue)
	intVal := listIntParams[1].(*IntValue)

	resultRawList := NewInkList()

	for listItem, listItemValue := range listVal.Value()._items {

		// Find + or - operation
		intOp := s._operationFuncs[ValueTypeInt].(BinaryOp[int])

		// Return value unknown until it's evaluated
		targetInt := intOp(listItemValue, intVal.Value()).(int)

		// Find this item's origin (linear search should be ok, should be short haha)
		var itemOrigin *ListDefinition
		for _, origin := range listVal.Value().Origins {
			if origin.Name() == listItem.OriginName() {
				itemOrigin = origin
				break
			}
		}

		if itemOrigin != nil {
			if incrementedItem, ok := itemOrigin.TryGetItemWithValue(targetInt); ok {
				resultRawList.Set(incrementedItem, targetInt)
			}
		}
	}

	return NewListValueFromList(resultRawList)
}

func (s *NativeFunctionCall) CoerceValuesToSingleType(parametersIn []Object) []Value {

	valType := ValueTypeInt

	var specialCaseList *ListValue

	// Find out what the output type is
	// "higher level" types infect both so that binary operations
	// use the same type on both sides. e.g. binary operation of
	// int and float causes the int to be casted to a float.
	for _, obj := range parametersIn {

		val := obj.(Value)
		if val.ValueType() > valType {
			valType = val.ValueType()
		}

		if val.ValueType() == ValueTypeList {
			specialCaseList, _ = val.(*ListValue) // C# as
		}
	}

	// Coerce to this chosen type
	parametersOut := []Value{}

	// Special case: Coercing to Ints to Lists
	// We have to do it early when we have both parameters
	// to hand - so that we can make use of the List's origin
	if valType == ValueTypeList {

		for _, param := range parametersIn {
			val := param.(Value)
			if val.ValueType() == ValueTypeList {
				parametersOut = append(parametersOut, val)
			} else if val.ValueType() == ValueTypeInt {
				intVal := val.ValueObject().(int)
				list := specialCaseList.Value().OriginOfMaxItem()

				if item, ok := list.TryGetItemWithValue(intVal); ok {
					castedValue := NewListValueFromInkListItem(item, intVal)
					parametersOut = append(parametersOut, castedValue)
				} else {
					// TODO: don't panic
					// Story Exception
					panic("Could not find List item with the value " + fmt.Sprint(intVal) + " in " + list.Name())
				}
			} else {
				// TODO: don't panic
				// Story Exception
				panic("Cannot mix Lists and " + fmt.Sprint(val.ValueType()) + " values in this operation")
			}
		}
	} else {
		for _, param := range parametersIn {
			val := param.(Value)
			castedValue := val.Cast(valType)
			parametersOut = append(parametersOut, castedValue)
		}
	}

	return parametersOut
}

func NewNativeFunctionCallFromName(name string) *NativeFunctionCall {

	GenerateNativeFunctionsIfNecessary()

	newNativeFunctionCall := new(NativeFunctionCall)
	newNativeFunctionCall.SetName(name)

	return newNativeFunctionCall
}

func NewNativeFunctionCallFromParams(name string, numberOfParameters int) *NativeFunctionCall {

	newNativeFunctionCall := new(NativeFunctionCall)
	newNativeFunctionCall._isPrototype = true
	newNativeFunctionCall.SetName(name)
	newNativeFunctionCall.SetNumberOfParameters(numberOfParameters)

	return newNativeFunctionCall
}

// Identity
// For defining operations that do nothing to the specific type
// (but are still supported), such as floor/ceil on int and float
// cast on float.
func Identity[T any](t T) interface{} {
	return t
}

func GenerateNativeFunctionsIfNecessary() {

	if _nativeFunctionInit == false {
		_nativeFunctionInit = true

		// Why no bool operations?
		// Before evaluation, all bools are coerced to ints in
		// CoerceValuesToSingleType (see default value for valType at top).
		// So, no operations are ever directly done in bools themselves.
		// This also means that 1 == true works, since true is always converted
		// to 1 first.
		// However, many operations return a "native" bool (equals, etc).

		// Int operations
		AddIntBinaryOp(Add, func(left int, right int) interface{} { return left + right })
		AddIntBinaryOp(Subtract, func(left int, right int) interface{} { return left - right })
		AddIntBinaryOp(Multiply, func(left int, right int) interface{} { return left * right })
		AddIntBinaryOp(Divide, func(left int, right int) interface{} { return left / right })
		AddIntBinaryOp(Mod, func(left int, right int) interface{} { return left % right })
		AddIntUnaryOp(Negate, func(val int) interface{} { return -val })

		AddIntBinaryOp(Equal, func(left int, right int) interface{} { return left == right })
		AddIntBinaryOp(Greater, func(left int, right int) interface{} { return left > right })
		AddIntBinaryOp(Less, func(left int, right int) interface{} { return left < right })
		AddIntBinaryOp(GreaterThanOrEquals, func(left int, right int) interface{} { return left >= right })
		AddIntBinaryOp(LessThanOrEquals, func(left int, right int) interface{} { return left <= right })
		AddIntBinaryOp(NotEquals, func(left int, right int) interface{} { return left != right })
		AddIntUnaryOp(Not, func(val int) interface{} { return val == 0 })

		AddIntBinaryOp(Add, func(left int, right int) interface{} { return left != 0 && right != 0 })
		AddIntBinaryOp(Or, func(left int, right int) interface{} { return left != 0 || right != 0 })

		AddIntBinaryOp(Max, func(left int, right int) interface{} { return int(math.Max(float64(left), float64(right))) })
		AddIntBinaryOp(Min, func(left int, right int) interface{} { return int(math.Min(float64(left), float64(right))) })

		// C#: Have to cast to float since you could do POW(2, -1), Go: math.Pow already gives back float
		AddIntBinaryOp(Pow, func(left int, right int) interface{} { return math.Pow(float64(left), float64(right)) })
		AddIntUnaryOp(Floor, Identity[int])
		AddIntUnaryOp(Ceiling, Identity[int])
		AddIntUnaryOp(Int, Identity[int])
		AddIntUnaryOp(Float, func(val int) interface{} { return float64(val) })

		// Float operations
		AddFloatBinaryOp(Add, func(left float64, right float64) interface{} { return left + right })
		AddFloatBinaryOp(Subtract, func(left float64, right float64) interface{} { return left - right })
		AddFloatBinaryOp(Multiply, func(left float64, right float64) interface{} { return left * right })
		AddFloatBinaryOp(Divide, func(left float64, right float64) interface{} { return left / right })
		AddFloatBinaryOp(Mod, func(left float64, right float64) interface{} { return math.Mod(left, right) })
		AddFloatUnaryOp(Negate, func(val float64) interface{} { return -val })

		AddFloatBinaryOp(Equal, func(left float64, right float64) interface{} { return left == right })
		AddFloatBinaryOp(Greater, func(left float64, right float64) interface{} { return left > right })
		AddFloatBinaryOp(Less, func(left float64, right float64) interface{} { return left < right })
		AddFloatBinaryOp(GreaterThanOrEquals, func(left float64, right float64) interface{} { return left >= right })
		AddFloatBinaryOp(LessThanOrEquals, func(left float64, right float64) interface{} { return left <= right })
		AddFloatBinaryOp(NotEquals, func(left float64, right float64) interface{} { return left != right })
		AddFloatUnaryOp(Not, func(val float64) interface{} { return val == 0.0 })

		AddFloatBinaryOp(And, func(left float64, right float64) interface{} { return left != 0.0 && right != 0.0 })
		AddFloatBinaryOp(Or, func(left float64, right float64) interface{} { return left != 0.0 || right != 0.0 })

		AddFloatBinaryOp(Max, func(left float64, right float64) interface{} { return math.Max(left, right) })
		AddFloatBinaryOp(Min, func(left float64, right float64) interface{} { return math.Min(left, right) })

		AddFloatBinaryOp(Pow, func(left float64, right float64) interface{} { return math.Pow(left, right) })
		AddFloatUnaryOp(Floor, func(val float64) interface{} { return math.Floor(val) })
		AddFloatUnaryOp(Ceiling, func(val float64) interface{} { return math.Ceil(val) })
		AddFloatUnaryOp(Int, func(val float64) interface{} { return int(val) })
		AddFloatUnaryOp(Float, Identity[float64])

		// String operations
		AddStringBinaryOp(Add, func(left string, right string) interface{} { return left + right })
		AddStringBinaryOp(Equal, func(left string, right string) interface{} { return left == right })
		AddStringBinaryOp(NotEquals, func(left string, right string) interface{} { return left != right })
		AddStringBinaryOp(Has, func(left string, right string) interface{} { return strings.Contains(left, right) })
		AddStringBinaryOp(Hasnt, func(left string, right string) interface{} { return !strings.Contains(left, right) })

		// List operations
		AddListBinaryOp(Add, func(left *InkList, right *InkList) interface{} { return left.Union(right) })
		AddListBinaryOp(Subtract, func(left *InkList, right *InkList) interface{} { return left.Without(right) })
		AddListBinaryOp(Has, func(left *InkList, right *InkList) interface{} { return left.Contains(right) })
		AddListBinaryOp(Hasnt, func(left *InkList, right *InkList) interface{} { return !left.Contains(right) })
		AddListBinaryOp(Intersect, func(left *InkList, right *InkList) interface{} { return left.Intersect(right) })

		AddListBinaryOp(Equal, func(left *InkList, right *InkList) interface{} { return left.Equals(right) })
		AddListBinaryOp(Greater, func(left *InkList, right *InkList) interface{} { return left.GreaterThan(right) })
		AddListBinaryOp(Less, func(left *InkList, right *InkList) interface{} { return left.LessThan(right) })
		AddListBinaryOp(GreaterThanOrEquals, func(left *InkList, right *InkList) interface{} { return left.GreaterThanOrEquals(right) })
		AddListBinaryOp(LessThanOrEquals, func(left *InkList, right *InkList) interface{} { return left.LessThanOrEquals(right) })
		AddListBinaryOp(NotEquals, func(left *InkList, right *InkList) interface{} { return !left.Equals(right) })

		AddListBinaryOp(And, func(left *InkList, right *InkList) interface{} { return left.Count() > 0 && right.Count() > 0 })
		AddListBinaryOp(Or, func(left *InkList, right *InkList) interface{} { return left.Count() > 0 || right.Count() > 0 })

		AddListUnaryOp(Not, func(val *InkList) interface{} {
			if val.Count() == 0 {
				return 1
			} else {
				return 0
			}
		})

		// Placeholders to ensure that these special case functions can exist,
		// since these function is never actually run, and is special cased in Call
		AddListUnaryOp(Invert, func(val *InkList) interface{} { return val.Inverse() })
		AddListUnaryOp(All, func(val *InkList) interface{} { return val.All() })
		AddListUnaryOp(ListMin, func(val *InkList) interface{} { return val.MinAsList() })
		AddListUnaryOp(ListMax, func(val *InkList) interface{} { return val.MaxAsList() })
		AddListUnaryOp(Count, func(val *InkList) interface{} { return val.Count() })
		AddListUnaryOp(ValueOfList, func(val *InkList) interface{} { return val.MaxItem().Value })

		// Special case: The only operations you can do on divert target values
		var divertTargetsEqual BinaryOp[*Path] = func(d1 *Path, d2 *Path) interface{} {
			return d1.Equals(d2)
		}

		var divertTargetsNotEqual BinaryOp[*Path] = func(d1 *Path, d2 *Path) interface{} {
			return !d1.Equals(d2)
		}

		AddOpToNativeFunc(Equal, 2, ValueTypeDivertTarget, divertTargetsEqual)
		AddOpToNativeFunc(NotEquals, 2, ValueTypeDivertTarget, divertTargetsNotEqual)
	}
}

func (s *NativeFunctionCall) AddOpFuncForType(valType ValueType, op interface{}) {

	if s._operationFuncs == nil {
		s._operationFuncs = make(map[ValueType]interface{}, 0)
	}

	s._operationFuncs[valType] = op
}

func AddOpToNativeFunc(name string, args int, valType ValueType, op interface{}) {

	var nativeFunc *NativeFunctionCall
	ok := false
	if nativeFunc, ok = _nativeFunctions[name]; !ok {
		nativeFunc = NewNativeFunctionCallFromParams(name, args)
		_nativeFunctions[name] = nativeFunc
	}

	nativeFunc.AddOpFuncForType(valType, op)
}

func AddIntBinaryOp(name string, op BinaryOp[int]) {
	AddOpToNativeFunc(name, 2, ValueTypeInt, op)
}

func AddIntUnaryOp(name string, op UnaryOp[int]) {
	AddOpToNativeFunc(name, 1, ValueTypeInt, op)
}

func AddFloatBinaryOp(name string, op BinaryOp[float64]) {
	AddOpToNativeFunc(name, 2, ValueTypeFloat, op)
}

func AddStringBinaryOp(name string, op BinaryOp[string]) {
	AddOpToNativeFunc(name, 2, ValueTypeString, op)
}

func AddListBinaryOp(name string, op BinaryOp[*InkList]) {
	AddOpToNativeFunc(name, 2, ValueTypeList, op)
}

func AddListUnaryOp(name string, op UnaryOp[*InkList]) {
	AddOpToNativeFunc(name, 1, ValueTypeList, op)
}

func AddFloatUnaryOp(name string, op UnaryOp[float64]) {
	AddOpToNativeFunc(name, 1, ValueTypeFloat, op)
}

func (s *NativeFunctionCall) String() string {
	return "Native '" + s.Name() + "'"
}

type BinaryOp[T any] func(left T, right T) interface{}

type UnaryOp[T any] func(val T) interface{}

var (
	_nativeFunctions    = map[string]*NativeFunctionCall{}
	_nativeFunctionInit = false
)
