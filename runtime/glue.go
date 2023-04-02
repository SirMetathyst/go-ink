package runtime

type Glue struct {
	ObjectImpl
}

func NewGlue() *Glue {

	newGlue := new(Glue)
	return newGlue
}

func (s *Glue) String() string {
	return "Glue"
}
