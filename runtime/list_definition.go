package runtime

type ListDefinition struct {

	// Private
	name             string
	items            map[*InkListItem]int
	itemNameToValues map[string]int
}

func (s *ListDefinition) Name() string {
	return s.name
}

func (s *ListDefinition) Items() map[*InkListItem]int {

	if s.items == nil {
		s.items = make(map[*InkListItem]int, 0)
		for k, v := range s.itemNameToValues {
			item := NewInkListItem(s.name, k)
			s.items[item] = v
		}
	}

	return s.items
}

func (s *ListDefinition) ValueForItem(item *InkListItem) int {

	if intVal, ok := s.itemNameToValues[item.ItemName()]; ok {
		return intVal
	} else {
		return 0
	}
}

func (s *ListDefinition) ContainsItem(item *InkListItem) bool {

	if item.OriginName() != s.Name() {
		return false
	}

	_, ok := s.itemNameToValues[item.ItemName()]
	return ok
}

func (s *ListDefinition) ContainsItemWithName(itemName string) bool {

	_, ok := s.itemNameToValues[itemName]
	return ok
}

func (s *ListDefinition) TryGetItemWithValue(val int) (*InkListItem, bool) {

	for k, v := range s.itemNameToValues {
		if v == val {
			item := NewInkListItem(s.name, k)
			return item, true
		}
	}

	return nil, false
}

func (s *ListDefinition) TryGetValueForItem(item *InkListItem) (int, bool) {

	v, ok := s.itemNameToValues[item.ItemName()]
	return v, ok
}

func NewListDefinition(name string, items map[string]int) *ListDefinition {

	newListDefintion := new(ListDefinition)
	newListDefintion.name = name
	newListDefintion.itemNameToValues = items

	return newListDefintion
}
