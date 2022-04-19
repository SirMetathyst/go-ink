package ink

import (
	"strconv"
	"strings"
)

// Note: In the original c# library path/path component uses Debug.Assert
// which is not exactly equivalent to go's panic because windows allows you to continue
// but if the debug.assert is hit then something has gone wrong. So will use panic as
// replacement.

var (
	parentId = "^"
)

type PathComponent struct {
	index int
	name  string
}

func (s *PathComponent) Index() int {
	return s.index
}

func (s *PathComponent) Name() string {
	return s.name
}

func (s *PathComponent) IsIndex() bool {
	return s.index >= 0
}

func (s *PathComponent) IsParent() bool {
	return s.name == parentId
}

func (s *PathComponent) String() string {
	if s.IsIndex() {
		return strconv.Itoa(s.Index())
	} else {
		return s.Name()
	}
}

func (s *PathComponent) Equals(otherComp *PathComponent) bool {
	if otherComp != nil && otherComp.IsIndex() == s.IsIndex() {
		if s.IsIndex() {
			return s.Index() == otherComp.Index()
		} else {
			return s.Name() == otherComp.Name()
		}
	}
	return false
}

// Go does not have built-in hash code functionality I have created in hash_code.go

func (s *PathComponent) HashCode() int {
	if s.IsIndex() {
		return s.Index()
	}
	return hashCodeFromString(s.Name())
}

// Go does not provide built-in functionality for constructors

func newPathComponentFromIndex(index int) *PathComponent {
	if index < 0 {
		panic("index is < 0")
	}
	return &PathComponent{
		index: index,
		name:  "",
	}
}

func newPathComponentFromString(name string) *PathComponent {
	if len(name) == 0 {
		panic("name == 0")
	}
	return &PathComponent{
		index: -1,
		name:  name,
	}
}

func newPathComponentFromParent() *PathComponent {
	return newPathComponentFromString(parentId)
}

type Path struct {
	isRelative       bool
	head             *PathComponent
	components       []*PathComponent
	componentsString string
}

func (s *Path) IsRelative() bool {
	return s.isRelative
}

func (s *Path) Head() *PathComponent {
	if len(s.components) > 0 {
		return s.components[0]
	}
	return nil
}

func (s *Path) Tail() *Path {
	if len(s.components) >= 2 {
		// GetRange (https://github.com/inkle/ink/blob/master/ink-engine-runtime/Path.cs#L104) returns a new list
		// but go slices use the same underlying array, so we make a copy before using it
		tailCompsOriginal := s.components[1 : len(s.components)-1]
		tailComps := make([]*PathComponent, len(tailCompsOriginal))
		copy(tailComps, tailCompsOriginal)
		return NewPathFromComponents(tailComps)
	}

	return NewPathSelf()
}

func (s *Path) LastComponent() *PathComponent {

	lastComponentIdx := len(s.components) - 1
	if lastComponentIdx >= 0 {
		return s.components[lastComponentIdx]
	}

	return nil
}

func (s *Path) Length() int {
	return len(s.components)
}

func (s *Path) Component(index int) *PathComponent {
	return s.components[index]
}

func (s *Path) ContainsNamedComponent() bool {
	for _, comp := range s.components {
		if !comp.IsIndex() {
			return true
		}
	}
	return false
}

func (s *Path) setComponentsString(str string) {
	s.components = s.components[:]
	s.componentsString = str

	// Empty path, empty components
	// (path is to root, like "/" in file system)
	if IsStringEmpty(s.componentsString) {
		return
	}

	// When components start with ".", it indicates a relative path, e.g.
	//   .^.^.hello.5
	// is equivalent to file system style path:
	//  ../../hello/5
	if s.componentsString[0] == '.' {
		s.isRelative = true
		// C# equivalent to String.Substring(1)
		s.componentsString = s.componentsString[1:]
	} else {
		s.isRelative = false
	}

	var componentStrings = strings.Split(s.componentsString, ".")
	for _, str := range componentStrings {
		if index, err := strconv.Atoi(str); err == nil {
			s.components = append(s.components, newPathComponentFromIndex(index))
		} else {
			s.components = append(s.components, newPathComponentFromString(str))
		}
	}
	// For some reason not adding this will cause the tests for paths with a dot to fail.
	s.componentsString = ""
}

func (s *Path) ComponentsString() string {
	if s.componentsString == "" {
		s.componentsString = JoinObjectsString(".", s.components)
		if s.IsRelative() {
			s.componentsString = "." + s.componentsString
		}
	}
	return s.componentsString
}

func (s *Path) PathByAppendingPath(pathToAppend *Path) *Path {
	p := NewPath()
	upwardMoves := 0

	for i := 1; i < len(pathToAppend.components); i++ {
		if pathToAppend.components[i].IsParent() {
			upwardMoves++
		} else {
			break
		}
	}

	for i := 1; i < len(s.components)-upwardMoves; i++ {
		p.components = append(p.components, s.components[i])
	}

	for i := upwardMoves + 1; i < len(pathToAppend.components); i++ {
		p.components = append(p.components, pathToAppend.components[i])
	}

	return p
}

func (s *Path) PathByAppendingComponent(c *PathComponent) *Path {
	p := NewPath()
	p.components = append(p.components, s.components...)
	p.components = append(p.components, c)
	return p
}

func (s *Path) Equals(otherPath *Path) bool {
	if otherPath == nil {
		return false
	}

	if len(otherPath.components) != len(s.components) {
		return false
	}

	if otherPath.IsRelative() != s.IsRelative() {
		return false
	}

	// C# linq SequenceEqual
	if len(otherPath.components) != len(s.components) {
		return false
	}

	for i := range otherPath.components {
		if s.components[i] != otherPath.components[i] {
			return false
		}
	}

	return true
}

func (s *Path) String() string {
	return s.ComponentsString()
}

func (s *Path) HashCode() int {
	return hashCodeFromString(s.String())
}

func NewPath() *Path {
	return &Path{}
}

func NewPathSelf() *Path {
	path := &Path{}
	path.isRelative = true
	return path
}

func NewPathFromHeadTail(head *PathComponent, tail *Path) *Path {
	path := &Path{}
	path.components = append(path.components, head)
	path.components = append(path.components, tail.components...)
	return path
}

// go does not support optional parameters, so I've separated these into two constructors

func NewPathFromComponents(components []*PathComponent) *Path {
	path := &Path{}
	path.components = append(path.components, components...)
	path.isRelative = false
	return path
}

func NewPathFromComponentsRelative(components []*PathComponent) *Path {
	path := &Path{}
	path.components = append(path.components, components...)
	path.isRelative = true
	return path
}

func NewPathFromComponentString(componentString string) *Path {
	path := &Path{}
	path.setComponentsString(componentString)
	return path
}
