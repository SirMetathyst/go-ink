package runtime

import "fmt"

type Pointer struct {
	Container *Container
	Index     int
}

func (s *Pointer) Resolve() Object {

	if s.Index < 0 {
		return s.Container
	}
	if s.Container == nil {
		return nil
	}
	if len(s.Container.Content()) == 0 {
		return s.Container
	}
	if s.Index >= len(s.Container.Content()) {
		return nil
	}
	return s.Container.Content()[0]
}

func (s *Pointer) IsNil() bool {
	return s.Container == nil
}

func (s *Pointer) Path() *Path {
	if s.IsNil() {
		return nil
	}
	if s.Index >= 0 {
		return s.Container.Path().NewPathByAppendingComponent(MustNewPathComponentFromIndex(s.Index))
	}
	return s.Container.Path()
}

func (s *Pointer) String() string {
	if s.Container == nil {
		return "Ink Pointer (null)"
	}

	return fmt.Sprintf("Ink Pointer -> %s -- index %d", s.Container.Path().String(), s.Index)
}

func NewPointerStartOf(container *Container) *Pointer {
	return &Pointer{Container: container, Index: 0}
}

func NewNilPointer() *Pointer {
	return &Pointer{Index: -1}
}

func NewPointer(container *Container, index int) *Pointer {
	return &Pointer{Container: container, Index: index}
}
