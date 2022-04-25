package runtime

type Divert struct {
	*objectImpl
	VariableDivertName string
	PushesToStack      bool
	StackPushType      PushPopType
	IsExternal         bool
	ExternalArgs       int
	IsConditional      bool
	targetPath         *Path
	targetPointer      *Pointer
}

func (s *Divert) TargetPath() *Path {

	if s.targetPath != nil && s.targetPath.IsRelative() {
		targetObj := s.TargetPointer().Resolve()
		if targetObj != nil {
			s.targetPath = targetObj.Path()
		}
	}

	return s.targetPath
}

func (s *Divert) SetTargetPath(path *Path) {
	s.targetPath = path
	s.targetPointer = NewNilPointer()
}

func (s *Divert) TargetPointer() *Pointer {

	if s.targetPointer.IsNil() {

		targetObj := s.ResolvePath(s.targetPath).Obj
		lastComponent, _ := s.targetPath.LastComponent()

		if lastComponent.IsIndex() {
			s.targetPointer.Container, _ = targetObj.Parent().(*Container)
			s.targetPointer.Index = lastComponent.Index()
		} else {
			targetObjContainer, _ := targetObj.(*Container)
			s.targetPointer = NewPointerStartOf(targetObjContainer)
		}
	}

	return s.targetPointer
}

func (s *Divert) TargetPathString() (string, bool) {

	if s.targetPath == nil {
		return "", false
	}

	return s.CompactPathString(s.targetPath), true
}

func (s *Divert) SetTargetPathString(path string) {

	if len(path) == 0 {
		s.targetPath = nil
	} else {
		s.targetPath = NewPathFromComponentsString(path)
	}
}

func (s *Divert) HasVariableTarget() bool {
	return len(s.VariableDivertName) != 0
}

func (s *Divert) Equals(otherDivert *Divert) bool {

	if otherDivert != nil {

		if s.HasVariableTarget() == otherDivert.HasVariableTarget() {
			if s.HasVariableTarget() {
				return s.VariableDivertName == otherDivert.VariableDivertName
			} else {
				return s.targetPath.Equals(otherDivert.targetPath)
			}
		}
	}

	return false
}

func (s *Divert) String() string {

	if s.HasVariableTarget() {
		return s.VariableDivertName
	}

	return s.targetPath.String()
}

func NewDivertWith(stackPushType PushPopType) *Divert {
	s := &Divert{PushesToStack: true, StackPushType: stackPushType}
	s.objectImpl = &objectImpl{instance: s}
	return s
}

func NewDivert() *Divert {
	s := &Divert{PushesToStack: false}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
