package runtime

type Choice struct {
	*objectImpl
	Text                string
	TargetPath          *Path
	SourcePath          string
	ThreadAtGeneration  *CallStackThread
	Index               int
	OriginalThreadIndex int
	IsInvisibleDefault  bool
}

func (s *Choice) PathStringOnChoice() string {
	return s.TargetPath.String()
}

func (s *Choice) SetPathStringOnChoice(path string) {
	s.TargetPath = NewPathFromComponentsString(path)
}

func NewChoice() *Choice {
	s := &Choice{}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
