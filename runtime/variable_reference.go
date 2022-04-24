package runtime

import "fmt"

type VariableReference struct {
	*objectImpl
	Name         string
	PathForCount *Path
}

func (s *VariableReference) PathStringForCount() (string, bool) {

	if s.PathForCount == nil {
		return "", false
	}

	return s.CompactPathString(s.PathForCount), true
}

func (s *VariableReference) SetPathStringForCount(value string) {

	if len(value) == 0 {
		s.PathForCount = nil
	} else {
		s.PathForCount = NewPathFromComponentsString(value)
	}
}

func (s *VariableReference) String() string {

	if len(s.Name) != 0 {
		return fmt.Sprintf("var(%s)", s.Name)
	}
	pathStr, _ := s.PathStringForCount()
	return fmt.Sprintf("read_count(%s)", pathStr)
}

func NewVariableReferenceWithName(name string) *VariableReference {
	s := &VariableReference{Name: name}
	s.objectImpl = &objectImpl{instance: s}
	return s
}

func NewVariableReference() *VariableReference {
	s := &VariableReference{}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
