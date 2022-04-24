package runtime

import "fmt"

type ChoicePoint struct {
	*objectImpl
	HasCondition         bool
	HasStartContent      bool
	HasChoiceOnlyContent bool
	OnceOnly             bool
	IsInvisibleDefault   bool
	pathOnChoice         *Path
	choiceTarget         *Container
}

func (s *ChoicePoint) Flags() int {
	flags := 0
	if s.HasCondition {
		flags |= 1
	}
	if s.HasStartContent {
		flags |= 2
	}
	if s.HasChoiceOnlyContent {
		flags |= 4
	}
	if s.IsInvisibleDefault {
		flags |= 8
	}
	if s.OnceOnly {
		flags |= 16
	}
	return flags
}

func (s *ChoicePoint) SetFlags(value int) {
	s.HasCondition = (value & 1) > 0
	s.HasStartContent = (value & 2) > 0
	s.HasChoiceOnlyContent = (value & 4) > 0
	s.IsInvisibleDefault = (value & 8) > 0
	s.OnceOnly = (value & 16) > 0
}

func (s *ChoicePoint) PathOnChoice() *Path {

	if s.pathOnChoice != nil && s.pathOnChoice.IsRelative() {
		choiceTargetObj := s.choiceTarget
		if choiceTargetObj != nil {
			s.pathOnChoice = choiceTargetObj.Path()
		}
	}

	return s.pathOnChoice
}

func (s *ChoicePoint) SetPathOnChoice(pathOnChoice *Path) {
	s.pathOnChoice = pathOnChoice
}

func (s *ChoicePoint) ChoiceTarget() *Container {
	return s.choiceTarget
}

func (s *ChoicePoint) PathStringOnChoice() string {
	return s.CompactPathString(s.pathOnChoice)
}

func (s *ChoicePoint) SetPathStringOnChoice(pathString string) {
	s.pathOnChoice = NewPathFromComponentsString(pathString)
}

func (s *ChoicePoint) String() string {
	targetLineNum, ok := s.DebugLineNumberOfPath(s.pathOnChoice)
	targetString := s.pathOnChoice.String()
	if ok {
		targetString = fmt.Sprintf("line %d (%s)", targetLineNum, targetString)
	}
	return "Choice: -> " + targetString
}

func NewChoicePoint(onceOnly bool) *ChoicePoint {
	s := &ChoicePoint{OnceOnly: onceOnly}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
