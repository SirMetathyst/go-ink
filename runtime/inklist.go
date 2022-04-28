package runtime

import (
	"math"
	"sort"
	"strings"
)

type KeyValuePair[Key any, Value any] struct {
	Key   Key
	Value Value
}

type InkList struct {
	inkListMap  map[InkListItem]int
	originNames []string
	origins     []*ListDefinition
}

func (s *InkList) Map() map[InkListItem]int {
	return s.inkListMap
}

func (s *InkList) Length() int {
	return len(s.inkListMap)
}

func (s *InkList) Remove(key InkListItem) bool {
	if _, ok := s.inkListMap[key]; !ok {
		return false
	}
	delete(s.inkListMap, key)
	return true
}

func (s InkList) Add(key InkListItem, value int) {
	if _, ok := s.inkListMap[key]; ok {
		panic("key already exists in ink list")
	}
	s.inkListMap[key] = value
}

func (s InkList) AddOrUpdate(key InkListItem, value int) {
	s.inkListMap[key] = value
}

func (s *InkList) AddItem(item InkListItem) {

	if len(item.OriginName()) != 0 {
		s.AddItemByName(item.ItemName())
		return
	}

	for _, origin := range s.origins {
		if origin.Name() == item.OriginName() {
			if value, ok := origin.TryGetValueForItem(item); ok {
				s.inkListMap[item] = value
				return
			}

			panic("Could not add the item " + item.String() + " to this list because it doesn't exist in the original list definition in ink.")
		}
	}

	panic("Failed to add item to list because the item was from a new list definition that wasn't previously known to this list. Only items from previously known lists can be used, so that the int value can be found.")
}

func (s *InkList) AddItemByName(itemName string) {

	var foundListDef *ListDefinition

	for _, origin := range s.origins {
		if origin.ContainsItemWithName(itemName) {
			if foundListDef != nil {
				panic("Could not add the item " + itemName + " to this list because it could come from either " + origin.name + " or " + foundListDef.name)
			}
			foundListDef = origin
		}
	}

	if foundListDef == nil {
		panic("Could not add the item " + itemName + " to this list because it isn't known to any list definitions previously associated with this list.")
	}

	item := NewInkListItem(foundListDef.Name(), itemName)
	itemVal := foundListDef.ValueForItem(item)
	s.inkListMap[item] = itemVal
}

func (s *InkList) ContainsInkListItemWithName(itemName string) bool {

	for itemKey, _ := range s.inkListMap {
		if itemKey.ItemName() == itemName {
			return true
		}
	}
	return false
}

func (s *InkList) GreaterThan(otherList *InkList) bool {

	if len(s.inkListMap) == 0 {
		return false
	}

	if otherList.Length() == 0 {
		return true
	}

	return s.MinItem().Value > otherList.MaxItem().Value
}

func (s *InkList) GreaterThanOrEquals(otherList *InkList) bool {

	if len(s.inkListMap) == 0 {
		return false
	}

	if otherList.Length() == 0 {
		return true
	}

	return s.MinItem().Value >= otherList.MinItem().Value &&
		s.MaxItem().Value >= otherList.MaxItem().Value
}

func (s InkList) LessThan(otherList *InkList) bool {

	if otherList.Length() == 0 {
		return false
	}

	if len(s.inkListMap) == 0 {
		return true
	}

	return s.MaxItem().Value < otherList.MinItem().Value
}

func (s *InkList) LessThanOrEquals(otherList *InkList) bool {

	if otherList.Length() == 0 {
		return false
	}

	if len(s.inkListMap) == 0 {
		return true
	}

	return s.MaxItem().Value <= otherList.MaxItem().Value &&
		s.MinItem().Value <= otherList.MinItem().Value
}

func (s *InkList) MaxAsList() *InkList {

	if len(s.inkListMap) > 0 {
		return NewInkListFromKeyValuePair(s.MaxItem())
	}

	return NewInkList()
}

func (s *InkList) MinAsList() *InkList {

	if len(s.inkListMap) > 0 {
		return NewInkListFromKeyValuePair(s.MinItem())
	}

	return NewInkList()
}

func (s *InkList) Equals(otherList *InkList) bool {

	if otherList == nil {
		return false
	}

	if len(s.inkListMap) != otherList.Length() {
		return false
	}

	for itemKey, _ := range s.inkListMap {
		if !otherList.ContainsInkListItem(itemKey) {
			return false
		}
	}

	return true
}

func (s *InkList) OrderedItems() []KeyValuePair[InkListItem, int] {
	var ordered []KeyValuePair[InkListItem, int]
	for itemKey, itemValue := range s.inkListMap {
		ordered = append(ordered, KeyValuePair[InkListItem, int]{Key: itemKey, Value: itemValue})
	}

	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Value == ordered[j].Value {
			return strings.Compare(strings.ToLower(ordered[i].Key.OriginName()), strings.ToLower(ordered[j].Key.OriginName())) == -1
		}
		return ordered[i].Value < ordered[j].Value
	})

	return ordered
}

