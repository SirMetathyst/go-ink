package main

import (
	"fmt"
	flatbuffers "github.com/google/flatbuffers/go"
)

func main() {

	// WRITE

	builder := flatbuffers.NewBuilder(0)

	// Start Instruction 1
	builder.StartObject(3)
	builder.PrependInt8Slot(0, int8(10), 0) // OpCode: 10
	builder.PrependInt32Slot(1, 20, 0)      // Operand 1: 20
	builder.PrependInt32Slot(2, 40, 0)      // Operand 2: 40
	element1 := builder.EndObject()
	// End Instruction 1

	// Start Instruction 2
	builder.StartObject(4)
	builder.PrependInt8Slot(0, int8(30), 0)   // OpCode: 30
	builder.PrependFloat32Slot(1, 128.555, 0) // Operand 1: 128.555
	builder.PrependFloat32Slot(2, 572.892, 0) // Operand 2: 572.892
	builder.PrependFloat32Slot(3, 0.892, 0)   // Operand 2: 572.892
	element2 := builder.EndObject()
	// End Instruction 2

	// Start Instruction 3
	builder.StartObject(3)
	builder.PrependInt8Slot(0, int8(10), 0) // OpCode: 10
	builder.PrependInt32Slot(1, 100, 0)     // Operand 1: 100
	builder.PrependInt32Slot(2, 200, 0)     // Operand 2: 200
	element3 := builder.EndObject()
	// End Instruction 3

	// Start Instruction Vector
	builder.StartVector(4, 3, 4)         // Length of 3 Instruction
	builder.PrependUOffsetT(element3)    // Prepend Element 3
	builder.PrependUOffsetT(element2)    // Prepend Element 2
	builder.PrependUOffsetT(element1)    // Prepend Element 1
	instructions := builder.EndVector(3) // Length of 3 Instruction
	// End Instruction Vector

	// Start Program
	builder.StartObject(2)

	// Prepend Version
	builder.PrependInt32Slot(0, 1, 0)

	// Prepend Instruction Vector
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(instructions), 0)

	// End Program
	program := builder.EndObject()

	// Complete flatbuffer
	builder.Finish(program)
	buf := builder.FinishedBytes()

	fmt.Println(buf)

	// READ
	n := flatbuffers.GetUOffsetT(buf[0:])
	tab := new(flatbuffers.Table)
	tab.Pos = n
	tab.Bytes = buf

	// Version
	if o := flatbuffers.UOffsetT(tab.Offset(4)); o != 0 {
		fmt.Println("Version: ", tab.GetInt32(o+tab.Pos))
	}

	// Instruction Length
	if o := flatbuffers.UOffsetT(tab.Offset(6)); o != 0 {
		fmt.Println("Instruction Length: ", tab.VectorLen(o))

		lastPos := tab.Pos // Remember original position
		// j: index of instruction
		for j := 0; j < tab.VectorLen(o); j++ {

			if o := flatbuffers.UOffsetT(tab.Offset(6)); o != 0 {

				x := tab.Vector(o)
				x += flatbuffers.UOffsetT(j) * 4
				x = tab.Indirect(x)
				tab.Pos = x // Goto offset

				// OpCode (If It Exists)
				if o := flatbuffers.UOffsetT(tab.Offset(4)); o != 0 {
					fmt.Println("OpCode: ", tab.GetInt8(o+tab.Pos))

					switch tab.GetInt8(o + tab.Pos) {

					case 10:
						// Operand 1 (If It Exists)
						if o := flatbuffers.UOffsetT(tab.Offset(6)); o != 0 {
							fmt.Println("Operand 1: ", tab.GetInt32(o+tab.Pos))
						}

						// Operand 2 (If It Exists)
						if o := flatbuffers.UOffsetT(tab.Offset(8)); o != 0 {
							fmt.Println("Operand 2: ", tab.GetInt32(o+tab.Pos))
						}

					case 30:
						// Operand 1 (If It Exists)
						if o := flatbuffers.UOffsetT(tab.Offset(6)); o != 0 {
							fmt.Println("Operand 1: ", tab.GetFloat32(o+tab.Pos))
						}

						// Operand 2 (If It Exists)
						if o := flatbuffers.UOffsetT(tab.Offset(8)); o != 0 {
							fmt.Println("Operand 2: ", tab.GetFloat32(o+tab.Pos))
						}

						// Operand 3 (If It Exists)
						if o := flatbuffers.UOffsetT(tab.Offset(10)); o != 0 {
							fmt.Println("Operand 3: ", tab.GetFloat32(o+tab.Pos))
						}
					}
				}
			}

			// Reset position
			tab.Pos = lastPos
		}
	}
}
