package ast

import (
	"github.com/ChernykhITMO/compiler/internal/frontend/token"
	"github.com/ChernykhITMO/compiler/internal/frontend/types"
)

type Expr interface {
	exprNode()
}

type Stmt interface {
	stmtNode()
}

type stmtBase struct{}

func (stmtBase) stmtNode() {}

type exprBase struct{}

func (exprBase) exprNode() {}

type Program struct {
	Functions []*FunctionDecl
}

type FunctionDecl struct {
	Name       string
	Params     []Param
	ReturnType types.Type
	Body       *BlockStmt
}

type Param struct {
	Name string
	Type types.Type
}

type IdentExpr struct {
	exprBase
	Name string
}

type BinaryExpr struct {
	exprBase
	Left  Expr
	Op    token.TokenType
	Right Expr
}

type UnaryExpr struct {
	exprBase
	Op   token.TokenType
	Expr Expr
}

type LiteralExpr struct {
	exprBase
	Lexeme string
	Token  token.TokenType
	Type   types.Type
}

type CallExpr struct {
	exprBase
	Callee Expr
	Args   []Expr
}

type VarDeclStmt struct {
	stmtBase
	Name string
	Type types.Type
	Init Expr
}

type BlockStmt struct {
	stmtBase
	Statements []Stmt
}

type ExprStmt struct {
	stmtBase
	Expr Expr
}

type AssignStmt struct {
	stmtBase
	Target Expr
	Value  Expr
}

type ReturnStmt struct {
	stmtBase
	Value Expr
}

type IfStmt struct {
	stmtBase
	Condition Expr
	ThenBlock *BlockStmt
	ElseBlock *BlockStmt
}

type WhileStmt struct {
	stmtBase
	Condition Expr
	Body      *BlockStmt
}

type ForStmt struct {
	stmtBase
	Init      Stmt
	Condition Expr
	Increment Stmt
	Body      *BlockStmt
}

type BreakStmt struct {
	stmtBase
}

type ContinueStmt struct {
	stmtBase
}
