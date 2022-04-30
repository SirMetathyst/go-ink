package runtime

type Stack[T any] []T

func (s *Stack[T]) IsEmpty() bool {
	return len(*s) == 0
}

func (s *Stack[T]) Push(str T) {
	*s = append(*s, str)
}

func (s *Stack[T]) Peek() T {
	if s.IsEmpty() {
		var result T
		return result
	} else {
		index := len(*s) - 1
		element := (*s)[index]
		return element
	}
}

func (s *Stack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var result T
		return result, false
	} else {
		index := len(*s) - 1
		element := (*s)[index]
		*s = (*s)[:index]
		return element, true
	}
}
