package runtime

type Void struct {
	ObjectImpl
}

func NewVoid() *Void {

	newVoid := new(Void)
	newVoid.this = newVoid

	return newVoid
}
