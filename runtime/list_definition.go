package runtime

type ListDefinition struct {

	// Private
	_name             string
	_items            map[InkListItem]int
	_itemNameToValues map[string]int
}

func (s *ListDefinition) Name() string {
	return s._name
}

func (s *ListDefinition) Items() map[InkListItem]int {

	if s._items != nil {
		return s._items
	}

	s._items = make(map[InkListItem]int)
	for key, value := range s._itemNameToValues {

		item := NewInkListItem(s.Name(), key)
		s._items[item] = value
	}

	return s._items
}

func (s *ListDefinition) TryGetItemWithValue(val int) (InkListItem, bool) {

	for key, value := range s._itemNameToValues {

		if value == val {
			item := NewInkListItem(s.Name(), key)
			return item, true
		}
	}

	return InkListItem{}, false
}

func NewListDefinition(name string, items map[string]int) *ListDefinition {

	newListDefinition := new(ListDefinition)
	newListDefinition._name = name
	newListDefinition._itemNameToValues = items

	return newListDefinition
}
