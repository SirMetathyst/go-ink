package runtime

import (
	"fmt"
	"reflect"
)

type VariableChanged func(variableName string, newValue Object)

type VariableChangedEvent struct {
	Event[VariableChanged]
}

func (s *VariableChangedEvent) Emit(variableName string, newValue Object) {
	for _, fn := range s.h {
		fn(variableName, newValue)
	}
}

type VariablesState struct {

	// Public

	//
	VariableChangedEvent *VariableChangedEvent
	Patch                *StatePatch

	// Private

	//
	_globalVariables               map[string]Object
	_defaultGlobalVariables        map[string]Object
	_batchObservingVariableChanges bool
	_callStack                     *CallStack
	_changedVariablesForBatchObs   map[string]struct{}
	_listDefsOrigin                *ListDefinitionsOrigin
}

func (s *VariablesState) SetBatchObservingVariableChanges(value bool) {

	s._batchObservingVariableChanges = value
	if value {

		s._changedVariablesForBatchObs = make(map[string]struct{})
	} else {

		// Finished observing variables in a batch - now send
		// notifications for changed variables all in one go.
		if s._changedVariablesForBatchObs != nil {
			for variableName, _ := range s._changedVariablesForBatchObs {
				currentValue := s._globalVariables[variableName]
				if s.VariableChangedEvent != nil {
					s.VariableChangedEvent.Emit(variableName, currentValue)
				}
			}
		}

		s._changedVariablesForBatchObs = nil
	}
}

// SetCallStack
// Allow StoryState to change the current callstack, e.g. for
// temporary function evaluation.
func (s *VariablesState) SetCallStack(value *CallStack) {
	s._callStack = value
}

// Get or set the value of a named global ink variable.
// The types available are the standard ink types. Certain
// types will be implicitly casted when setting.
// For example, doubles to floats, longs to ints, and bools
// to ints.
func (s *VariablesState) GetVariable(variableName string) interface{} {

	if s.Patch != nil {
		if varContents, ok := s.Patch.TryGetGlobal(variableName); ok {
			val, _ := varContents.(Value)
			return val.ValueObject()
		}
	}

	// Search main dictionary first.
	// If it's not found, it might be because the story content has changed,
	// and the original default value hasn't be instantiated.
	// Should really warn somehow, but it's difficult to see how...!

	if varContents, ok := s._globalVariables[variableName]; ok {
		val, _ := varContents.(Value)
		return val.ValueObject()
	}

	if varContents, ok := s._defaultGlobalVariables[variableName]; ok {
		val, _ := varContents.(Value)
		return val.ValueObject()
	}

	return nil
}

func (s *VariablesState) Set(variableName string, value interface{}) {

	if _, ok := s._defaultGlobalVariables[variableName]; !ok {
		// StoryException
		// TODO: probably shouldn't panic, return error instead
		panic("Cannot assign to a variable (" + variableName + ") that hasn't been declared in the story")
	}

	val := CreateValue(value)
	if val == nil {
		if value == nil {
			panic("Cannot pass null to VariableState")
		}

		panic(fmt.Sprintf("Invalid value passed to VariableState: %v", value))
	}

	s.SetGlobal(variableName, val)
}

func NewVariablesState(callStack *CallStack, listDefsOrigin *ListDefinitionsOrigin) *VariablesState {

	newVariablesState := new(VariablesState)
	newVariablesState._globalVariables = make(map[string]Object)
	newVariablesState._callStack = callStack
	newVariablesState._listDefsOrigin = listDefsOrigin

	return newVariablesState
}

func (s *VariablesState) ApplyPatch() {

	for namedVarKey, namedVarValue := range s.Patch.Globals() {
		s._globalVariables[namedVarKey] = namedVarValue
	}

	if s._changedVariablesForBatchObs != nil {
		for name, _ := range s.Patch.ChangedVariables() {
			AddToMap(s._changedVariablesForBatchObs, name, struct{}{})
		}
	}

	s.Patch = nil
}

func (s *VariablesState) SetJsonToken(jToken map[string]interface{}) {

	ClearMap(s._globalVariables)

	for varValKey, varValValue := range s._defaultGlobalVariables {
		if loadedToken, ok := jToken[varValKey]; ok {
			s._globalVariables[varValKey] = JTokenToRuntimeObject(loadedToken)
		} else {
			s._globalVariables[varValKey] = varValValue
		}
	}
}

