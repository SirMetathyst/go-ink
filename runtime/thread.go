package runtime

type Thread struct {
	Callstack       []*Element
	ThreadIndex     int
	PreviousPointer *Pointer
}

func (s *Thread) Copy() *Thread {
	copy := NewThread()
	copy.ThreadIndex = s.ThreadIndex
	for _, elm := range s.Callstack {
		copy.Callstack = append(copy.Callstack, elm.Copy())
	}
	copy.PreviousPointer = s.PreviousPointer
	return copy
}

func NewThread() *Thread {
	return &Thread{}
}

func NewThreadFromMap(jThreadObj map[string]interface{}, storyContext *Story) *Thread {

	newThread := &Thread{}
	newThread.ThreadIndex = jThreadObj["threadIndex"].(int)
	jThreadCallstack := jThreadObj["callstack"].([]interface{})

	for _, jElTok := range jThreadCallstack {

		jElementObj := jElTok.(map[string]interface{})
		pushPopType := jElementObj["type"].(int)
		pointer := NewNilPointer()
		currentContainerPathStr := ""

		if currentContainerPathStrToken, ok := jElementObj["cPath"]; ok {

			currentContainerPathStr = currentContainerPathStrToken.(string)
			threadPointerResult := storyContext.ContentAtPath(NewPathFromComponentsString(currentContainerPathStr))
			pointer.Container = threadPointerResult.Container
			pointer.Index = jElementObj["idx"].(int)

			if threadPointerResult.Obj == nil {
				panic("When loading state, internal story location couldn't be found: " + currentContainerPathStr + ". Has the story changed since this save data was created?")
			} else if threadPointerResult.Approximate {
				storyContext.Warning("When loading state, exact internal story location couldn't be found: '" + currentContainerPathStr + "', so it was approximated to '"+pointer.container.path.ToString()+"' to recover. Has the story changed since this save data was created?)
			}
		}

		inExpressionEvaluation := jElementObj["exp"].(bool)
		el := NewElementWith(pushPopType, pointer, inExpressionEvaluation)

		if temps, ok := jElementObj["temp"]; ok {
			el.TemporaryVariables = JObjectToMapRuntimeObjs(temps.(map[string]interface{}))
		} else {
			el.TemporaryVariables = map[string]Object{}
		}

		newThread.Callstack = append(newThread.Callstack, el)
	}

	if prevContentObjPath, ok := jThreadObj["previousContentObject"]; ok {
		prevPath := NewPathFromComponentsString(prevContentObjPath.(string))
		newThread.PreviousPointer = storyContext.PointerAtPath(prevPath)
	}

	return newThread
}
