package runtime

import (
	"fmt"
	"reflect"
)

type CountFlag int

const (
	CountFlagVisits         CountFlag = 1
	CountFlagTurns          CountFlag = 2
	CountFlagCountStartOnly CountFlag = 4
)

var _ NamedContent = (*Container)(nil)
var _ Object = (*Container)(nil)

type Container struct {
	*objectImpl
	name                     string
	content                  []Object
	namedContent             map[string]NamedContent
	visitsShouldBeCounted    bool
	turnIndexShouldBeCounted bool
	countingAtStartOnly      bool
}

func (s *Container) VisitsShouldBeCounted() bool {
	return s.visitsShouldBeCounted
}

func (s *Container) SetVisitsShouldBeCounted(state bool) {
	s.visitsShouldBeCounted = state
}

func (s *Container) TurnIndexShouldBeCounted() bool {
	return s.turnIndexShouldBeCounted
}

func (s *Container) SetTurnIndexShouldBeCounted(state bool) {
	s.turnIndexShouldBeCounted = state
}

func (s *Container) CountingAtStartOnly() bool {
	return s.countingAtStartOnly
}

func (s *Container) SetCountingAtStartOnly(state bool) {
	s.countingAtStartOnly = state
}

func (s *Container) CountFlags() int {

	var flags CountFlag = 0
	if s.visitsShouldBeCounted {
		flags |= CountFlagVisits
	}
	if s.turnIndexShouldBeCounted {
		flags |= CountFlagTurns
	}
	if s.countingAtStartOnly {
		flags |= CountFlagCountStartOnly
	}

	//// If we're only storing CountStartOnly, it serves no purpose,
	//// since it's dependent on the other two to be used at all.
	//// (e.g. for setting the fact that *if* a gather or choice's
	//// content is counted, then is should only be counter at the start)
	//// So this is just an optimisation for storage.
	if flags == CountFlagCountStartOnly {
		flags = 0
	}

	return int(flags)
}

func (s *Container) SetCountFlags(n int) {

	var flag = CountFlag(n)
	if (flag & CountFlagVisits) > 0 {
		s.visitsShouldBeCounted = true
	}
	if (flag & CountFlagTurns) > 0 {
		s.turnIndexShouldBeCounted = true
	}
	if (flag & CountFlagCountStartOnly) > 0 {
		s.countingAtStartOnly = true
	}
}

func (s *Container) Name() string {
	return s.name
}

func (s *Container) SetName(name string) {
	s.name = name
}

func (s *Container) HasValidName() bool {
	if len(s.name) > 0 {
		return true
	}
	return false
}

func (s *Container) Content() []Object {
	return s.content
}

func (s *Container) SetContent(content []Object) {
	s.AddContent(content...)
}

func (s *Container) AddContent(contentObjects ...Object) {

	for _, contentObj := range contentObjects {

		s.content = append(s.content, contentObj)

		if contentObj.Parent() != nil {
			panic(fmt.Sprintf("content is already in %v", reflect.TypeOf(contentObj.Parent()).String()))
		}

		contentObj.SetParent(s)
		s.TryAddNamedContent(contentObj)
	}
}

func (s *Container) TryAddNamedContent(contentObj Object) {

	var namedContentObj, ok = contentObj.(NamedContent)
	if ok && namedContentObj.HasValidName() {
		s.AddToNamedContentOnly(namedContentObj)
	}
}

func (s *Container) AddToNamedContentOnly(namedContentObj NamedContent) {

	runtimeObj := namedContentObj.(Object)
	runtimeObj.SetParent(s)
	s.namedContent[namedContentObj.Name()] = namedContentObj
}

func (s *Container) NamedContent() map[string]NamedContent {
	return s.namedContent
}

func (s *Container) SetNamedContent(namedContent map[string]NamedContent) {
	s.namedContent = namedContent
}

