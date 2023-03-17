package runtime

type VariableAssignment struct {
	ObjectImpl
	variableName     string
	isNewDeclaration bool
	IsGlobal         bool
}

func (s *VariableAssignment) VariableName() string {
	return s.variableName
}

func (s *VariableAssignment) IsNewDeclaration() bool {
	return s.isNewDeclaration
}

func NewVariableAssignment(variableName string, isNewDeclaration bool) *VariableAssignment {

	newVariableAssignment := new(VariableAssignment)
	newVariableAssignment.variableName = variableName
	newVariableAssignment.isNewDeclaration = isNewDeclaration
	newVariableAssignment.this = newVariableAssignment

	return newVariableAssignment
}

func (s *VariableAssignment) String() string {
	return "VarAssign to " + s.variableName
}
