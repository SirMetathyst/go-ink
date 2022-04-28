package runtime

type StatePatch struct {
	globals          map[string]Object
	changedVariables []string
	visitCounts      map[*Container]int
	turnIndices      map[*Container]int
}

func (s *StatePatch) Globals() map[string]Object {
	return s.globals
}

func (s *StatePatch) ChangedVariables() []string {
	return s.changedVariables
}

func (s *StatePatch) VisitCounts() map[*Container]int {
	return s.visitCounts
}

func (s *StatePatch) TurnIndices() map[*Container]int {
	return s.turnIndices
}

func (s *StatePatch) TryGetGlobal(name string) (Object, bool) {
	v, ok := s.globals[name]
	return v, ok
}

func (s *StatePatch) SetGlobal(name string, value Object) {
	s.globals[name] = value
}

func (s *StatePatch) AddChangedVariable(name string) {
	s.changedVariables = append(s.changedVariables, name)
}

func (s *StatePatch) TryGetVisitCount(container *Container) (int, bool) {
	v, ok := s.visitCounts[container]
	return v, ok
}

func (s *StatePatch) SetVisitCount(container *Container, count int) {
	s.visitCounts[container] = count
}

func (s *StatePatch) SetTurnIndex(container *Container, index int) {
	s.turnIndices[container] = index
}

func (s *StatePatch) TryGetTurnIndex(container *Container) (int, bool) {
	v, ok := s.turnIndices[container]
	return v, ok
}

func NewStatePatch(copy *StatePatch) *StatePatch {
	s := &StatePatch{}
	s.globals = make(map[string]Object, 0)
	s.visitCounts = make(map[*Container]int, 0)
	s.turnIndices = make(map[*Container]int, 0)

	if copy != nil {

		for key, value := range copy.Globals() {
			s.globals[key] = value
		}

		for key, value := range copy.VisitCounts() {
			s.visitCounts[key] = value
		}

		for key, value := range copy.TurnIndices() {
			s.turnIndices[key] = value
		}

		for _, value := range copy.ChangedVariables() {
			s.changedVariables = append(s.changedVariables, value)
		}
	}

	return s
}
