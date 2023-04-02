package runtime

type StatePatch struct {

	// Private
	_globals          map[string]Object
	_changedVariables map[string]struct{}
	_visitCounts      map[*Container]int
	_turnIndices      map[*Container]int
}

func (s *StatePatch) Globals() map[string]Object {
	return s._globals
}

func (s *StatePatch) ChangedVariables() map[string]struct{} {
	return s._changedVariables
}

func (s *StatePatch) VisitCounts() map[*Container]int {
	return s._visitCounts
}

func (s *StatePatch) TurnIndices() map[*Container]int {
	return s._turnIndices
}

func NewStatePatchFromStatePatch(toCopy *StatePatch) *StatePatch {

	newStatePatch := new(StatePatch)
	newStatePatch._globals = make(map[string]Object)
	newStatePatch._changedVariables = make(map[string]struct{})
	newStatePatch._visitCounts = make(map[*Container]int)
	newStatePatch._turnIndices = make(map[*Container]int)

	if toCopy != nil {

		for k, v := range toCopy.Globals() {
			AddToMap(newStatePatch.Globals(), k, v)
		}

		for k, v := range toCopy.ChangedVariables() {
			AddToMap(newStatePatch.ChangedVariables(), k, v)
		}

		for k, v := range toCopy.VisitCounts() {
			AddToMap(newStatePatch.VisitCounts(), k, v)
		}

		for k, v := range toCopy.TurnIndices() {
			AddToMap(newStatePatch.TurnIndices(), k, v)
		}
	}

	return newStatePatch
}

func (s *StatePatch) TryGetGlobal(name string) (Object, bool) {
	v, ok := s._globals[name]
	return v, ok
}

func (s *StatePatch) SetGlobal(name string, value Object) {
	s._globals[name] = value
}

func (s *StatePatch) AddChangedVariable(name string) {
	AddToMap(s._changedVariables, name, struct{}{})
}

func (s *StatePatch) TryGetVisitCount(container *Container) (int, bool) {
	v, ok := s._visitCounts[container]
	return v, ok
}

func (s *StatePatch) SetVisitCount(container *Container, index int) {
	s._visitCounts[container] = index
}

func (s *StatePatch) SetTurnIndex(container *Container, index int) {
	s._turnIndices[container] = index
}

func (s *StatePatch) TryGetTurnIndex(container *Container) (int, bool) {
	v, ok := s._turnIndices[container]
	return v, ok
}
