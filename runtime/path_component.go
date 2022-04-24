package runtime

import (
	"errors"
	"strconv"
)

var (
	ErrPathComponentIndex = errors.New("path component: index must be zero or greater")
	ErrPathComponentName  = errors.New("path component: name must not be empty")
)

var (
	ParentPathComponent = MustNewPathComponentFromName("^")
)

type PathComponent struct {
	index int
	name  string
}

func (s PathComponent) Index() int {
	return s.index
}

func (s PathComponent) Name() string {
	return s.name
}

func (s PathComponent) IsIndex() bool {
	if s.index >= 0 {
		return true
	}
	return false
}

func (s PathComponent) IsParent() bool {
	if s.name == PathParentID {
		return true
	}
	return false
}

func (s PathComponent) String() string {
	if s.IsIndex() {
		return strconv.Itoa(s.index)
	} else {
		return s.name
	}
}

func NewPathComponentFromIndex(index int) (PathComponent, error) {
	if index < 0 {
		return PathComponent{}, ErrPathComponentIndex
	}
	return PathComponent{index: index, name: ""}, nil
}

func MustNewPathComponentFromIndex(index int) PathComponent {
	c, err := NewPathComponentFromIndex(index)
	if err != nil {
		panic(err)
	}
	return c
}

func NewPathComponentFromName(name string) (PathComponent, error) {
	if len(name) == 0 {
		return PathComponent{}, ErrPathComponentName
	}
	return PathComponent{index: -1, name: name}, nil
}

func MustNewPathComponentFromName(name string) PathComponent {
	c, err := NewPathComponentFromName(name)
	if err != nil {
		panic(err)
	}
	return c
}
