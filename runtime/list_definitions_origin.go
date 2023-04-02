package runtime

type ListDefinitionsOrigin struct {

	// Private
	_lists                        map[string]*ListDefinition
	_allUnambiguousListValueCache map[string]*ListValue
}

func (s *ListDefinitionsOrigin) Lists() []*ListDefinition {

	listOfLists := []*ListDefinition{}
	for _, value := range s._lists {

		listOfLists = append(listOfLists, value)
	}

	return listOfLists
}

func NewListDefinitionsOrigin(lists []*ListDefinition) *ListDefinitionsOrigin {

	newListDefinitionsOrigin := new(ListDefinitionsOrigin)
	newListDefinitionsOrigin._lists = make(map[string]*ListDefinition)
	newListDefinitionsOrigin._allUnambiguousListValueCache = make(map[string]*ListValue)

	for _, list := range lists {

		newListDefinitionsOrigin._lists[list.Name()] = list

		for item, val := range list.Items() {

			listValue := NewListValueFromInkListItem(item, val)

			// May be ambiguous, but compiler should've caught that,
			// so we may be doing some replacement here, but that's okay.
			newListDefinitionsOrigin._allUnambiguousListValueCache[item.ItemName()] = listValue
			newListDefinitionsOrigin._allUnambiguousListValueCache[item.Fullname()] = listValue
		}
	}

	return newListDefinitionsOrigin
}

func (s *ListDefinitionsOrigin) TryListGetDefinition(name string) (*ListDefinition, bool) {

	v, ok := s._lists[name]
	return v, ok
}

func (s *ListDefinitionsOrigin) FindSingleItemListWithName(name string) (val *ListValue) {

	val, _ = s._allUnambiguousListValueCache[name]
	return
}
