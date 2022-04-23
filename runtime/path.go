package runtime

import (
	"strconv"
	"strings"
)

const PathParentID = "^"

type Path struct {
	components       []PathComponent
	componentsString string
	isRelative       bool
}

func (s Path) IsRelative() bool {
	return s.isRelative
}

func (s Path) Length() int {
	return len(s.components)
}

func (s Path) Component(index int) (PathComponent, bool) {
	if index <= -1 || index > len(s.components)-1 {
		return PathComponent{}, false
	}
	return s.components[index], true
}

func (s Path) ContainsNamedComponent() bool {
	for _, comp := range s.components {
		if comp.IsIndex() == false {
			return true
		}
	}
	return false
}

func (s Path) FirstComponent() (PathComponent, bool) {
	if len(s.components) > 0 {
		return s.components[0], true
	} else {
		return PathComponent{}, false
	}
}

func (s Path) LastComponent() (PathComponent, bool) {
	lastComponentIdx := len(s.components) - 1
	if lastComponentIdx >= 0 {
		return s.components[lastComponentIdx], true
	} else {
		return PathComponent{}, false
	}
}

func (s Path) Tail() Path {
	if len(s.components) >= 2 {
		tailComponents := s.components[1:]
		tailComponentsCopy := make([]PathComponent, len(tailComponents))
		copy(tailComponentsCopy, tailComponents)
		return NewPathFromComponents(tailComponentsCopy)
	} else {
		path := NewPath()
		path.isRelative = true
		return path
	}
}

func (s Path) NewPathByAppendingPath(pathToAppend Path) Path {

	var components []PathComponent

	upwardMoves := 0
	for i := 0; i < len(pathToAppend.components); i++ {
		if pathToAppend.components[i].IsParent() {
			upwardMoves++
		} else {
			break
		}
	}

	for i := 0; i < len(s.components)-upwardMoves; i++ {
		components = append(components, s.components[i])
	}

	for i := upwardMoves; i < len(pathToAppend.components); i++ {
		components = append(components, pathToAppend.components[i])
	}

	if s.IsRelative() {
		return NewRelativePathFromComponents(components)
	}

	return NewPathFromComponents(components)
}

func (s Path) NewPathByAppendingComponent(c PathComponent) Path {

	var x []PathComponent
	x = append(x, s.components...)
	x = append(x, c)

	if s.IsRelative() {
		return NewRelativePathFromComponents(x)
	}

	return NewPathFromComponents(x)
}

func (s Path) String() string {
	return s.componentsString
}

func (s Path) Equals(otherPath Path) bool {

	if len(otherPath.components) != len(s.components) {
		return false
	}

	if otherPath.IsRelative() != s.IsRelative() {
		return false
	}

	for i := 0; i < len(otherPath.components); i++ {
		if otherPath.components[i] != s.components[i] {
			return false
		}
	}

	return true
}

func joinComponents(components []PathComponent, sep string) string {

	sb := strings.Builder{}
	isFirst := true

	for _, c := range components {
		if isFirst == false {
			sb.WriteString(sep)
		}
		sb.WriteString(c.String())
		isFirst = false
	}

	return sb.String()
}

func NewPath() Path {
	return Path{}
}

func NewRelativePath() Path {
	return Path{isRelative: true, componentsString: "."}
}

func NewPathFromComponentsString(componentsString string) Path {
	p := Path{}

	p.componentsString = componentsString
	substring := componentsString

	// Empty path, empty components
	// (path is to root, like "/" in file system)
	if len(p.componentsString) == 0 {
		return p
	}

	// When components start with ".", it indicates a relative path, e.g.
	//   .^.^.hello.5
	// is equivalent to file system style path:
	//  ../../hello/5
	if p.componentsString[0] == '.' {
		p.isRelative = true
		//s.componentsString = s.componentsString[1:]
		substring = p.componentsString[1:]
	} else {
		p.isRelative = false
	}

	//componentStrings := strings.Split(s.componentsString, ".")
	componentStrings := strings.Split(substring, ".")
	for _, str := range componentStrings {
		index, err := strconv.Atoi(str)
		if err == nil {
			newComponent, err := NewPathComponentFromIndex(index)
			if err == nil {
				p.components = append(p.components, newComponent)
			}
		} else {
			newComponent, err := NewPathComponentFromName(str)
			if err == nil {
				p.components = append(p.components, newComponent)
			}
		}
	}

	return p
}

func NewPathFromComponents(components []PathComponent) Path {
	componentsString := joinComponents(components, ".")
	return Path{components: components, componentsString: componentsString}
}

func NewRelativePathFromComponents(components []PathComponent) Path {
	componentsString := joinComponents(components, ".")
	componentsString = "." + componentsString
	return Path{components: components, isRelative: true, componentsString: componentsString}
}
