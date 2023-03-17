package runtime

type NamedContent interface {
	Name() string
	HasValidName() bool
}
