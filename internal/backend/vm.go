package backend

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/ChernykhITMO/compiler/internal/backend/jit"
	"github.com/ChernykhITMO/compiler/internal/backend/jit/infrastructure/arm"
	"github.com/ChernykhITMO/compiler/internal/bytecode"
)

type rootSet struct {
	locals *[]bytecode.Value
	stack  *[]bytecode.Value
}

type jitKey struct {
	fn *bytecode.FunctionInfo
	ip int
}

type VM struct {
	mod   *bytecode.Module
	heap  bytecode.Heap
	roots []rootSet

	jitBlocks     map[jitKey]*jit.BasicBlock
	jitHotCounter map[jitKey]uint32

	// Чтобы не пытаться компилить один и тот же блок бесконечно при “не компилится”.
	jitTried map[jitKey]bool

	jitThreshold uint32
}

func NewVM(mod *bytecode.Module) *VM {
	return &VM{
		mod:           mod,
		jitBlocks:     make(map[jitKey]*jit.BasicBlock),
		jitHotCounter: make(map[jitKey]uint32),
		jitTried:      make(map[jitKey]bool),
		jitThreshold:  400,
	}
}

func (vm *VM) Call(name string, args []bytecode.Value) (bytecode.Value, error) {
	fn, ok := vm.mod.Functions[name]
	if !ok {
		return bytecode.Value{}, fmt.Errorf("unknown function %q", name)
	}
	if len(args) != fn.ParamCount {
		return bytecode.Value{}, fmt.Errorf("function %q: expected %d args, got %d",
			name, fn.ParamCount, len(args))
	}
	return vm.runFunction(fn, args)
}