func (s *Container) NamedOnlyContent() map[string]Object {

	namedOnlyContentMap := map[string]Object{}

	for k, v := range s.namedContent {
		namedOnlyContentMap[k] = v.(Object)
	}

	for _, c := range s.content {
		if named, ok := c.(NamedContent); ok && named.HasValidName() {
			delete(namedOnlyContentMap, named.Name())
		}
	}

	if len(namedOnlyContentMap) == 0 {
		namedOnlyContentMap = nil
	}

	return namedOnlyContentMap
}

func (s *Container) SetNamedOnlyContent(namedOnlyContent map[string]Object) {

	var existingNamedOnly = s.NamedOnlyContent()
	if existingNamedOnly != nil {
		for k := range existingNamedOnly {
			delete(s.namedContent, k)
		}
	}

	if namedOnlyContent == nil {
		return
	}

	for _, v := range namedOnlyContent {
		if named, ok := v.(NamedContent); ok {
			s.AddToNamedContentOnly(named)
		}
	}
}

func (s *Container) InsertContent(contentObj Object, index int) {

	// Insert at index
	s.content = append(s.content[:index+1], s.content[index:]...)
	s.content[index] = contentObj

	if contentObj.Parent() != nil {
		// This threw an exception before (in the c# version)
		panic(fmt.Sprintf("content is already in %v", reflect.TypeOf(contentObj.Parent()).String()))
	}

	contentObj.SetParent(s)
	s.TryAddNamedContent(contentObj)
}

func (s *Container) AddContentsOfContainer(otherContainer *Container) {

	s.content = append(s.content, otherContainer.Content()...)
	for _, obj := range otherContainer.Content() {
		obj.SetParent(s)
		s.TryAddNamedContent(obj)
	}
}

func (s *Container) contentWithPathComponent(component PathComponent) Object {

	if component.IsIndex() {

		if component.Index() >= 0 && component.Index() < len(s.content) {
			return s.content[component.Index()]
		}

		// When path is out of range, quietly return nil
		// (useful as we step/increment forwards through content)
		return nil

	}

	if component.IsParent() {
		return s.Parent()
	}

	if foundContent, ok := s.namedContent[component.Name()]; ok {
		return foundContent.(Object)
	}

	return nil
}

func (s *Container) ContentAtPathWithPathStartPathLength(path *Path, partialPathStart int /*= 0*/, partialPathLength int /*= -1*/) SearchResult {

	if partialPathLength == -1 {
		partialPathLength = path.Length()
	}

	result := SearchResult{Approximate: false}

	var currentContainer = s
	var currentObj Object = s

	for i := partialPathStart + 1; i < partialPathLength; i++ {
		var comp, _ = path.Component(i)

		// Path component was wrong type
		if currentContainer == nil {
			result.Approximate = true
			break
		}

		var foundObj = currentContainer.contentWithPathComponent(comp)

		// Couldn't resolve entire path?
		if foundObj == nil {
			result.Approximate = true
			break
		}

		currentObj = foundObj
		currentContainer, _ = foundObj.(*Container)
	}

	result.Obj = currentObj

	return result
}

func (s *Container) ContentAtPathWithPathStart(path *Path, partialPathStart int /*= 0*/) SearchResult {
	return s.ContentAtPathWithPathStartPathLength(path, partialPathStart, -1)
}

func (s *Container) ContentAtPath(path *Path) SearchResult {
	return s.ContentAtPathWithPathStartPathLength(path, 0, -1)
}

func NewContainer() *Container {
	s := &Container{}
	s.objectImpl = &objectImpl{instance: s}
	return s
}

//func (s *Container) InternalPathToFirstLeafContent() *Path {
//	var components []*PathComponent
//	container := s
//	for container != nil {
//		if len(container.Content()) > 0 {
//			components = append(components, newPathComponentFromIndex(0))
//			con, ok := container.content[0].(*Container)
//			if !ok {
//				panic("should explicitly cast to container like the c# version")
//			}
//			container = con
//		}
//	}
//	return NewPathFromComponents(components)
//}

//func (s *Container) PathToFirstLeafContent() *Path {
//
//	if s.pathToFirstLeafContent == nil {
//		s.pathToFirstLeafContent = s.path.PathByAppendingPath(s.internalPathToFirstLeafContent)
//	}
//
//	return s.pathToFirstLeafContent
//}
