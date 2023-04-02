package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

type Path struct {
	_isRelative       bool
	_components       []*PathComponent
	_componentsString string
}

func (s *Path) Component(index int) *PathComponent {
	return s._components[index]
}

func (s *Path) IsRelative() bool {
	return s._isRelative
}

func (s *Path) Tail() *Path {

	if len(s._components) >= 2 {

		//tailComps := s._components.GetRange(1, s._components.Count()-1)
		tailComps := s._components[1:]

		return NewPathFromComponents(tailComps, false)

	}

	newPath := NewPath()
	newPath._isRelative = true

	return newPath
}

func (s *Path) Length() int {
	return len(s._components)
}

func (s *Path) LastComponent() *PathComponent {

	lastComponentIdx := len(s._components) - 1

	if lastComponentIdx >= 0 {
		return s._components[lastComponentIdx]
	}

	return nil
}

func NewPath() *Path {

	newPath := new(Path)
	newPath._components = []*PathComponent{}

	return newPath
}

func NewPathFromComponents(components []*PathComponent, relative bool) *Path {

	newPath := NewPath()
	newPath._components = append([]*PathComponent{}, components...)
	newPath._isRelative = relative

	return newPath
}

func NewPathFromString(componentsString string) *Path {

	newPath := NewPath()
	newPath.SetComponentsString(componentsString)

	return newPath
}

func (s *Path) PathByAppendingPath(pathToAppend *Path) *Path {

	p := NewPath()

	upwardMoves := 0
	for i := 0; i < len(pathToAppend._components); i++ {
		if pathToAppend._components[i].IsParent() {
			upwardMoves++
		} else {
			break
		}
	}

	for i := 0; i < len(s._components)-upwardMoves; i++ {
		p._components = append(p._components, s._components[i])
	}

	for i := upwardMoves; i < len(pathToAppend._components); i++ {
		p._components = append(p._components, pathToAppend._components[i])
	}

	return p
}

func (s *Path) PathByAppendingComponent(c *PathComponent) *Path {

	p := NewPath()
	p._components = append(s._components, c)

	return p
}

func (s *Path) ComponentsString() string {

	//if s._componentsString == "" {
	s._componentsString = s.join(".", s._components)
	if s.IsRelative() {
		s._componentsString = "." + s._componentsString
	}
	//}

	return s._componentsString
}

func (s *Path) join(separator string, components []*PathComponent) string {

	var sb strings.Builder

	isFirst := true
	for _, component := range components {

		if !isFirst {
			sb.WriteString(separator)
		}

		sb.WriteString(component.String())
		isFirst = false
	}

	return sb.String()
}

func (s *Path) SetComponentsString(value string) {

	s._components = s._components[:0]
	s._componentsString = value

	// Empty path, empty components
	// (path is to root, like "/" in file system)
	if s._componentsString == "" {
		return
	}

	// When components start with ".", it indicates a relative path, e.g.
	//   .^.^.hello.5
	// is equivalent to file system style path:
	//  ../../hello/5
	if s._componentsString[0] == '.' {
		s._isRelative = true
		s._componentsString = s._componentsString[1:]
	} else {
		s._isRelative = false
	}

	componentStrings := strings.Split(s._componentsString, ".")

	for _, str := range componentStrings {

		if index, err := strconv.Atoi(str); err == nil {
			s._components = append(s._components, NewPathComponentFromIndex(index))
		} else {
			s._components = append(s._components, NewPathComponentFromName(str))
		}
	}
}

func (s *Path) String() string {
	return s.ComponentsString()
}

func (s *Path) Equals(otherPath *Path) bool {

	if otherPath == nil {
		return false
	}

	if len(otherPath._components) != len(s._components) {
		return false
	}

	if otherPath.IsRelative() != s.IsRelative() {
		return false
	}

	for i, otherPathComponent := range otherPath._components {
		if !otherPathComponent.Equals(s._components[i]) {
			return false
		}
	}

	return true
}

type PathComponent struct {
	_index int
	_name  string
}

func (s *PathComponent) Index() int {
	return s._index
}

func (s *PathComponent) Name() string {
	return s._name
}

func (s *PathComponent) IsIndex() bool {
	return s._index >= 0
}

func (s *PathComponent) IsParent() bool {
	return s._name == "^"
}

func NewPathComponentFromIndex(index int) *PathComponent {

	// Debug.Assert(index >= 0);
	if index < 0 {
		panic("index is not greater or equal to 0")
	}

	newPathComponent := new(PathComponent)
	newPathComponent._index = index
	newPathComponent._name = ""

	return newPathComponent
}

func NewPathComponentFromName(name string) *PathComponent {

	if len(name) <= 0 {
		panic("name length is zero")
	}

	newPathComponent := new(PathComponent)
	newPathComponent._name = name
	newPathComponent._index = -1

	return newPathComponent
}

func PathComponentToParent() *PathComponent {
	return NewPathComponentFromName("^")
}

func (s *PathComponent) String() string {

	if s.IsIndex() {
		return fmt.Sprint(s._index)
	}

	return s._name
}

func (s *PathComponent) Equals(otherComp *PathComponent) bool {

	if otherComp != nil && otherComp.IsIndex() == s.IsIndex() {
		if s.IsIndex() {
			return s.Index() == otherComp.Index()
		}

		return s.Name() == otherComp.Name()
	}

	return false
}
