package runtime

import "fmt"

var NullPointer = NewPointer(nil, -1)

// Pointer
// Internal structure used to point to a particular / current point in the story.
// Where Path is a set of components that make content fully addressable, this is
// a reference to the current container, and the index of the current piece of
// content within that container. This scheme makes it as fast and efficient as
// possible to increment the pointer (move the story forwards) in a way that's as
// native to the internal engine as possible.
type Pointer struct {
	Container *Container
	Index     int
}

func NewPointer(container *Container, index int) Pointer {

	newPointer := Pointer{}
	newPointer.Container = container
	newPointer.Index = index

	return newPointer
}

func (s Pointer) Resolve() Object {

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

	return s.Container.Content()[s.Index]
}

func (s Pointer) IsNull() bool {
	return s.Container == nil
}

func (s Pointer) Path() *Path {

	if s.IsNull() {
		return nil
	}

	if s.Index >= 0 {
		return s.Container.Path(s.Container).PathByAppendingComponent(NewPathComponentFromIndex(s.Index))
	}

	return s.Container.Path(s.Container)
}

func (s Pointer) String() string {

	if s.Container == nil {
		return "Ink Pointer (null)"
	}

	return "Ink Pointer -> " + s.Container.Path(s.Container).String() + " -- index " + fmt.Sprint(s.Index)
}

func StartOfPointer(container *Container) Pointer {
	return NewPointer(container, 0)
}
