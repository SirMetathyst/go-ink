package runtime

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// InkListItem
// The underlying type for a list item in ink. It stores the original list definition
// name as well as the item name, but without the value of the item. When the value is
// stored, it's stored in a KeyValuePair of InkListItem and int.
type InkListItem struct {

	// Private
	originName string
	itemName   string
}

// OriginName
// The name of the list where the item was originally defined.
func (s *InkListItem) OriginName() string {
	return s.originName
}

// ItemName
// The main name of the item as defined in ink.
func (s *InkListItem) ItemName() string {
	return s.itemName
}

// NewInkListItem
// Create an item with the given original list definition name, and the name of this item.
func NewInkListItem(originName string, itemName string) *InkListItem {

	inkListItem := new(InkListItem)
	inkListItem.originName = originName
	inkListItem.itemName = itemName

	return inkListItem
}

// NewInkListFromFullname
// Create an item from a dot-separted string of the form "listDefinitionName.listItemName".
func NewInkListFromFullname(fullName string) *InkListItem {

	inkListItem := new(InkListItem)
	nameParts := strings.Split(fullName, ".")
	inkListItem.originName = nameParts[0]
	inkListItem.itemName = nameParts[1]

	return inkListItem
}

func NullInkListItem() *InkListItem {
	return NewInkListItem("", "")
}

func (s *InkListItem) IsNull() bool {

	return s.originName == "" && s.itemName == ""
}

// Fullname
// Get the full dot-separated name of the item, in the form "listDefinitionName.itemName".
func (s *InkListItem) Fullname() string {

	v := ""
	if s.originName == "" {
		v = "?"
	} else {
		v = s.originName
	}

	return v + "." + s.itemName
}

// String
// Get the full dot-separated name of the item, in the form "listDefinitionName.itemName".
// Calls fullName internally.
func (s *InkListItem) String() string {

	return s.Fullname()
}

func (s *InkListItem) Equals(otherInkListItem *InkListItem) bool {

	return otherInkListItem.ItemName() == s.ItemName() && otherInkListItem.OriginName() == s.OriginName()
}

func (s *InkListItem) HashCode() string {
	return fmt.Sprintf("%s-%s", s.originName, s.itemName)
}

type InkList struct {

	// Private
	originNames []string

	// Public
	Dict    map[*InkListItem]int
	Origins []*ListDefinition
}

func NewInkList() *InkList {

	newInkList := new(InkList)
	newInkList.Dict = make(map[*InkListItem]int)

	return newInkList
}

func (s *InkList) Length() int {
	return len(s.Dict)
}

// NewInkListFromList
// Create a new ink list that contains the same contents as another list.
func NewInkListFromList(otherList *InkList) *InkList {

	newInkList := new(InkList)
	newInkList.Dict = make(map[*InkListItem]int)

	for k, v := range otherList.Dict {
		newInkList.Dict[k] = v
	}

	otherOriginNames := otherList.OriginNames()
	if otherOriginNames != nil {
		newInkList.originNames = append([]string{}, otherList.originNames...)
	}

	if otherList.Origins != nil {
		newInkList.Origins = append([]*ListDefinition{}, otherList.Origins...)
	}

	return newInkList
}

// NewInkListFromOriginStory
// Create a new empty ink list that's intended to hold items from a particular origin
// list definition. The origin Story is needed in order to be able to look up that definition.
func NewInkListFromOriginStory(singleOriginListName string, originStory *Story) *InkList {

	newInkList := new(InkList)
	newInkList.Dict = make(map[*InkListItem]int)
	newInkList.SetInitialOriginName(singleOriginListName)

	if def, ok := originStory.ListDefinitions().TryListGetDefinition(singleOriginListName); ok {
		newInkList.Origins = append([]*ListDefinition{}, def)
	} else {
		panic("InkList origin could not be found in story when constructing new list: " + singleOriginListName)
	}

	return newInkList
}

func NewInkListFromListItem(listItem *InkListItem, val int) *InkList {

	newInkList := new(InkList)
	newInkList.Dict = make(map[*InkListItem]int)
	newInkList.Dict[listItem] = val

	return newInkList
}

