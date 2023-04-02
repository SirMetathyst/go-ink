package runtime

import "fmt"

type CommandType int

const (
	CommandTypeNotSet CommandType = iota - 1
	CommandTypeEvalStart
	CommandTypeEvalOutput
	CommandTypeEvalEnd
	CommandTypeDuplicate
	CommandTypePopEvaluatedValue
	CommandTypePopFunction
	CommandTypePopTunnel
	CommandTypeBeginString
	CommandTypeEndString
	CommandTypeNoOp
	CommandTypeChoiceCount
	CommandTypeTurns
	CommandTypeTurnsSince
	CommandTypeReadCount
	CommandTypeRandom
	CommandTypeSeedRandom
	CommandTypeVisitIndex
	CommandTypeSequenceShuffleIndex
	CommandTypeStartThread
	CommandTypeDone
	CommandTypeEnd
	CommandTypeListFromInt
	CommandTypeListRange
	CommandTypeListRandom
	CommandTypeBeginTag
	CommandTypeEndTag
	//----
	CommandTypeTotalValues
)

type ControlCommand struct {
	ObjectImpl

	// Public
	CommandType CommandType
}

func NewControlCommand(commandType CommandType) *ControlCommand {

	newControlCommand := new(ControlCommand)
	newControlCommand.CommandType = commandType

	return newControlCommand
}

func (s *ControlCommand) Copy() Object {
	return NewControlCommand(s.CommandType)
}

func (s *ControlCommand) String() string {
	return fmt.Sprint(s.CommandType)
}

// The following static factory methods are to make generating these objects
// slightly more succinct. Without these, the code gets pretty massive! e.g.
//
//     var c = new Runtime.ControlCommand(Runtime.ControlCommand.CommandType.EvalStart)
//
// as opposed to
//
//     var c = Runtime.ControlCommand.EvalStart()

func NewEvalStartCommand() *ControlCommand {
	return NewControlCommand(CommandTypeEvalStart)
}

func NewEvalOutputCommand() *ControlCommand {
	return NewControlCommand(CommandTypeEvalOutput)
}

func NewEvalEndCommand() *ControlCommand {
	return NewControlCommand(CommandTypeEvalEnd)
}

func NewDuplicateCommand() *ControlCommand {
	return NewControlCommand(CommandTypeDuplicate)
}

func NewPopEvaluatedValueCommand() *ControlCommand {
	return NewControlCommand(CommandTypePopEvaluatedValue)
}

func NewPopFunctionCommand() *ControlCommand {
	return NewControlCommand(CommandTypePopFunction)
}

func NewPopTunnelCommand() *ControlCommand {
	return NewControlCommand(CommandTypePopTunnel)
}

func NewBeginStringCommand() *ControlCommand {
	return NewControlCommand(CommandTypeBeginString)
}

func NewEndStringCommand() *ControlCommand {
	return NewControlCommand(CommandTypeEndString)
}

func NewNoOpCommand() *ControlCommand {
	return NewControlCommand(CommandTypeNoOp)
}

func NewChoiceCountCommand() *ControlCommand {
	return NewControlCommand(CommandTypeChoiceCount)
}

func NewTurnsCommand() *ControlCommand {
	return NewControlCommand(CommandTypeTurns)
}

func NewTurnsSinceCommand() *ControlCommand {
	return NewControlCommand(CommandTypeTurnsSince)
}

func NewReadCountCommand() *ControlCommand {
	return NewControlCommand(CommandTypeReadCount)
}

func NewRandomCommand() *ControlCommand {
	return NewControlCommand(CommandTypeRandom)
}

func NewSeedRandomCommand() *ControlCommand {
	return NewControlCommand(CommandTypeSeedRandom)
}

func NewVisitIndexCommand() *ControlCommand {
	return NewControlCommand(CommandTypeVisitIndex)
}

func NewSequenceShuffleIndexCommand() *ControlCommand {
	return NewControlCommand(CommandTypeSequenceShuffleIndex)
}

func NewStartThreadCommand() *ControlCommand {
	return NewControlCommand(CommandTypeStartThread)
}

func NewDoneCommand() *ControlCommand {
	return NewControlCommand(CommandTypeDone)
}

func NewEndCommand() *ControlCommand {
	return NewControlCommand(CommandTypeEnd)
}

func NewListFromIntCommand() *ControlCommand {
	return NewControlCommand(CommandTypeListFromInt)
}

func NewListRangeCommand() *ControlCommand {
	return NewControlCommand(CommandTypeListRange)
}

func NewListRandomCommand() *ControlCommand {
	return NewControlCommand(CommandTypeListRandom)
}

func NewBeginTagCommand() *ControlCommand {
	return NewControlCommand(CommandTypeBeginTag)
}

func NewEndTagCommand() *ControlCommand {
	return NewControlCommand(CommandTypeEndTag)
}
