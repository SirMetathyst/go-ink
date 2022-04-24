package runtime

type Void struct {
	*objectImpl
}

func NewVoid() *Void {
	s := &Void{}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
