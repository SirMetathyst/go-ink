package runtime

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	ParentID = "^"
)

type Path struct {
	isRelative       bool
	components       []*PathComponent
	componentsString string
}

func (s *Path) Component(index int) *PathComponent {
	return s.components[index]
}

func (s *Path) IsRelative() bool {
	return s.isRelative
}

func (s *Path) Head() *PathComponent {

	if len(s.components) > 0 {
		return s.components[0]
	} else {
		return nil
	}
}

func (s *Path) Tail() *Path {

	if len(s.components) >= 2 {
		tailComps := s.components[1:]
		return NewPathFromComponents(tailComps, false)
	} else {
		return NewSelfPath()
	}
}

func (s *Path) Length() int {
	return len(s.components)
}

func (s *Path) LastComponent() *PathComponent {

	lastComponentIdx := len(s.components) - 1
	if lastComponentIdx >= 0 {
		return s.components[lastComponentIdx]
	} else {
		return nil
	}
}

func (s *Path) ContainsNamedComponent() bool {

	for _, comp := range s.components {
		if !comp.IsIndex() {
			return true
		}
	}

	return false
}

func NewPath() *Path {
	return new(Path)
}

func NewPathFromHeadTail(head *PathComponent, tail *Path) *Path {

	newPath := new(Path)
	newPath.components = append(newPath.components, head)
	newPath.components = append(newPath.components, tail.components...)

	return newPath
}

func NewPathFromComponents(components []*PathComponent, relative bool) *Path {

	newPath := new(Path)
	newPath.components = append(newPath.components, components...)
	newPath.isRelative = relative

	return newPath
}

func NewPathFromComponentString(componentsString string) *Path {

	newPath := new(Path)
	newPath.SetComponentsString(componentsString)

	return newPath
}

func NewSelfPath() *Path {

	newPath := new(Path)
	newPath.isRelative = true

	return newPath
}

func (s *Path) NewPathByAppendingPath(pathToAppend *Path) *Path {

	p := new(Path)

	upwardMoves := 0
	for i := 0; i < len(pathToAppend.components); i++ {
		if pathToAppend.components[i].IsParent() {
			upwardMoves++
		} else {
			break
		}
	}

	for i := 0; i < len(s.components)-upwardMoves; i++ {
		p.components = append(p.components, s.components[i])
	}

	for i := upwardMoves; i < len(pathToAppend.components); i++ {
		p.components = append(p.components, pathToAppend.components[i])
	}

	return p
}

func (s *Path) NewPathByAppendingComponent(c *PathComponent) *Path {

	p := new(Path)
	p.components = append(p.components, s.components...)
	p.components = append(p.components, c)

	return p
}

func (s *Path) ComponentsString() string {

	if s.componentsString == "" {
		s.componentsString = s.join(".", s.components)
		if s.isRelative {
			s.componentsString = "." + s.componentsString
		}
	}

	return s.componentsString
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

	s.components = nil
	s.componentsString = value

	// Empty path, empty components
	// (path is to root, like "/" in file system)
	if s.componentsString == "" {
		return
	}

	// When components start with ".", it indicates a relative path, e.g.
	//   .^.^.hello.5
	// is equivalent to file system style path:
	//  ../../hello/5
	if s.componentsString[0] == '.' {
		s.isRelative = true
		s.componentsString = s.componentsString[1:]
	} else {
		s.isRelative = false
	}

	componentStrings := strings.Split(s.componentsString, ".")
	for _, str := range componentStrings {
		index, err := strconv.Atoi(str)
		if err == nil {
			s.components = append(s.components, NewPathComponentFromIndex(index))
		} else {
			s.components = append(s.components, NewPathComponentFromName(str))
		}
	}
}

func (s *Path) String() string {
	return s.componentsString
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

	for index, otherPathComponent := range otherPath.components {
		if !otherPathComponent.Equals(s.components[index]) {
			return false
		}
	}

	return true
}

func (s *Path) HashCode() string {
	return s.String()
}

type PathComponent struct {
	index    int
	name     string
	isIndex  bool
	isParent bool
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
	return s.name == ParentID
}

func NewPathComponentFromIndex(index int) *PathComponent {

	// Debug.Assert(index >= 0);
	if index < 0 {
		panic("index is not greater or equal to 0")
	}

	newPathComponent := new(PathComponent)
	newPathComponent.index = index
	newPathComponent.name = ""

	return newPathComponent
}

func NewPathComponentFromName(name string) *PathComponent {

	if len(name) <= 0 {
		panic("name length is zero")
	}

	newPathComponent := new(PathComponent)
	newPathComponent.index = -1
	newPathComponent.name = name

	return newPathComponent
}

func PathComponentToParent() *PathComponent {
	return NewPathComponentFromName(ParentID)
}

func (s *PathComponent) String() string {

	if s.isIndex {
		return fmt.Sprint(s.index)
	} else {
		return s.name
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

func (s *PathComponent) HashCode() string {

	if s.isIndex {
		return fmt.Sprint(s.index)
	} else {
		return s.name
	}
}
