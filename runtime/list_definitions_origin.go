package runtime

type ListDefinitionsOrigin struct {
	lists                        map[string]*ListDefinition
	allUnambiguousListValueCache map[string]*ListValue
}

func (s *ListDefinitionsOrigin) Lists() []*ListDefinition {

	var listOfLists []*ListDefinition
	for _, v := range s.lists {
		listOfLists = append(listOfLists, v)
	}

	return listOfLists
}

func NewListDefinitionsOrigin(lists []*ListDefinition) *ListDefinitionsOrigin {

	newListDefinitionsOrigin := new(ListDefinitionsOrigin)
	_lists := make(map[string]*ListDefinition, 0)
	allUnambiguousListValueCache := make(map[string]*ListValue)

	for _, list := range lists {

		_lists[list.Name()] = list

		for item, val := range list.Items() {
			listValue := NewListValueFromItem(item, val)

			// May be ambiguous, but compiler should've caught that,
			// so we may be doing some replacement here, but that's okay.
			allUnambiguousListValueCache[item.ItemName()] = listValue
			allUnambiguousListValueCache[item.Fullname()] = listValue
		}
	}

	return newListDefinitionsOrigin
}

func (s *ListDefinitionsOrigin) TryListGetDefinition(name string) (*ListDefinition, bool) {
	def, ok := s.lists[name]
	return def, ok
}

func (s *ListDefinitionsOrigin) FindSingleItemListName(name string) *ListValue {

	val, _ := s.allUnambiguousListValueCache[name]
	return val
}
