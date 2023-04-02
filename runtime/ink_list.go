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
	_originName string
	_itemName   string
}

// OriginName
// The name of the list where the item was originally defined.
func (s InkListItem) OriginName() string {
	return s._originName
}

// ItemName
// The main name of the item as defined in ink.
func (s InkListItem) ItemName() string {
	return s._itemName
}

// NewInkListItem
// Create an item with the given original list definition name, and the name of this item.
func NewInkListItem(originName string, itemName string) InkListItem {

	inkListItem := InkListItem{}
	inkListItem._originName = originName
	inkListItem._itemName = itemName

	return inkListItem
}

// NewInkListFromFullname
// Create an item from a dot-separted string of the form "listDefinitionName.listItemName".
func NewInkListFromFullname(fullName string) InkListItem {

	inkListItem := InkListItem{}
	nameParts := strings.Split(fullName, ".")
	inkListItem._originName = nameParts[0]
	inkListItem._itemName = nameParts[1]

	return inkListItem
}

func (s InkListItem) IsNull() bool {

	return s._originName == "" && s._itemName == ""
}

// Fullname
// Get the full dot-separated name of the item, in the form "listDefinitionName.itemName".
func (s InkListItem) Fullname() string {

	v := s.OriginName()
	if v == "" {
		v = "?"
	}

	return v + "." + s.ItemName()
}

type InkList struct {
	//Dictionary[InkListItem, int]

	_items map[InkListItem]int

	// Private
	_originNames []string

	// Public
	Origins []*ListDefinition
}

func (s *InkList) Count() int {
	return len(s._items)
}

func (s *InkList) Set(key InkListItem, value int) {
	s._items[key] = value
}

func (s *InkList) Remove(key InkListItem) {
	delete(s._items, key)
}

func (s *InkList) ContainsKey(key InkListItem) bool {
	_, ok := s._items[key]
	return ok
}

func (s *InkList) Add(key InkListItem, value int) {
	if _, ok := s._items[key]; ok == false {
		s._items[key] = value
	} else {
		panic("An item with the same key has already been added. Key: " + fmt.Sprint(key))
	}
}

func NewInkList() *InkList {

	newInkList := new(InkList)

	return newInkList
}

// NewInkListFromInkList
// Create a new ink list that contains the same contents as another list.
func NewInkListFromInkList(otherList *InkList) *InkList {

	newInkList := NewInkList()

	otherOriginNames := otherList.OriginNames()
	if otherOriginNames != nil {
		newInkList._originNames = NewSliceFromSlice(otherOriginNames)
	}

	if otherList.Origins != nil {
		newInkList.Origins = NewSliceFromSlice(otherList.Origins)
	}

	return newInkList
}

// NewInkListFromOriginStory
// Create a new empty ink list that's intended to hold items from a particular origin
// list definition. The origin Story is needed in order to be able to look up that definition.
func NewInkListFromOriginStory(singleOriginListName string, originStory *Story) *InkList {

	newInkList := NewInkList()
	newInkList.SetInitialOriginName(singleOriginListName)

	if def, ok := originStory.ListDefinitions().TryListGetDefinition(singleOriginListName); ok {

		newInkList.Origins = []*ListDefinition{def}
		return newInkList
	}

	panic("InkList origin could not be found in story when constructing new list: " + singleOriginListName)
}

func NewInkListFromSingleElement(singleElement KeyValuePair[InkListItem, int]) *InkList {

	newInkList := NewInkList()
	newInkList.Add(singleElement.Key, singleElement.Value)

	return newInkList
}

