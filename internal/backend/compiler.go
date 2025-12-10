package backend

import (
	"fmt"
	"strconv"

	"github.com/ChernykhITMO/compiler/internal/bytecode"
	"github.com/ChernykhITMO/compiler/internal/frontend/ast"
	"github.com/ChernykhITMO/compiler/internal/frontend/token"
	"github.com/ChernykhITMO/compiler/internal/frontend/types"
)

type localVar struct {
	name string
	slot int // индекс в locals во фрейме vm
	typ  bytecode.TypeKind
}

type Compiler struct {
	mod    *bytecode.Module
	fn     *bytecode.FunctionInfo
	locals []localVar

	breakStack    [][]int
	continueStack [][]int
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

func mapTypeName(t types.Type) bytecode.TypeKind {
	switch t.Kind {
	case types.TypeInt:
		return bytecode.TypeInt
	case types.TypeFloat:
		return bytecode.TypeFloat
	case types.TypeString:
		return bytecode.TypeString
	case types.TypeBool:
		return bytecode.TypeBool
	case types.TypeChar:
		return bytecode.TypeChar
	case types.TypeVoid:
		return bytecode.TypeVoid
	case types.TypeArray:
		return bytecode.TypeArray
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

func (c *Compiler) CompileProgram(p *ast.Program) (*bytecode.Module, error) {

	for _, fn := range p.Functions {
		if _, exists := c.mod.Functions[fn.Name]; exists {
			return nil, fmt.Errorf("duplicate function: %s", fn.Name)
		}

		bfn := &bytecode.FunctionInfo{
			Name:       fn.Name,
			ParamCount: len(fn.Params),
			ParamTypes: make([]bytecode.TypeKind, len(fn.Params)),
			ReturnType: mapTypeName(fn.ReturnType),
		}

		for i, p := range fn.Params {
			bfn.ParamTypes[i] = mapTypeName(p.Type)
		}

		c.mod.Functions[bfn.Name] = bfn
	}

	for _, fn := range p.Functions {
		if err := c.compileFunction(fn); err != nil {
			return nil, err
		}
	}
	return c.mod, nil
}

func (c *Compiler) compileFunction(fn *ast.FunctionDecl) error {

	bfn, ok := c.mod.Functions[fn.Name]
	if !ok {
		return fmt.Errorf("error: function %s not registered", fn.Name)
	}

	c.fn = bfn
	c.locals = nil

	bfn.Chunk = bytecode.Chunk{}
	bfn.NumLocals = 0

	for i, p := range fn.Params {
		bfn.ParamTypes[i] = mapTypeName(p.Type)
	}

	for _, p := range fn.Params {
		c.addLocal(p.Name, mapTypeName(p.Type))
	}

	c.compileBlock(fn.Body)

	ch := c.chunk()
	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(bytecode.Value{Kind: bytecode.ValNull})
	ch.WriteUint16(uint16(idx), 0)
	ch.Write(bytecode.OpReturn, 0)

	return nil
}

func (c *Compiler) compileBlock(b *ast.BlockStmt) {
	for _, stmt := range b.Statements {
		c.compileStmt(stmt)
	}
}

func (c *Compiler) compileStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.VarDeclStmt:
		c.compileVarDecl(st)
	case *ast.AssignStmt:
		c.compileAssign(st)
	case *ast.ExprStmt:
		c.compileExpr(st.Expr)
		c.chunk().Write(bytecode.OpPop, 0)
	case *ast.ReturnStmt:
		c.compileReturn(st)
	case *ast.IfStmt:
		c.compileIf(st)
	case *ast.WhileStmt:
		c.compileWhile(st)
	case *ast.ForStmt:
		c.compileFor(st)
	case *ast.BreakStmt:
		c.compileBreak(st)
	case *ast.ContinueStmt:
		c.compileContinue(st)
	default:
		panic(fmt.Sprintf("unknown stmt %T", st))
	}
}

func (c *Compiler) compileVarDecl(s *ast.VarDeclStmt) {
	ch := c.chunk()
	typ := mapTypeName(s.Type)

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

func (c *Compiler) compileAssign(s *ast.AssignStmt) {
	ch := c.chunk()

	switch target := s.Target.(type) {
	case *ast.IdentExpr:
		c.compileExpr(s.Value)

		if slot, ok := c.resolveLocal(target.Name); ok {
			ch.Write(bytecode.OpStoreLocal, 0)
			ch.WriteByte(byte(slot), 0)
		} else {
			panic("unknown variable " + target.Name)
		}

	case *ast.IndexExpr:
		c.compileExpr(target.Array)
		c.compileExpr(target.Index)
		c.compileExpr(s.Value)
		ch.Write(bytecode.OpArraySet, 0)

	default:
		panic("assignment to unsupported target")
	}
}

func (c *Compiler) compileReturn(s *ast.ReturnStmt) {
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

func (c *Compiler) compileIf(s *ast.IfStmt) {
	ch := c.chunk()

	c.compileExpr(s.Condition)

	ch.Write(bytecode.OpJumpIfFalse, 0) // если false, перескакиваем в else
	jumpToElse := len(ch.Code)          // индекс где будут лежать аргументы этой инструкции
	ch.WriteUint16(0, 0)                //заглушка

	ch.Write(bytecode.OpPop, 0) // убираем условие, если true

	c.compileBlock(s.ThenBlock)

	ch.Write(bytecode.OpJump, 0)  // перескачить else при true
	jumpAfterElse := len(ch.Code) // индекс где будут лежать аргументы этой инструкции
	ch.WriteUint16(0, 0)

	elsePos := len(ch.Code) //здесь начинается else часть
	ch.PatchUint16(jumpToElse, uint16(elsePos))

	ch.Write(bytecode.OpPop, 0)

	if s.ElseBlock != nil {
		c.compileBlock(s.ElseBlock)
	}

	endPos := len(ch.Code) // конец блока if
	ch.PatchUint16(jumpAfterElse, uint16(endPos))
}

func (c *Compiler) compileWhile(s *ast.WhileStmt) {
	ch := c.chunk()

	loopStart := len(ch.Code)
	c.beginLoop()

	c.compileExpr(s.Condition)

	ch.Write(bytecode.OpJumpIfFalse, 0)
	exitJump := len(ch.Code)
	ch.WriteUint16(0, 0)

	ch.Write(bytecode.OpPop, 0)

	c.compileBlock(s.Body)

	ch.Write(bytecode.OpJump, 0)
	ch.WriteUint16(uint16(loopStart), 0)

	exitLabel := len(ch.Code)
	ch.PatchUint16(exitJump, uint16(exitLabel))
	ch.Write(bytecode.OpPop, 0)

	afterLoop := len(ch.Code)
	c.endLoop(loopStart, afterLoop)

}

func (c *Compiler) compileExpr(e ast.Expr) {
	switch ex := e.(type) {
	case *ast.IdentExpr:
		c.compileIdent(ex)
	case *ast.LiteralExpr:
		c.compileLiteral(ex)
	case *ast.UnaryExpr:
		c.compileUnary(ex)
	case *ast.BinaryExpr:
		c.compileBinary(ex)
	case *ast.CallExpr:
		c.compileCall(ex)
	case *ast.IndexExpr:
		c.compileExpr(ex.Array)
		c.compileExpr(ex.Index)
		c.chunk().Write(bytecode.OpArrayGet, 0)
	case *ast.NewArrayExpr:
		c.compileExpr(ex.Length)
		c.chunk().Write(bytecode.OpArrayNew, 0)
	default:
		panic(fmt.Sprintf("unknown expr %T", ex))
	}
}

func (c *Compiler) compileLiteral(l *ast.LiteralExpr) {

	switch l.Type.Kind {
	case types.TypeInt:
		c.compileInt(l)
	case types.TypeFloat:
		c.compileFloat(l)
	case types.TypeString:
		c.compileString(l)
	case types.TypeBool:
		c.compileBool(l)
	case types.TypeChar:
		c.compileChar(l)
	case types.TypeNull:
		c.compileNull()

	default:
		panic(fmt.Sprintf("unknown type %T", l.Type))
	}

}

func (c *Compiler) compileInt(l *ast.LiteralExpr) {
	ch := c.chunk()
	intVal, _ := strconv.Atoi(l.Lexeme)
	v := bytecode.Value{Kind: bytecode.ValInt, I: int64(intVal)}

	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(v)
	ch.WriteUint16(uint16(idx), 0)
}

func (c *Compiler) compileFloat(l *ast.LiteralExpr) {
	ch := c.chunk()
	floatVal, _ := strconv.ParseFloat(l.Lexeme, 32)
	v := bytecode.Value{Kind: bytecode.ValFloat, F: floatVal}

	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(v)
	ch.WriteUint16(uint16(idx), 0)
}

func (c *Compiler) compileString(l *ast.LiteralExpr) {
	ch := c.chunk()
	v := bytecode.Value{Kind: bytecode.ValString, S: l.Lexeme}

	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(v)
	ch.WriteUint16(uint16(idx), 0)
}

func (c *Compiler) compileBool(l *ast.LiteralExpr) {
	ch := c.chunk()
	boolV, _ := strconv.ParseBool(l.Lexeme)
	v := bytecode.Value{Kind: bytecode.ValBool, B: boolV}

	ch.Write(bytecode.OpConst, 0)
	idx := ch.AddConstant(v)
	ch.WriteUint16(uint16(idx), 0)
}

func (c *Compiler) compileChar(l *ast.LiteralExpr) {
	ch := c.chunk()
	v := bytecode.Value{Kind: bytecode.ValChar, C: l.Lexeme[0]}

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

func (c *Compiler) compileIdent(e *ast.IdentExpr) {
	ch := c.chunk()

	if slot, ok := c.resolveLocal(e.Name); ok {
		ch.Write(bytecode.OpLoadLocal, 0)
		ch.WriteByte(byte(slot), 0)
		return
	}

	panic("unknown variable: " + e.Name)
}

func (c *Compiler) compileUnary(e *ast.UnaryExpr) {
	c.compileExpr(e.Expr)

	ch := c.chunk()

	switch e.Op {
	case token.TokenMinus:
		ch.Write(bytecode.OpNeg, 0)
	case token.TokenNot:
		ch.Write(bytecode.OpNot, 0)
	default:
		panic("unknown unary op")
	}
}

func (c *Compiler) compileBinary(e *ast.BinaryExpr) {
	ch := c.chunk()

	switch e.Op {

	case token.TokenAnd:

		c.compileExpr(e.Left)

		ch.Write(bytecode.OpJumpIfFalse, 0)
		jumpToEnd := len(ch.Code)
		ch.WriteUint16(0, 0)

		ch.Write(bytecode.OpPop, 0)

		c.compileExpr(e.Right)

		end := len(ch.Code)
		ch.PatchUint16(jumpToEnd, uint16(end))
		return

	case token.TokenOr:

		c.compileExpr(e.Left)

		ch.Write(bytecode.OpJumpIfFalse, 0)
		jumpToRight := len(ch.Code)
		ch.WriteUint16(0, 0)

		ch.Write(bytecode.OpJump, 0)
		jumpAfterTrue := len(ch.Code)
		ch.WriteUint16(0, 0)

		rightPos := len(ch.Code)
		ch.PatchUint16(jumpToRight, uint16(rightPos))

		ch.Write(bytecode.OpPop, 0)

		c.compileExpr(e.Right)

		end := len(ch.Code)
		ch.PatchUint16(jumpAfterTrue, uint16(end))
		return

	default:
		c.compileExpr(e.Left)
		c.compileExpr(e.Right)

		switch e.Op {
		case token.TokenPlus:
			ch.Write(bytecode.OpAdd, 0)
		case token.TokenMinus:
			ch.Write(bytecode.OpSub, 0)
		case token.TokenMultiply:
			ch.Write(bytecode.OpMul, 0)
		case token.TokenDivide:
			ch.Write(bytecode.OpDiv, 0)
		case token.TokenModulo:
			ch.Write(bytecode.OpMod, 0)
		case token.TokenPower:
			ch.Write(bytecode.OpPow, 0)

		case token.TokenEqual:
			ch.Write(bytecode.OpEq, 0)
		case token.TokenNotEqual:
			ch.Write(bytecode.OpNe, 0)
		case token.TokenLess:
			ch.Write(bytecode.OpLt, 0)
		case token.TokenLessEqual:
			ch.Write(bytecode.OpLe, 0)
		case token.TokenGreater:
			ch.Write(bytecode.OpGt, 0)
		case token.TokenGreaterEqual:
			ch.Write(bytecode.OpGe, 0)

		default:
			panic("unknown binary op")
		}
	}
}

func (c *Compiler) compileCall(e *ast.CallExpr) {
	ch := c.chunk()

	id, ok := e.Callee.(*ast.IdentExpr)
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

func (c *Compiler) compileFor(s *ast.ForStmt) {
	ch := c.chunk()

	if s.Init != nil {
		c.compileStmt(s.Init)
	}

	loopStart := len(ch.Code)
	c.beginLoop()

	var exitJumpPos int
	hasCond := s.Condition != nil

	if hasCond {
		c.compileExpr(s.Condition)

		ch.Write(bytecode.OpJumpIfFalse, 0)
		exitJumpPos = len(ch.Code)
		ch.WriteUint16(0, 0)

		ch.Write(bytecode.OpPop, 0)
	}

	c.compileBlock(s.Body)

	incrementTarget := loopStart
	if s.Increment != nil {
		incrementTarget = len(ch.Code)
		c.compileStmt(s.Increment)
	}

	ch.Write(bytecode.OpJump, 0)
	ch.WriteUint16(uint16(loopStart), 0)

	if hasCond {
		exitLable := len(ch.Code)
		ch.PatchUint16(exitJumpPos, uint16(exitLable))

		ch.Write(bytecode.OpPop, 0)

		afterLoop := len(ch.Code)
		c.endLoop(incrementTarget, afterLoop)
	} else {
		afterLoop := len(ch.Code)
		c.endLoop(incrementTarget, afterLoop)
	}
}

func (c *Compiler) beginLoop() {
	c.breakStack = append(c.breakStack, nil)
	c.continueStack = append(c.continueStack, nil)
}

func (c *Compiler) endLoop(continueTarget, breakTarget int) {
	ch := c.chunk()

	bi := len(c.breakStack) - 1
	for _, pos := range c.breakStack[bi] {
		ch.PatchUint16(pos, uint16(breakTarget))
	}
	c.breakStack = c.breakStack[:bi]

	ci := len(c.continueStack) - 1
	for _, pos := range c.continueStack[ci] {
		ch.PatchUint16(pos, uint16(continueTarget))
	}
	c.continueStack = c.continueStack[:ci]
}

func (c *Compiler) compileBreak(_ *ast.BreakStmt) {
	if len(c.breakStack) == 0 {
		panic("break outside of loop")
	}
	ch := c.chunk()
	ch.Write(bytecode.OpJump, 0)
	pos := len(ch.Code)
	ch.WriteUint16(0, 0)

	i := len(c.breakStack) - 1
	c.breakStack[i] = append(c.breakStack[i], pos)
}

func (c *Compiler) compileContinue(_ *ast.ContinueStmt) {
	if len(c.continueStack) == 0 {
		panic("continue outside of loop")
	}
	ch := c.chunk()
	ch.Write(bytecode.OpJump, 0)
	pos := len(ch.Code)
	ch.WriteUint16(0, 0)

	i := len(c.continueStack) - 1
	c.continueStack[i] = append(c.continueStack[i], pos)
}
