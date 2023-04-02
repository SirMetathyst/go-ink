package runtime

import "fmt"

// The ChoicePoint represents the point within the Story where
// a Choice instance gets generated. The distinction is made
// because the text of the Choice can be dynamically generated.
type ChoicePoint struct {
	ObjectImpl

	// Private
	_pathOnChoice *Path

	// Public
	HasCondition         bool
	HasStartContent      bool
	HasChoiceOnlyContent bool
	OnceOnly             bool
	IsInvisibleDefault   bool
}

func (s *ChoicePoint) Flags() int {

	rflags := 0

	if s.HasCondition {
		rflags |= 1
	}
	if s.HasStartContent {
		rflags |= 2
	}
	if s.HasChoiceOnlyContent {
		rflags |= 4
	}
	if s.IsInvisibleDefault {
		rflags |= 8
	}
	if s.OnceOnly {
		rflags |= 16
	}

	return rflags
}

func (s *ChoicePoint) SetFlags(value int) {

	s.HasCondition = (value & 1) > 0
	s.HasStartContent = (value & 2) > 0
	s.HasChoiceOnlyContent = (value & 4) > 0
	s.IsInvisibleDefault = (value & 8) > 0
	s.OnceOnly = (value & 16) > 0

}

func (s *ChoicePoint) PathOnChoice() *Path {

	// Resolve any relative paths to global ones as we come across them
	if s._pathOnChoice != nil && s._pathOnChoice.IsRelative() {

		choiceTargetObj := s.ChoiceTarget()

		if choiceTargetObj != nil {
			s._pathOnChoice = choiceTargetObj.Path(s)
		}
	}

	return s._pathOnChoice
}

func (s *ChoicePoint) SetPathOnChoice(value *Path) {

	s._pathOnChoice = value
}

func (s *ChoicePoint) ChoiceTarget() *Container {

	return ResolvePath(s, s._pathOnChoice).Container()
}

func (s *ChoicePoint) PathStringOnChoice() string {

	return CompactPathString(s, s.PathOnChoice())
}

func (s *ChoicePoint) SetPathStringOnChoice(value string) {

	s.SetPathOnChoice(NewPathFromString(value))
}

func NewChoicePointOnceOnly(onceOnly bool) *ChoicePoint {

	newChoicePoint := new(ChoicePoint)
	newChoicePoint.OnceOnly = onceOnly

	return newChoicePoint
}

func NewChoicePoint() *ChoicePoint {

	newChoicePoint := new(ChoicePoint)
	newChoicePoint.OnceOnly = true

	return newChoicePoint
}

func (s *ChoicePoint) String() string {

	targetLineNum, ok := s.DebugLineNumberOfPath(s, s.PathOnChoice())
	targetString := s.PathOnChoice().String()

	if ok {
		targetString = " line " + fmt.Sprint(targetLineNum) + "(" + targetString + ")"
	}

	return "Choice: -> " + targetString
}