func (s *InkList) OriginOfMaxItem() *ListDefinition {

	if s.Origins == nil {

		return nil
	}

	maxOriginName := s.MaxItem().Key.OriginName()
	for _, origin := range s.Origins {

		if origin.Name() == maxOriginName {

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

	if s.Count() > 0 {

		if s._originNames == nil && s.Count() > 0 {

			s._originNames = []string{}
		} else {

			s._originNames = s._originNames[:0]
		}

		for item, _ := range s._items {
			s._originNames = append(s._originNames, item.OriginName())
		}
	}

	return s._originNames
}

func (s *InkList) SetInitialOriginName(initialOriginName string) {

	s._originNames = []string{initialOriginName}
}

func (s *InkList) SetInitialOriginNames(initialOriginNames []string) {

	if initialOriginNames == nil {

		s._originNames = nil
		return
	}

	s._originNames = initialOriginNames
}

// MaxItem
// Get the maximum item in the list, equivalent to calling LIST_MAX(list) in ink.
func (s *InkList) MaxItem() KeyValuePair[InkListItem, int] {

	max := KeyValuePair[InkListItem, int]{}

	for key, value := range s._items {

		if max.Key.IsNull() || value > max.Value {

			max.Key = key
			max.Value = value
		}
	}

	return max
}

// MinItem
// Get the minimum item in the list, equivalent to calling LIST_MIN(list) in ink.
func (s *InkList) MinItem() KeyValuePair[InkListItem, int] {

	min := KeyValuePair[InkListItem, int]{}

	for key, value := range s._items {

		if min.Key.IsNull() || value < min.Value {

			min.Key = key
			min.Value = value
		}
	}

	return min
}

// Inverse
// The inverse of the list, equivalent to calling LIST_INVERSE(list) in ink
func (s *InkList) Inverse() *InkList {

	list := NewInkList()

	if s.Origins != nil {

		for _, origin := range s.Origins {

			for key, value := range origin.Items() {

				if s.ContainsKey(key) == false {
					list.Add(key, value)
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

			for key, value := range origin.Items() {

				list.Set(key, value)
			}
		}
	}
	return list
}

// Union
// Returns a new list that is the combination of the current list and one that's
// passed in. Equivalent to calling (list1 + list2) in ink.
func (s *InkList) Union(otherList *InkList) *InkList {

	union := NewInkListFromInkList(s)

	for key, value := range otherList._items {

		union.Set(key, value)
	}

	return union
}

// Intersect
// Returns a new list that is the intersection of the current list with another
// list that's passed in - i.e. a list of the items that are shared between the
// two other lists. Equivalent to calling (list1 ^ list2) in ink.
func (s *InkList) Intersect(otherList *InkList) *InkList {

	intersection := NewInkList()
	for key, value := range s._items {

		if otherList.ContainsKey(key) {

			intersection.Add(key, value)
		}
	}

	return intersection
}

// Without
// Returns a new list that's the same as the current one, except with the given items
// removed that are in the passed in list. Equivalent to calling (list1 - list2) in ink.
func (s *InkList) Without(listToRemove *InkList) *InkList {

	result := NewInkListFromInkList(s)

	for key, _ := range listToRemove._items {

		result.Remove(key)
	}

	return result
}

// Contains
// Returns true if the current list contains all the items that are in the list that
// is passed in. Equivalent to calling (list1 ? list2) in ink.
func (s *InkList) Contains(otherList *InkList) bool {

	if otherList.Count() == 0 || s.Count() == 0 {
		return false
	}

	for key, _ := range otherList._items {

		if s.ContainsKey(key) == false {
			return false
		}
	}

	return true
}

// GreaterThan
// Returns true if all the item values in the current list are greater than all the
// item values in the passed in list. Equivalent to calling (list1 > list2) in ink.
func (s *InkList) GreaterThan(otherList *InkList) bool {

	if s.Count() == 0 {
		return false
	}

	if otherList.Count() == 0 {

		return true
	}

	return s.MinItem().Value > otherList.MaxItem().Value
}

// GreaterThanOrEquals
// Returns true if the item values in the current list overlap or are all greater than
// the item values in the passed in list. None of the item values in the current list must
// fall below the item values in the passed in list. Equivalent to (list1 >= list2) in ink,
// or LIST_MIN(list1) >= LIST_MIN(list2) &amp;&amp; LIST_MAX(list1) >= LIST_MAX(list2).
func (s *InkList) GreaterThanOrEquals(otherList *InkList) bool {

	if s.Count() == 0 {
		return false
	}

	if otherList.Count() == 0 {
		return true
	}

	return s.MinItem().Value >= otherList.MinItem().Value &&
		s.MaxItem().Value >= otherList.MaxItem().Value
}

// LessThan
// Returns true if all the item values in the current list are less than all the
// item values in the passed in list. Equivalent to calling (list1 &lt; list2) in ink.
func (s *InkList) LessThan(otherList *InkList) bool {

	if otherList.Count() == 0 {
		return false
	}

	if s.Count() == 0 {
		return true
	}

	return s.MaxItem().Value < otherList.MinItem().Value
}

// LessThanOrEquals
// Returns true if the item values in the current list overlap or are all less than
// the item values in the passed in list. None of the item values in the current list must
// go above the item values in the passed in list. Equivalent to (list1 &lt;= list2) in ink,
// or LIST_MAX(list1) &lt;= LIST_MAX(list2) &amp;&amp; LIST_MIN(list1) &lt;= LIST_MIN(list2).
func (s *InkList) LessThanOrEquals(otherList *InkList) bool {

	if otherList.Count() == 0 {
		return false
	}

	if s.Count() == 0 {
		return true
	}

	return s.MaxItem().Value <= otherList.MaxItem().Value &&
		s.MinItem().Value <= otherList.MinItem().Value
}

func (s *InkList) MaxAsList() *InkList {

	if s.Count() > 0 {

		return NewInkListFromSingleElement(s.MaxItem())
	}

	return NewInkList()
}

func (s *InkList) MinAsList() *InkList {

	if s.Count() > 0 {

		return NewInkListFromSingleElement(s.MinItem())
	}

	return NewInkList()
}

// ListWithSubRange
// Returns a sublist with the elements given the minimum and maxmimum bounds.
// The bounds can either be ints which are indices into the entire (sorted) list,
// or they can be InkLists themselves. These are intended to be single-item lists so
// you can specify the upper and lower bounds. If you pass in multi-item lists, it'll
// use the minimum and maximum items in those lists respectively.
// WARNING: Calling this method requires a full sort of all the elements in the list.
func (s *InkList) ListWithSubRange(minBound interface{}, maxBound interface{}) *InkList {

	if s.Count() == 0 {

		return NewInkList()
	}

	ordered := s.OrderedItems()

	minValue := 0
	maxValue := math.MaxInt32

	if v, isInt := minBound.(int); isInt {

		minValue = v
	} else {

		if v, isInkList := minBound.(*InkList); isInkList && v.Count() > 0 {
			minValue = v.MinItem().Value
		}
	}

	if v, isInt := maxBound.(int); isInt {

		maxValue = v
	} else {

		if v, isInkList := minBound.(*InkList); isInkList && v.Count() > 0 {
			maxValue = v.MaxItem().Value
		}
	}

	subList := NewInkList()
	subList.SetInitialOriginNames(s.OriginNames())

	for _, item := range ordered {

		if item.Value >= minValue && item.Value <= maxValue {

			subList.Add(item.Key, item.Value)
		}
	}

	return subList
}

func (s *InkList) OrderedItems() []KeyValuePair[InkListItem, int] {

	ordered := []KeyValuePair[InkListItem, int]{}
	for key, value := range s._items {
		ordered = append(ordered, KeyValuePair[InkListItem, int]{key, value})
	}

	sort.Slice(ordered, func(i, j int) bool {

		if ordered[i].Key.OriginName() == ordered[j].Key.OriginName() {
			return ordered[i].Key.OriginName() < ordered[j].Key.OriginName()
		}

		return ordered[i].Value < ordered[j].Value
	})

	return ordered
}

/*
   public override bool Equals (object other)
     {
         var otherRawList = other as InkList;
         if (otherRawList == null) return false;
         if (otherRawList.Count != Count) return false;

         foreach (var kv in this) {
             if (!otherRawList.ContainsKey (kv.Key))
                 return false;
         }

         return true;
     }
*/

func (s *InkList) Equals(other *InkList) bool {

	if other == nil {
		return false
	}

	if other.Count() != s.Count() {
		return false
	}

	for key := range s._items {
		if other.ContainsKey(key) == false {
			return false
		}
	}

	return true
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