// DontSaveDefaultValues
// When saving out JSON state, we can skip saving global values that
// remain equal to the initial values that were declared in ink.
// This makes the save file (potentially) much smaller assuming that
// at least a portion of the globals haven't changed. However, it
// can also take marginally longer to save in the case that the
// majority HAVE changed, since it has to compare all globals.
// It may also be useful to turn this off for testing worst case
// save timing.
var DontSaveDefaultValues = true

func (s *VariablesState) WriteJson(writer *Writer) {

	writer.WriteObjectStart()
	for name, val := range s._globalVariables {

		if DontSaveDefaultValues {
			// Don't write out values that are the same as the default global values
			if defaultVal, ok := s._defaultGlobalVariables[name]; ok {
				if s.RuntimeObjectsEqual(val, defaultVal) {
					continue
				}
			}
		}

		writer.WritePropertyStart(name)
		WriteRuntimeObject(writer, val)
		writer.WritePropertyEnd()
	}
	writer.WriteObjectEnd()
}

func (s *VariablesState) RuntimeObjectsEqual(obj1 Object, obj2 Object) bool {

	if reflect.TypeOf(obj1).Name() != reflect.TypeOf(obj2).Name() {
		return false
	}

	boolVal, ok := obj1.(*BoolValue)
	if ok {
		return boolVal.Value() == obj2.(*BoolValue).Value()
	}

	intVal, ok := obj1.(*IntValue)
	if ok {
		return intVal.Value() == obj2.(*IntValue).Value()
	}

	floatVal, ok := obj1.(*FloatValue)
	if ok {
		return floatVal.Value() == obj2.(*FloatValue).Value()
	}

	// TODO: clean this up
	// e.g. InkList does not implement correct equals
	val1, _ := obj1.(Value)
	val2, _ := obj2.(Value)

	if val1 != nil {

		equalsObj1, ok1 := val1.ValueObject().(Equals)
		equalsObj2, ok2 := val2.ValueObject().(Equals)

		if ok1 && ok2 {
			return equalsObj1.Equals(equalsObj2)
		} else {
			return equalsObj1 == equalsObj2
		}
	}

	panic("FastRoughDefinitelyEquals: Unsupported runtime object type: " + reflect.TypeOf(obj1).Name())
}

func (s *VariablesState) GlobalVariableExistsWithName(name string) bool {

	if _, ok := s._globalVariables[name]; ok {
		return true
	}

	if s._defaultGlobalVariables != nil {
		if _, ok := s._defaultGlobalVariables[name]; ok {
			return true
		}
	}

	return false
}

// GetVariableWithName
// (default) contextIndex: -1
func (s *VariablesState) GetVariableWithName(name string, contextIndex int) Object {

	varValue := s.GetRawVariableWithName(name, contextIndex)

	varPointer, _ := varValue.(*VariablePointerValue)
	if varPointer != nil {
		varValue = s.ValueAtVariablePointer(varPointer)
	}

	return varValue
}

func (s *VariablesState) GetRawVariableWithName(name string, contextIndex int) Object {

	var varValue Object
	ok := false

	// 0 context = global
	if contextIndex == 0 || contextIndex == -1 {

		if s.Patch != nil {
			if varValue, ok = s.Patch.TryGetGlobal(name); ok {
				return varValue
			}
		}

		if varValue, ok = s._globalVariables[name]; ok {
			return varValue
		}

		// Getting variables can actually happen during globals set up since you can do
		//  VAR x = A_LIST_ITEM
		// So _defaultGlobalVariables may be null.
		// We need to do this check though in case a new global is added, so we need to
		// revert to the default globals dictionary since an initial value hasn't yet been set.
		if s._defaultGlobalVariables != nil {
			if varValue, ok = s._defaultGlobalVariables[name]; ok {
				return varValue
			}
		}

		listItemValue := s._listDefsOrigin.FindSingleItemListWithName(name)
		if listItemValue != nil {
			return listItemValue
		}
	}

	// Temporary
	varValue = s._callStack.GetTemporaryVariableWithName(name, contextIndex)
	return varValue
}

func (s *VariablesState) ValueAtVariablePointer(pointer *VariablePointerValue) Object {

	return s.GetVariableWithName(pointer.Value(), pointer.ContextIndex())
}

