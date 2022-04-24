package runtime

type Glue struct {
	*objectImpl
}

func (s *Glue) String() string {
	return "Glue"
}

func NewGlue() *Glue {
	s := &Glue{}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
