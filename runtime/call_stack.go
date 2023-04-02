package runtime

import (
	"fmt"
	"strings"
)

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

type Thread struct {
	// Private
	_callStack []*Element

	// Public
	ThreadIndex     int
	PreviousPointer Pointer
}

func (s *Thread) CallStack() []*Element {

	return s._callStack
}

func (s *Thread) CallStackAdd(value *Element) {
	s._callStack = append(s._callStack, value)
}

func (s *Thread) CallStackRemoveLast() {
	s._callStack = s._callStack[:len(s._callStack)-1]
}

func NewThread() *Thread {

	newThread := new(Thread)
	newThread._callStack = []*Element{}

	return newThread
}

func NewThreadFromJObject(jThreadObj map[string]interface{}, storyContext *Story) *Thread {

	newThread := new(Thread)
	newThread.ThreadIndex = jThreadObj["threadIndex"].(int)
	jThreadCallstack := jThreadObj["callstack"].([]interface{})

	for _, jElTok := range jThreadCallstack {

		jElementObj := jElTok.(map[string]interface{})
		pushPopType := PushPopType(jElementObj["type"].(int))
		pointer := NullPointer

		if currentContainerPathStrToken, ok := jElementObj["cPath"]; ok {

			currentContainerPathStr := currentContainerPathStrToken.(fmt.Stringer).String()
			threadPointerResult := storyContext.ContentAtPath(NewPathFromString(currentContainerPathStr))
			pointer.Container = threadPointerResult.Container()
			pointer.Index = jElementObj["idx"].(int)

			if threadPointerResult.Obj == nil {
				panic("When loading state, internal story location couldn't be found: " + currentContainerPathStr + ". Has the story changed since this save data was created?")
			}

			if threadPointerResult.Approximate {
				storyContext.Warning("When loading state, exact internal story location couldn't be found: '" + currentContainerPathStr + "', so it was approximated to '" + pointer.Container.Path(pointer.Container).String() + "' to recover. Has the story changed since this save data was created?")
			}
		}

		inExpressionEvaluation := jElementObj["exp"].(bool)
		el := NewElement(pushPopType, pointer, inExpressionEvaluation)

		if temps, ok := jElementObj["temp"]; ok {
			el.TemporaryVariables = JObjectToDictionaryRuntimeObjs(temps.(map[string]interface{}))
		} else {
			ClearMap(el.TemporaryVariables)
		}

		newThread.CallStackAdd(el)
	}

	if prevContentObjPath, ok := jThreadObj["previousContentObject"]; ok {
		prevPath := NewPathFromString(prevContentObjPath.(string))
		newThread.PreviousPointer = storyContext.PointerAtPath(prevPath)
	}

	return newThread
}

func (s *Thread) Copy() *Thread {

	threadCopy := NewThread()
	threadCopy.ThreadIndex = s.ThreadIndex
	threadCopy.PreviousPointer = s.PreviousPointer

	for _, e := range s.CallStack() {
		threadCopy.CallStackAdd(e.Copy())
	}

	return threadCopy
}

// TODO: Thread.WriteJson

type CallStack struct {

	// Private
	_threads       []*Thread
	_threadCounter int
	_startOfRoot   Pointer
}

func (s *CallStack) Depth() int {

	return len(s.Elements())
}

func (s *CallStack) CurrentElementIndex() int {

	return len(s.Elements()) - 1
}

func (s *CallStack) Elements() []*Element {

	return s.CurrentThread().CallStack()
}

func (s *CallStack) CurrentElement() *Element {

	thread := s._threads[len(s._threads)-1]
	cs := thread.CallStack()
	return cs[len(cs)-1]
}

func (s *CallStack) CurrentThread() *Thread {

	return s._threads[len(s._threads)-1]
}

func (s *CallStack) SetCurrentThread(value *Thread) {

	// Debug.Assert (_threads.Count == 1, "Shouldn't be directly setting the current thread when we have a stack of them");
	s._threads = append(s._threads, value)
}

func (s *CallStack) CanPop() bool {

	return len(s.Elements()) > 1
}

func NewCallStack(storyContext *Story) *CallStack {

	newCallStack := new(CallStack)
	newCallStack._startOfRoot = StartOfPointer(RootContentContainer(storyContext))
	newCallStack.Reset()

	return newCallStack
}

func NewCallStackFromCallStack(toCopy *CallStack) *CallStack {

	newCallStack := new(CallStack)
	newCallStack._threadCounter = toCopy._threadCounter
	newCallStack._startOfRoot = toCopy._startOfRoot

	for _, otherThread := range toCopy._threads {
		newCallStack._threads = append(newCallStack._threads, otherThread.Copy())
	}

	return newCallStack
}

func (s *CallStack) Reset() {

	s._threads = []*Thread{NewThread()}
	s._threads[0].CallStackAdd(NewElement(Tunnel, s._startOfRoot, false))
}

