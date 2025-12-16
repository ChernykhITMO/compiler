package jit

import (
	"fmt"
	"unsafe"

	"github.com/ChernykhITMO/compiler/internal/backend/jit/infrastructure/arm"
	"github.com/ChernykhITMO/compiler/internal/bytecode"
)

type BasicBlock struct {
	EntryPoint uintptr
	HeaderIp   int
	TailIp     int
	Memory     *arm.ExecuteMemory
}

func CompileBasicBlockJitArm(fn *bytecode.FunctionInfo, headerIp int) (*BasicBlock, bool, error) {
	tailIp, isCompileInJit := AnalyzeNumericBasicBlockForJit(fn, headerIp)
	if !isCompileInJit {
		return nil, false, nil
	}

	ch := &fn.Chunk
	code := ch.Code

	var context arm.ContextVM
	offsetLocalsBase := int(unsafe.Offsetof(context.LocalsBase))
	offsetStackBase := int(unsafe.Offsetof(context.StackBase))
	offsetStackSize := int(unsafe.Offsetof(context.StackSize))
	offsetDidReturn := int(unsafe.Offsetof(context.DidReturn))

	var bytecodeVal bytecode.Value
	valueSize := int(unsafe.Sizeof(bytecodeVal))
	if valueSize%8 != 0 {
		return nil, false, fmt.Errorf("Value size must be multiple of 8, got %d", valueSize)
	}

	offKind := int(unsafe.Offsetof(bytecodeVal.Kind))
	offInt := int(unsafe.Offsetof(bytecodeVal.I))
	offBool := int(unsafe.Offsetof(bytecodeVal.B))

	memory, err := arm.AllocateMemoryMmap(4096)
	if err != nil {
		return nil, false, err
	}

	// X0 – указатель на ContextVM
	// X1 – localsBase context.LocalsBase
	// X2 – stackBase context.StackBase
	// W3 – stackSize context.StackSize
	// X4 – valueSize константа размера Value, чтобы считать адреса
	// X5,X6 – временные регистры для адресов, умножений
	// X7,X8,X9 – временные регистры для операндов или результатов
	// W10,W11 – временные 32-битные регистры (kind/bool/index)
	arm.EmitLoad64DataFromBase(memory, 1, 0, offsetLocalsBase)
	arm.EmitLoad64DataFromBase(memory, 2, 0, offsetStackBase)
	arm.EmitLoad32DataFromBase(memory, 3, 0, offsetStackSize)

	if valueSize > 0xFFFF {
		return nil, false, fmt.Errorf("Value size too big for imm16: %d", valueSize)
	}

	arm.EmitMove16ToX64(memory, 4, uint16(valueSize))

	computeAddress := func(baseX, indexW, outX uint8) {
		arm.EmitMultiplyRegisters64(memory, 5, indexW, 4) // X5 = X(indexW) * X4
		arm.EmitAddRegisters64(memory, outX, baseX, 5)    // outX = baseX + X5
	}
	clearValueSlot := func(addrX uint8) {
		for off := 0; off < valueSize; off += 8 {
			arm.EmitStore64DataToBase(memory, 31, addrX, off)
		}
	}

	pushInt16Values := func(imm16 uint16) {
		computeAddress(2, 3, 6) // X6 = stackBase + stackSize*valueSize
		clearValueSlot(6)

		arm.EmitMove16ToW32(memory, 10, uint16(bytecode.ValInt)) // W10 = ValInt
		arm.EmitStore8DataToBase(memory, 10, 6, offKind)         // slot.Kind = ValInt

		arm.EmitMove16ToX64(memory, 7, imm16)           // X7 = imm
		arm.EmitStore64DataToBase(memory, 7, 6, offInt) // slot.I = imm

		arm.EmitAdd32(memory, 3, 3, 1) // stackSize++
	}

	pushBoolFromW10 := func() {
		computeAddress(2, 3, 6) // X6 = адрес вершины стека
		clearValueSlot(6)

		arm.EmitMove16ToW32(memory, 11, uint16(bytecode.ValBool)) // W11 = ValBool
		arm.EmitStore8DataToBase(memory, 11, 6, offKind)          // slot.Kind = ValBool
		arm.EmitStore8DataToBase(memory, 10, 6, offBool)          // slot.B = W10 (0/1)

		arm.EmitAdd32(memory, 3, 3, 1) // stackSize++
	}

	readUint16 := func(ip *int) uint16 {
		hi := uint16(code[*ip])
		lo := uint16(code[*ip+1])
		*ip += 2
		return (hi << 8) | lo
	}

	emitPop2LoadIntsIntoX7X8 := func() {
		arm.EmitSub32(memory, 3, 3, 1)                   // stackSize--
		computeAddress(2, 3, 6)                          // X6 = addr(b)
		arm.EmitLoad64DataFromBase(memory, 7, 6, offInt) // X7 = b

		arm.EmitSub32(memory, 3, 3, 1)                   // stackSize--
		computeAddress(2, 3, 6)                          // X6 = addr(a)
		arm.EmitLoad64DataFromBase(memory, 8, 6, offInt) // X8 = a
	}

	emitWriteIntResultToTopAndInc := func(resultX uint8) {
		computeAddress(2, 3, 6) // X6 = addr(top)
		clearValueSlot(6)

		arm.EmitMove16ToW32(memory, 10, uint16(bytecode.ValInt)) // W10 = ValInt
		arm.EmitStore8DataToBase(memory, 10, 6, offKind)         // slot.Kind = ValInt

		arm.EmitStore64DataToBase(memory, resultX, 6, offInt) // slot.I = resultX
		arm.EmitAdd32(memory, 3, 3, 1)                        // stackSize++
	}

	ip := headerIp
	for {
		op := bytecode.OpCode(ch.Code[ip])
		ip++

		switch op {
		case bytecode.OpConst:
			index := int(readUint16(&ip))
			value := ch.Constants[index]

			if value.Kind != bytecode.ValInt {
				_ = memory.FreeMemoryMmap()
				return nil, false, nil
			}
			if value.I < 0 || value.I > 0xFFFF {
				_ = memory.FreeMemoryMmap()
				return nil, false, nil
			}

			pushInt16Values(uint16(value.I))

		case bytecode.OpLoadLocal:
			slot := uint16(code[ip])
			ip++

			computeAddress(2, 3, 6)                   // X6 = addr(stackTop)
			arm.EmitMove16ToW32(memory, 10, slot)     // W10 = slot
			computeAddress(1, 10, 7)                  // X7 = addr(local[slot])
			for off := 0; off < valueSize; off += 8 { // копируем весь Value
				arm.EmitLoad64DataFromBase(memory, 8, 7, off) // X8 = local + off
				arm.EmitStore64DataToBase(memory, 8, 6, off)  // stack + off = 8
			}
			arm.EmitAdd32(memory, 3, 3, 1) // stackSize++

		case bytecode.OpStoreLocal:
			slot := uint16(code[ip])
			ip++

			arm.EmitSub32(memory, 3, 3, 1)        // stackSize-- (pop)
			computeAddress(2, 3, 6)               // X6 = addr(pop value)
			arm.EmitMove16ToW32(memory, 10, slot) // W10 = slot
			computeAddress(1, 10, 7)              // X7 = addr(local[slot])

			for off := 0; off < valueSize; off += 8 {
				arm.EmitLoad64DataFromBase(memory, 8, 6, off) // X8 = stack+off
				arm.EmitStore64DataToBase(memory, 8, 7, off)  // local + off = X8
			}

		case bytecode.OpAdd, bytecode.OpSub, bytecode.OpMul:
			emitPop2LoadIntsIntoX7X8() // X7 = b, X8 = a

			switch op {
			case bytecode.OpAdd:
				arm.EmitAddRegisters64(memory, 9, 8, 7) // X9 = a + b
			case bytecode.OpSub:
				arm.EmitSubRegisters64(memory, 9, 8, 7) // X9 = a - b
			case bytecode.OpMul:
				arm.EmitMultiplyRegisters64(memory, 9, 8, 7) // X9 = a * b
			}
			emitWriteIntResultToTopAndInc(9) // push X9

		case bytecode.OpEq, bytecode.OpNe, bytecode.OpLt, bytecode.OpLe, bytecode.OpGt, bytecode.OpGe:
			emitPop2LoadIntsIntoX7X8() // X7 = b, X8 = a

			arm.EmitCompareRegisters64(memory, 8, 7) // выставляем флаги сравнения (a ? b)

			branchTrue := arm.EmitConditionalJump(memory, mapCompareToCondition(op)) // if true -> переход на true блок
			arm.EmitMove16ToW32(memory, 10, 0)                                       // W10 = 0 (false)
			branchEnd := arm.EmitUnconditionalJump(memory)

			truePos := memory.GetUsedByte()    // позиция true-block в машинном коде
			arm.EmitMove16ToW32(memory, 10, 1) // W10 = 1 (true)

			endPos := memory.GetUsedByte()
			arm.PatchConditional(memory, branchTrue, truePos, mapCompareToCondition(op))
			arm.PatchUnconditional(memory, branchEnd, endPos)

			pushBoolFromW10() // push bool(W10)

		case bytecode.OpPop:
			arm.EmitSub32(memory, 3, 3, 1) // stackSize--

		case bytecode.OpJump:
			target := readUint16(&ip)

			arm.EmitMove16ToW32(memory, 10, 0)                        // W10 = 0
			arm.EmitStore32DataToBase(memory, 10, 0, offsetDidReturn) // ctx.DidReturn = 0

			arm.EmitStore32DataToBase(memory, 3, 0, offsetStackSize) // ctx.StackSize = W3
			arm.EmitMove16ToW32(memory, 0, target)                   // W0 = nextIp
			arm.EmitReturn(memory)                                   // RET
			goto done

		case bytecode.OpJumpIfFalse:
			target := readUint16(&ip)      // куда перейти если false
			fallthroughValue := uint16(ip) // куда перейти если true

			arm.EmitSub32(memory, 10, 3, 1) // W10 = stackSize--
			computeAddress(2, 10, 6)        // X6 = addr(top-1)

			arm.EmitLoad8DataFromBase(memory, 11, 6, offBool) // W11 = top.B (0/1)

			// если W11 == 0 -> вернуть target, иначе вернуть fallthrough
			arm.EmitCompareRegisters64(memory, 11, 31)
			branchFalse := arm.EmitConditionalJump(memory, arm.ConditionEq) // EQ => false

			arm.EmitMove16ToW32(memory, 10, 0)
			arm.EmitStore32DataToBase(memory, 10, 0, offsetDidReturn) // ctx.DidReturn = 0
			arm.EmitStore32DataToBase(memory, 3, 0, offsetStackSize)  // ctx.StackSize = W3
			arm.EmitMove16ToW32(memory, 0, fallthroughValue)          // W0 = fallthrough
			arm.EmitReturn(memory)                                    // RET

			falsePos := memory.GetUsedByte() // false-block
			arm.EmitMove16ToW32(memory, 10, 0)
			arm.EmitStore32DataToBase(memory, 10, 0, offsetDidReturn) // ctx.DidReturn = 0
			arm.EmitStore32DataToBase(memory, 3, 0, offsetStackSize)  // ctx.StackSize = W3
			arm.EmitMove16ToW32(memory, 0, target)                    // W0 = target
			arm.EmitReturn(memory)                                    // RET

			arm.PatchConditional(memory, branchFalse, falsePos, arm.ConditionEq)
			goto done

		case bytecode.OpReturn:
			arm.EmitMove16ToW32(memory, 10, 1)                        // W10 = 1
			arm.EmitStore32DataToBase(memory, 10, 0, offsetDidReturn) // ctx.DidReturn = 1

			arm.EmitStore32DataToBase(memory, 3, 0, offsetStackSize) // ctx.StackSize = W3
			arm.EmitMove16ToW32(memory, 0, 0)                        // W0 = 0 (не важно)
			arm.EmitReturn(memory)                                   // RET
			goto done

		default:
			_ = memory.FreeMemoryMmap()
			return nil, false, nil

		}
	}

done:
	if err := memory.MakeReadExecute(); err != nil { // RW -> RX
		_ = memory.FreeMemoryMmap()
		return nil, false, err
	}

	return &BasicBlock{
		EntryPoint: memory.GetPtrBaseBuf(),
		HeaderIp:   headerIp,
		TailIp:     tailIp,
		Memory:     memory,
	}, true, nil
}

func mapCompareToCondition(op bytecode.OpCode) uint8 {
	switch op {
	case bytecode.OpEq:
		return arm.ConditionEq
	case bytecode.OpNe:
		return arm.ConditionNe
	case bytecode.OpLt:
		return arm.ConditionLt
	case bytecode.OpLe:
		return arm.ConditionLe
	case bytecode.OpGt:
		return arm.ConditionGt
	case bytecode.OpGe:
		return arm.ConditionGe
	default:
		return arm.ConditionEq
	}
}
