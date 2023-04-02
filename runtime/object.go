package runtime

import (
	"fmt"
	"math"
)

type Object interface {
	Parent() Object
	SetParent(parent Object)
	DebugMetadata() *DebugMetadata
	SetDebugMetadata(debugMetadata *DebugMetadata)
	OwnDebugMetadata() *DebugMetadata
	Path(child Object) *Path
	Copy() Object
}

// ObjectImpl base behaviour for all ink runtime content.
type ObjectImpl struct {

	// Private
	_parent        Object
	_debugMetadata *DebugMetadata
	_path          *Path
}

// Parent
// Runtime.Objects can be included in the main Story as a hierarchy.
// Usually parents are Container objects. (TODO: Always?)
func (s *ObjectImpl) Parent() Object {
	return s._parent
}

// SetParent
// Runtime.Objects can be included in the main Story as a hierarchy.
// Usually parents are Container objects. (TODO: Always?)
func (s *ObjectImpl) SetParent(parent Object) {
	s._parent = parent
}

// DebugMetadata
// TODO: Come up with some clever solution for not having
// to have debug metadata on the object itself, perhaps
// for serialisation purposes at least.
func (s *ObjectImpl) DebugMetadata() *DebugMetadata {

	if s._debugMetadata == nil {

		if s.Parent() != nil {

			return s.Parent().DebugMetadata()
		}
	}

	return s._debugMetadata
}

func (s *ObjectImpl) SetDebugMetadata(value *DebugMetadata) {

	s._debugMetadata = value
}

func (s *ObjectImpl) OwnDebugMetadata() *DebugMetadata {

	return s._debugMetadata
}

func (s *ObjectImpl) Path(child Object) *Path {

	if s._path == nil {

		if s.Parent() == nil {

			s._path = NewPath()
		} else {

			// Maintain a Stack so that the order of the components
			// is reversed when they're added to the Path.
			// We're iterating up the hierarchy from the leaves/children to the root.
			comps := NewStack[*PathComponent]()

			container, _ := child.Parent().(*Container)

			for container != nil {

				namedChild, _ := child.(NamedContent)
				if namedChild != nil && namedChild.HasValidName() {
					comps.Push(NewPathComponentFromName(namedChild.Name()))
				} else {
					comps.Push(NewPathComponentFromIndex(container.ContentIndexOf(child)))
				}

				child = container
				container, _ = container.Parent().(*Container)
			}
			s._path = NewPathFromComponents(comps.items, false)
		}
	}

	return s._path
}

func (s *ObjectImpl) Copy() Object {

	panic(fmt.Sprintf("%v doesn't support copying", s))
}

func ResolvePath(target Object, pathToResolve *Path) *SearchResult {

	if pathToResolve.IsRelative() {

		var nearestContainer, _ = target.(*Container)
		if nearestContainer == nil {

			nearestContainer, _ = target.Parent().(*Container)
			pathToResolve = pathToResolve.Tail()
		}

		return nearestContainer.ContentAtPath(pathToResolve, 0, -1)
	}

	return RootContentContainer(target).ContentAtPath(pathToResolve, 0, -1)
}

func RootContentContainer(target Object) (c *Container) {

	for target.Parent() != nil {
		target = target.Parent()
	}

	c, _ = target.(*Container)
	return
}

func (s *ObjectImpl) DebugLineNumberOfPath(target Object, path *Path) (int, bool) {

	if path == nil {

		return -1, false
	}

	root := RootContentContainer(target)
	if root != nil {

		targetContent := root.ContentAtPath(path, 0, -1).Obj
		if targetContent != nil {

			if dm := targetContent.DebugMetadata(); dm != nil {
				return dm.StartLineNumber, true
			}
		}
	}

	return -1, false
}

func CompactPathString(target Object, otherPath *Path) string {

	globalPathStr := ""
	relativePathStr := ""

	if otherPath.IsRelative() {

		relativePathStr = otherPath.ComponentsString()
		globalPathStr = target.Path(target).PathByAppendingPath(otherPath).ComponentsString()
	} else {

		relativePath := ConvertPathToRelative(target, otherPath)
		relativePathStr = relativePath.ComponentsString()
		globalPathStr = otherPath.ComponentsString()
	}

	if len(relativePathStr) < len(globalPathStr) {
		return relativePathStr
	}

	return globalPathStr
}

func ConvertPathToRelative(target Object, globalPath *Path) *Path {

	// 1. Find last shared ancestor
	// 2. Drill up using ".." style (actually represented as "^")
	// 3. Re-build downward chain from common ancestor

	ownPath := target.Path(target)

	minPathLength := int(math.Min(float64(globalPath.Length()), float64(ownPath.Length())))
	lastSharedPathCompIndex := -1

	for i := 0; i < minPathLength; i++ {

		ownComp := ownPath.Component(i)
		otherComp := globalPath.Component(i)

		if ownComp.Equals(otherComp) {
			lastSharedPathCompIndex = i
		} else {
			break
		}
	}

	// No shared path components, so just use global path
	if lastSharedPathCompIndex == -1 {
		return globalPath
	}

	numUpwardsMoves := (ownPath.Length() - 1) - lastSharedPathCompIndex

	var newPathComps []*PathComponent

	for up := 0; up < numUpwardsMoves; up++ {
		newPathComps = append(newPathComps, PathComponentToParent())
	}

	for down := lastSharedPathCompIndex + 1; down < globalPath.Length(); down++ {
		newPathComps = append(newPathComps, globalPath.Component(down))
	}
	relativePath := NewPathFromComponents(newPathComps, true)
	return relativePath
}