func (vm *VM) runFunction(fn *bytecode.FunctionInfo, args []bytecode.Value) (bytecode.Value, error) {
	ch := &fn.Chunk

	locals := make([]bytecode.Value, fn.NumLocals)
	copy(locals, args)

	const maxStackCapacity = 4096
	stackBacking := make([]bytecode.Value, maxStackCapacity)
	stack := stackBacking[:0]

	vm.roots = append(vm.roots, rootSet{
		locals: &locals,
		stack:  &stack,
	})
	defer func() {
		vm.roots = vm.roots[:len(vm.roots)-1]
	}()

	ip := 0

	readUint16 := func() uint16 {
		hi := uint16(ch.Code[ip])
		lo := uint16(ch.Code[ip+1])
		ip += 2
		return (hi << 8) | lo
	}

	pop := func() bytecode.Value {
		if len(stack) == 0 {
			panic("stack underflow")
		}
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return v
	}

	push := func(v bytecode.Value) {
		if len(stack) >= len(stackBacking) {
			panic("stack overflow")
		}
		stack = stack[:len(stack)+1]
		stack[len(stack)-1] = v
	}

	for {
		if ip >= len(ch.Code) {
			return bytecode.Value{Kind: bytecode.ValNull}, nil
		}

		if jitBlocks := vm.jitBlocks[jitKey{fn: fn, ip: ip}]; jitBlocks != nil {
			var ctx arm.ContextVM
			if len(locals) > 0 {
				ctx.LocalsBase = unsafe.Pointer(&locals[0])
			}

			ctx.StackBase = unsafe.Pointer(&stackBacking[0])
			ctx.StackSize = uint32(len(stack))
			ctx.DidReturn = 0
			nextIP := arm.CallJitBlock(jitBlocks.EntryPoint, &ctx)
			stack = stackBacking[:ctx.StackSize]

			if ctx.DidReturn != 0 {
				if len(stack) == 0 {
					return bytecode.Value{Kind: bytecode.ValNull}, nil
				}
				return stack[len(stack)-1], nil
			}

			if int(nextIP) < 0 || int(nextIP) > len(ch.Code) {
				return bytecode.Value{}, fmt.Errorf("jit: bad nextIP=%d", nextIP)
			}
			ip = int(nextIP)

			continue
		}

		op := bytecode.OpCode(ch.Code[ip])
		ip++

		switch op {
		case bytecode.OpConst:
			idx := readUint16()
			if int(idx) >= len(ch.Constants) {
				return bytecode.Value{}, fmt.Errorf("const index out of range: %d", idx)
			}
			push(ch.Constants[idx])

		case bytecode.OpLoadLocal:
			slot := int(ch.Code[ip])
			ip++
			if slot < 0 || slot >= len(locals) {
				return bytecode.Value{}, fmt.Errorf("load local: bad slot %d", slot)
			}
			push(locals[slot])

		case bytecode.OpStoreLocal:
			slot := int(ch.Code[ip])
			ip++
			if slot < 0 || slot >= len(locals) {
				return bytecode.Value{}, fmt.Errorf("store local: bad slot %d", slot)
			}
			v := pop()
			locals[slot] = v

		case bytecode.OpAdd:
			b := pop()
			a := pop()
			res, err := vm.binaryNumberOp("+", a, b)
			if err != nil {
				return bytecode.Value{}, err
			}
			push(res)

		case bytecode.OpSub:
			b := pop()
			a := pop()
			res, err := vm.binaryNumberOp("-", a, b)
			if err != nil {
				return bytecode.Value{}, err
			}
			push(res)

		case bytecode.OpMul:
			b := pop()
			a := pop()
			res, err := vm.binaryNumberOp("*", a, b)
			if err != nil {
				return bytecode.Value{}, err
			}
			push(res)

		case bytecode.OpDiv:
			b := pop()
			a := pop()
			res, err := vm.binaryNumberOp("/", a, b)
			if err != nil {
				return bytecode.Value{}, err
			}
			push(res)

		case bytecode.OpMod:
			b := pop()
			a := pop()
			res, err := vm.binaryNumberOp("%", a, b)
			if err != nil {
				return bytecode.Value{}, err
			}
			push(res)

		case bytecode.OpPow:
			b := pop()
			a := pop()
			res, err := vm.binaryNumberOp("^", a, b)
			if err != nil {
				return bytecode.Value{}, err
			}
			push(res)

		case bytecode.OpEq:
			b := pop()
			a := pop()
			push(boolValue(vm.equal(a, b)))

		case bytecode.OpNe:
			b := pop()
			a := pop()
			push(boolValue(!vm.equal(a, b)))

		case bytecode.OpLt, bytecode.OpLe, bytecode.OpGt, bytecode.OpGe:
			b := pop()
			a := pop()
			res, err := vm.compareNumbers(op, a, b)
			if err != nil {
				return bytecode.Value{}, err
			}
			push(boolValue(res))

		case bytecode.OpNeg:
			v := pop()
			if v.Kind != bytecode.ValFloat && v.Kind != bytecode.ValInt {
				return bytecode.Value{}, fmt.Errorf("unary - on non-number")
			}
			if v.Kind == bytecode.ValFloat {
				v.F = -v.F
			} else {
				v.I = -v.I
			}
			push(v)

		case bytecode.OpNot:
			v := pop()
			push(boolValue(!vm.isTruthy(v)))

		case bytecode.OpJump:
			target := int(readUint16())
			if target < 0 || target > len(ch.Code) {
				return bytecode.Value{}, fmt.Errorf("jump: bad target %d", target)
			}

			if target < ip {
				key := jitKey{fn: fn, ip: target}

				if !vm.jitTried[key] {
					vm.jitHotCounter[key]++

					if vm.jitHotCounter[key] >= vm.jitThreshold {
						vm.jitTried[key] = true

						jitBlock, ok, err := jit.CompileBasicBlockJitArm(fn, target)
						if err != nil {
							return bytecode.Value{}, err
						}
						if ok && jitBlock != nil {
							vm.jitBlocks[key] = jitBlock
						}
					}
				}
			}
			ip = target

		case bytecode.OpJumpIfFalse:
			target := int(readUint16())
			top := stack[len(stack)-1]
			if !vm.isTruthy(top) {
				if target < 0 || target > len(ch.Code) {
					return bytecode.Value{}, fmt.Errorf("jump-if-false: bad target %d", target)
				}
				ip = target
			}

		case bytecode.OpPop:
			_ = pop()

		case bytecode.OpCall:
			idx := readUint16()
			if int(idx) >= len(ch.Constants) {
				return bytecode.Value{}, fmt.Errorf("call: const index out of range %d", idx)
			}
			constVal := ch.Constants[idx]
			if constVal.Kind != bytecode.ValString {
				return bytecode.Value{}, fmt.Errorf("call: const is not string (function name)")
			}
			calleeName := constVal.S
			callee, ok := vm.mod.Functions[calleeName]
			if !ok {
				return bytecode.Value{}, fmt.Errorf("unknown function %q", calleeName)
			}

			n := callee.ParamCount
			if len(stack) < n {
				return bytecode.Value{}, fmt.Errorf("call %q: stack has %d values, want %d args",
					calleeName, len(stack), n)
			}

			argsVals := make([]bytecode.Value, n)
			copy(argsVals, stack[len(stack)-n:])
			stack = stack[:len(stack)-n]

			ret, err := vm.runFunction(callee, argsVals)
			if err != nil {
				return bytecode.Value{}, err
			}
			push(ret)

		case bytecode.OpReturn:
			if len(stack) == 0 {
				return bytecode.Value{Kind: bytecode.ValNull}, nil
			}
			return stack[len(stack)-1], nil

		case bytecode.OpArrayNew:
			lenVal := pop()
			if lenVal.Kind != bytecode.ValInt {
				return bytecode.Value{}, fmt.Errorf("array new: length must be int")
			}
			if lenVal.I < 0 {
				return bytecode.Value{}, fmt.Errorf("array new: length must be >= 0")
			}
			n := int(lenVal.I)

			obj := vm.newObject(bytecode.ObjArray)
			obj.Items = make([]bytecode.Value, n)

			push(bytecode.Value{
				Kind: bytecode.ValObject,
				Obj:  obj,
			})

		case bytecode.OpArrayGet:
			idxVal := pop()
			arrVal := pop()

			if arrVal.Kind != bytecode.ValObject || arrVal.Obj == nil || arrVal.Obj.Type != bytecode.ObjArray {
				return bytecode.Value{}, fmt.Errorf("array get: value is not array")
			}
			if idxVal.Kind != bytecode.ValInt {
				return bytecode.Value{}, fmt.Errorf("array get: index must be int")
			}
			idx := int(idxVal.I)
			if idx < 0 || idx >= len(arrVal.Obj.Items) {
				return bytecode.Value{}, fmt.Errorf("array get: index %d out of range [0,%d)", idx, len(arrVal.Obj.Items))
			}

			push(arrVal.Obj.Items[idx])

		case bytecode.OpArraySet:
			val := pop()
			idxVal := pop()
			arrVal := pop()

			if arrVal.Kind != bytecode.ValObject || arrVal.Obj == nil || arrVal.Obj.Type != bytecode.ObjArray {
				return bytecode.Value{}, fmt.Errorf("array set: value is not array")
			}
			if idxVal.Kind != bytecode.ValInt {
				return bytecode.Value{}, fmt.Errorf("array set: index must be int")
			}
			idx := int(idxVal.I)
			if idx < 0 || idx >= len(arrVal.Obj.Items) {
				return bytecode.Value{}, fmt.Errorf("array set: index %d out of range [0,%d)", idx, len(arrVal.Obj.Items))
			}

			arrVal.Obj.Items[idx] = val

		default:
			return bytecode.Value{}, fmt.Errorf("unknown opcode %d", op)
		}
	}
}

