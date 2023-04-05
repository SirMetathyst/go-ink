package runtime

type Element struct {

	// Private
	_pushPopType PushPopType

	// Public
	CurrentPointer         Pointer
	InExpressionEvaluation bool
	TemporaryVariables     map[string]Object

	// When this callstack element is actually a function evaluation called from the game,
	// we need to keep track of the size of the evaluation stack when it was called
	// so that we know whether there was any return value.
	EvaluationStackHeightWhenPushed int

	// When functions are called, we trim whitespace from the start and end of what
	// they generate, so we make sure know where the function's start and end are.
	FunctionStartInOutputStream int
}

func (s *Element) PushPopType() PushPopType {

	return s._pushPopType
}

func NewElement(newPushPopType PushPopType, newPointer Pointer, newInExpressionEvaluation bool) *Element {

	newElement := new(Element)
	newElement.CurrentPointer = newPointer
	newElement.InExpressionEvaluation = newInExpressionEvaluation
	newElement.TemporaryVariables = make(map[string]Object, 0)
	newElement._pushPopType = newPushPopType

	return newElement
}

func (s *Element) Copy() *Element {

	elementCopy := NewElement(s.PushPopType(), s.CurrentPointer, s.InExpressionEvaluation)
	elementCopy.TemporaryVariables = NewMapFromMap(s.TemporaryVariables)
	elementCopy.EvaluationStackHeightWhenPushed = s.EvaluationStackHeightWhenPushed
	elementCopy.FunctionStartInOutputStream = s.FunctionStartInOutputStream

	return elementCopy
}