// NewInkListFromString
// Converts a string to an ink list and returns for use in the story.
func NewInkListFromString(myListItem string, originStory *Story) *InkList {
	listValue := originStory.ListDefinitions().FindSingleItemListWithName(myListItem)
	if listValue != nil {
		return NewInkListFromList(listValue)
	} else {
		panic("Could not find the InkListItem from the string '" + myListItem + "' to create an InkList because it doesn't exist in the original list definition in ink.")
	}
}

// AddItem
// Adds the given item to the ink list. Note that the item must come from a list definition that
// is already "known" to this list, so that the item's value can be looked up. By "known", we mean
// that it already has items in it from that source, or it did at one point - it can't be a
// completely fresh empty list, or a list that only contains items from a different list definition.
func (s *InkList) AddItem(item *InkListItem) {

	if item.OriginName() == "" {
		s.AddItemName(item.ItemName())
		return
	}

	for _, origin := range s.Origins {

		if origin.Name() == item.OriginName() {
			if intVal, ok := origin.TryGetValueForItem(item); ok {
				s.Dict[item] = intVal
				return
			} else {
				panic("Could not add the item " + item.String() + " to this list because it doesn't exist in the original list definition in ink.")
			}
		}
	}

	panic("Failed to add item to list because the item was from a new list definition that wasn't previously known to this list. Only items from previously known lists can be used, so that the int value can be found.")
}

// AddItemName
// Adds the given item to the ink list, attempting to find the origin list definition that it belongs to.
// The item must therefore come from a list definition that is already "known" to this list, so that the
// item's value can be looked up. By "known", we mean that it already has items in it from that source, or
// it did at one point - it can't be a completely fresh empty list, or a list that only contains items from
// a different list definition.
func (s *InkList) AddItemName(itemName string) {

	var foundListDef *ListDefinition

	for _, origin := range s.Origins {
		if origin.ContainsItemWithName(itemName) {
			if foundListDef != nil {
				panic("Could not add the item " + itemName + " to this list because it could come from either " + origin.name + " or " + foundListDef.name)
			} else {
				foundListDef = origin
			}
		}
	}

	if foundListDef == nil {
		panic("Could not add the item " + itemName + " to this list because it isn't known to any list definitions previously associated with this list.")
	}

	item := NewInkListItem(foundListDef.Name(), itemName)
	itemVal := foundListDef.ValueForItem(item)
	s.Dict[item] = itemVal
}

// ContainsItemNamed
// Returns true if this ink list contains an item with the given short name
// (ignoring the original list where it was defined).
func (s *InkList) ContainsItemNamed(itemName string) bool {

	for k, _ := range s.Dict {
		if k.ItemName() == itemName {
			return true
		}
	}

	return false
}

func (s *InkList) OriginOfMaxItem() *ListDefinition {

	if s.Origins == nil {
		return nil
	}

	maxOrigin, _ := s.MaxItem()
	for _, origin := range s.Origins {
		if origin.Name() == maxOrigin.OriginName() {
			return origin
		}
	}

	return nil
}

// OriginNames
// Origin name needs to be serialised when content is empty,
// assuming a name is availble, for list definitions with variable
// that is currently empty.
func (s *InkList) OriginNames() []string {

	if s.Length() > 0 {
		if s.originNames == nil && s.Length() > 0 {
			s.originNames = nil
		} else {
			s.originNames = nil
		}

		for k, _ := range s.Dict {
			s.originNames = append(s.originNames, k.OriginName())
		}
	}

	return s.originNames
}

func (s *InkList) SetInitialOriginName(initialOriginName string) {
	s.originNames = append([]string{}, initialOriginName)
}

func (s *InkList) SetInitialOriginNames(initialOriginNames []string) {

	if initialOriginNames == nil {
		s.originNames = nil
	} else {
		s.originNames = append([]string{}, initialOriginNames...)
	}
}

// MaxItem
// Get the maximum item in the list, equivalent to calling LIST_MAX(list) in ink.
func (s *InkList) MaxItem() (*InkListItem, int) {

	newListItem := NewInkListItem("", "")
	newListValue := 0

	for k, v := range s.Dict {
		if newListItem.IsNull() || v > newListValue {
			newListItem = k
			newListValue = v
		}
	}

	return newListItem, newListValue
}

// MinItem
// Get the minimum item in the list, equivalent to calling LIST_MIN(list) in ink.
func (s *InkList) MinItem() (*InkListItem, int) {

	newListItem := NewInkListItem("", "")
	newListValue := 0

	for k, v := range s.Dict {
		if newListItem.IsNull() || v < newListValue {
			newListItem = k
			newListValue = v
		}
	}

	return newListItem, newListValue
}

