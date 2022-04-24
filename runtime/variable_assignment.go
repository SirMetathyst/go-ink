package runtime

import "fmt"

type VariableAssignment struct {
	*objectImpl
	IsGlobal         bool
	variableName     string
	isNewDeclaration bool
}

func (s *VariableAssignment) VariableName() string {
	return s.variableName
}

func (s *VariableAssignment) IsNewDeclaration() bool {
	return s.isNewDeclaration
}

func (s *VariableAssignment) String() string {
	return fmt.Sprintf("VarAssign to %s", s.variableName)
}

func NewVariableAssignmentWith(variableName string, isNewDeclaration bool) *VariableAssignment {
	s := &VariableAssignment{variableName: variableName, isNewDeclaration: isNewDeclaration}
	s.objectImpl = &objectImpl{instance: s}
	return s
}

func NewVariableAssignment() *VariableAssignment {
	s := &VariableAssignment{}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
