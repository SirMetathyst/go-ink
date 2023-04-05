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

type Equals interface {
	Equals(v interface{}) bool
}

func NewMapFromMap[TKey comparable, TValue any](v map[TKey]TValue) map[TKey]TValue {
	c := make(map[TKey]TValue, len(v))
	for key, value := range v {
		c[key] = value
	}
	return c
}

func NewSliceFromSlice[T any](s []T) []T {
	n := make([]T, len(s))
	copy(n, s)
	return n
}

func ClearMap[TKey comparable, TValue any](v map[TKey]TValue) {
	for key, _ := range v {
		delete(v, key)
	}
}

func Remove[T comparable](s *[]T, v interface{}) {
	n := *s
	c := len(*s)
	for index, vv := range n {
		if vv == v {
			n = append(n[:index], n[index+1:]...)
		}
	}
	nc := len(*s)
	if nc == c {
		panic("should not happen")
	}
}

func AddToMap[TKey comparable, TValue any](s map[TKey]TValue, key TKey, value TValue) {
	if _, ok := s[key]; !ok {
		s[key] = value
		return
	}
	panic("key already in map")
}

type KeyValuePair[TKey any, TValue any] struct {
	Key   TKey
	Value TValue
}
