package runtime

type ErrorHandler func(message string, errType ErrorType)

type ErrorType int

const (
	ErrTypeAuthor ErrorType = iota
	ErrTypeWarning
	ErrTypeError
)
