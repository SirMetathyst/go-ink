package runtime

import (
	"fmt"
)

type Container struct {
	ObjectImpl

	// Private
	_name         string
	_content      []Object
	_namedContent map[string]NamedContent

	// Public
	VisitsShouldBeCounted    bool
	TurnIndexShouldBeCounted bool
	CountingAtStartOnly      bool
}

func (s *Container) Name() string {
	return s._name
}

func (s *Container) SetName(value string) {
	s._name = value
}

func (s *Container) Content() []Object {
	return s._content
}

func (s *Container) ContentIndexOf(v Object) int {

	for index, vv := range s._content {
		if vv == v {
			return index
		}
	}

	return -1
}

func (s *Container) NamedContent() map[string]NamedContent {
	return s._namedContent
}

func (s *Container) NamedOnlyContent() map[string]Object {

	namedOnlyContentDict := make(map[string]Object, 0)

	for key, value := range s.NamedContent() {

		namedOnlyContentDict[key] = value.(Object)
	}

	for _, c := range s.Content() {

		var named, _ = c.(NamedContent)
		if named != nil && named.HasValidName() {
			delete(namedOnlyContentDict, named.Name())
		}
	}

	if len(namedOnlyContentDict) == 0 {
		namedOnlyContentDict = nil
	}

	return namedOnlyContentDict
}

func (s *Container) SetNamedOnlyContent(value map[string]Object) {

	existingNamedOnly := s.NamedOnlyContent()
	if existingNamedOnly != nil {
		for key, _ := range existingNamedOnly {
			delete(s.NamedContent(), key)
		}
	}

	if value == nil {
		return
	}

	for _, vv := range value {
		named, _ := vv.(NamedContent)
		if named != nil {
			s.AddToNamedContentOnly(named)
		}
	}
}

func (s *Container) CountFlags() int {

	flags := 0

	if s.VisitsShouldBeCounted {
		flags |= 1
	}

	if s.TurnIndexShouldBeCounted {
		flags |= 2
	}

	if s.CountingAtStartOnly {
		flags |= 4
	}

	// If we're only storing CountStartOnly, it serves no purpose,
	// since it's dependent on the other two to be used at all.
	// (e.g. for setting the fact that *if* a gather or choice's
	// content is counted, then is should only be counter at the start)
	// So this is just an optimisation for storage.
	if flags == 4 {
		flags = 0
	}

	return flags
}

func (s *Container) SetCountFlags(value int) {

	if (value & 1) > 0 {
		s.VisitsShouldBeCounted = true
	}

	if (value & 2) > 0 {
		s.TurnIndexShouldBeCounted = true
	}

	if (value & 4) > 0 {
		s.CountingAtStartOnly = true
	}
}

func (s *Container) HasValidName() bool {

	return len(s.Name()) > 0
}

func NewContainer() *Container {

	newContainer := new(Container)
	newContainer._content = []Object{}
	newContainer._namedContent = make(map[string]NamedContent)

	return newContainer
}

func (s *Container) AddContent(contentObj Object) {

	// Add
	s._content = append(s._content, contentObj)

	if contentObj.Parent() != nil {
		panic(fmt.Sprintf("content is already in %v", contentObj.Parent()))
	}

	contentObj.SetParent(s)

	s.TryAddNamedContent(contentObj)
}

func (s *Container) InsertContent(contentObj Object, index int) {

	// Insert
	s._content = append(s._content[:index], append([]Object{contentObj}, s._content[index:]...)...)

	if contentObj.Parent() != nil {
		panic(fmt.Sprintf("content is already in %v", contentObj.Parent()))
	}

	contentObj.SetParent(s)

	s.TryAddNamedContent(contentObj)
}

func (s *Container) AddToNamedContentOnly(namedContentObj NamedContent) {

	//Debug.Assert (namedContentObj is Runtime.Object, "Can only add Runtime.Objects to a Runtime.Container");

	runtimeObj := namedContentObj.(Object)
	runtimeObj.SetParent(s)

	s._namedContent[namedContentObj.Name()] = namedContentObj
}

func (s *Container) AddContentsOfContainer(otherContainer *Container) {

	// AddRange
	s._content = append(s._content, otherContainer.Content()...)

	for _, obj := range otherContainer.Content() {

		obj.SetParent(s)
		s.TryAddNamedContent(obj)
	}
}

func (s *Container) ContentAtPath(path *Path, partialPathStart int, partialPathLength int) SearchResult {

	if partialPathLength == -1 {
		partialPathLength = path.Length()
	}

	var result = NewSearchResult()
	result.Approximate = false

	currentContainer := s
	var currentObj Object = s

	for i := partialPathStart; i < partialPathLength; i++ {

		comp := path.Component(i)

		// Path component was wrong type
		if currentContainer == nil {
			result.Approximate = true
			break
		}

		foundObj := currentContainer.ContentWithPathComponent(comp)

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

func (s *Container) TryAddNamedContent(contentObj Object) {

	namedContentObj, _ := contentObj.(NamedContent)

	if namedContentObj != nil && namedContentObj.HasValidName() {

		s.AddToNamedContentOnly(namedContentObj)
	}
}

func (s *Container) ContentWithPathComponent(component *PathComponent) Object {

	if component.IsIndex() {

		if component.Index() >= 0 && component.Index() < len(s.Content()) {
			return s._content[component.Index()]
		}

		// When path is out of range, quietly return nil
		// (useful as we step/increment forwards through content)
		return nil
	}

	if component.IsParent() {
		return s.Parent()
	}

	if foundContent, ok := s.NamedContent()[component.Name()]; ok {
		return foundContent.(Object)
	}

	return nil
}
