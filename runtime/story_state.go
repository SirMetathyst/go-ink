package runtime

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// Private

//
const kDefaultFlowName = "DEFAULT_FLOW"

// Backward compatible changes since v8:
// v10: dynamic tags
// v9:  multi-flows
const kMinCompatibleLoadVersion = 8

// Public

// Backward compatible changes since v8:
// v10: dynamic tags
// v9:  multi-flows
const KInkSaveStateVersion = 10

type Event[T any] struct {
	h []T
}

func (s *Event[T]) Register(v T) {
	s.h = append(s.h, v)
}

type Action func()

type ActionEvent struct {
	Event[Action]
}

func (s *ActionEvent) Emit() {
	for _, fn := range s.h {
		fn()
	}
}

type ActionT1[T any] func(t T)

type ActionT1Event[T any] struct {
	Event[ActionT1[T]]
}

func (s *ActionT1Event[T]) Emit(v T) {
	for _, fn := range s.h {
		fn(v)
	}
}

// StoryState
// contains All story state information is included in the StoryState class,
// including global variables, read counts, the pointer to the current
// point in the story, the call stack (for tunnels, functions, etc),
// and a few other smaller bits and pieces. You can save the current
// state using the json serialisation functions ToJson and LoadJson.
type StoryState struct {

	// Private
	_currentText           string
	_visitCounts           map[string]int
	_turnIndices           map[string]int
	_outputStreamTextDirty bool // must be set true in constructor
	_outputStreamTagsDirty bool // must be set true in constructor
	_patch                 *StatePatch
	_currentFlow           *Flow
	_namedFlows            map[string]*Flow
	_aliveFlowNamesDirty   bool // must be set true in constructor
	_currentErrors         []string
	_currentWarnings       []string
	_variablesState        *VariablesState
	_aliveFlowNames        []string
	_evaluationStack       []Object
	_story                 *Story
	_currentTags           []string
	_currentTurnIndex      int

	// Public
	OnDidLoadState  *ActionEvent
	DivertedPointer Pointer
	StorySeed       int
	PreviousRandom  int
	DidSafeExit     bool
}

func (s *StoryState) OutputStream() []Object {

	return s._currentFlow.OutputStream
}

func (s *StoryState) CurrentChoices() []*Choice {

	// If we can continue generating text content rather than choices,
	// then we reflect the choice list as being empty, since choices
	// should always come at the end.
	if s.CanContinue() {
		return []*Choice{}
	}

	return s._currentFlow.CurrentChoices
}

func (s *StoryState) GeneratedChoices() []*Choice {

	return s._currentFlow.CurrentChoices
}

// TODO: Consider removing currentErrors / currentWarnings altogether
// and relying on client error handler code immediately handling StoryExceptions etc
// Or is there a specific reason we need to collect potentially multiple
// errors before throwing/exiting?
func (s *StoryState) CurrentErrors() []string {

	return s._currentErrors
}

func (s *StoryState) CurrentWarnings() []string {

	return s._currentWarnings
}

func (s *StoryState) VariablesState() *VariablesState {

	return s._variablesState
}

func (s *StoryState) CallStack() *CallStack {

	return s._currentFlow.CallStack
}

func (s *StoryState) EvaluationStack() []Object {

	return s._evaluationStack
}

func (s *StoryState) CurrentTurnIndex() int {

	return s._currentTurnIndex
}

// String representation of the location where the story currently is.
func (s *StoryState) CurrentPathString() string {

	pointer := s.CurrentPointer()
	if pointer.IsNull() {
		return ""
	}

	return pointer.Path().String()
}

func (s *StoryState) CurrentPointer() Pointer {

	return s.CallStack().CurrentElement().CurrentPointer
}

func (s *StoryState) SetCurrentPointer(value Pointer) {

	s.CallStack().CurrentElement().CurrentPointer = value
}

func (s *StoryState) PreviousPointer() Pointer {

	return s.CallStack().CurrentThread().PreviousPointer
}

func (s *StoryState) SetPreviousPointer(value Pointer) {

	s.CallStack().CurrentThread().PreviousPointer = value
}

func (s *StoryState) CanContinue() bool {

	return s.CurrentPointer().IsNull() == false && s.HasError() == false
}

func (s *StoryState) HasError() bool {

	return s.CurrentErrors() != nil && len(s.CurrentErrors()) > 0
}

func (s *StoryState) HasWarning() bool {

	return s.CurrentWarnings() != nil && len(s.CurrentWarnings()) > 0
}

func (s *StoryState) CurrentText() string {

	if s._outputStreamTextDirty {

		var sb strings.Builder

		inTag := false
		for _, outputObj := range s.OutputStream() {

			textContent, _ := outputObj.(*StringValue) // C# as
			if inTag == false && textContent != nil {

				sb.WriteString(textContent.Value())
			} else {

				controlCommand, _ := outputObj.(*ControlCommand) // C# as
				if controlCommand != nil {

					if controlCommand.CommandType == CommandTypeBeginTag {
						inTag = true
					} else if controlCommand.CommandType == CommandTypeEndTag {
						inTag = false
					}
				}
			}
		}

		s._currentText = s.CleanOutputWhitespace(sb.String())
		s._outputStreamTextDirty = false
	}

	return s._currentText
}

// ToJson
// exports the current state to json format, in order to save the game.
func (s *StoryState) ToJson() string {

	writer := new(Writer)
	s.WriteJson(writer)

	return writer.String()
}

// LoadJson
// loads a previously saved state in JSON format.
func (s *StoryState) LoadJson(json string) {

	jObject := TextToDictionary(json)
	s.LoadJsonObj(jObject)

	if s.OnDidLoadState != nil {
		s.OnDidLoadState.Emit()
	}
}

