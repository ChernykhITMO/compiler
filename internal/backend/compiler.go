package backend

import (
	"fmt"
	"github.com/ChernykhITMO/compiler/internal/bytecode"
	"github.com/ChernykhITMO/compiler/internal/frontend"
)

type localVar struct {
	name string
	slot int
	typ  bytecode.TypeKind
}

type Compiler struct {
	mod    *bytecode.Module
	fn     *bytecode.FunctionInfo
	locals []localVar
}

func NewCompiler() *Compiler {
	functions := make(map[string]*bytecode.FunctionInfo)
	module := &bytecode.Module{Functions: functions}

	return &Compiler{mod: module}
}

func (c *Compiler) chunk() *bytecode.Chunk {
	return &c.fn.Chunk
}

func (c *Compiler) Module() *bytecode.Module {
	return c.mod
}

func mapTypeName(t string) bytecode.TypeKind {
	switch t {
	case "int":
		return bytecode.TypeInt
	case "float":
		return bytecode.TypeFloat
	case "string":
		return bytecode.TypeString
	case "bool":
		return bytecode.TypeBool
	case "char":
		return bytecode.TypeChar
	case "void":
		return bytecode.TypeVoid
	default:
		return bytecode.TypeInvalid
	}
}

func (c *Compiler) addLocal(name string, typ bytecode.TypeKind) int {
	slot := len(c.locals)
	c.locals = append(c.locals, localVar{name: name, slot: slot, typ: typ})

	if c.fn != nil && slot+1 > c.fn.NumLocals {
		c.fn.NumLocals = slot + 1
	}
	return slot
}

func (c *Compiler) resolveLocal(name string) (int, bool) {
	for i := len(c.locals) - 1; i >= 0; i-- {
		if c.locals[i].name == name {
			return c.locals[i].slot, true
		}
	}
	return 0, false
}

func (c *Compiler) CompileProgram(p *frontend.Program) (*bytecode.Module, error) {
	for _, fn := range p.Functions {
		if err := c.compileFunction(fn); err != nil {
			return nil, err
		}
	}
	return c.mod, nil
}

func (c *Compiler) compileFunction(fn *frontend.FunctionDecl) error {
	bfn := &bytecode.FunctionInfo{
		Name:       fn.Name,
		ParamCount: len(fn.Params),
		ParamTypes: make([]bytecode.TypeKind, len(fn.Params)),
		ReturnType: mapTypeName(fn.ReturnType),
	}

	for i, p := range fn.Params {
		bfn.ParamTypes[i] = mapTypeName(p.TypeName)
	}

	c.fn = bfn
	c.locals = nil

	for _, p := range fn.Params {
		c.addLocal(p.Name, mapTypeName(p.TypeName))
	}

	c.compileBlock(fn.Body)

	ch := c.chunk()
	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValNull})
	ch.WriteUint16(uint16(idx), 0)
	ch.Write(bytecode.OpReturn, 0)

	if c.mod.Functions == nil {
		c.mod.Functions = make(map[string]*bytecode.FunctionInfo)
	}
	c.mod.Functions[bfn.Name] = bfn
	return nil
}

func (c *Compiler) compileBlock(b *frontend.BlockStmt) {
	for _, stmt := range b.Statements {
		c.compileStmt(stmt)
	}
}

func (c *Compiler) compileStmt(s frontend.Stmt) {
	switch st := s.(type) {
	case *frontend.VarDeclStmt:
		c.compileVarDecl(st)
	case *frontend.AssignStmt:
		c.compileAssign(st)
	case *frontend.ExprStmt:
		c.compileExpr(st.Expr)
		c.chunk().Write(bytecode.OpPop, 0)
	case *frontend.ReturnStmt:
		c.compileReturn(st)
	case *frontend.IfStmt:
		c.compileIf(st)
	case *frontend.WhileStmt:
		c.compileWhile(st)
	default:
		panic(fmt.Sprintf("unknown stmt %T", st))
	}
}

