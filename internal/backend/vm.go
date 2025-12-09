package backend

import (
	"fmt"
	"github.com/ChernykhITMO/compiler/internal/bytecode"
	"math"
)

type VM struct {
	mod *bytecode.Module
}

func NewVM(mod *bytecode.Module) *VM {
	return &VM{mod: mod}
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

	stack := make([]bytecode.Value, 0, 256)

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
		stack = append(stack, v)
	}

	for {
		if ip >= len(ch.Code) {
			return bytecode.Value{Kind: bytecode.ValNull}, nil
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

		default:
			return bytecode.Value{}, fmt.Errorf("unknown opcode %d", op)
		}
	}
}

// TODO подумать над тем что делаем только bool или что то еще
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
