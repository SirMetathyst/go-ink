package runtime

type VariableAssignment struct {
	ObjectImpl

	// Private
	_variableName     string
	_isNewDeclaration bool

	// Public
	IsGlobal bool
}

func (s *VariableAssignment) VariableName() string {
	return s._variableName
}

func (s *VariableAssignment) IsNewDeclaration() bool {
	return s._isNewDeclaration
}

func NewVariableAssignment(variableName string, isNewDeclaration bool) *VariableAssignment {

	newVariableAssignment := new(VariableAssignment)
	newVariableAssignment._variableName = variableName
	newVariableAssignment._isNewDeclaration = isNewDeclaration

	return newVariableAssignment
}

func (s *VariableAssignment) String() string {
	return "VarAssign to " + s._variableName
}
