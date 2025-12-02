package bytecode

type FunctionInfo struct {
	Name       string
	ParamCount int
	ParamTypes []TypeKind
	ReturnType TypeKind

	Chunk     Chunk
	NumLocals int
}

type Module struct {
	Functions map[string]*FunctionInfo
}
