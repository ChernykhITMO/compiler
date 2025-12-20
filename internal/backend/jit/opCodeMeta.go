package jit

import "github.com/ChernykhITMO/compiler/internal/bytecode"

func OpCodeSizeByte(op bytecode.OpCode) int {
	switch op {
	case bytecode.OpConst, bytecode.OpJump, bytecode.OpJumpIfFalse, bytecode.OpCall:
		return 1 + 2
	case bytecode.OpLoadLocal, bytecode.OpStoreLocal:
		return 1 + 1
	default:
		return 1
	}
}
