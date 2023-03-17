package runtime

import "sync"

type Stack[T any] struct {
	items []T
	mutex sync.Mutex
}

func NewStack[T any]() *Stack[T] {
	return new(Stack[T])
}

func (s *Stack[T]) Dump() []T {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var copiedStack = make([]T, len(s.items))
	copy(copiedStack, s.items)

	return copiedStack
}

func (s *Stack[T]) Peek() (T, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.items) == 0 {
		var item T
		return item, false
	}

	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = nil
}

func (s *Stack[T]) Push(item T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.items = append(s.items, item)
}

func (s *Stack[T]) Len() int {
	return len(s.items)
}

func (s *Stack[T]) IsEmpty() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.items) == 0
}

func (s *Stack[T]) Pop() (T, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.items) == 0 {
		var item T
		return item, false
	}

	lastItem := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]

	return lastItem, true
}
