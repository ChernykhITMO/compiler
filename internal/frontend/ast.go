package frontend

type Expr interface {
	exprNode()
}

type Stmt interface {
	stmtNode()
}

type Program struct {
	Functions []*FunctionDecl
}

type FunctionDecl struct {
	Name       string
	Params     []Param
	ReturnType string
	Body       *BlockStmt
}

type Param struct {
	TypeName string
	Name     string
}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (*BinaryExpr) exprNode() {}

type UnaryExpr struct {
	Op   string
	Expr Expr
}

func (*UnaryExpr) exprNode() {}

type NumberExpr struct {
	Value float64
}

func (*NumberExpr) exprNode() {}

type StringExpr struct {
	Value string
}

func (*StringExpr) exprNode() {}

type BoolExpr struct {
	Value bool
}

func (*BoolExpr) exprNode() {}

type NullExpr struct{}

func (*NullExpr) exprNode() {}

type IdentExpr struct {
	Name string
}

func (*IdentExpr) exprNode() {}

type CallExpr struct {
	Callee Expr
	Args   []Expr
}

func (*CallExpr) exprNode() {}

type BlockStmt struct {
	Statements []Stmt
}

func (*BlockStmt) stmtNode() {}

type ExprStmt struct {
	Expr Expr
}

func (*ExprStmt) stmtNode() {}

type VarDeclStmt struct {
	TypeName string
	Name     string
	Init     Expr
}

func (*VarDeclStmt) stmtNode() {}

type AssignStmt struct {
	Target Expr
	Value  Expr
}

func (*AssignStmt) stmtNode() {}

type ReturnStmt struct {
	Value Expr
}

func (*ReturnStmt) stmtNode() {}

type IfStmt struct {
	Condition Expr
	ThenBlock *BlockStmt
	ElseBlock *BlockStmt
}

func (*IfStmt) stmtNode() {}

type WhileStmt struct {
	Condition Expr
	Body      *BlockStmt
}

func (*WhileStmt) stmtNode() {}
