package runtime

type OnEvaluateFunctionEvent struct {
	Event[OnEvaluateFunction]
}

func (s *OnEvaluateFunctionEvent) Emit(functionName string, arguments []interface{}) {
	for _, fn := range s.h {
		fn(functionName, arguments)
	}
}

type OnChoosePathStringEvent struct {
	Event[OnChoosePathString]
}

func (s *OnChoosePathStringEvent) Emit(path string, arguments []interface{}) {
	for _, fn := range s.h {
		fn(path, arguments)
	}
}

type OnCompleteEvaluateFunctionEvent struct {
	Event[OnCompleteEvaluateFunction]
}

func (s *OnCompleteEvaluateFunctionEvent) Emit(functionName string, arguments []interface{}, textOutput string, result interface{}) {
	for _, fn := range s.h {
		fn(functionName, arguments, textOutput, result)
	}
}
