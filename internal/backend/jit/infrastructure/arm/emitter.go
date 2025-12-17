package arm

const (
	ConditionEq = 0  // =
	ConditionNe = 1  // !=
	ConditionLt = 11 // <
	ConditionLe = 13 // <=
	ConditionGt = 12 // >
	ConditionGe = 10 // >=
)

func EmitReturn(mem *ExecuteMemory) {
	mem.WriteUint32JitInstruction(0xD65F03C0) // RET X30
}

func EmitMove16ToX64(mem *ExecuteMemory, destinationX64 uint8, imm16 uint16) {
	mem.WriteUint32JitInstruction(0xD2800000 | (uint32(imm16) << 5) | uint32(destinationX64))
}
func EmitMove16ToW32(mem *ExecuteMemory, destinationW32 uint8, imm16 uint16) {
	mem.WriteUint32JitInstruction(0x52800000 | (uint32(imm16) << 5) | uint32(destinationW32))
}

// LDR Xt, [Xn, #offsetBytes]
func EmitLoad64DataFromBase(mem *ExecuteMemory, destinationX64, baseX64 uint8, offsetBytes int) {
	imm := offsetBytes / 8
	mem.WriteUint32JitInstruction(0xF9400000 | (uint32(imm) << 10) | (uint32(baseX64) << 5) | uint32(destinationX64))
}

// LDR Wt, [Xn, #offsetBytes]
func EmitLoad32DataFromBase(mem *ExecuteMemory, destinationW, baseX64 uint8, offsetBytes int) {
	imm := offsetBytes / 4
	mem.WriteUint32JitInstruction(0xB9400000 | (uint32(imm) << 10) | (uint32(baseX64) << 5) | uint32(destinationW))
}

// LDRB Wt, [Xn, #offsetBytes]
func EmitLoad8DataFromBase(mem *ExecuteMemory, destinationW, baseX64 uint8, offsetBytes int) {
	mem.WriteUint32JitInstruction(0x39400000 | (uint32(offsetBytes) << 10) | (uint32(baseX64) << 5) | uint32(destinationW))
}

// STR Xt, [Xn, #offsetBytes]
func EmitStore64DataToBase(mem *ExecuteMemory, sourceX64, baseX64 uint8, offsetBytes int) {
	imm := offsetBytes / 8
	mem.WriteUint32JitInstruction(0xF9000000 | (uint32(imm) << 10) | (uint32(baseX64) << 5) | uint32(sourceX64))
}

// STR Wt, [Xn, #offsetBytes]
func EmitStore32DataToBase(mem *ExecuteMemory, sourceW32, baseX64 uint8, offsetBytes int) {
	imm := offsetBytes / 4
	mem.WriteUint32JitInstruction(0xB9000000 | (uint32(imm) << 10) | (uint32(baseX64) << 5) | uint32(sourceW32))
}

// STRB Wt, [Xn, #offsetBytes]
func EmitStore8DataToBase(mem *ExecuteMemory, sourceW32, baseX64 uint8, offsetBytes int) {
	mem.WriteUint32JitInstruction(0x39000000 | (uint32(offsetBytes) << 10) | (uint32(baseX64) << 5) | uint32(sourceW32))
}

// ADDW $imm12, Wn -> Wd
func EmitAdd32(mem *ExecuteMemory, destinationW, baseW32 uint8, imm12 uint16) {
	mem.WriteUint32JitInstruction(0x11000000 | (uint32(imm12) << 10) | (uint32(baseW32) << 5) | uint32(destinationW))
}

// ADD Xm + Xn -> Xd
func EmitAddRegisters64(mem *ExecuteMemory, destinationX64, leftX64, rightX64 uint8) {
	mem.WriteUint32JitInstruction(0x8B000000 | (uint32(rightX64) << 16) | (uint32(leftX64) << 5) | uint32(destinationX64))
}

// SUBW $imm12, Wn -> Wd
func EmitSub32(mem *ExecuteMemory, destinationW, baseW32 uint8, imm12 uint16) {
	mem.WriteUint32JitInstruction(0x51000000 | (uint32(imm12) << 10) | (uint32(baseW32) << 5) | uint32(destinationW))
}

// SUB Xm from Xn -> Xd (Xd = Xn - Xm)
func EmitSubRegisters64(mem *ExecuteMemory, destinationX64, leftX64, rightX64 uint8) {
	mem.WriteUint32JitInstruction(0xCB000000 | (uint32(rightX64) << 16) | (uint32(leftX64) << 5) | uint32(destinationX64))
}

func EmitMultiplyRegisters64(mem *ExecuteMemory, destinationX64, leftX64, rightX64 uint8) {
	mem.WriteUint32JitInstruction(0x9B007C00 | (uint32(rightX64) << 16) | (uint32(leftX64) << 5) | uint32(destinationX64))
}

// CMP Xm, Xn (flag Xn - Xm)
func EmitCompareRegisters64(mem *ExecuteMemory, leftX64, rightX64 uint8) {
	mem.WriteUint32JitInstruction(0xEB00001F | (uint32(rightX64) << 16) | (uint32(leftX64) << 5))
}

// B.cond placeholder -> потом PatchConditional(...)
func EmitConditionalJump(mem *ExecuteMemory, condition uint8) int {
	bytePos := mem.GetUsedByte()
	mem.WriteUint32JitInstruction(0x54000000 | uint32(condition)) // imm19=0 пока
	return bytePos
}

func PatchConditional(mem *ExecuteMemory, branchBytePos int, targetBytePos int, condition uint8) {
	imm := (targetBytePos - branchBytePos) / 4
	instr := uint32(0x54000000) | (uint32(imm&0x7FFFF) << 5) | uint32(condition)
	mem.PatchUint32At(branchBytePos, instr)
}

// placeholder -> потом PatchUnconditional(...)
func EmitUnconditionalJump(mem *ExecuteMemory) int {
	bytePos := mem.GetUsedByte()
	mem.WriteUint32JitInstruction(0x14000000) // imm26=0 пока
	return bytePos
}

func PatchUnconditional(mem *ExecuteMemory, branchBytePos int, targetBytePos int) {
	imm := (targetBytePos - branchBytePos) / 4
	instr := uint32(0x14000000) | uint32(imm&0x03FFFFFF)
	mem.PatchUint32At(branchBytePos, instr)
}
