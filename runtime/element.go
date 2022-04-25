package runtime

type Element struct {
	CurrentPointer                  *Pointer
	InExpressionEvaluation          bool
	TemporaryVariables              map[string]Object
	Type                            int
	EvaluationStackHeightWhenPushed int
	FunctionStartInOuputStream      int
}

func (s *Element) Copy() *Element {
	elm := NewElementWith(s.Type, s.CurrentPointer, s.InExpressionEvaluation)
	newTemporaryVariables := map[string]Object{}
	for k, v := range s.TemporaryVariables {
		newTemporaryVariables[k] = v
	}
	elm.TemporaryVariables = newTemporaryVariables
	elm.EvaluationStackHeightWhenPushed = s.EvaluationStackHeightWhenPushed
	elm.FunctionStartInOuputStream = s.FunctionStartInOuputStream
	return elm
}

func NewElementWith(pushPopType int, pointer *Pointer, inExpressionEvaluation bool) *Element {
	return &Element{CurrentPointer: pointer, InExpressionEvaluation: inExpressionEvaluation, Type: pushPopType, TemporaryVariables: map[string]Object{}}
}

func NewElement(pushPopType int, pointer *Pointer) *Element {
	return &Element{CurrentPointer: pointer, Type: pushPopType, TemporaryVariables: map[string]Object{}}

}
