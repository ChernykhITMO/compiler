package bytecode

type TypeKind byte

const (
	TypeInvalid TypeKind = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeString
	TypeChar
	TypeVoid
	TypeNull
)

type ValueKind byte

const (
	ValInt ValueKind = iota
	ValFloat
	ValBool
	ValString
	ValChar
	ValNull
)

type Value struct {
	Kind ValueKind
	I    int64
	F    float64
	B    bool
	S    string
	C    byte
}

type OpCode byte

const (
	OpConst OpCode = iota
	OpLoadLocal
	OpStoreLocal
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpPow

	OpEq
	OpNe
	OpLt
	OpLe
	OpGt
	OpGe

	OpNeg
	OpNot

	OpJump
	OpJumpIfFalse
	OpPop

	OpCall
	OpReturn
)
