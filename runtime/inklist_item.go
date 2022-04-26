package runtime

import "strings"

type InkListItem struct {
	originName string
	itemName   string
}

func (s *InkListItem) OriginName() string {
	return s.originName
}

func (s *InkListItem) ItemName() string {
	return s.itemName
}

func (s *InkListItem) IsNil() bool {
	return len(s.originName) == 0 && len(s.itemName) == 0
}

func (s *InkListItem) FullName() string {
	v := s.originName
	if len(v) == 0 {
		v = "?"
	}
	return v + "." + s.itemName
}

func (s *InkListItem) Equals(value *InkListItem) bool {

	if value == nil {
		return false
	}

	return value.ItemName() == s.ItemName() && value.OriginName() == s.OriginName()
}

func (s *InkListItem) String() string {
	return s.FullName()
}

func NewNilInkListItem() InkListItem {
	return NewInkListItem("", "")
}

func NewInkListItem(originName string, itemName string) InkListItem {
	return InkListItem{originName: originName, itemName: itemName}
}

func NewInkListItemFromFullName(fullname string) InkListItem {
	nameParts := strings.Split(fullname, ".")
	return NewInkListItem(nameParts[0], nameParts[1])
}