func (s *VariablesState) Assign(varAss *VariableAssignment, value Object) {

	name := varAss.VariableName()
	contextIndex := -1

	// Are we assigning to a global variable?
	setGlobal := false
	if varAss.IsNewDeclaration() {
		setGlobal = varAss.IsGlobal
	} else {
		setGlobal = s.GlobalVariableExistsWithName(name)
	}

	// Constructing new variable pointer reference
	if varAss.IsNewDeclaration() {
		varPointer, _ := value.(*VariablePointerValue) // C# as
		if varPointer != nil {
			fullyResolvedVariablePointer := s.ResolveVariablePointer(varPointer)
			value = fullyResolvedVariablePointer
		}
	} else {
		// Assign to existing variable pointer?
		// Then assign to the variable that the pointer is pointing to by name.
		var existingPointer *VariablePointerValue
		for do := true; do; do = existingPointer != nil {
			existingPointer, _ = s.GetRawVariableWithName(name, contextIndex).(*VariablePointerValue)
			if existingPointer != nil {
				name = existingPointer.Value()
				contextIndex = existingPointer.ContextIndex()
				setGlobal = contextIndex == 0
			}
		}
	}

	if setGlobal {
		s.SetGlobal(name, value)
	} else {
		s._callStack.SetTemporaryVariable(name, value, varAss.IsNewDeclaration(), contextIndex)
	}
}

func (s *VariablesState) SnapshotDefaultGlobals() {

	s._defaultGlobalVariables = NewMapFromMap(s._globalVariables)
}
func (s *VariablesState) RetainListOriginsForAssignment(oldValue Object, newValue Object) {

	oldList, _ := oldValue.(*ListValue)
	newList, _ := newValue.(*ListValue)
	if oldList != nil && newList != nil && newList.Value().Count() > 0 {
		newList.Value().SetInitialOriginNames(oldList.Value().OriginNames())
	}
}

func (s *VariablesState) SetGlobal(variableName string, value Object) {

	var oldValue Object
	ok := false
	if s.Patch != nil {
		if oldValue, ok = s.Patch.TryGetGlobal(variableName); !ok {
			oldValue, ok = s._globalVariables[variableName]
		}
	}

	RetainListOriginsForAssignment(oldValue, value)

	if s.Patch != nil {
		s.Patch.SetGlobal(variableName, value)
	} else {
		s._globalVariables[variableName] = value
	}

	if s.VariableChangedEvent != nil && value != oldValue {
		if s._batchObservingVariableChanges {
			if s.Patch != nil {
				s.Patch.AddChangedVariable(variableName)
			} else if s._changedVariablesForBatchObs != nil {
				AddToMap(s._changedVariablesForBatchObs, variableName, struct{}{})
			}
		} else {
			s.VariableChangedEvent.Emit(variableName, value)
		}
	}
}

// ResolveVariablePointer
// Given a variable pointer with just the name of the target known, resolve to a variable
// pointer that more specifically points to the exact instance: whether it's global,
// or the exact position of a temporary on the callstack.
func (s *VariablesState) ResolveVariablePointer(varPointer *VariablePointerValue) *VariablePointerValue {

	contextIndex := varPointer.ContextIndex()

	if contextIndex == -1 {
		contextIndex = s.GetContextIndexOfVariableNamed(varPointer.Value())
	}

	valueOfVariablePointedTo := s.GetRawVariableWithName(varPointer.Value(), contextIndex)

	// Extra layer of indirection:
	// When accessing a pointer to a pointer (e.g. when calling nested or
	// recursive functions that take a variable references, ensure we don't create
	// a chain of indirection by just returning the final target.
	doubleRedirectionPointer, _ := valueOfVariablePointedTo.(*VariablePointerValue) // C# as
	if doubleRedirectionPointer != nil {
		return doubleRedirectionPointer
	} else {
		// Make copy of the variable pointer so we're not using the value direct from
		// the runtime. Temporary must be local to the current scope.
		return NewVariablePointerValueFromValue(varPointer.Value(), contextIndex)
	}
}

// GetContextIndexOfVariableNamed
// 0  if named variable is global
// 1+ if named variable is a temporary in a particular call stack element
func (s *VariablesState) GetContextIndexOfVariableNamed(varName string) int {

	if s.GlobalVariableExistsWithName(varName) {
		return 0
	}

	return s._callStack.CurrentElementIndex()
}