func (c *Compiler) compileVarDecl(s *frontend.VarDeclStmt) {
	ch := c.chunk()
	typ := mapTypeName(s.TypeName)

	if s.Init != nil {
		c.compileExpr(s.Init)
	} else {
		ch.Write(bytecode.OpConst, 0)
		idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValNull})
		ch.WriteUint16(uint16(idx), 0)
	}

	slot := c.addLocal(s.Name, typ)

	ch.Write(bytecode.OpStoreLocal, 0)
	ch.WriteByte(byte(slot), 0)
}

func (c *Compiler) compileAssign(s *frontend.AssignStmt) {
	ch := c.chunk()

	c.compileExpr(s.Value)

	ident, ok := s.Target.(*frontend.IdentExpr)
	if !ok {
		panic("assignment to non-identifier not supported")
	}

	if slot, ok := c.resolveLocal(ident.Name); ok {
		ch.Write(bytecode.OpStoreLocal, 0)
		ch.WriteByte(byte(slot), 0)
	} else {
		panic("unknown variable " + ident.Name)
	}
}

func (c *Compiler) compileReturn(s *frontend.ReturnStmt) {
	ch := c.chunk()
	if s.Value != nil {

		c.compileExpr(s.Value)
	} else {

		ch.Write(bytecode.OpConst, 0)
		idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValNull})
		ch.WriteUint16(uint16(idx), 0)
	}
	ch.Write(bytecode.OpReturn, 0)
}

func (c *Compiler) compileIf(s *frontend.IfStmt) {
	ch := c.chunk()

	c.compileExpr(s.Condition)

	ch.Write(bytecode.OpJumpIfFalse, 0)
	jumpToElse := len(ch.Code)
	ch.WriteUint16(0, 0)

	ch.Write(bytecode.OpPop, 0)

	c.compileBlock(s.ThenBlock)

	ch.Write(bytecode.OpJump, 0)
	jumpAfterElse := len(ch.Code)
	ch.WriteUint16(0, 0)

	elsePos := len(ch.Code)
	ch.PatchUint16(jumpToElse, uint16(elsePos))

	ch.Write(bytecode.OpPop, 0)

	if s.ElseBlock != nil {
		c.compileBlock(s.ElseBlock)
	}

	endPos := len(ch.Code)
	ch.PatchUint16(jumpAfterElse, uint16(endPos))
}

func (c *Compiler) compileWhile(s *frontend.WhileStmt) {
	ch := c.chunk()

	loopStart := len(ch.Code)

	c.compileExpr(s.Condition)

	ch.Write(bytecode.OpJumpIfFalse, 0)
	exitJump := len(ch.Code)
	ch.WriteUint16(0, 0)

	ch.Write(bytecode.OpPop, 0)

	c.compileBlock(s.Body)

	ch.Write(bytecode.OpJump, 0)
	ch.WriteUint16(uint16(loopStart), 0)

	afterLoop := len(ch.Code)
	ch.PatchUint16(exitJump, uint16(afterLoop))
	ch.Write(bytecode.OpPop, 0)
}

func (c *Compiler) compileExpr(e frontend.Expr) {
	switch ex := e.(type) {

	case *frontend.NumberExpr:
		c.compileNumber(ex)

	case *frontend.StringExpr:
		c.compileString(ex)

	case *frontend.BoolExpr:
		c.compileBool(ex)

	case *frontend.NullExpr:
		c.compileNull()

	case *frontend.IdentExpr:
		c.compileIdent(ex)

	case *frontend.UnaryExpr:
		c.compileUnary(ex)

	case *frontend.BinaryExpr:
		c.compileBinary(ex)

	case *frontend.CallExpr:
		c.compileCall(ex)

	default:
		panic(fmt.Sprintf("unknown expr %T", e))
	}
}

func (c *Compiler) compileNumber(e *frontend.NumberExpr) {
	ch := c.chunk()
	v := bytecode.Value{Kind: bytecode.ValFloat, F: e.Value}

	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(v)
	ch.WriteUint16(uint16(idx), 0)
}

