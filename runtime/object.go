package runtime

import (
	"fmt"
	"log"
)

type Object interface {
	SetParent(p Object)
	Parent() Object
	SetDebugMetadata(d *DebugMetadata)
	DebugMetadata() *DebugMetadata
	OwnDebugMetadata() *DebugMetadata
	DebugLineNumberOfPath(p *Path) (int, bool)
	Path() *Path
	ResolvePath(p *Path) SearchResult
	ConvertPathToRelative(globalPath *Path) *Path
	CompactPathString(otherPath *Path) string
	RootContentContainer() *Container
}

var _ Object = (*objectImpl)(nil)

type objectImpl struct {
	instance             Object
	debugMetadata        *DebugMetadata
	path                 *Path
	parent               Object
	rootContentContainer *Container
}

func (s *objectImpl) SetParent(p Object) {
	s.parent = p
}

func (s *objectImpl) Parent() Object {
	return s.parent
}

func (s *objectImpl) SetDebugMetadata(d *DebugMetadata) {
	s.debugMetadata = d
}

func (s *objectImpl) DebugMetadata() *DebugMetadata {
	if s.debugMetadata == nil {
		if s.parent != nil {
			return s.parent.DebugMetadata()
		}
	}
	return s.debugMetadata
}

func (s *objectImpl) OwnDebugMetadata() *DebugMetadata {
	return s.debugMetadata
}

func (s *objectImpl) Path() *Path {

	if s.path == nil {

		if s.parent == nil {
			s.path = NewPath()
		} else {
			var comps []PathComponent
			child := s.instance
			container, isContainer := child.Parent().(*Container)

			for isContainer {

				namedChild, isNamedContent := child.(NamedContent)
				if isNamedContent && namedChild.HasValidName() {
					component, err := NewPathComponentFromName(namedChild.Name())
					if err != nil {
						panic(err)
					}
					comps = append([]PathComponent{component}, comps...)
				} else {
					index := -1
					for i, elm := range container.Content() {
						if elm == child {
							index = i
						}
					}
					component, err := NewPathComponentFromIndex(index)
					if err != nil {
						panic(err)
					}
					comps = append([]PathComponent{component}, comps...)
				}

				child = container
				_, isContainer = container.Parent().(*Container)
			}

			s.path = NewPathFromComponents(comps)
		}
	}

	return s.path
}

func (s *objectImpl) RootContentContainer() *Container {
	ancestor := s.instance
	for ancestor.Parent() != nil {
		ancestor = ancestor.Parent()
	}
	ancestorContainer := ancestor.(*Container)
	return ancestorContainer
}

func (s *objectImpl) DebugLineNumberOfPath(p *Path) (int, bool) {

	if s.path == nil {
		return -1, false
	}

	root := s.rootContentContainer
	if root != nil {
		targetContent := root.ContentAtPath(p).Obj
		if targetContent != nil {
			dm := targetContent.DebugMetadata()
			if dm != nil {
				return dm.StartLineNumber, true
			}
		}
	}

	return -1, false
}

func (s *objectImpl) ResolvePath(p *Path) SearchResult {

	if p.IsRelative() {
		nearestContainer, isContainer := s.instance.(*Container)
		if !isContainer {
			if s.Parent() == nil {
				log.Println("Can't resolve relative path because we don't have a parent")
			}
			nearestContainer, isContainer = s.Parent().(*Container)
			if !isContainer {
				log.Println("Expected parent to be a container")
				component, ok := p.Component(0)
				if ok {
					log.Println(component.IsParent())
				}
			}
			p = p.Tail()
		}

		return nearestContainer.ContentAtPath(p)
	}

	return s.rootContentContainer.ContentAtPath(p)
}

func (s *objectImpl) ConvertPathToRelative(globalPath *Path) *Path {

	ownPath := s.path
	minPathLength := min(globalPath.Length(), ownPath.Length())
	lastSharedPathCompIndex := -1

	for i := 0; i < minPathLength; i++ {
		ownComp, _ := ownPath.Component(i)
		otherComp, _ := globalPath.Component(i)

		if ownComp == otherComp {
			lastSharedPathCompIndex = i
		} else {
			break
		}
	}

	if lastSharedPathCompIndex == -1 {
		return globalPath
	}

	numUpwardsMoves := (ownPath.Length() - 1) - lastSharedPathCompIndex
	var newPathComps []PathComponent

	for up := 0; up < numUpwardsMoves; up++ {
		newPathComps = append(newPathComps, ParentPathComponent)
	}

	for down := lastSharedPathCompIndex + 1; down < globalPath.Length(); down++ {
		comp, ok := globalPath.Component(down)
		if !ok {
			panic(fmt.Sprintf("component not found: %d of %s", down, globalPath.String()))
		}
		newPathComps = append(newPathComps, comp)
	}

	relativePath := NewRelativePathFromComponents(newPathComps)
	return relativePath
}

func (s *objectImpl) CompactPathString(otherPath *Path) string {

	globalPathStr := ""
	relativePathStr := ""

	if otherPath.IsRelative() {
		relativePathStr = otherPath.String()
		globalPathStr = s.path.NewPathByAppendingPath(otherPath).String()
	} else {
		relativePath := s.ConvertPathToRelative(otherPath)
		relativePathStr = relativePath.String()
		globalPathStr = otherPath.String()
	}

	if len(relativePathStr) < len(globalPathStr) {
		return relativePathStr
	} else {
		return globalPathStr
	}
}
