package runtime

import "fmt"

// The ChoicePoint represents the point within the Story where
// a Choice instance gets generated. The distinction is made
// because the text of the Choice can be dynamically generated.
type ChoicePoint struct {
	ObjectImpl

	// Private
	pathOnChoice *Path

	// Public
	HasCondition         bool
	HasStartContent      bool
	HasChoiceOnlyContent bool
	OneOnly              bool
	IsInvisibleDefault   bool
}

func (s *ChoicePoint) PathOnChoice() *Path {

	if s.pathOnChoice != nil && s.pathOnChoice.IsRelative() {
		choiceTargetObj := s.ChoiceTarget()
		if choiceTargetObj != nil {
			s.pathOnChoice = choiceTargetObj.Path()
		}
	}

	return s.pathOnChoice
}

func (s *ChoicePoint) ChoiceTarget() *Container {
	return s.ResolvePath(s.pathOnChoice).Container()
}

func (s *ChoicePoint) PathStringOnChoice() string {

	return s.CompactPathString(s.PathOnChoice())
}

func (s *ChoicePoint) SetPathStringOnChoice(value string) {
	s.pathOnChoice = NewPathFromComponentString(value)
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
	if s.OneOnly {
		flags |= 16
	}
	return flags
}

func (s *ChoicePoint) SetFlags(value int) {
	s.HasCondition = (value & 1) > 0
	s.HasStartContent = (value & 2) > 0
	s.HasChoiceOnlyContent = (value & 4) > 0
	s.IsInvisibleDefault = (value & 8) > 0
	s.OneOnly = (value & 16) > 0
}

func NewChoicePointOneOnly(oneOnly bool) *ChoicePoint {

	newChoicePoint := new(ChoicePoint)
	newChoicePoint.this = newChoicePoint
	newChoicePoint.OneOnly = oneOnly

	return newChoicePoint
}

func NewChoicePoint() *ChoicePoint {

	newChoicePoint := new(ChoicePoint)
	newChoicePoint.this = newChoicePoint
	newChoicePoint.OneOnly = true

	return newChoicePoint
}

func (s *ChoicePoint) String() string {

	targetLineNum, ok := s.DebugLineNumberOfPath(s.pathOnChoice)
	targetString := s.pathOnChoice.String()

	if ok {
		targetString = " line " + fmt.Sprint(targetLineNum) + "(" + targetString + ")"
	}

	return "Choice: -> " + targetString
}
