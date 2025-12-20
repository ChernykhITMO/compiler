package jit

import "github.com/ChernykhITMO/compiler/internal/bytecode"

type CodeReader struct {
	code []byte
	ip   int
}

func (r *CodeReader) NextInstruction() (Instruction, bool) {
	instr, ok := Decode(r.code, r.ip)
	if !ok {
		return Instruction{}, false
	}
	r.ip += instr.Size
	return instr, true
}

func (r *CodeReader) ExpectInstruction(opCode bytecode.OpCode) bool {
	instr, ok := r.NextInstruction()
	return ok && instr.OpCode == opCode
}

func (r *CodeReader) ExpectArgument(opCode bytecode.OpCode) (int, bool) {
	instr, ok := r.NextInstruction()
	if !ok || instr.OpCode != opCode {
		return 0, false
	}
	return instr.Argument, true
}
