// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package inkjet

import (
	"strconv"

	flatbuffers "github.com/google/flatbuffers/go"
)

type OpCode int8

const (
	OpCodeNOP OpCode = 0
	OpCodeTXT OpCode = 1
	OpCodeJMP OpCode = 2
	OpCodeHLT OpCode = 3
)

var EnumNamesOpCode = map[OpCode]string{
	OpCodeNOP: "NOP",
	OpCodeTXT: "TXT",
	OpCodeJMP: "JMP",
	OpCodeHLT: "HLT",
}

var EnumValuesOpCode = map[string]OpCode{
	"NOP": OpCodeNOP,
	"TXT": OpCodeTXT,
	"JMP": OpCodeJMP,
	"HLT": OpCodeHLT,
}

func (v OpCode) String() string {
	if s, ok := EnumNamesOpCode[v]; ok {
		return s
	}
	return "OpCode(" + strconv.FormatInt(int64(v), 10) + ")"
}

type Instruction struct {
	_tab flatbuffers.Table
}

func GetRootAsInstruction(buf []byte, offset flatbuffers.UOffsetT) *Instruction {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Instruction{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Instruction) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Instruction) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Instruction) Op() OpCode {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return OpCode(rcv._tab.GetInt8(o + rcv._tab.Pos))
	}
	return 0
}

func (rcv *Instruction) MutateOp(n OpCode) bool {
	return rcv._tab.MutateInt8Slot(4, int8(n))
}

func (rcv *Instruction) Oprand1() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Instruction) MutateOprand1(n int32) bool {
	return rcv._tab.MutateInt32Slot(6, n)
}

func (rcv *Instruction) Oprand2() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func InstructionStart(builder *flatbuffers.Builder) {
	builder.StartObject(3)
}
func InstructionAddOp(builder *flatbuffers.Builder, op OpCode) {
	builder.PrependInt8Slot(0, int8(op), 0)
}
func InstructionAddOprand1(builder *flatbuffers.Builder, oprand1 int32) {
	builder.PrependInt32Slot(1, oprand1, 0)
}
func InstructionAddOprand2(builder *flatbuffers.Builder, oprand2 flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(oprand2), 0)
}
func InstructionEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}

type Program struct {
	_tab flatbuffers.Table
}

func GetRootAsProgram(buf []byte, offset flatbuffers.UOffsetT) *Program {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Program{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Program) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Program) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Program) Version() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Program) MutateVersion(n int32) bool {
	return rcv._tab.MutateInt32Slot(4, n)
}

func (rcv *Program) Instructions(obj *Instruction, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Program) InstructionsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func ProgramStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func ProgramAddVersion(builder *flatbuffers.Builder, version int32) {
	builder.PrependInt32Slot(0, version, 0)
}
func ProgramAddInstructions(builder *flatbuffers.Builder, instructions flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(instructions), 0)
}
func ProgramStartInstructionsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func ProgramEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}