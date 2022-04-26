package runtime

type ListDefinitionsOrigin struct {
	lists                        map[string]*ListDefinition
	allUnambiguousListValueCache map[string]*ListValue
}

func (s *ListDefinitionsOrigin) Lists() []*ListDefinition {

	var listOfLists []*ListDefinition
	for _, namedListValue := range s.lists {
		listOfLists = append(listOfLists, namedListValue)
	}
	return listOfLists
}

func (s *ListDefinitionsOrigin) TryListGetDefinition(name string) (def *ListDefinition, ok bool) {
	def, ok = s.lists[name]
	return
}

func (s *ListDefinitionsOrigin) FindSingleItemListWithName(name string) *ListValue {
	listValue, _ := s.allUnambiguousListValueCache[name]
	return listValue
}

func NewListDefinitionsOrigin(lists []*ListDefinition) *ListDefinitionsOrigin {

	s := &ListDefinitionsOrigin{}
	s.lists = make(map[string]*ListDefinition, 0)
	s.allUnambiguousListValueCache = make(map[string]*ListValue)

	for _, list := range lists {

		s.lists[list.Name()] = list

		for inkListItem, value := range list.Items() {
			listValue := NewListValueFromSingleInkListItem(inkListItem, value)
			s.allUnambiguousListValueCache[inkListItem.ItemName()] = listValue
			s.allUnambiguousListValueCache[inkListItem.FullName()] = listValue
		}
	}

	return s
}