// VisitCountAtPathString
// Gets the visit/read count of a particular Container at the given path.
// For a knot or stitch, that path string will be in the form:
//
//     knot
//     knot.stitch
func (s *StoryState) VisitCountAtPathString(pathString string) int {

	if s._patch != nil {

		container := s._story.ContentAtPath(NewPathFromString(pathString)).Container()
		if container == nil {
			panic("Content at path not found: " + pathString)
		}

		if visitCount, ok := s._patch.TryGetVisitCount(container); ok {
			return visitCount
		}
	}

	if visitCount, ok := s._visitCounts[pathString]; ok {
		return visitCount
	}

	return 0
}

func (s *StoryState) VisitCountForContainer(container *Container) int {

	if container.VisitsShouldBeCounted == false {
		s._story.Error("Read count for target (" + container.Name() + " - on " + container.DebugMetadata().String() + ") unknown.")
		return 0
	}

	if s._patch != nil {
		if count, ok := s._patch.TryGetVisitCount(container); ok {
			return count
		}
	}

	containerPathStr := container.Path(container).String()
	count, _ := s._visitCounts[containerPathStr]
	return count
}

func (s *StoryState) IncrementVisitCountForContainer(container *Container) {

	if s._patch != nil {
		currCount := s.VisitCountForContainer(container)
		currCount++
		s._patch.SetVisitCount(container, currCount)
		return
	}

	containerPathStr := container.Path(container).String()
	count, _ := s._visitCounts[containerPathStr] // C# as
	count++
	s._visitCounts[containerPathStr] = count
}

func (s *StoryState) RecordTurnIndexVisitToContainer(container *Container) {

	if s._patch != nil {
		s._patch.SetTurnIndex(container, s.CurrentTurnIndex())
		return
	}

	containerPathStr := container.Path(container).String()
	s._turnIndices[containerPathStr] = s.CurrentTurnIndex()
}

func (s *StoryState) TurnsSinceForContainer(container *Container) int {

	if container.TurnIndexShouldBeCounted == false {
		s._story.Error("TURNS_SINCE() for target (" + container.Name() + " - on " + container.DebugMetadata().String() + ") unknown.")
	}

	if s._patch != nil {
		if index, ok := s._patch.TryGetTurnIndex(container); ok {
			return s.CurrentTurnIndex() - index
		}
	}

	containerPathStr := container.Path(container).String()
	if index, ok := s._turnIndices[containerPathStr]; ok {
		return s.CurrentTurnIndex() - index
	}

	return -1
}

// CleanOutputWhitespace
// Cleans inline whitespace in the following way:
//  - Removes all whitespace from the start and end of line (including just before a \n)
//  - Turns all consecutive space and tab runs into single spaces (HTML style)
func (s *StoryState) CleanOutputWhitespace(str string) string {

	var sb strings.Builder

	currentWhitespaceStart := -1
	startOfLine := 0

	for i := 0; i < len(str); i++ {

		c := str[i]
		isInlineWhitespace := c == ' ' || c == '\t'

		if isInlineWhitespace && currentWhitespaceStart == -1 {
			currentWhitespaceStart = i
		}

		if !isInlineWhitespace {
			if c != '\n' && currentWhitespaceStart > 0 && currentWhitespaceStart != startOfLine {
				sb.WriteRune(' ')
			}
			currentWhitespaceStart = -1
		}

		if c == '\n' {
			startOfLine = i + 1
		}

		if !isInlineWhitespace {
			sb.WriteByte(c)
		}
	}

	return sb.String()
}

func (s *StoryState) CurrentTags() []string {

	if s._outputStreamTagsDirty {

		inTag := false
		var sb strings.Builder

		for _, outputObj := range s.OutputStream() {
			controlCommand, _ := outputObj.(*ControlCommand) // C# as

			if controlCommand != nil {
				if controlCommand.CommandType == CommandTypeBeginTag {
					if inTag && sb.Len() > 0 {
						txt := s.CleanOutputWhitespace(sb.String())
						s._currentTags = append(s._currentTags, txt)
						sb.Reset()
					}
					inTag = true
				} else if controlCommand.CommandType == CommandTypeEndTag {
					if sb.Len() > 0 {
						txt := s.CleanOutputWhitespace(sb.String())
						s._currentTags = append(s._currentTags, txt)
						sb.Reset()
					}
					inTag = false
				}
			} else if inTag {
				strVal, _ := outputObj.(*StringValue) // C# as
				if strVal != nil {
					sb.WriteString(strVal.Value())
				}
			} else {
				tag, _ := outputObj.(*Tag)
				if tag != nil && len(tag.Text()) > 0 {
					s._currentTags = append(s._currentTags, tag.Text()) // tag.text has whitespae already cleaned
				}
			}
		}

		if sb.Len() > 0 {
			txt := s.CleanOutputWhitespace(sb.String())
			s._currentTags = append(s._currentTags, txt)
			sb.Reset()
		}

		s._outputStreamTagsDirty = false
	}

	return s._currentTags
}

func (s *StoryState) CurrentFlowName() string {

	return s._currentFlow.Name
}

func (s *StoryState) CurrentFlowIsDefaultFlow() bool {

	return s._currentFlow.Name == kDefaultFlowName
}

func (s *StoryState) AliveFlowNames() []string {

	if s._aliveFlowNamesDirty {

		s._aliveFlowNames = []string{}

		if s._namedFlows != nil {
			for flowName, _ := range s._namedFlows {
				if flowName != kDefaultFlowName {
					s._aliveFlowNames = append(s._aliveFlowNames, flowName)
				}
			}
		}

		s._aliveFlowNamesDirty = false
	}

	return s._aliveFlowNames
}