// Inverse
// The inverse of the list, equivalent to calling LIST_INVERSE(list) in ink
func (s *InkList) Inverse() *InkList {
	list := NewInkList()
	if s.Origins != nil {
		for _, origin := range s.Origins {
			for k, v := range origin.Items() {
				if _, ok := s.Dict[k]; !ok {
					list.Dict[k] = v
				}
			}
		}
	}
	return list
}

// All
// The list of all items from the original list definition, equivalent to calling
// LIST_ALL(list) in ink.
func (s *InkList) All() *InkList {

	list := NewInkList()
	if s.Origins != nil {
		for _, origin := range s.Origins {
			for k, v := range origin.Items() {
				list.Dict[k] = v
			}
		}
	}
	return list
}

// Union
// Returns a new list that is the combination of the current list and one that's
// passed in. Equivalent to calling (list1 + list2) in ink.
func (s *InkList) Union(otherList *InkList) *InkList {
	union := NewInkListFromList(s)
	for k, v := range otherList.Dict {
		union.Dict[k] = v
	}
	return union
}

// Intersect
// Returns a new list that is the intersection of the current list with another
// list that's passed in - i.e. a list of the items that are shared between the
// two other lists. Equivalent to calling (list1 ^ list2) in ink.
func (s *InkList) Intersect(otherList *InkList) *InkList {
	intersection := NewInkList()
	for k, v := range s.Dict {
		if _, ok := otherList.Dict[k]; ok {
			intersection.Dict[k] = v
		}
	}
	return intersection
}

// HasIntersection
// Fast test for the existence of any intersection between the current list and another
func (s *InkList) HasIntersection(otherList *InkList) bool {

	for k, _ := range s.Dict {
		if _, ok := otherList.Dict[k]; !ok {
			return false
		}
	}

	return true
}

// Without
// Returns a new list that's the same as the current one, except with the given items
// removed that are in the passed in list. Equivalent to calling (list1 - list2) in ink.
func (s *InkList) Without(listToRemove *InkList) *InkList {

	result := NewInkListFromList(s)
	for k, _ := range listToRemove.Dict {
		delete(result.Dict, k)
	}
	return result
}

// Contains
// Returns true if the current list contains all the items that are in the list that
// is passed in. Equivalent to calling (list1 ? list2) in ink.
func (s *InkList) Contains(otherList *InkList) bool {

	if otherList.Length() == 0 || s.Length() == 0 {
		return false
	}
	for k, _ := range otherList.Dict {
		if _, ok := s.Dict[k]; !ok {
			return false
		}
	}
	return true
}

// ContainsItemName
// Returns true if the current list contains an item matching the given name.
func (s *InkList) ContainsItemName(listItemName string) bool {

	for k, _ := range s.Dict {
		if k.ItemName() == listItemName {
			return true
		}
	}

	return false
}

// GreaterThan
// Returns true if all the item values in the current list are greater than all the
// item values in the passed in list. Equivalent to calling (list1 > list2) in ink.
func (s *InkList) GreaterThan(otherList *InkList) bool {
	if s.Length() == 0 {
		return false
	}
	if otherList.Length() == 0 {
		return true
	}

	_, value := s.MinItem()
	_, otherValue := otherList.MaxItem()

	return value > otherValue
}

// GreaterThanOrEquals
// Returns true if the item values in the current list overlap or are all greater than
// the item values in the passed in list. None of the item values in the current list must
// fall below the item values in the passed in list. Equivalent to (list1 >= list2) in ink,
// or LIST_MIN(list1) >= LIST_MIN(list2) &amp;&amp; LIST_MAX(list1) >= LIST_MAX(list2).
func (s *InkList) GreaterThanOrEquals(otherList *InkList) bool {

	if s.Length() == 0 {
		return false
	}
	if otherList.Length() == 0 {
		return true
	}

	_, valueMin := s.MinItem()
	_, otherValueMin := otherList.MinItem()
	_, valueMax := s.MaxItem()
	_, otherValueMax := otherList.MaxItem()

	return valueMin >= otherValueMin && valueMax >= otherValueMax
}

