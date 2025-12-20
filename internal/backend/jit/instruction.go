package jit

import "github.com/ChernykhITMO/compiler/internal/bytecode"

type Instruction struct {
	OpCode   bytecode.OpCode
	Argument int
	Size     int
}

func Decode(code []byte, ip int) (Instruction, bool) {
	if ip >= len(code) {
		return Instruction{}, false
	}

	OpCode := bytecode.OpCode(code[ip])
	switch OpCode {
	case bytecode.OpConst, bytecode.OpJump, bytecode.OpJumpIfFalse, bytecode.OpCall:
		if ip+2 >= len(code) {
			return Instruction{}, false
		}
		Argument := int(uint16(code[ip+1])<<8 | uint16(code[ip+2]))
		return Instruction{OpCode: OpCode, Argument: Argument, Size: 3}, true

	case bytecode.OpLoadLocal, bytecode.OpStoreLocal:
		if ip+1 >= len(code) {
			return Instruction{}, false
		}
		return Instruction{OpCode: OpCode, Argument: int(code[ip+1]), Size: 2}, true

	default:
		return Instruction{OpCode: OpCode, Argument: 0, Size: 1}, true
	}
}
