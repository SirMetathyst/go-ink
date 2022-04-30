package runtime

import "strconv"

type Flow struct {
	Name           string
	CallStack      *CallStack
	OutputStream   []Object
	CurrentChoices []*Choice
}

func (s *Flow) LoadFlowChoiceThreads(jChoiceThreads map[string]interface{}, story *Story) {

	for _, choice := range s.CurrentChoices {
		foundActiveThread := s.CallStack.ThreadWithIndex(choice.OriginalThreadIndex)
		if foundActiveThread != nil {
			choice.ThreadAtGeneration = foundActiveThread.Copy()
		} else {
			jSavedChoiceThread := jChoiceThreads[strconv.Itoa(choice.OriginalThreadIndex)].(map[string]interface{})
			choice.ThreadAtGeneration = NewThreadFromMap(jSavedChoiceThread, story)
		}
	}
}

func NewFlow(name string, story *Story) *Flow {
	s := &Flow{Name: name}
	s.CallStack = NewCallStack(story)
	return s
}

func NewFlowWithMap(name string, story *Story, jObject map[string]interface{}) *Flow {

	s := &Flow{Name: name}
	s.CallStack = NewCallStack(story)
	s.CallStack.SetJsonToken(jObject["callstack"].(map[string]interface{}), story)
	s.OutputStream = JArrayToRuntimeObjList(jObject["outputStream"].([]interface{}))
	s.CurrentChoices = JArrayToRuntimeObjList[*Choice](jObject["currentChoices"].([]interface{}))

	jChoiceThreadsObj, _ := jObject["choiceThreads"]
	s.LoadFlowChoiceThreads(jChoiceThreadsObj.(map[string]interface{}), story)
	return s
}