// LessThan
// Returns true if all the item values in the current list are less than all the
// item values in the passed in list. Equivalent to calling (list1 &lt; list2) in ink.
func (s *InkList) LessThan(otherList *InkList) bool {

	if otherList.Length() == 0 {
		return false
	}
	if s.Length() == 0 {
		return true
	}

	_, valueMax := s.MaxItem()
	_, otherValueMin := otherList.MinItem()

	return valueMax < otherValueMin
}

// LessThanOrEquals
// Returns true if the item values in the current list overlap or are all less than
// the item values in the passed in list. None of the item values in the current list must
// go above the item values in the passed in list. Equivalent to (list1 &lt;= list2) in ink,
// or LIST_MAX(list1) &lt;= LIST_MAX(list2) &amp;&amp; LIST_MIN(list1) &lt;= LIST_MIN(list2).
func (s *InkList) LessThanOrEquals(otherList *InkList) bool {

	if otherList.Length() == 0 {
		return false
	}
	if s.Length() == 0 {
		return true
	}

	_, valueMin := s.MinItem()
	_, otherValueMin := otherList.MinItem()
	_, valueMax := s.MaxItem()
	_, otherValueMax := otherList.MaxItem()

	return valueMax <= otherValueMax && valueMin <= otherValueMin
}

func (s *InkList) MaxAsList() *InkList {

	if s.Length() > 0 {
		k, v := s.MaxItem()
		return NewInkListFromListItem(k, v)
	} else {
		return NewInkList()
	}
}

// ListWithSubRange
// Returns a sublist with the elements given the minimum and maxmimum bounds.
// The bounds can either be ints which are indices into the entire (sorted) list,
// or they can be InkLists themselves. These are intended to be single-item lists so
// you can specify the upper and lower bounds. If you pass in multi-item lists, it'll
// use the minimum and maximum items in those lists respectively.
// WARNING: Calling this method requires a full sort of all the elements in the list.
func (s *InkList) ListWithSubRange(minBound interface{}, maxBound interface{}) *InkList {

	if s.Length() == 0 {
		return NewInkList()
	}

	ordered := s.OrderedItems()
	minValue := 0
	maxValue := math.MaxInt32

	if v, ok := minBound.(int); ok {
		minValue = v
	} else {

		if v, ok := minBound.(*InkList); ok && v.Length() > 0 {
			_, minValue = v.MinItem()
		}
	}

	if v, ok := maxBound.(int); ok {
		maxValue = v
	} else {

		if v, ok := minBound.(*InkList); ok && v.Length() > 0 {
			_, maxValue = v.MaxItem()
		}
	}

	subList := NewInkList()
	subList.SetInitialOriginNames(s.OriginNames())

	for _, item := range ordered {
		if item.Value >= minValue && item.Value <= maxValue {
			subList.Dict[item.Key] = item.Value
		}
	}

	return subList
}

// Equals
// Returns true if the passed object is also an ink list that contains
// the same items as the current list, false otherwise.
func (s *InkList) Equals(otherList *InkList) bool {

	if otherList == nil {
		return false
	}

	if otherList.Length() != s.Length() {
		return false
	}

	for k, _ := range s.Dict {
		if _, ok := otherList.Dict[k]; !ok {
			return false
		}
	}

	return true
}

// HashCode
// Return the hashcode for this object,
// used for comparisons and inserting into dictionaries.
func (s *InkList) HashCode() string {

	var sb strings.Builder

	for k, _ := range s.Dict {
		sb.WriteString(">>")
		sb.WriteString(k.HashCode())
		sb.WriteString("<<")
	}

	return sb.String()
}

type InkListItemPair struct {
	Key   *InkListItem
	Value int
}

func (s *InkList) OrderedItems() []InkListItemPair {

	var ordered []InkListItemPair
	for k, v := range s.Dict {
		ordered = append(ordered, InkListItemPair{Key: k, Value: v})
	}

	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Value < ordered[j].Value
	})

	return ordered
}

// String
// Returns a string in the form "a, b, c" with the names of the items in the list, without
// the origin list definition names. Equivalent to writing {list} in ink.
func (s *InkList) String() string {

	ordered := s.OrderedItems()

	var sb strings.Builder

	for i := 0; i < len(ordered); i++ {
		if i > 0 {
			sb.WriteString(", ")
		}

		item := ordered[i].Key
		sb.WriteString(item.ItemName())
	}

	return sb.String()
}