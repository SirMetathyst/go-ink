package runtime

// Choice
// A generated Choice from the story.
// A single ChoicePoint in the Story could potentially generate
// different Choices dynamically dependent on state, so they're
// separated.
type Choice struct {
	ObjectImpl

	// Public
	Text string

	// Get the path to the original choice point -
	// where was this choice defined in the story?
	SourcePath string

	// The original index into currentChoices list on the Story when
	// this Choice was generated, for convenience.
	Index              int
	TargetPath         *Path
	ThreadAtGeneration *Thread
	OriginalTheadIndex int
	IsInvisibleDefault bool
	Tags               []string
}

func (s *Choice) PathStringOnChoice() string {
	return s.TargetPath.String()
}

func (s *Choice) SetPathStringOnChoice(value string) {
	s.TargetPath = NewPathFromComponentString(value)
}

func NewChoice() *Choice {

	newChoice := new(Choice)
	newChoice.this = newChoice

	return newChoice
}