func (vm *VM) isTruthy(v bytecode.Value) bool {
	switch v.Kind {
	case bytecode.ValBool:
		return v.B
	default:
		panic("non-bool used in boolean context")
	}
}

func (vm *VM) equal(a, b bytecode.Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case bytecode.ValNull:
		return true
	case bytecode.ValBool:
		return a.B == b.B
	case bytecode.ValInt:
		return a.I == b.I
	case bytecode.ValFloat:
		return a.F == b.F
	case bytecode.ValString:
		return a.S == b.S
	case bytecode.ValChar:
		return a.C == b.C
	default:
		return false
	}
}

func boolValue(b bool) bytecode.Value {
	return bytecode.Value{Kind: bytecode.ValBool, B: b}
}

func (vm *VM) binaryNumberOp(op string, a, b bytecode.Value) (bytecode.Value, error) {
	if a.Kind != b.Kind {
		return bytecode.Value{}, fmt.Errorf("numeric op %s: mixed types %v and %v", op, a.Kind, b.Kind)
	}

	switch a.Kind {
	case bytecode.ValInt:
		v, err := intOp(op, a.I, b.I)
		if err != nil {
			return bytecode.Value{}, err
		}
		return bytecode.Value{Kind: bytecode.ValInt, I: v}, nil

	case bytecode.ValFloat:
		v, err := floatOp(op, a.F, b.F)
		if err != nil {
			return bytecode.Value{}, err
		}
		return bytecode.Value{Kind: bytecode.ValFloat, F: v}, nil

	default:
		return bytecode.Value{}, fmt.Errorf("numeric op %s: unsupported kind %v", op, a.Kind)
	}
}

func intOp(op string, a, b int64) (int64, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return a / b, nil
	case "%":
		if b == 0 {
			return 0, fmt.Errorf("modulo by zero")
		}
		return a % b, nil
	case "^":
		return int64(math.Pow(float64(a), float64(b))), nil
	default:
		return 0, fmt.Errorf("unknown int op %q", op)
	}
}

func floatOp(op string, a, b float64) (float64, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		return a / b, nil
	case "%":
		return math.Mod(a, b), nil
	case "^":
		return math.Pow(a, b), nil
	default:
		return 0, fmt.Errorf("unknown float op %q", op)
	}
}

func (vm *VM) compareNumbers(op bytecode.OpCode, a, b bytecode.Value) (bool, error) {
	if a.Kind != b.Kind {
		return false, fmt.Errorf("compare: mixed types %v and %v", a.Kind, b.Kind)
	}

	switch a.Kind {
	case bytecode.ValInt:
		return compareInt(op, a.I, b.I)

	case bytecode.ValFloat:
		return compareFloat(op, a.F, b.F)

	default:
		return false, fmt.Errorf("compare: not a number")
	}
}

func compareInt(op bytecode.OpCode, a, b int64) (bool, error) {
	switch op {
	case bytecode.OpLt:
		return a < b, nil
	case bytecode.OpLe:
		return a <= b, nil
	case bytecode.OpGt:
		return a > b, nil
	case bytecode.OpGe:
		return a >= b, nil
	default:
		return false, fmt.Errorf("unknown compare op %d", op)
	}
}

func compareFloat(op bytecode.OpCode, a, b float64) (bool, error) {
	switch op {
	case bytecode.OpLt:
		return a < b, nil
	case bytecode.OpLe:
		return a <= b, nil
	case bytecode.OpGt:
		return a > b, nil
	case bytecode.OpGe:
		return a >= b, nil
	default:
		return false, fmt.Errorf("unknown compare op %d", op)
	}
}
