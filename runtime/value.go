package runtime

type ValueType int

const (
	ValueTypeBool ValueType = iota - 1
	ValueTypeInt
	ValueTypeFloat
	ValueTypeList
	ValueTypeString
	ValueTypeDivertTarget
	ValueTypeVariablePointer
)

type Value interface {
	ValueType() ValueType
	IsTruthy() bool
	Cast(newType ValueType) (Value, error)
	ValueObject() interface{}
	Copy() interface{}
}

func CreateValue(value interface{}) Value {

	switch v := value.(type) {
	case bool:
		return &BoolValue{Value: v}
	case int:
		return &IntValue{Value: v}
	case int64: // C# long
		return &IntValue{Value: int(v)}
	case float64:
		return &FloatValue{Value: v}
	case string:
		return &StringValue{Value: v}
	case *Path:
		return &DivertTargetValue{Value: v}
	case *InkList:
		return &ListValue{Value: v}
	}

	return nil
}