func (c *Compiler) compileString(e *frontend.StringExpr) {
	ch := c.chunk()
	v := bytecode.Value{Kind: bytecode.ValString, S: e.Value}

	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(v)
	ch.WriteUint16(uint16(idx), 0)
}

func (c *Compiler) compileBool(e *frontend.BoolExpr) {
	ch := c.chunk()
	v := bytecode.Value{Kind: bytecode.ValBool, B: e.Value}

	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(v)
	ch.WriteUint16(uint16(idx), 0)
}

func (c *Compiler) compileNull() {
	ch := c.chunk()

	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValNull})
	ch.WriteUint16(uint16(idx), 0)
}

func (c *Compiler) compileIdent(e *frontend.IdentExpr) {
	ch := c.chunk()

	if slot, ok := c.resolveLocal(e.Name); ok {
		ch.Write(bytecode.OpLoadLocal, 0)
		ch.WriteByte(byte(slot), 0)
		return
	}

	panic("unknown variable: " + e.Name)
}

func (c *Compiler) compileUnary(e *frontend.UnaryExpr) {
	c.compileExpr(e.Expr)

	ch := c.chunk()

	switch e.Op {
	case "-":
		ch.Write(bytecode.OpNeg, 0)
	case "!":
		ch.Write(bytecode.OpNot, 0)
	default:
		panic("unknown unary op: " + e.Op)
	}
}

func (c *Compiler) compileBinary(e *frontend.BinaryExpr) {
	ch := c.chunk()

	c.compileExpr(e.Left)
	c.compileExpr(e.Right)

	switch e.Op {
	case "+":
		ch.Write(bytecode.OpAdd, 0)
	case "-":
		ch.Write(bytecode.OpSub, 0)
	case "*":
		ch.Write(bytecode.OpMul, 0)
	case "/":
		ch.Write(bytecode.OpDiv, 0)
	case "%":
		ch.Write(bytecode.OpMod, 0)
	case "^":
		ch.Write(bytecode.OpPow, 0)

	case "==":
		ch.Write(bytecode.OpEq, 0)
	case "!=":
		ch.Write(bytecode.OpNe, 0)
	case "<":
		ch.Write(bytecode.OpLt, 0)
	case "<=":
		ch.Write(bytecode.OpLe, 0)
	case ">":
		ch.Write(bytecode.OpGt, 0)
	case ">=":
		ch.Write(bytecode.OpGe, 0)

	case "&&":
		c.compileExpr(e.Left)

		ch.Write(bytecode.OpJumpIfFalse, 0)
		jumpToEnd := len(ch.Code)
		ch.WriteUint16(0, 0)

		ch.Write(bytecode.OpPop, 0)

		c.compileExpr(e.Right)

		end := len(ch.Code)
		ch.PatchUint16(jumpToEnd, uint16(end))
		return

	case "||":
		c.compileExpr(e.Left)

		ch.Write(bytecode.OpNot, 0)

		ch.Write(bytecode.OpJumpIfFalse, 0)
		jumpToEnd := len(ch.Code)
		ch.WriteUint16(0, 0)

		ch.Write(bytecode.OpNot, 0)

		ch.Write(bytecode.OpPop, 0)

		c.compileExpr(e.Right)
		
		end := len(ch.Code)
		ch.PatchUint16(jumpToEnd, uint16(end))
		return

	default:
		panic("unknown binary op: " + e.Op)
	}
}

func (c *Compiler) compileCall(e *frontend.CallExpr) {
	ch := c.chunk()

	id, ok := e.Callee.(*frontend.IdentExpr)
	if !ok {
		panic("call of non-identifier is not supported")
	}

	for _, arg := range e.Args {
		c.compileExpr(arg)
	}

	name := id.Name
	_, ok = c.mod.Functions[name]
	if !ok {
		panic("unknown function: " + name)
	}

	ch.Write(bytecode.OpCall, 0)

	idx := ch.AddConstant(bytecode.Value{
		Kind: bytecode.ValString,
		S:    name,
	})
	ch.WriteUint16(uint16(idx), 0)
}