func (s *StoryState) InExpressionEvaluation() bool {

	return s.CallStack().CurrentElement().InExpressionEvaluation
}

func (s *StoryState) SetInExpressionEvaluation(value bool) {

	s.CallStack().CurrentElement().InExpressionEvaluation = value
}

func NewStoryState(story *Story) *StoryState {

	newStoryState := new(StoryState)
	newStoryState._outputStreamTextDirty = true // set true in c# class def
	newStoryState._outputStreamTagsDirty = true // set true in c# class def
	newStoryState._story = story
	newStoryState._currentFlow = NewFlow(kDefaultFlowName, story)
	newStoryState.OutputStreamDirty()
	newStoryState._aliveFlowNamesDirty = true // set true in c# class def
	newStoryState._evaluationStack = []Object{}
	newStoryState._variablesState = NewVariablesState(newStoryState.CallStack(), story.ListDefinitions())
	newStoryState._visitCounts = make(map[string]int)
	newStoryState._turnIndices = make(map[string]int)
	newStoryState._currentTurnIndex = -1

	// Seed the shuffle random numbers
	timeSeed := time.Now().UnixNano() / int64(time.Millisecond)
	newStoryState.StorySeed = (rand.New(rand.NewSource(timeSeed))).Intn(100)
	newStoryState.PreviousRandom = 0

	newStoryState.GoToStart()

	return newStoryState
}

func (s *StoryState) GoToStart() {

	s.CallStack().CurrentElement().CurrentPointer = StartOfPointer(s._story.MainContentContainer())
}

func (s *StoryState) switchFlow_Internal(flowName string) {

	if flowName == "" {
		panic("Must pass a non-null string to Story.SwitchFlow")
	}

	if s._namedFlows == nil {
		s._namedFlows = make(map[string]*Flow)
		s._namedFlows[kDefaultFlowName] = s._currentFlow
	}

	if flowName == s._currentFlow.Name {
		return
	}

	var flow *Flow
	ok := false
	if flow, ok = s._namedFlows[flowName]; !ok {
		flow = NewFlow(flowName, s._story)
		s._namedFlows[flowName] = flow
		s._aliveFlowNamesDirty = true
	}

	s._currentFlow = flow
	s._variablesState.SetCallStack(s._currentFlow.CallStack)

	// Cause text to be regenerated from output stream if necessary
	s.OutputStreamDirty()
}

func (s *StoryState) switchToDefaultFlow_Internal() {

	if s._namedFlows == nil {
		return
	}

	s.switchFlow_Internal(kDefaultFlowName)
}

func (s *StoryState) removeFlow_Internal(flowName string) {

	if flowName == "" {
		panic("Must pass a non-null string to Story.DestroyFlow")
	}

	if flowName == kDefaultFlowName {
		panic("Cannot destroy default flow")
	}

	// If we're currently in the flow that's being removed, switch back to default
	if s._currentFlow.Name == flowName {
		s.switchToDefaultFlow_Internal()
	}
	delete(s._namedFlows, flowName)
	s._aliveFlowNamesDirty = true
}

// Warning: Any Runtime.Object content referenced within the StoryState will
// be re-referenced rather than cloned. This is generally okay though since
// Runtime.Objects are treated as immutable after they've been set up.
// (e.g. we don't edit a Runtime.StringValue after it's been created an added.)
// I wonder if there's a sensible way to enforce that..??
func (s *StoryState) CopyAndStartPatching() *StoryState {

	storyStateCopy := NewStoryState(s._story)
	storyStateCopy._patch = NewStatePatchFromStatePatch(s._patch)

	// Hijack the new default flow to become a copy of our current one
	// If the patch is applied, then this new flow will replace the old one in _namedFlows
	storyStateCopy._currentFlow.Name = s._currentFlow.Name
	storyStateCopy._currentFlow.CallStack = NewCallStackFromCallStack(s._currentFlow.CallStack)
	storyStateCopy._currentFlow.CurrentChoices = append(storyStateCopy._currentFlow.CurrentChoices, s._currentFlow.CurrentChoices...)
	storyStateCopy._currentFlow.OutputStream = append(storyStateCopy._currentFlow.OutputStream, s._currentFlow.OutputStream...)
	storyStateCopy.OutputStreamDirty()

	// The copy of the state has its own copy of the named flows dictionary,
	// except with the current flow replaced with the copy above
	// (Assuming we're in multi-flow mode at all. If we're not then
	// the above copy is simply the default flow copy and we're done)
	if s._namedFlows != nil {

		storyStateCopy._namedFlows = make(map[string]*Flow)
		for namedFlowKey, namedFlowValue := range s._namedFlows {
			storyStateCopy._namedFlows[namedFlowKey] = namedFlowValue
		}

		storyStateCopy._namedFlows[s._currentFlow.Name] = storyStateCopy._currentFlow
		storyStateCopy._aliveFlowNamesDirty = true
	}

	if s.HasError() {
		storyStateCopy._currentErrors = append(storyStateCopy._currentErrors, s._currentErrors...)
	}

	if s.HasWarning() {
		storyStateCopy._currentWarnings = append(storyStateCopy._currentWarnings, s._currentWarnings...)
	}

	// ref copy - exactly the same variables state!
	// we're expecting not to read it only while in patch mode
	// (though the callstack will be modified)
	storyStateCopy._variablesState = s._variablesState
	storyStateCopy._variablesState.SetCallStack(storyStateCopy.CallStack())
	storyStateCopy._variablesState.Patch = storyStateCopy._patch

	storyStateCopy._evaluationStack = append(storyStateCopy._evaluationStack, s._evaluationStack...)

	if s.DivertedPointer.IsNull() == false {

		storyStateCopy.DivertedPointer = s.DivertedPointer
	}

	storyStateCopy.SetPreviousPointer(s.PreviousPointer())

	// visit counts and turn indicies will be read only, not modified
	// while in patch mode
	storyStateCopy._visitCounts = s._visitCounts
	storyStateCopy._turnIndices = s._turnIndices

	storyStateCopy._currentTurnIndex = s._currentTurnIndex
	storyStateCopy.StorySeed = s.StorySeed
	storyStateCopy.PreviousRandom = s.PreviousRandom

	storyStateCopy.DidSafeExit = s.DidSafeExit

	return storyStateCopy
}

