package runtime

type CommandType int

const (
	NotSet CommandType = iota - 1
	EvalStart
	EvalOutput
	EvalEnd
	Duplicate
	PopEvaluatedValue
	PopFunction
	PopTunnel
	BeginString
	EndString
	NoOp
	ChoiceCount
	Turns
	TurnsSince
	ReadCount
	Random
	SeedRandom
	VisitIndex
	SequenceShuffleIndex
	StartThread
	Done
	End
	ListFromInt
	ListRange
	ListRandom
	//----
	TOTAL_VALUES
)

func (s CommandType) String() string {
	switch s {
	case NotSet:
		return "NotSet"
	case EvalStart:
		return "EvalStart"
	case EvalOutput:
		return "EvalOutput"
	case EvalEnd:
		return "EvalEnd"
	case Duplicate:
		return "Duplicate"
	case PopEvaluatedValue:
		return "PopEvaluatedValue"
	case PopFunction:
		return "PopFunction"
	case PopTunnel:
		return "PopTunnel"
	case BeginString:
		return "BeginString"
	case EndString:
		return "EndString"
	case NoOp:
		return "NoOp"
	case ChoiceCount:
		return "ChoiceCount"
	case Turns:
		return "Turns"
	case TurnsSince:
		return "TurnsSince"
	case ReadCount:
		return "ReadCount"
	case Random:
		return "Random"
	case SeedRandom:
		return "SeedRandom"
	case VisitIndex:
		return "VisitIndex"
	case SequenceShuffleIndex:
		return "SequenceShuffleIndex"
	case StartThread:
		return "StartThread"
	case Done:
		return "Done"
	case End:
		return "End"
	case ListFromInt:
		return "ListFromInt"
	case ListRange:
		return "ListRange"
	case ListRandom:
		return "ListRandom"
	default:
		return "Unknown"
	}
}

type ControlCommand struct {
	*objectImpl
	commandType CommandType
}

func (s *ControlCommand) CommandType() CommandType {
	return s.commandType
}

func (s *ControlCommand) Copy() *ControlCommand {
	return NewControlCommandWith(s.commandType)
}

func (s *ControlCommand) String() string {
	return s.commandType.String()
}

func NewControlCommandWith(commandType CommandType) *ControlCommand {
	s := &ControlCommand{commandType: commandType}
	s.objectImpl = &objectImpl{instance: s}
	return s
}

func NewControlCommand() *ControlCommand {
	s := &ControlCommand{commandType: NotSet}
	s.objectImpl = &objectImpl{instance: s}
	return s
}