func (s *InkList) ListWithSubRange(minBound interface{}, maxBound interface{}) *InkList {

	if len(s.inkListMap) == 0 {
		return NewInkList()
	}

	ordered := s.OrderedItems()
	minValue := 0
	maxValue := math.MaxInt

	if v, ok := minBound.(int); ok {
		minValue = v
	} else {
		if v, ok := minBound.(*InkList); ok && v.Length() > 0 {
			minValue = v.MinItem().Value
		}
	}

	if v, ok := maxBound.(int); ok {
		maxValue = v
	} else {
		if v, ok := maxBound.(*InkList); ok && v.Length() > 0 {
			maxValue = v.MaxItem().Value
		}
	}

	sublist := NewInkList()
	sublist.SetInitialOriginNames(s.originNames)
	for _, pair := range ordered {
		if pair.Value >= minValue && pair.Value <= maxValue {
			sublist.Add(pair.Key, pair.Value)
		}
	}

	return sublist
}

func (s *InkList) MaxItem() KeyValuePair[InkListItem, int] {

	max := KeyValuePair[InkListItem, int]{}

	for key, value := range s.inkListMap {
		if max.Key.IsNil() || value > max.Value {
			max.Key = key
			max.Value = value
		}
	}

	return max
}

func (s *InkList) MinItem() KeyValuePair[InkListItem, int] {

	min := KeyValuePair[InkListItem, int]{}

	for key, value := range s.inkListMap {
		if min.Key.IsNil() || value < min.Value {
			min.Key = key
			min.Value = value
		}
	}

	return min
}

func (s *InkList) ContainsInkListItem(key InkListItem) bool {
	_, ok := s.inkListMap[key]
	return ok
}

func (s *InkList) ContainsInkList(otherList *InkList) bool {

	if len(otherList.Map()) == 0 || len(s.inkListMap) == 0 {
		return false
	}

	for itemKey, _ := range otherList.Map() {
		if !s.ContainsInkListItem(itemKey) {
			return false
		}
	}
	return true
}

func (s *InkList) Inverse() *InkList {
	list := NewInkList()
	if s.origins != nil {
		for _, origin := range s.origins {
			for itemKey, itemValue := range origin.Items() {
				if s.ContainsInkListItem(itemKey) {
					list.Add(itemKey, itemValue)
				}
			}
		}
	}
	return list
}

func (s *InkList) All() *InkList {

	list := NewInkList()
	if s.origins != nil {
		for _, origin := range s.origins {
			for itemKey, itemValue := range origin.Items() {
				list.AddOrUpdate(itemKey, itemValue)
			}
		}
	}

	return list
}

func (s *InkList) Union(otherList *InkList) *InkList {

	union := NewInkListFromInkList(s)

	for listKey, listValue := range otherList.Map() {
		union.AddOrUpdate(listKey, listValue)
	}

	return union
}

func (s *InkList) Intersect(otherList *InkList) *InkList {

	intersection := NewInkList()

	for itemKey, itemValue := range s.inkListMap {
		if otherList.ContainsInkListItem(itemKey) {
			intersection.Add(itemKey, itemValue)
		}
	}

	return intersection
}

func (s *InkList) HasIntersection(otherList *InkList) bool {

	for itemKey, _ := range s.inkListMap {
		if otherList.ContainsInkListItem(itemKey) {
			return true
		}
	}

	return false
}

func (s *InkList) Without(listToRemove *InkList) *InkList {

	result := NewInkListFromInkList(s)

	for itemKey, _ := range listToRemove.Map() {
		result.Remove(itemKey)
	}

	return result
}

func (s *InkList) OriginOfMaxItem() *ListDefinition {

	if s.origins == nil {
		return nil
	}

	maxOrigin := s.MaxItem().Key
	maxOriginName := maxOrigin.OriginName()

	for _, origin := range s.origins {
		if origin.Name() == maxOriginName {
			return origin
		}
	}

	return nil
}

func (s *InkList) OriginNames() []string {
	if len(s.inkListMap) > 0 {
		s.originNames = s.originNames[:0]

		for key, _ := range s.inkListMap {
			s.originNames = append(s.originNames, key.OriginName())
		}
	}
	return s.originNames
}

func (s *InkList) SetInitialOriginName(initialOriginName string) {

	s.originNames = s.originNames[:0]
	s.originNames = append(s.originNames, initialOriginName)
}

func (s *InkList) SetInitialOriginNames(initialOriginNames []string) {

	if len(initialOriginNames) == 0 {
		s.originNames = s.originNames[:0]
	} else {
		s.originNames = s.originNames[:0]
		for _, name := range initialOriginNames {
			s.originNames = append(s.originNames, name)
		}
	}
}

func (s *InkList) String() string {

	ordered := s.OrderedItems()

	sb := strings.Builder{}
	for i := 0; i < len(ordered); i++ {
		if i > 0 {
			sb.WriteString(", ")
		}

		item := ordered[i].Key
		sb.WriteString(item.ItemName())
	}

	return sb.String()
}

func NewInkListFromString(listItem string, originStory *Story) *InkList {

	listValue := originStory.ListDefinitions().FindSingleItemListWithName(listItem)
	if listValue != nil {
		return NewInkListFromInkList(listValue.Value)
	}

	panic("Could not find the InkListItem from the string '" + listItem + "' to create an InkList because it doesn't exist in the original list definition in ink.")
}

func NewInkListFromInkList(otherList *InkList) *InkList {
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
