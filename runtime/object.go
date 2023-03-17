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
	DebugLineNumberOfPath(path *Path) (int, bool)
	Path() *Path
	ResolvePath(path *Path) *SearchResult
	ConvertPathToRelative(globalPath *Path) *Path
	CompactPathString(otherPath *Path) string
	RootContentContainer() *Container
	Copy() Object
}

// ObjectImpl base behaviour for all ink runtime content.
type ObjectImpl struct {
	parent        Object
	debugMetadata *DebugMetadata
	path          *Path
	this          Object
}

// Parent
// Runtime.Objects can be included in the main Story as a hierarchy.
// Usually parents are Container objects. (TODO: Always?)
func (s *ObjectImpl) Parent() Object {
	return s.parent
}

// SetParent
// Runtime.Objects can be included in the main Story as a hierarchy.
// Usually parents are Container objects. (TODO: Always?)
func (s *ObjectImpl) SetParent(parent Object) {
	s.parent = parent
}

// DebugMetadata
// TODO: Come up with some clever solution for not having
// to have debug metadata on the object itself, perhaps
// for serialisation purposes at least.
func (s *ObjectImpl) DebugMetadata() *DebugMetadata {
	if s.debugMetadata == nil {
		if s.this.Parent() != nil {
			return s.this.Parent().DebugMetadata()
		}
	}

	return s.debugMetadata
}

func (s *ObjectImpl) SetDebugMetadata(debugMetadata *DebugMetadata) {
	s.debugMetadata = debugMetadata
}

func (s *ObjectImpl) DebugLineNumberOfPath(path *Path) (int, bool) {

	if path == nil {
		return -1, false
	}

	root := s.this.RootContentContainer()
	if root != nil {
		var targetContent Object
		targetContent = root.ContentAtPath(path, 0, -1).Obj
		if targetContent != nil {
			dm := targetContent.DebugMetadata()
			if dm != nil {
				return dm.StartLineNumber, true
			}
		}
	}

	return -1, false
}

func (s *ObjectImpl) Path() *Path {

	if s.path == nil {

		if s.Parent() == nil {
			s.path = NewPath()
		} else {
			// Maintain a Stack so that the order of the components
			// is reversed when they're added to the Path.
			// We're iterating up the hierarchy from the leaves/children to the root.
			comps := NewStack[*PathComponent]()
			var child Object
			child = s.this
			container, okContainer := child.Parent().(*Container)

			for okContainer {
				namedChild, okNamedChild := child.(NamedContent)
				if okNamedChild && namedChild.HasValidName() {
					comps.Push(NewPathComponentFromName(namedChild.Name()))
				} else {

					index := -1
					for contentIndex, contentValue := range container.Content() {
						if contentValue == child {
							index = contentIndex
							break
						}
					}

					comps.Push(NewPathComponentFromIndex(index))
				}
				var containerObject Object
				containerObject = container
				child = containerObject
				container, okContainer = container.Parent().(*Container)
			}
		}
	}

	return s.path
}

func (s *ObjectImpl) ResolvePath(path *Path) *SearchResult {

	if path.IsRelative() {

		var nearestContainer, _ = s.this.(*Container)
		if nearestContainer != nil {

			//Debug.Assert(this.parent != null, "Can't resolve relative path because we don't have a parent")
			if s.this.Parent() == nil {
				panic("Can't resolve relative path because we don't have a parent")
			}

			nearestContainer, _ = s.this.Parent().(*Container)

			//Debug.Assert(nearestContainer != null, "Expected parent to be a container")
			if nearestContainer == nil {
				panic("Expected parent to be a container")
			}

			//Debug.Assert(path.GetComponent(0).isParent)
			if !path.Component(0).IsParent() {
				panic("not parent")
			}

			path = path.Tail()
		}

		return nearestContainer.ContentAtPath(path, 0, -1)
	} else {
		return s.this.RootContentContainer().ContentAtPath(path, 0, -1)
	}
}

func (s *ObjectImpl) ConvertPathToRelative(globalPath *Path) *Path {

	// 1. Find last shared ancestor
	// 2. Drill up using ".." style (actually represented as "^")
	// 3. Re-build downward chain from common ancestor

	ownPath := s.this.Path()

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

func (s *ObjectImpl) CompactPathString(otherPath *Path) string {

	globalPathStr := ""
	relativePathStr := ""

	if otherPath.IsRelative() {
		relativePathStr = otherPath.ComponentsString()
		globalPathStr = s.this.Path().NewPathByAppendingPath(otherPath).ComponentsString()
	} else {
		relativePath := s.this.ConvertPathToRelative(otherPath)
		relativePathStr = relativePath.componentsString
		globalPathStr = otherPath.componentsString
	}

	if len(relativePathStr) < len(globalPathStr) {
		return relativePathStr
	} else {
		return globalPathStr
	}
}

func (s *ObjectImpl) RootContentContainer() *Container {

	var ancestor Object
	ancestor = s.this
	for ancestor.Parent() != nil {
		ancestor = ancestor.Parent()
	}

	if ancestorContainer, ok := ancestor.(*Container); ok {
		return ancestorContainer
	}

	return nil
}

func (s *ObjectImpl) Copy() Object {

	panic(fmt.Sprintf("%v doesn't support copying", s))
}
