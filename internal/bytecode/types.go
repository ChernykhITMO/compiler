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
	TypeArray
)

type ValueKind byte

const (
	ValInt ValueKind = iota
	ValFloat
	ValBool
	ValString
	ValChar
	ValNull
	ValObject
)

type ObjectType byte

const (
	ObjArray ObjectType = iota
)

type Object struct {
	Mark  bool
	Type  ObjectType
	Next  *Object // односвязный список всех объектов в куче
	Items []Value // для массивов: элементы
}

type Heap struct {
	Head       *Object // голова списка объектов
	NumObjects int     // сколько сейчас объектов
	MaxObjects int     // порог, при котором запускаем GC
}

type Value struct {
	Kind ValueKind
	I    int64
	F    float64
	B    bool
	S    string
	C    byte
	Obj  *Object
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

	OpArrayNew // выделить память под массив
	OpArrayGet // получит значение по индексу
	OpArraySet // присовить значение по индексу

	OpArraySwapJit

	OpPrint
)
