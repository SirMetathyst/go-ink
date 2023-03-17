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

	// Private
	commandType CommandType
}

func NewControlCommandWithCommand(commandType CommandType) *ControlCommand {

	newControlCommand := new(ControlCommand)
	newControlCommand.this = newControlCommand
	newControlCommand.commandType = commandType

	return newControlCommand
}

func NewControlCommand() *ControlCommand {

	newControlCommand := new(ControlCommand)
	newControlCommand.this = newControlCommand
	newControlCommand.commandType = CommandTypeNotSet

	return newControlCommand
}

func (s *ControlCommand) Copy() Object {
	return NewControlCommandWithCommand(s.commandType)
}

func (s *ControlCommand) String() string {
	return fmt.Sprint(s.commandType)
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
	return NewControlCommandWithCommand(CommandTypeEvalStart)
}

func NewEvalOutputCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeEvalOutput)
}

func NewEvalEndCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeEvalEnd)
}

func NewDuplicateCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeDuplicate)
}

func NewPopEvaluatedValueCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypePopEvaluatedValue)
}

func NewPopFunctionCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypePopFunction)
}

func NewPopTunnelCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypePopTunnel)
}

func NewBeginStringCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeBeginString)
}

func NewEndStringCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeEndString)
}

func NewNoOpCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeNoOp)
}

func NewChoiceCountCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeChoiceCount)
}

func NewTurnsCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeTurns)
}

func NewTurnsSinceCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeTurnsSince)
}

func NewReadCountCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeReadCount)
}

func NewRandomCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeRandom)
}

func NewSeedRandomCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeSeedRandom)
}

func NewVisitIndexCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeVisitIndex)
}

func NewSequenceShuffleIndexCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeSequenceShuffleIndex)
}

func NewStartThreadCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeStartThread)
}

func NewDoneCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeDone)
}

func NewEndCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeEnd)
}

func NewListFromIntCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeListFromInt)
}

func NewListRangeCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeListRange)
}

func NewListRandomCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeListRandom)
}

func NewBeginTagCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeBeginTag)
}

func NewEndTagCommand() *ControlCommand {
	return NewControlCommandWithCommand(CommandTypeEndTag)
}
