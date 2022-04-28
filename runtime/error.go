package runtime

import "fmt"

type ErrorHandler func(message string, errType ErrorType)

type ErrorType int

const (
	ErrTypeAuthor ErrorType = iota
	ErrTypeWarning
	ErrTypeError
)

type StoryError struct {
	UseEndLineNumber bool
	message          string
}

func (s *StoryError) Error() string {
	return fmt.Sprintf("story runtime error: %s", s.message)
}

func NewStoryError(message string) *StoryError {
	return &StoryError{message: message}
}
