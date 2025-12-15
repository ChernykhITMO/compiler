package jit

import (
	"github.com/ChernykhITMO/compiler/internal/bytecode"
)

func AnalyzeNumericBasicBlockForJit(fn *bytecode.FunctionInfo, headerIp int) (tailIp int, isCompileInJit bool) {
	ch := &fn.Chunk
	code := ch.Code

	if headerIp < 0 || headerIp >= len(code) {
		return 0, false
	}

	ip := headerIp

	readUint16 := func() uint16 {
		if ip+1 >= len(code) {
			return 0
		}
		hi := uint16(code[ip])
		lo := uint16(code[ip+1])
		ip += 2
		return (hi << 8) | lo
	}
	readByte := func() byte {
		b := code[ip]
		ip++
		return b
	}

	for {
		if ip >= len(code) {
			return 0, false
		}

		op := bytecode.OpCode(code[ip])
		ip++

		switch op {
		// разрешенные операции для jit, только int
		case bytecode.OpConst:
			idx := readUint16()
			if int(idx) >= len(ch.Constants) {
				return 0, false
			}
			if ch.Constants[idx].Kind != bytecode.ValInt {
				return 0, false
			}

		case bytecode.OpLoadLocal, bytecode.OpStoreLocal:
			_ = readByte()

		case bytecode.OpAdd, bytecode.OpSub, bytecode.OpMul:
			// ok

		case bytecode.OpEq, bytecode.OpNe, bytecode.OpLt, bytecode.OpLe, bytecode.OpGt, bytecode.OpGe:
			// ok

		case bytecode.OpPop:
			// ok

		// Терминаторы = заканчивает блок
		case bytecode.OpJump:
			_ = readUint16()
			return ip, true

		case bytecode.OpJumpIfFalse:
			_ = readUint16()
			return ip, true

		case bytecode.OpReturn:
			return ip, true

		// запрещенные операции для jit
		case bytecode.OpCall, bytecode.OpArrayNew, bytecode.OpArrayGet, bytecode.OpArraySet,
			bytecode.OpDiv, bytecode.OpMod, bytecode.OpPow, bytecode.OpNeg, bytecode.OpNot:
			return 0, false

		default:
			return 0, false
		}
	}
}
