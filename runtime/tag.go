package runtime

type Tag struct {
	*objectImpl
	text string
}

func (s *Tag) Text() string {
	return s.text
}

func (s *Tag) String() string {
	return "# " + s.text
}

func NewTag(text string) *Tag {
	s := &Tag{text: text}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