func (s *StoryState) RestoreAfterPatch() {

	// VariablesState was being borrowed by the patched
	// state, so restore it with our own callstack.
	// _patch will be null normally, but if you're in the
	// middle of a save, it may contain a _patch for save purpsoes.
	s._variablesState._callStack = s.CallStack()
	s._variablesState.Patch = s._patch // usually null
}

func (s *StoryState) ApplyAnyPatch() {

	if s._patch == nil {
		return
	}

	s._variablesState.ApplyPatch()
	for pathToCountKey, pathToCountValue := range s._patch._visitCounts {
		s.ApplyCountChanges(pathToCountKey, pathToCountValue, true)
	}

	for pathToIndexKey, pathToIndexValue := range s._patch._turnIndices {
		s.ApplyCountChanges(pathToIndexKey, pathToIndexValue, false)
	}

	s._patch = nil
}

func (s *StoryState) ApplyCountChanges(container *Container, newCount int, isVisit bool) {

	var counts map[string]int
	if isVisit {
		counts = s._visitCounts
	} else {
		counts = s._turnIndices
	}

	counts[container.Path(container).String()] = newCount
}

func (s *StoryState) WriteJson(writer *Writer) {

	writer.WriteObjectStart()

	// Flows
	writer.WritePropertyStart("flows")
	writer.WriteObjectStart()

	// Multi-flow
	if s._namedFlows != nil {
		for namedFlowKey, namedFlowValue := range s._namedFlows {
			writer.WritePropertyStart(namedFlowKey)
			namedFlowValue.WriteJson(writer)
			writer.WritePropertyEnd()
		}
	} else {
		// Single flow
		writer.WritePropertyStart(s._currentFlow.Name)
		s._currentFlow.WriteJson(writer)
		writer.WritePropertyEnd()
	}

	writer.WriteObjectEnd()
	writer.WritePropertyEnd() // end of flows

	writer.WriteStringProperty("currentFlowName", s._currentFlow.Name)

	writer.WritePropertyStart("variablesState")
	s._variablesState.WriteJson(writer)
	writer.WritePropertyEnd()

	writer.WritePropertyStart("evalStack")
	WriteListRuntimeObjs(writer, s._evaluationStack)
	writer.WritePropertyEnd()

	if s.DivertedPointer.IsNull() == false {
		writer.WriteStringProperty("currentDivertTarget", s.DivertedPointer.Path().ComponentsString())
	}

	writer.WritePropertyStart("visitCounts")
	WriteIntDictionary(writer, s._visitCounts)
	writer.WritePropertyEnd()

	writer.WritePropertyStart("turnIndices")
	WriteIntDictionary(writer, s._turnIndices)
	writer.WritePropertyEnd()

	writer.WriteIntProperty("turnIdx", s._currentTurnIndex)
	writer.WriteIntProperty("storySeed", s.StorySeed)
	writer.WriteIntProperty("previousRandom", s.PreviousRandom)

	writer.WriteIntProperty("inkSaveVersion", KInkSaveStateVersion)

	// Not using this right now, but could do in future.
	writer.WriteIntProperty("inkFormatVersion", InkVersionCurrent)

	writer.WriteObjectEnd()
}

