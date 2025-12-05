package bytecode

type TypeKind byte

const (
	TypeInvalid TypeKind = iota //инвалидный тип
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
	OpConst      OpCode = iota // положить константу на стек
	OpLoadLocal                //загрузить локальную переменную на стек
	OpStoreLocal               //сохранить в локальную переменную вершину стека
	OpAdd                      // сложение
	OpSub                      // вычитание
	OpMul                      // умножение
	OpDiv                      // деление
	OpMod                      // остаток
	OpPow                      // вовзедение в степень

	OpEq // =
	OpNe // !=
	OpLt // <
	OpLe // <=
	OpGt // >
	OpGe // >=

	OpNeg // -
	OpNot // !

	OpJump        // безусловный переход по адрессу
	OpJumpIfFalse // переход если вершина стека false
	OpPop         // удаление вершины со стека

	OpCall   // вызов функции
	OpReturn // вернуть из функции
)
