package runtime

import "fmt"

type VariableReference struct {
	ObjectImpl
	// Normal named variable
	Name         string
	PathForCount *Path
}

func (s *VariableReference) ContainerForCount() *Container {
	return ResolvePath(s, s.PathForCount).Container()
}

// PathStringForCount
// Variable reference is actually a path for a visit (read) count
func (s *VariableReference) PathStringForCount() (string, bool) {

	if s.PathForCount == nil {
		return "", false // Nil string over empty string is important here
	}

	x := CompactPathString(s, s.PathForCount)

	return x, true
}

func (s *VariableReference) SetPathStringForCount(value string) {

	if value == "" {
		s.PathForCount = nil
	} else {
		s.PathForCount = NewPathFromString(value)
	}
}

func NewVariableReference() *VariableReference {

	newVariableReference := new(VariableReference)

	return newVariableReference
}

func NewVariableReferenceFromName(name string) *VariableReference {

	newVariableReference := new(VariableReference)
	newVariableReference.Name = name

	return newVariableReference
}

func (s *VariableReference) String() string {

	if s.Name != "" {
		return fmt.Sprintf("var(%s)", s.Name)
	} else {
		pathStr, _ := s.PathStringForCount()
		return fmt.Sprintf("read_count(%s)", pathStr)
	}
}