func (s *StoryState) LoadJsonObj(jObject map[string]interface{}) {

	var jSaveVersion interface{}
	ok := false

	if jSaveVersion, ok = jObject["inkSaveVersion"]; !ok {
		panic("ink save format incorrect, can't load.")
	}

	if jSaveVersion.(int) < KInkSaveStateVersion {
		panic("Ink save format isn't compatible with the current version (saw '" + fmt.Sprint(jSaveVersion) + "', but minimum is " + fmt.Sprint(kMinCompatibleLoadVersion) + "), so can't load.")
	}

	// Flows: Always exists in latest format (even if there's just one default)
	// but this dictionary doesn't exist in prev format
	if flowsObj, ok := jObject["flows"]; ok {

		flowsObjDict := flowsObj.(map[string]interface{})

		// Single default flow
		if len(flowsObjDict) == 1 {
			s._namedFlows = nil
		} else if s._namedFlows == nil {
			// Multi-flow, need to create flows dict
			s._namedFlows = make(map[string]*Flow)
		} else {
			// Multi-flow, already have a flows dict
			for key, _ := range s._namedFlows {
				delete(s._namedFlows, key)
			}
		}

		// Load up each flow (there may only be one)
		for name, namedFlowObjValue := range flowsObjDict {
			flowObj := namedFlowObjValue.(map[string]interface{})

			// Load up this flow using JSON data
			flow := NewFlowFromJObject(name, s._story, flowObj)

			if len(flowsObjDict) == 1 {
				s._currentFlow = NewFlowFromJObject(name, s._story, flowObj)
			} else {
				s._namedFlows[name] = flow
			}
		}

		if s._namedFlows != nil && len(s._namedFlows) > 1 {
			currFlowName := jObject["currentFlowName"].(string)
			s._currentFlow = s._namedFlows[currFlowName]
		}
	} else {
		// Old format: individually load up callstack, output stream, choices in current/default flow

		s._namedFlows = nil
		s._currentFlow.Name = kDefaultFlowName
		s._currentFlow.CallStack.SetJsonToken(jObject["callstackThreads"].(map[string]interface{}), s._story)
		s._currentFlow.OutputStream = JArrayToRuntimeObjList[Object](jObject["outputStream"].([]interface{}), false)
		s._currentFlow.CurrentChoices = JArrayToRuntimeObjList[*Choice](jObject["currentChoices"].([]interface{}), false)

		jChoiceThreadsObj := jObject["choiceThreads"]
		s._currentFlow.LoadFlowChoiceThreads(jChoiceThreadsObj.(map[string]interface{}), s._story)
	}

	s.OutputStreamDirty()
	s._aliveFlowNamesDirty = true

	s._variablesState.SetJsonToken(jObject["variablesState"].(map[string]interface{}))
	s._variablesState.SetCallStack(s._currentFlow.CallStack)

	s._evaluationStack = JArrayToRuntimeObjList[Object](jObject["evalStack"].([]interface{}), false)

	var currentDivertTargetPath interface{}
	ok = false
	if currentDivertTargetPath, ok = jObject["currentDivertTarget"]; ok {
		divertPath := NewPathFromString(currentDivertTargetPath.(fmt.Stringer).String())
		s.DivertedPointer = s._story.PointerAtPath(divertPath)
	}

	s._visitCounts = JObjectToIntDictionary(jObject["visitCounts"].(map[string]interface{}))
	s._turnIndices = JObjectToIntDictionary(jObject["turnIndices"].(map[string]interface{}))

	s._currentTurnIndex = jObject["turnIdx"].(int)
	s.StorySeed = jObject["storySeed"].(int)

	// Not optional, but bug in inkjs means it's actually missing in inkjs saves
	var previousRandomObj interface{}
	ok = false
	if previousRandomObj, ok = jObject["previousRandom"]; ok {
		s.PreviousRandom = previousRandomObj.(int)
	} else {
		s.PreviousRandom = 0
	}
}

func (s *StoryState) ResetErrors() {
	s._currentErrors = nil
	s._currentWarnings = nil
}

// ResetOutput
// (default) objs: nil
func (s *StoryState) ResetOutput(objs []Object) {
	//OutputStream().Clear() // C#
	s._currentFlow.OutputStream = s._currentFlow.OutputStream[:0]

	if objs != nil {
		s._currentFlow.OutputStream = append(s._currentFlow.OutputStream, objs...)
	}
	s.OutputStreamDirty()
}

// PushToOutputStream
// Push to output stream, but split out newlines in text for consistency
// in dealing with them later.
func (s *StoryState) PushToOutputStream(obj Object) {

	text, _ := obj.(*StringValue)
	if text != nil {

		listText := s.TrySplittingHeadTailWhitespace(text)
		if listText != nil {

			for _, textObj := range listText {
				s.PushToOutputStreamIndividual(textObj)
			}
			s.OutputStreamDirty()
			return
		}
	}

	s.PushToOutputStreamIndividual(obj)

	s.OutputStreamDirty()
}

func (s *StoryState) PopFromOutputStream(count int) {

	//s.OutputStream().RemoveRange(s.OutputStream().Count()-count, count)
	s._currentFlow.OutputStream = s._currentFlow.OutputStream[:len(s._currentFlow.OutputStream)-count]
	s.OutputStreamDirty()
}

// TrySplittingHeadTailWhitespace
// At both the start and the end of the string, split out the new lines like so:
//
//  "   \n  \n     \n  the string \n is awesome \n     \n     "
//      ^-----------^                           ^-------^
//
// Excess newlines are converted into single newlines, and spaces discarded.
// Outside spaces are significant and retained. "Interior" newlines within
// the main string are ignored, since this is for the purpose of gluing only.
//
//  - If no splitting is necessary, null is returned.
//  - A newline on its own is returned in a list for consistency.
func (s *StoryState) TrySplittingHeadTailWhitespace(single *StringValue) []*StringValue {

	str := single.Value()

	headFirstNewlineIdx := -1
	headLastNewlineIdx := -1
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c == '\n' {
			if headFirstNewlineIdx == -1 {
				headFirstNewlineIdx = i
			}
			headLastNewlineIdx = i
		} else if c == ' ' || c == '\t' {
			continue
		} else {
			break
		}
	}

	tailLastNewlineIdx := -1
	tailFirstNewlineIdx := -1
	for i := len(str) - 1; i >= 0; i-- {
		c := str[i]
		if c == '\n' {
			if tailLastNewlineIdx == -1 {
				tailLastNewlineIdx = i
			}
			tailFirstNewlineIdx = i
		} else if c == ' ' || c == '\t' {
			continue
		} else {
			break
		}
	}

	// No splitting to be done?
	if headFirstNewlineIdx == -1 && tailLastNewlineIdx == -1 {
		return nil
	}

	listTexts := []*StringValue{}
	innerStrStart := 0
	innerStrEnd := len(str)

	if headFirstNewlineIdx != -1 {
		if headFirstNewlineIdx > 0 {
			leadingSpaces := NewStringValueFromString(str[0:headFirstNewlineIdx])
			//listTexts = append(listTexts, leadingSpaces)
			listTexts = append(listTexts, leadingSpaces)
		}
		//listTexts = append(listTexts, NewStringValueFromString("\n"))
		listTexts = append(listTexts, NewStringValueFromString("\n"))
		innerStrStart = headLastNewlineIdx + 1
	}

	if tailLastNewlineIdx != -1 {
		innerStrEnd = tailFirstNewlineIdx
	}

	if innerStrEnd > innerStrStart {
		innerStrText := str[innerStrStart : innerStrEnd-innerStrStart]
		//listTexts = append(listTexts, NewStringValueFromString(innerStrText))
		listTexts = append(listTexts, NewStringValueFromString(innerStrText))
	}

	if tailLastNewlineIdx != -1 && tailFirstNewlineIdx > headLastNewlineIdx {
		//listTexts = append(listTexts, NewStringValueFromString("\n"))
		listTexts = append(listTexts, NewStringValueFromString("\n"))
		if tailLastNewlineIdx < len(str)-1 {
			numSpaces := (len(str) - tailLastNewlineIdx) - 1
			trailingSpaces := NewStringValueFromString(str[tailLastNewlineIdx+1 : numSpaces])
			//listTexts = append(listTexts, trailingSpaces)
			listTexts = append(listTexts, trailingSpaces)
		}
	}

	return listTexts
}

