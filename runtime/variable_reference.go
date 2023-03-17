package runtime

import "fmt"

type VariableReference struct {
	ObjectImpl
	// Normal named variable
	Name         string
	PathForCount *Path
}

func (s *VariableReference) ContainerForCount() *Container {
	return s.ResolvePath(s.PathForCount).Container()
}

// PathStringForCount
// Variable reference is actually a path for a visit (read) count
func (s *VariableReference) PathStringForCount() string {

	if s.PathForCount == nil {
		return ""
	}

	return s.CompactPathString(s.PathForCount)
}

func (s *VariableReference) SetPathStringForCount(value string) {

	if value == "" {
		s.PathForCount = nil
	} else {
		s.PathForCount = NewPathFromComponentString(value)
	}
}

func NewVariableReference() *VariableReference {

	newVariableReference := new(VariableReference)
	newVariableReference.this = newVariableReference

	return newVariableReference
}

func NewVariableReferenceFromName(name string) *VariableReference {

	newVariableReference := new(VariableReference)
	newVariableReference.Name = name
	newVariableReference.this = newVariableReference

	return newVariableReference
}

func (s *VariableReference) String() string {

	if s.Name != "" {
		return fmt.Sprintf("var(%s)", s.Name)
	} else {
		pathStr := s.PathStringForCount()
		return fmt.Sprintf("read_count(%s)", pathStr)
	}
}
