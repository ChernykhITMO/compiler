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

	readUint16 := func() (uint16, bool) {
		if ip+1 >= len(code) {
			return 0, false
		}
		hi := uint16(code[ip])
		lo := uint16(code[ip+1])
		ip += 2
		return (hi << 8) | lo, true
	}
	readByte := func() (byte, bool) {
		if ip == len(code) {
			return 0, false
		}
		b := code[ip]
		ip++
		return b, true
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
			idx, ok := readUint16()
			if !ok {
				return 0, false
			}
			if int(idx) >= len(ch.Constants) {
				return 0, false
			}
			if ch.Constants[idx].Kind != bytecode.ValInt {
				return 0, false
			}

		case bytecode.OpLoadLocal, bytecode.OpStoreLocal:
			_, ok := readByte()
			if !ok {
				return 0, false
			}

		case bytecode.OpAdd, bytecode.OpSub, bytecode.OpMul,
			bytecode.OpEq, bytecode.OpNe, bytecode.OpLt, bytecode.OpLe, bytecode.OpGt, bytecode.OpGe,
			bytecode.OpPop:
			// ok

		// Терминаторы = разрешение на переход к jit, ели не встретиилось неудовл команды
		case bytecode.OpJump, bytecode.OpJumpIfFalse, bytecode.OpCall:
			_, ok := readUint16()
			if !ok {
				return 0, false
			}
			return ip, true

		case bytecode.OpReturn, bytecode.OpArrayNew, bytecode.OpArrayGet, bytecode.OpArraySet:
			return ip, true

		case bytecode.OpDiv, bytecode.OpMod, bytecode.OpPow, bytecode.OpNeg, bytecode.OpNot:
			return 0, false
			
		default:
			return 0, false
		}
	}
}
