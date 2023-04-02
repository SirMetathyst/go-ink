package runtime

import "fmt"

type Flow struct {
	Name           string
	CallStack      *CallStack
	OutputStream   []Object
	CurrentChoices []*Choice
}

func NewFlow(name string, story *Story) *Flow {

	newFlow := new(Flow)
	newFlow.Name = name
	newFlow.CallStack = NewCallStack(story)
	newFlow.OutputStream = []Object{}
	newFlow.CurrentChoices = []*Choice{}

	return newFlow
}

func NewFlowFromJObject(name string, story *Story, jObject map[string]interface{}) *Flow {

	newFlow := new(Flow)
	newFlow.Name = name
	newFlow.CallStack = NewCallStack(story)
	newFlow.CallStack.SetJsonToken(jObject["callstack"].(map[string]interface{}), story)
	newFlow.OutputStream = JArrayToRuntimeObjList[Object](jObject["outputStream"].([]interface{}), false)
	newFlow.CurrentChoices = JArrayToRuntimeObjList[*Choice](jObject["currentChoices"].([]interface{}), false)

	jChoiceThreadsObj, _ := jObject["choiceThreads"] // C# as
	newFlow.LoadFlowChoiceThreads(jChoiceThreadsObj.(map[string]interface{}), story)

	return newFlow
}

func (s *Flow) WriteJson(writer *Writer) {

	writer.WriteObjectStart()

	writer.WritePropertyStart("callstack")
	//s.CallStack.WriteJson(writer)
	writer.WritePropertyEnd()

	writer.WritePropertyStart("outputStream")
	WriteListRuntimeObjs(writer, s.OutputStream)
	writer.WritePropertyEnd()

	// choiceThreads: optional
	// Has to come BEFORE the choices themselves are written out
	// since the originalThreadIndex of each choice needs to be set
	hasChoiceThreads := false
	for _, c := range s.CurrentChoices {

		c.OriginalTheadIndex = c.ThreadAtGeneration.ThreadIndex

		if s.CallStack.ThreadWithIndex(c.OriginalTheadIndex) == nil {

			if !hasChoiceThreads {
				hasChoiceThreads = true
				writer.WritePropertyStart("choiceThreads")
				writer.WriteObjectStart()
			}

			writer.WritePropertyStart(c.OriginalTheadIndex)
			//c.ThreadAtGeneration.WriteJson(writer)
			writer.WritePropertyEnd()
		}
	}

	if hasChoiceThreads {
		writer.WriteObjectEnd()
		writer.WritePropertyEnd()
	}

	writer.WritePropertyStart("currentChoices")
	writer.WriteArrayStart()
	for _, c := range s.CurrentChoices {
		WriteChoice(writer, c)
	}
	writer.WriteArrayEnd()
	writer.WritePropertyEnd()

	writer.WriteObjectEnd()
}

// LoadFlowChoiceThreads
// Used both to load old format and current
func (s *Flow) LoadFlowChoiceThreads(jChoiceThreads map[string]interface{}, story *Story) {

	for _, choice := range s.CurrentChoices {
		foundActiveThread := s.CallStack.ThreadWithIndex(choice.OriginalTheadIndex)
		if foundActiveThread != nil {
			choice.ThreadAtGeneration = foundActiveThread.Copy()
		} else {
			jSavedChoiceThread := jChoiceThreads[fmt.Sprint(choice.OriginalTheadIndex)].(map[string]interface{})
			choice.ThreadAtGeneration = NewThreadFromJObject(jSavedChoiceThread, story)
		}
	}
}
