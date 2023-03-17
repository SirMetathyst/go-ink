package runtime

type Glue struct {
	ObjectImpl
}

func NewGlue() *Glue {

	newGlue := new(Glue)
	newGlue.this = newGlue

	return newGlue
}

func (s *Glue) String() string {
	return "Glue"
}
