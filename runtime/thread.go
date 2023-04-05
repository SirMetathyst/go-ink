package runtime

type Thread struct {

	// Private
	_elements []*Element

	// Public
	ThreadIndex     int
	PreviousPointer Pointer
}

func (s *Thread) Elements() []*Element {
	return s._elements
}

func (s *Thread) Add(value *Element) {
	s._elements = append(s._elements, value)
}

func (s *Thread) RemoveLast() {
	s._elements = s._elements[:len(s._elements)-1]
}

func NewThread() *Thread {

	newThread := new(Thread)
	newThread._elements = []*Element{}

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

			currentContainerPathStr := currentContainerPathStrToken.(string)
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

		newThread.Add(el)
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

	for _, e := range s.Elements() {
		threadCopy.Add(e.Copy())
	}

	return threadCopy
}

// TODO: Thread.WriteJson
