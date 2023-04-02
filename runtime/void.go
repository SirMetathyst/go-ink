package runtime

type Void struct {
	ObjectImpl
}

func NewVoid() *Void {

	newVoid := new(Void)

	return newVoid
}
