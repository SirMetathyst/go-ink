package runtime

import (
	"fmt"
	"strings"
)

type Container struct {
	ObjectImpl

	// Public
	NamedContent             map[string]NamedContent
	VisitsShouldBeCounted    bool
	TurnIndexShouldBeCounted bool
	CountingAtStartOnly      bool

	// Private
	name                   string
	content                []Object
	namedOnlyContent       map[string]Object
	countFlags             CountFlags
	pathToFirstLeafContent *Path
}

func (s *Container) Content() []Object {
	return s.content
}

func (s *Container) SetContent(value []Object) {
	s.AddContent(value...)
}

func (s *Container) NamedOnlyContent() map[string]Object {

	namedOnlyContentDict := make(map[string]Object, 0)
	for key, value := range s.NamedContent {
		namedOnlyContentDict[key] = value.(Object)
	}

	for _, c := range s.content {
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

	existingNamedOnly := s.namedOnlyContent
	if existingNamedOnly != nil {
		for key, _ := range existingNamedOnly {
			delete(s.NamedContent, key)
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

type CountFlags uint8

const (
	Visits CountFlags = 1 << iota
	Turns
	CountStartOnly
)

func (s *Container) CountFlags() int {

	var flags CountFlags
	flags = 0

	if s.VisitsShouldBeCounted {
		flags |= Visits
	}

	if s.TurnIndexShouldBeCounted {
		flags |= Turns
	}

	if s.CountingAtStartOnly {
		flags |= CountStartOnly
	}

	// If we're only storing CountStartOnly, it serves no purpose,
	// since it's dependent on the other two to be used at all.
	// (e.g. for setting the fact that *if* a gather or choice's
	// content is counted, then is should only be counter at the start)
	// So this is just an optimisation for storage.
	if flags == CountStartOnly {
		flags = 0
	}

	return int(flags)
}

func (s *Container) SetCountFlags(value int) {

	flag := CountFlags(value)

	if (flag & Visits) > 0 {
		s.VisitsShouldBeCounted = true
	}

	if (flag & Turns) > 0 {
		s.TurnIndexShouldBeCounted = true
	}

	if (flag & CountStartOnly) > 0 {
		s.CountingAtStartOnly = true
	}
}

func (s *Container) HasValidName() bool {
	return len(s.name) > 0
}

func (s *Container) Name() string {
	return s.name
}

func (s *Container) SetName(value string) {
	s.name = value
}

func (s *Container) PathToFirstLeafContent() *Path {

	if s.pathToFirstLeafContent == nil {
		s.pathToFirstLeafContent = s.Path().NewPathByAppendingPath(s.internalPathToFirstLeafContent())
	}

	return s.pathToFirstLeafContent
}

func (s *Container) internalPathToFirstLeafContent() *Path {

	var components []*PathComponent
	container := s

	for container != nil {
		if len(container.content) > 0 {
			components = append(components, NewPathComponentFromIndex(0))
			container, _ = container.content[0].(*Container)
		}
	}

	return NewPathFromComponents(components, false)
}

func NewContainer() *Container {

	newContainer := new(Container)
	newContainer.this = newContainer
	newContainer.NamedContent = make(map[string]NamedContent, 0)

	return newContainer
}

func (s *Container) AddContent(contentList ...Object) {

	for _, contentObj := range contentList {

		s.content = append(s.content, contentObj)

		if contentObj.Parent() != nil {
			panic(fmt.Sprintf("content is already in %v", contentObj.Parent()))
		}

		contentObj.SetParent(s)

		s.TryAddNamedContent(contentObj)
	}
}

func (s *Container) InsertContent(contentObj Object, index int) {

	//content.Insert (index, contentObj)
	s.content = append(s.content[:index], append([]Object{contentObj}, s.content[index:]...)...)

	if contentObj.Parent() != nil {
		panic(fmt.Sprintf("content is already in %v", contentObj.Parent()))
	}

	contentObj.SetParent(s)

	s.TryAddNamedContent(contentObj)
}

func (s *Container) TryAddNamedContent(contentObj Object) {

	namedContentObj, _ := contentObj.(NamedContent)
	if namedContentObj != nil && namedContentObj.HasValidName() {
		s.AddToNamedContentOnly(namedContentObj)
	}
}

func (s *Container) AddToNamedContentOnly(namedContentObj NamedContent) {

	//Debug.Assert (namedContentObj is Runtime.Object, "Can only add Runtime.Objects to a Runtime.Container");
	if _, ok := namedContentObj.(Object); !ok {
		panic("Can only add Runtime.Objects to a Runtime.Container")
	}

	runtimeObj := namedContentObj.(Object)
	runtimeObj.SetParent(s)

	s.NamedContent[namedContentObj.Name()] = namedContentObj
}

func (s *Container) AddContentsOfContainer(otherContainer *Container) {

	s.content = append(s.content, otherContainer.content...)
	for _, obj := range otherContainer.content {
		obj.SetParent(s)
		s.TryAddNamedContent(obj)
	}
}

func (s *Container) ContentWithPathComponent(component *PathComponent) Object {

	if component.IsIndex() {

		if component.Index() >= 0 && component.Index() < len(s.Content()) {
			return s.Content()[component.Index()]
		} else {
			// When path is out of range, quietly return nil
			// (useful as we step/increment forwards through content)
			return nil
		}

	} else if component.IsParent() {
		return s.Parent()
	} else {
		if foundContent, ok := s.NamedContent[component.Name()]; ok {
			return foundContent.(Object)
		} else {
			return nil
		}
	}
}

// ContentAtPath
// partialPathStart (default: 0)
// partialPathLength (default: -1)
func (s *Container) ContentAtPath(path *Path, partialPathStart int, partialPathLength int) *SearchResult {

	if partialPathLength == -1 {
		partialPathLength = path.Length()
	}

	var result = NewSearchResult()
	result.Approximate = false

	currentContainer := s
	var currentObj Object
	currentObj = s

	for i := partialPathStart; i < partialPathLength; i++ {

		comp := path.Component(i)

		if currentContainer == nil {
			result.Approximate = true
			break
		}

		foundObj := currentContainer.ContentWithPathComponent(comp)
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

func (s *Container) BuildStringOfHierarchy(sb *strings.Builder, indentation int, pointedObj Object) {

	appendIndentation := func() {
		spacesPerIndent := 4
		for i := 0; i < spacesPerIndent*indentation; i++ {
			sb.WriteString(" ")
		}
	}

	appendIndentation()
	sb.WriteString("[")

	if s.HasValidName() {
		sb.WriteString(fmt.Sprintf(" (%s)", s.Name()))
	}

	if s == pointedObj {
		sb.WriteString("  <---")
	}

	sb.WriteString("\n")

	indentation++

	for i := 0; i < len(s.Content()); i++ {

		obj := s.Content()[0]

		if container, ok := obj.(*Container); ok {
			container.BuildStringOfHierarchy(sb, indentation, pointedObj)
		} else {
			appendIndentation()
			if strObj, ok := obj.(*StringValue); ok {
				sb.WriteString("\"")
				sb.WriteString(strings.ReplaceAll(strObj.String(), "\n", "\\n"))
				sb.WriteString("\"")
			} else {
				sb.WriteString(strObj.String())
			}
		}

		if i != len(s.Content())-1 {
			sb.WriteString(",")
		}

		if _, ok := obj.(*Container); ok && obj == pointedObj {
			sb.WriteString("  <---")
		}

		sb.WriteString("\n")
	}

	onlyNamed := make(map[string]NamedContent, 0)
	contains := false

	for key, value := range s.NamedContent {

		contains = false
		for _, v := range s.content {
			if v == value.(Object) {
				contains = true
			}
		}

		if contains {
			continue
		} else {
			onlyNamed[key] = value
		}
	}

	if len(onlyNamed) > 0 {
		appendIndentation()
		sb.WriteString("-- named: --\n")

		for _, value := range onlyNamed {

			//Debug.Assert (objKV.Value is Container, "Can only print out named Containers");
			if _, ok := value.(*Container); !ok {
				panic("Can only print out named Containers")
			}

			container := value.(*Container)
			container.BuildStringOfHierarchy(sb, indentation, pointedObj)
			sb.WriteString("\n")
		}
	}

	indentation--

	appendIndentation()
	sb.WriteString("]")
}

func (s *Container) BuildStringOfHierarchyEx() string {

	sb := new(strings.Builder)
	s.BuildStringOfHierarchy(sb, 0, nil)

	return sb.String()
}
