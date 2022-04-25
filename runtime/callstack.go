package runtime

import (
	"fmt"
	"strings"
)

type CallStack struct {
	elements       []*Element
	currentElement *Element
	threads        []*Thread
	threadCounter  int
	startOfRoot    *Pointer
}

func (s *CallStack) Depth() int {
	return len(s.elements)
}

func (s *CallStack) CurrentElement() *Element {
	thread := s.threads[len(s.threads)-1]
	cs := thread.Callstack
	return cs[len(cs)-1]
}

func (s *CallStack) CurrentThread() *Thread {
	return s.threads[len(s.threads)-1]
}

func (s *CallStack) SetCurrentThread(value *Thread) {

	if len(s.threads) == 1 {
		panic("Shouldn't be directly setting the current thread when we have a stack of them")
	}

	s.threads = s.threads[:0]
	s.threads = append(s.threads, value)
}

func (s *CallStack) CallStack() []*Element {
	return s.CurrentThread().Callstack
}

func (s *CallStack) CurrentElementIndex() int {
	return len(s.CallStack()) - 1
}

func (s *CallStack) CanPop() bool {
	return len(s.CallStack()) > 1
}

func (s *CallStack) Reset() {
	s.threads = s.threads[:0]
	s.threads = append(s.threads, NewThread())
	s.threads[0].Callstack = append(s.threads[0].Callstack, NewElement(int(PushPopTunnel), s.startOfRoot))
}

func (s *CallStack) ElementIsEvaluateFromGame() bool {
	return s.CurrentElement().Type == int(PushPopFunctionEvaluationFromGame)
}

func (s *CallStack) CanPopThread() bool {
	return len(s.threads) > 1 && !s.ElementIsEvaluateFromGame()
}

func (s *CallStack) SetJsonToken(jObject map[string]interface{}, storyContext *Story) {

	s.threads = s.threads[:0]

	jThreads := jObject["threads"].([]interface{})

	for _, jThreadTok := range jThreads {
		jThreadObj := jThreadTok.(map[string]interface{})
		thread := NewThreadFromMap(jThreadObj, storyContext)
		s.threads = append(s.threads, thread)
	}

	s.threadCounter = jObject["threadCounter"].(int)
	s.startOfRoot = NewPointerStartOf(storyContext.RootContentContainer())
}

func (s *CallStack) PushThread() {
	newThread := s.CurrentThread().Copy()
	s.threadCounter++
	newThread.ThreadIndex = s.threadCounter
	s.threads = append(s.threads, newThread)
}

func (s *CallStack) ForkThread() *Thread {
	forkedThread := s.CurrentThread().Copy()
	s.threadCounter++
	forkedThread.ThreadIndex = s.threadCounter
	return forkedThread
}

func (s *CallStack) PopThread() {
	if s.CanPopThread() {
		removeIndex := -1
		for i, elm := range s.threads {
			if elm == s.CurrentThread() {
				removeIndex = i
			}
		}
		if removeIndex != -1 {
			s.threads = append(s.threads[:removeIndex], s.threads[removeIndex+1:]...)
		}
	}
}

func (s *CallStack) PushWith(pushPopType PushPopType, externalEvaluationStackHeight int, outputStreamLengthWithPushed int) {

	element := NewElementWith(int(pushPopType), s.CurrentElement().CurrentPointer, false)
	element.EvaluationStackHeightWhenPushed = externalEvaluationStackHeight
	element.FunctionStartInOuputStream = outputStreamLengthWithPushed
	cs := s.CallStack()
	cs = append(cs, element)
}

func (s *CallStack) Push(pushPopType PushPopType) {
	s.PushWith(pushPopType, 0, 0)
}

func (s *CallStack) CanPopWith(pushPopType PushPopType) bool {
	if !s.CanPop() {
		return false
	}

	if int(pushPopType) < 0 {
		return true
	}

	return s.CurrentElement().Type == int(pushPopType)
}

func (s *CallStack) Pop(pushPopType PushPopType) {
	if s.CanPopWith(pushPopType) {
		cs := s.CallStack()
		cs = cs[:len(cs)-1]
		return
	}
	panic("Mismatched push/pop in Callstack")
}

func (s *CallStack) GetTemporaryVariableWithNameAndIndex(name string, contextIndex int) Object {

	if contextIndex == -1 {
		contextIndex = s.CurrentElementIndex() + 1
	}

	contextElement := s.CallStack()[contextIndex-1]

	if varValue, ok := contextElement.TemporaryVariables[name]; ok {
		return varValue
	}

	return nil
}

func (s *CallStack) GetTemporaryVariableWithName(name string) Object {
	return s.GetTemporaryVariableWithNameAndIndex(name, -1)
}

func (s *CallStack) SetTemporaryVariableWithContextIndex(name string, value Object, declareNew bool, contextIndex int) {

	if contextIndex == -1 {
		contextIndex = s.CurrentElementIndex() + 1
	}

	contextElement := s.CallStack()[contextIndex-1]

	if _, ok := contextElement.TemporaryVariables[name]; !ok && !declareNew {
		panic("Could not find temporary variable to set: " + name)
	}

	if oldValue, ok := contextElement.TemporaryVariables[name]; ok {
		RetainListOriginsForAssignment(oldValue, value)
	}

	contextElement.TemporaryVariables[name] = value
}

func (s *CallStack) SetTemporaryVariable(name string, value Object, declareNew bool, contextIndex int) {
	s.SetTemporaryVariableWithContextIndex(name, value, declareNew, -1)
}

func (s *CallStack) ContextForVariableNamed(name string) int {

	if _, ok := s.currentElement.TemporaryVariables[name]; ok {
		return s.CurrentElementIndex() + 1
	}

	return 0
}

func (s *CallStack) ThreadWithIndex(index int) *Thread {
	for _, thread := range s.threads {
		if thread.ThreadIndex == index {
			return thread
		}
	}
	return NewThread()
}

func (s *CallStack) CallStackTrace() string {

	sb := strings.Builder{}

	for t := 0; t < len(s.threads); t++ {

		thread := s.threads[t]
		isCurrent := t == len(s.threads)-1
		isCurrentStr := ""
		if isCurrent {
			isCurrentStr = "(current)"
		}
		sb.WriteString(fmt.Sprintf("=== THREAD %d/%d %s===\n", t+1, len(s.threads), isCurrentStr))

		for i := 0; i < len(thread.Callstack); i++ {

			if thread.Callstack[i].Type == int(PushPopFunction) {
				sb.WriteString("  [FUNCTION")
			} else {
				sb.WriteString("  [TUNNEL]")
			}

			pointer := thread.Callstack[i].CurrentPointer
			if !pointer.IsNil() {
				sb.WriteString("<SOMEWHERE IN ")
				sb.WriteString(pointer.Container.Path().String())
				sb.WriteString(">")
			}
		}
	}

	return sb.String()
}

func NewCallStack(storyContext *Story) *CallStack {
	startOfRoot := NewPointerStartOf(storyContext.RootContentContainer())
	cs := &CallStack{startOfRoot: startOfRoot}
	cs.Reset()
	return cs
}
