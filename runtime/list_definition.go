package runtime

type ListDefinition struct {
	name             string
	items            map[InkListItem]int
	itemNameToValues map[string]int
}

func (s *ListDefinition) Name() string {
	return s.name
}

func (s *ListDefinition) Items() map[InkListItem]int {

	if s.items == nil {
		s.items = map[InkListItem]int{}

		for key, value := range s.itemNameToValues {
			item := NewInkListItem(s.name, key)
			s.items[item] = value
		}
	}

	return s.items
}

func (s *ListDefinition) ValueForItem(item InkListItem) int {

	if value, ok := s.itemNameToValues[item.ItemName()]; ok {
		return value
	}

	return 0
}

func (s *ListDefinition) ContainsItem(item InkListItem) bool {

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

func (s *ListDefinition) TryGetItemWithValue(value int) (InkListItem, bool) {

	for itemKey, itemValue := range s.itemNameToValues {
		if itemValue == value {
			item := NewInkListItem(s.name, itemKey)
			return item, true
		}
	}

	return NewNilInkListItem(), false
}

func (s *ListDefinition) TryGetValueForItem(item InkListItem) (value int, ok bool) {
	value, ok = s.itemNameToValues[item.ItemName()]
	return
}

func NewListDefinition(name string, items map[string]int) *ListDefinition {
	return &ListDefinition{name: name, itemNameToValues: items}
}
