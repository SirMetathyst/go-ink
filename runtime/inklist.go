package runtime

type KeyValuePair[K, V] struct {
	Key   K
	Value V
}

type InkList struct {
	inkListMap  map[InkListItem]int
	originNames []string
	origins     []*ListDefinition
}

func (s *InkList) Length() int {
	return len(s.inkListMap)
}

func NewInkListFromOriginal(otherList *InkList) *InkList {
	s := NewInkList()
	s.originNames = otherList.originNames

	for _, origin := range otherList.origins {
		s.origins = append(s.origins, origin)
	}

	return s
}

func NewInkListFromKeyValuePair(singleElement KeyValuePair[InkListItem, int]) *InkList {
	s := NewInkList()
	s.inkListMap[singleElement.Key] = singleElement.Value
	return s
}

func NewInkListFromSingleOriginListName(singleOriginListName string, originStory *Story) *InkList {
	s := NewInkList()
	s.SetInitialOriginName(singleOriginListName)

	if def, ok := originStory.ListDefinitions().TryListGetDefinition(singleOriginListName); ok {
		s.origins = append(s.origins, def)
	} else {
		panic("InkList origin could not be found in story when constructing new list: " + singleOriginListName)
	}

	return s
}

func NewInkList() *InkList {
	return &InkList{inkListMap: map[InkListItem]int{}}
}