// SetJsonToken
// Unfortunately it's not possible to implement jsonToken since
// the setter needs to take a Story as a context in order to
// look up objects from paths for currentContainer within elements.
func (s *CallStack) SetJsonToken(jObject map[string]interface{}, storyContext *Story) {

	s._threads = s._threads[:0]

	jThreads := jObject["threads"].([]interface{})

	for _, jThreadTok := range jThreads {

		jThreadObj := jThreadTok.(map[string]interface{})
		thread := NewThreadFromJObject(jThreadObj, storyContext)
		s._threads = append(s._threads, thread)
	}

	s._threadCounter = jObject["threadCounter"].(int)
	s._startOfRoot = StartOfPointer(RootContentContainer(storyContext))
}

// TODO: CallStack.WriteJson

func (s *CallStack) PushThread() {

	newThread := s.CurrentThread().Copy()
	newThread.ThreadIndex = s._threadCounter
	s._threadCounter++
	s._threads = append(s._threads, newThread)
}

func (s *CallStack) ForkThread() *Thread {

	forkedThread := s.CurrentThread().Copy()
	forkedThread.ThreadIndex = s._threadCounter
	s._threadCounter++
	return forkedThread
}

func (s *CallStack) PopThread() {

	if s.CanPopThread() {
		Remove(&s._threads, s.CurrentThread())
		return
	}

	panic("can't pop thread")
}

func (s *CallStack) CanPopThread() bool {

	return len(s._threads) > 1 && s.ElementIsEvaluateFromGame() == false
}

func (s *CallStack) ElementIsEvaluateFromGame() bool {

	return s.CurrentElement().PushPopType() == FunctionEvaluationFromGame
}

// (default) externalEvaluationStackHeight: 0
// (default) outputStreamLengthWithPushed: 0
func (s *CallStack) Push(newType PushPopType, newExternalEvaluationStackHeight int, newOutputStreamLengthWithPushed int) {

	element := NewElement(newType, s.CurrentElement().CurrentPointer, false)
	element.EvaluationStackHeightWhenPushed = newExternalEvaluationStackHeight
	element.FunctionStartInOutputStream = newOutputStreamLengthWithPushed

	//s.Elements().Add(element)
	s.CurrentThread().CallStackAdd(element)
}

func (s *CallStack) CanPopWith(pushPopType PushPopType) bool {

	if s.CanPop() == false {
		return false
	}

	if pushPopType == -1 {
		return true
	}

	return s.CurrentElement().PushPopType() == pushPopType
}

func (s *CallStack) Pop(pushPopType PushPopType) {

	if s.CanPopWith(pushPopType) {

		//s.Elements().RemoveAt(s.Elements().Count() - 1)
		s.CurrentThread().CallStackRemoveLast()

		return
	}

	panic("Mismatched push/pop in Callstack")
}

func (s *CallStack) GetTemporaryVariableWithName(name string, contextIndex int) Object {

	if contextIndex == -1 {

		contextIndex = s.CurrentElementIndex() + 1
	}

	contextElement := s.Elements()[contextIndex-1]

	if varValue, ok := contextElement.TemporaryVariables[name]; ok {

		return varValue
	}

	return nil
}

func (s *CallStack) SetTemporaryVariable(name string, value Object, declareNew bool, contextIndex int) {

	if contextIndex == -1 {

		contextIndex = s.CurrentElementIndex() + 1
	}

	contextElement := s.Elements()[contextIndex-1]

	if _, ok := contextElement.TemporaryVariables[name]; ok && declareNew == false {

		panic("Could not find temporary variable to set: " + name)
	}

	if oldValue, ok := contextElement.TemporaryVariables[name]; ok {

		RetainListOriginsForAssignment(oldValue, value)
	}

	contextElement.TemporaryVariables[name] = value
}

// Find the most appropriate context for this variable.
// Are we referencing a temporary or global variable?
// Note that the compiler will have warned us about possible conflicts,
// so anything that happens here should be safe!
func (s *CallStack) ContextForVariableNamed(name string) int {

	// Current temporary context?
	// (Shouldn't attempt to access contexts higher in the callstack.)
	if _, ok := s.CurrentElement().TemporaryVariables[name]; ok {

		return s.CurrentElementIndex() + 1
	}

	return 0
}

func (s *CallStack) ThreadWithIndex(index int) *Thread {

	for _, t := range s._threads {
		if t.ThreadIndex == index {
			return t
		}
	}

	return nil
}

func (s *CallStack) CallStackTrace() string {
	var sb strings.Builder

	for t := 0; t < len(s._threads); t++ {

		thread := s._threads[t]
		isCurrent := t == len(s._threads)-1
		isCurrentStr := ""
		if isCurrent {
			isCurrentStr = "(current)"
		}
		sb.WriteString(fmt.Sprintf("=== THREAD %d/%d %s===\n", t+1, len(s._threads), isCurrentStr))

		for i := 0; i < len(thread.CallStack()); i++ {

			if thread.CallStack()[i].PushPopType() == Function {
				sb.WriteString("  [FUNCTION] \n")
			} else {
				sb.WriteString("  [TUNNEL] \n")
			}

			pointer := thread.CallStack()[i].CurrentPointer
			if !pointer.IsNull() {
				sb.WriteString("<SOMEWHERE IN \n")
				sb.WriteString(pointer.Container.Path(pointer.Container).String())
				sb.WriteString(">\n")
			}
		}
	}

	return sb.String()
}
