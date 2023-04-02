package runtime

type ErrorHandler func(message string, typ ErrorType)

type ErrorHandlerEvent struct {
	Event[ErrorHandler]
}

func (s *ErrorHandlerEvent) Emit(message string, typ ErrorType) {
	for _, fn := range s.h {
		fn(message, typ)
	}
}

type ErrorType int

const (
	ErrorTypeAuthor ErrorType = iota
	ErrorTypeWarning
	ErrorTypeError
)
