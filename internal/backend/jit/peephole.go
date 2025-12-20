package jit

import "github.com/ChernykhITMO/compiler/internal/bytecode"

type replacementCode struct {
	oldStartIP int
	oldEndIP   int
	newCode    []byte
}

func OptimizePeephole(fn *bytecode.FunctionInfo) {
	ch := &fn.Chunk
	code := ch.Code

	var reps []replacementCode
	for ip := 0; ip < len(code); {
		op := bytecode.OpCode(code[ip])
		size := OpCodeSizeByte(op)
		if size <= 0 || ip+size > len(code) {
			break
		}

		if ok, newCode, oldSpanBytes := matchBubbleSortSwap(code, ip); ok {
			reps = append(reps, replacementCode{
				oldStartIP: ip,
				oldEndIP:   ip + oldSpanBytes,
				newCode:    newCode,
			})
			ip += oldSpanBytes
			continue
		}

		ip += size
	}

	if len(reps) == 0 {
		return
	}

	// mapping old ip -> new ip
	oldToNewIPMap := make(map[int]int, 1024)
	replaceIndex := 0
	newIP := 0
	for ip := 0; ip < len(code); {
		if replaceIndex < len(reps) && ip == reps[replaceIndex].oldStartIP {
			oldToNewIPMap[ip] = newIP
			newIP += len(reps[replaceIndex].newCode)
			ip = reps[replaceIndex].oldEndIP
			replaceIndex++
			continue
		}

		op := bytecode.OpCode(code[ip])
		size := OpCodeSizeByte(op)
		if size <= 0 || ip+size > len(code) {
			break
		}
		oldToNewIPMap[ip] = newIP
		newIP += size
		ip += size
	}

	// собираем новый код и переписываем jump target
	out := make([]byte, 0, newIP)
	replaceIndex = 0
	for ip := 0; ip < len(code); {
		if replaceIndex < len(reps) && ip == reps[replaceIndex].oldStartIP {
			out = append(out, reps[replaceIndex].newCode...)
			ip = reps[replaceIndex].oldEndIP
			replaceIndex++
			continue
		}

		op := bytecode.OpCode(code[ip])
		ip++

		out = append(out, byte(op))

		switch op {
		case bytecode.OpConst, bytecode.OpCall:
			if ip+1 >= len(code) {
				ch.Code = out
				return
			}
			out = append(out, code[ip], code[ip+1])
			ip += 2

		case bytecode.OpJump, bytecode.OpJumpIfFalse:
			if ip+1 >= len(code) {
				ch.Code = out
				return
			}
			oldTargetIP := int(uint16(code[ip])<<8 | uint16(code[ip+1]))
			ip += 2

			newTargetIP, ok := oldToNewIPMap[oldTargetIP]
			if !ok {
				ch.Code = code
				return
			}

			out = append(out, byte(uint16(newTargetIP)>>8), byte(uint16(newTargetIP)))

		case bytecode.OpLoadLocal, bytecode.OpStoreLocal:
			if ip >= len(code) {
				ch.Code = out
				return
			}
			out = append(out, code[ip])
			ip++
		}
	}

	ch.Code = out
}

// просмотр байткода и проверка паттерна = условно регулярное выражение
func matchBubbleSortSwap(code []byte, start int) (bool, []byte, int) {
	r := CodeReader{code: code, ip: start}

	// arr[j]
	arrSlot, ok := r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok {
		return false, nil, 0
	}
	jSlot, ok := r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || !r.ExpectInstruction(bytecode.OpArrayGet) {
		return false, nil, 0
	}

	// arr[j+1]
	s, ok := r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != arrSlot {
		return false, nil, 0
	}
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != jSlot ||
		!r.ExpectInstruction(bytecode.OpConst) ||
		!r.ExpectInstruction(bytecode.OpAdd) ||
		!r.ExpectInstruction(bytecode.OpArrayGet) {
		return false, nil, 0
	}

	// arr[j] > arr[j+1]
	if !r.ExpectInstruction(bytecode.OpGt) {
		return false, nil, 0
	}
	skipIP, ok := r.ExpectArgument(bytecode.OpJumpIfFalse)
	if !ok || !r.ExpectInstruction(bytecode.OpPop) {
		return false, nil, 0
	}

	// tmp = arr[j]
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != arrSlot {
		return false, nil, 0
	}
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != jSlot || !r.ExpectInstruction(bytecode.OpArrayGet) {
		return false, nil, 0
	}
	tmpSlot, ok := r.ExpectArgument(bytecode.OpStoreLocal)
	if !ok {
		return false, nil, 0
	}

	// arr[j] = arr[j+1]
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != arrSlot {
		return false, nil, 0
	}
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != jSlot {
		return false, nil, 0
	}
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != arrSlot {
		return false, nil, 0
	}
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != jSlot ||
		!r.ExpectInstruction(bytecode.OpConst) ||
		!r.ExpectInstruction(bytecode.OpAdd) ||
		!r.ExpectInstruction(bytecode.OpArrayGet) ||
		!r.ExpectInstruction(bytecode.OpArraySet) {
		return false, nil, 0
	}

	// arr[j+1] = tmp
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != arrSlot {
		return false, nil, 0
	}
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != jSlot || !r.ExpectInstruction(bytecode.OpConst) || !r.ExpectInstruction(bytecode.OpAdd) {
		return false, nil, 0
	}
	s, ok = r.ExpectArgument(bytecode.OpLoadLocal)
	if !ok || s != tmpSlot || !r.ExpectInstruction(bytecode.OpArraySet) {
		return false, nil, 0
	}

	// Jump end
	endIP, ok := r.ExpectArgument(bytecode.OpJump)
	if !ok {
		return false, nil, 0
	}

	// Проверка структуры if
	if skipIP < 0 || skipIP >= len(code) || bytecode.OpCode(code[skipIP]) != bytecode.OpPop || endIP != skipIP+1 {
		return false, nil, 0
	}

	newCode := []byte{
		byte(bytecode.OpLoadLocal), byte(arrSlot),
		byte(bytecode.OpLoadLocal), byte(jSlot),
		byte(bytecode.OpArraySwapJit),
	}

	span := endIP - start
	if span <= 0 {
		return false, nil, 0
	}
	return true, newCode, span
}