func (s *StoryState) PushToOutputStreamIndividual(obj Object) {

	glue, _ := obj.(*Glue)
	text, _ := obj.(*StringValue)

	includeInOutput := true

	// New glue, so chomp away any whitespace from the end of the stream
	if glue != nil {
		s.TrimNewlinesFromOutputStream()
		includeInOutput = true
	} else if text != nil {
		// New text: do we really want to append it, if it's whitespace?
		// Two different reasons for whitespace to be thrown away:
		//   - Function start/end trimming
		//   - User defined glue: <>
		// We also need to know when to stop trimming, when there's non-whitespace.

		// Where does the current function call begin?
		functionTrimIndex := -1
		currEl := s.CallStack().CurrentElement()
		if currEl.PushPopType() == Function {
			functionTrimIndex = currEl.FunctionStartInOutputStream
		}

		// Do 2 things:
		//  - Find latest glue
		//  - Check whether we're in the middle of string evaluation
		// If we're in string eval within the current function, we
		// don't want to trim back further than the length of the current string.
		glueTrimIndex := -1
		for i := len(s.OutputStream()) - 1; i >= 0; i-- {
			o := s.OutputStream()[i]
			c, _ := o.(*ControlCommand) // C# as
			g, _ := o.(*Glue)           // C# as

			// Find latest glue
			if g != nil {
				glueTrimIndex = i
				break

			} else if c != nil && c.CommandType == CommandTypeBeginString {
				// Don't function-trim past the start of a string evaluation section
				if i >= functionTrimIndex {
					functionTrimIndex = -1
				}
				break
			}
		}

		// Where is the most agressive (earliest) trim point?
		trimIndex := -1
		if glueTrimIndex != -1 && functionTrimIndex != -1 {
			trimIndex = int(math.Min(float64(functionTrimIndex), float64(glueTrimIndex)))
		} else if glueTrimIndex != -1 {
			trimIndex = glueTrimIndex
		} else {
			trimIndex = functionTrimIndex
		}

		// So, are we trimming then?
		if trimIndex != -1 {

			// While trimming, we want to throw all newlines away,
			// whether due to glue or the start of a function
			if text.IsNewline() {
				includeInOutput = false
			} else if text.IsNonWhitespace() {
				// Able to completely reset when normal text is pushed

				if glueTrimIndex > -1 {
					s.RemoveExistingGlue()
				}

				// Tell all functions in callstack that we have seen proper text,
				// so trimming whitespace at the start is done.
				if functionTrimIndex > -1 {
					callstackElements := s.CallStack().Elements()
					for i := len(callstackElements); i >= 0; i-- {
						el := callstackElements[i]
						if el.PushPopType() == Function {
							el.FunctionStartInOutputStream = -1
						} else {
							break
						}
					}
				}
			}
		} else if text.IsNewline() {
			// De-duplicate newlines, and don't ever lead with a newline
			if s.OutputStreamEndsInNewline() || !s.OutputStreamContainsContent() {
				includeInOutput = false
			}
		}
	}

	if includeInOutput {
		// outputStream.Add(obj) C#
		s._currentFlow.OutputStream = append(s._currentFlow.OutputStream, obj)
		s.OutputStreamDirty()
	}
}

func (s *StoryState) TrimNewlinesFromOutputStream() {

	removeWhitespaceFrom := -1

	// Work back from the end, and try to find the point where
	// we need to start removing content.
	//  - Simply work backwards to find the first newline in a string of whitespace
	// e.g. This is the content   \n   \n\n
	//                            ^---------^ whitespace to remove
	//                        ^--- first while loop stops here
	i := len(s.OutputStream()) - 1
	for i >= 0 {
		obj := s.OutputStream()[i]
		cmd, _ := obj.(*ControlCommand) // C# as
		txt, _ := obj.(*StringValue)    // C# as

		if cmd != nil && (txt != nil && txt.IsNonWhitespace()) {
			break
		} else if txt != nil && txt.IsNewline() {
			removeWhitespaceFrom = i
		}
		i--
	}

	// Remove the whitespace
	if removeWhitespaceFrom >= 0 {
		i = removeWhitespaceFrom
		for i < len(s.OutputStream()) {
			text, _ := s.OutputStream()[i].(*StringValue) // C# as
			if text != nil {
				//s.OutputStream().RemoveAt(i)
				s._currentFlow.OutputStream = append(s._currentFlow.OutputStream[:i], s._currentFlow.OutputStream[i+1:]...)
				// outputStream.RemoveAt (i); C#
				//s._currentFlow.OutputStream = append(s._currentFlow.OutputStream[:i], s._currentFlow.OutputStream[i+1:]...)
			} else {
				i++
			}
		}
	}

	s.OutputStreamDirty()
}

