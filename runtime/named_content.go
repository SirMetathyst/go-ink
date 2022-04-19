package ink

type NamedContent interface {
	Name() string
	HasValidName() bool
}