// RemoveExistingGlue
// Only called when non-whitespace is appended
func (s *StoryState) RemoveExistingGlue() {

	for i := len(s.OutputStream()) - 1; i >= 0; i-- {
		c := s.OutputStream()[i]
		if _, isGlue := c.(*Glue); isGlue {
			// outputStream.RemoveAt (i); C#
			//s.OutputStream().RemoveAt(i)
			s._currentFlow.OutputStream = append(s._currentFlow.OutputStream[:i], s._currentFlow.OutputStream[i+1:]...)
			//s._currentFlow.OutputStream = append(s._currentFlow.OutputStream[:i], s._currentFlow.OutputStream[i+1:]...)
		} else if _, isControlCommand := c.(*ControlCommand); isControlCommand {
			break
		}
	}

	s.OutputStreamDirty()
}

func (s *StoryState) OutputStreamEndsInNewline() bool {

	if len(s.OutputStream()) > 0 {

		for i := len(s.OutputStream()) - 1; i >= 0; i-- {
			obj := s.OutputStream()[i]
			if _, isControlCommand := obj.(*ControlCommand); isControlCommand {
				break
			}
			text, _ := s.OutputStream()[i].(*StringValue) // C# as
			if text != nil {
				if text.IsNewline() {
					return true
				} else if text.IsNonWhitespace() {
					break
				}
			}
		}
	}

	return false
}

func (s *StoryState) OutputStreamContainsContent() bool {

	for _, content := range s.OutputStream() {
		if _, isStringValue := content.(*StringValue); isStringValue {
			return true
		}
	}
	return false
}

func (s *StoryState) InStringEvaluation() bool {

	for i := len(s.OutputStream()) - 1; i >= 0; i-- {
		cmd, _ := s.OutputStream()[i].(*ControlCommand)
		if cmd != nil && cmd.CommandType == CommandTypeBeginString {
			return true
		}
	}
	return false
}

func (s *StoryState) PushEvaluationStack(obj Object) {

	// Include metadata about the origin List for list values when
	// they're used, so that lower level functions can make use
	// of the origin list to get related items, or make comparisons
	// with the integer values etc.
	listValue, _ := obj.(*ListValue) // C# as
	if listValue != nil {

		// Update origin when list is has something to indicate the list origin
		rawList := listValue.Value()
		if rawList.OriginNames() != nil {
			// if( rawList.origins == null )
			//	 rawList.origins = new List<ListDefinition>(); C#
			rawList.Origins = nil

			for _, n := range rawList.OriginNames() {
				def, _ := s._story.ListDefinitions().TryListGetDefinition(n)
				contains := false
				for _, v := range rawList.Origins {
					if v == def {
						contains = true
					}
				}
				if !contains {
					//rawList.Origins = append(rawList.Origins, def)
					//rawList.Origins.Add(def)
					rawList.Origins = append(rawList.Origins, def)
				}
			}
		}
	}

	s._evaluationStack = append(s._evaluationStack, obj)
	//s._evaluationStack.Add(obj)
}

func (s *StoryState) PopEvaluationStack() Object {

	obj := s._evaluationStack[len(s._evaluationStack)-1]
	//    evaluationStack.RemoveAt (evaluationStack.Count - 1); C#
	//s._evaluationStack.RemoveAt(s._evaluationStack.Count() - 1)
	s._evaluationStack = s._evaluationStack[:len(s._evaluationStack)-1]
	return obj
}

func (s *StoryState) PeekEvaluationStack() Object {

	return s._evaluationStack[len(s._evaluationStack)-1]
}

func (s *StoryState) PopEvaluationStackEx(numberOfObjects int) []Object {

	if numberOfObjects > len(s._evaluationStack) {
		panic("trying to pop too many objects")
	}

	popped := s._evaluationStack[len(s._evaluationStack)-numberOfObjects:]
	//popped := s._evaluationStack.GetRange(s._evaluationStack.Count()-numberOfObjects, numberOfObjects)
	// evaluationStack.RemoveRange (evaluationStack.Count - numberOfObjects, numberOfObjects); C#
	s._evaluationStack = s._evaluationStack[:len(s._evaluationStack)-numberOfObjects]
	//s._evaluationStack.RemoveRange(s._evaluationStack.Count()-numberOfObjects, numberOfObjects)
	return popped
}

// ForceEnd
// Ends the current ink flow, unwrapping the callstack but without
// affecting any variables. Useful if the ink is (say) in the middle
// a nested tunnel, and you want it to reset so that you can divert
// elsewhere using ChoosePathString(). Otherwise, after finishing
// the content you diverted to, it would continue where it left off.
// Calling this is equivalent to calling -> END in ink.
func (s *StoryState) ForceEnd() {

	s.CallStack().Reset()

	s._currentFlow.CurrentChoices = s._currentFlow.CurrentChoices[:0] // _currentFlow.currentChoices.Clear();
	//s._currentFlow.CurrentChoices.Clear()

	s.SetCurrentPointer(NullPointer)
	s.SetPreviousPointer(NullPointer)

	s.DidSafeExit = true
}

// TrimWhitespaceFromFunctionEnd
// Add the end of a function call, trim any whitespace from the end.
// We always trim the start and end of the text that a function produces.
// The start whitespace is discard as it is generated, and the end
// whitespace is trimmed in one go here when we pop the function.
func (s *StoryState) TrimWhitespaceFromFunctionEnd() {

	// Debug.Assert (callStack.currentElement.type == PushPopType.Function); C#

	functionStartPoint := s.CallStack().CurrentElement().FunctionStartInOutputStream

	// If the start point has become -1, it means that some non-whitespace
	// text has been pushed, so it's safe to go as far back as we're able.
	if functionStartPoint == -1 {
		functionStartPoint = 0
	}

	// Trim whitespace from END of function call
	for i := len(s.OutputStream()) - 1; i >= functionStartPoint; i-- {
		obj := s.OutputStream()[i]
		txt, _ := obj.(*StringValue)
		cmd, _ := obj.(*ControlCommand)
		if txt == nil {
			continue
		}
		if cmd != nil {
			break
		}

		if txt.IsNewline() || txt.IsInlineWhitespace() {
			s._currentFlow.OutputStream = append(s._currentFlow.OutputStream[:i], s._currentFlow.OutputStream[i+1:]...) //  outputStream.RemoveAt (i); C#
			//s.OutputStream().RemoveAt(i)
			s.OutputStreamDirty()
		} else {
			break
		}
	}
}

// PopCallstack
// (default) popType: nil
func (s *StoryState) PopCallstack(popType PushPopType) {

	// Add the end of a function call, trim any whitespace from the end.
	if s.CallStack().CurrentElement().PushPopType() == Function {
		s.TrimWhitespaceFromFunctionEnd()
	}

	s.CallStack().Pop(popType)
}

// SetChosenPath
// Don't make public since the method need to be wrapped in Story for visit counting
func (s *StoryState) SetChosenPath(path *Path, incrementingTurnIndex bool) {

	// Changing direction, assume we need to clear current set of choices
	s._currentFlow.CurrentChoices = s._currentFlow.CurrentChoices[:0] // _currentFlow.currentChoices.Clear (); C#
	//s._currentFlow.CurrentChoices.Clear()

	newPointer := s._story.PointerAtPath(path)
	if !newPointer.IsNull() && newPointer.Index == -1 {
		newPointer.Index = 0
	}

	s.SetCurrentPointer(newPointer)

	if incrementingTurnIndex {
		s._currentTurnIndex = s.CurrentTurnIndex() + 1
	}
}

func (s *StoryState) StartFunctionEvaluationFromGame(funcContainer *Container, arguments ...interface{}) {

	s.CallStack().Push(FunctionEvaluationFromGame, len(s._evaluationStack), 0)
	s.CallStack().CurrentElement().CurrentPointer = StartOfPointer(funcContainer)

	s.PassArgumentsToEvaluationStack(arguments...)
}

func (s *StoryState) PassArgumentsToEvaluationStack(arguments ...interface{}) {

	// Pass arguments onto the evaluation stack
	if arguments != nil {
		for i := 0; i < len(arguments); i++ {
			_, isInt := arguments[i].(int)
			_, isFloat := arguments[i].(float64)
			_, isString := arguments[i].(string)
			_, isBool := arguments[i].(bool)
			_, isInkList := arguments[i].(*InkList)
			if !(isInt || isFloat || isString || isBool || isInkList) {
				panic("ink arguments when calling EvaluateFunction / ChoosePathStringWithParameters must be int, float, string, bool or InkList. Argument was " + reflect.TypeOf(arguments[i]).Name())
			}

			s.PushEvaluationStack(CreateValue(arguments[i]))
		}
	}
}

func (s *StoryState) TryExitFunctionEvaluationFromGame() bool {

	if s.CallStack().CurrentElement().PushPopType() == FunctionEvaluationFromGame {
		s.SetCurrentPointer(NullPointer)
		s.DidSafeExit = true
		return true
	}

	return false
}

func (s *StoryState) CompleteFunctionEvaluationFromGame() interface{} {

	if s.CallStack().CurrentElement().PushPopType() == FunctionEvaluationFromGame {
		panic("Expected external function evaluation to be complete. Stack trace: " + s.CallStack().CallStackTrace())
	}

	originalEvaluationStackHeight := s.CallStack().CurrentElement().EvaluationStackHeightWhenPushed

	// Do we have a returned value?
	// Potentially pop multiple values off the stack, in case we need
	// to clean up after ourselves (e.g. caller of EvaluateFunction may
	// have passed too many arguments, and we currently have no way to check for that)
	var returnedObj Object
	for len(s._evaluationStack) > originalEvaluationStackHeight {
		poppedObj := s.PopEvaluationStack()
		if returnedObj == nil {
			returnedObj = poppedObj
		}
	}

	// Finally, pop the external function evaluation
	s.PopCallstack(FunctionEvaluationFromGame)

	if returnedObj != nil {
		if _, isVoid := returnedObj.(*Void); isVoid {
			return nil
		}

		// Some kind of value, if not void
		returnVal, _ := returnedObj.(Value)

		// DivertTargets get returned as the string of components
		// (rather than a Path, which isn't public)
		if returnVal.ValueType() == ValueTypeDivertTarget {
			return returnVal.ValueObject().(fmt.Stringer).String()
		}

		// Other types can just have their exact object type:
		// int, float, string. VariablePointers get returned as strings.
		return returnVal.ValueObject()
	}

	return nil
}

func (s *StoryState) AddError(message string, isWarning bool) {
	if !isWarning {
		s._currentErrors = append(s._currentErrors, message)
	} else {
		s._currentWarnings = append(s._currentWarnings, message)
	}
}

func (s *StoryState) OutputStreamDirty() {
	s._outputStreamTextDirty = true
	s._outputStreamTagsDirty = true
}
