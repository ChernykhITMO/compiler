package semantics

import (
	"fmt"

	"github.com/ChernykhITMO/compiler/internal/frontend"
)

type SemanticError struct {
	Type    string
	Message string
}

type Checker struct {
	functions map[string]struct{}
	variables map[string]struct{}
	errors    []SemanticError
}

func NewChecker() *Checker {
	return &Checker{
		functions: make(map[string]struct{}),
		variables: make(map[string]struct{}),
		errors:    make([]SemanticError, 0),
	}
}

func (c *Checker) Check(program *frontend.Program) []SemanticError {
	c.errors = []SemanticError{}

	for _, fn := range program.Functions {
		c.functions[fn.Name] = struct{}{}
	}

	for _, fn := range program.Functions {
		c.checkFunction(fn)
	}

	return c.errors
}

func (c *Checker) checkFunction(fn *frontend.FunctionDecl) {
	c.variables = make(map[string]struct{})

	for _, param := range fn.Params {
		c.variables[param.Name] = struct{}{}
	}

	c.checkBlock(fn.Body)
}

func (c *Checker) checkBlock(block *frontend.BlockStmt) {
	if block == nil {
		return
	}

	for _, stmt := range block.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkStatement(stmt frontend.Stmt) {
	switch s := stmt.(type) {
	case *frontend.VarDeclStmt:
		// Объявление переменной
		if _, ok := c.variables[s.Name]; ok {
			c.addError("DuplicateVariable",
				fmt.Sprintf("Variable '%s' already declared", s.Name))
		} else {
			c.variables[s.Name] = struct{}{}
		}

	case *frontend.AssignStmt:
		c.checkExpression(s.Target)
		c.checkExpression(s.Value)

	case *frontend.ExprStmt:
		c.checkExpression(s.Expr)

	case *frontend.ReturnStmt:
		if s.Value != nil {
			c.checkExpression(s.Value)
		}

	case *frontend.IfStmt:
		c.checkExpression(s.Condition)
		c.checkBlock(s.ThenBlock)
		if s.ElseBlock != nil {
			c.checkBlock(s.ElseBlock)
		}

	case *frontend.WhileStmt:
		c.checkExpression(s.Condition)
		c.checkBlock(s.Body)

	case *frontend.ForStmt:
		if s.Init != nil {
			c.checkStatement(s.Init)
		}
		if s.Condition != nil {
			c.checkExpression(s.Condition)
		}
		if s.Increment != nil {
			c.checkStatement(s.Increment)
		}
		c.checkBlock(s.Body)
	}
}

// Проверяем выражение
func (c *Checker) checkExpression(expr frontend.Expr) {
	switch e := expr.(type) {
	case *frontend.IdentExpr:
		// Проверяем, что переменная объявлена
		_, ok1 := c.variables[e.Name]
		_, ok2 := c.functions[e.Name]
		if !ok1 && !ok2 {
			c.addError("UndeclaredVariable",
				fmt.Sprintf("'%s' is not declared", e.Name))
		}

	case *frontend.CallExpr:
		if ident, ok := e.Callee.(*frontend.IdentExpr); ok {
			if _, ok := c.functions[ident.Name]; !ok {
				c.addError("UndeclaredFunction",
					fmt.Sprintf("Function '%s' is not declared", ident.Name))
			}
		}
		for _, arg := range e.Args {
			c.checkExpression(arg)
		}

	case *frontend.BinaryExpr:
		c.checkExpression(e.Left)
		c.checkExpression(e.Right)

	case *frontend.UnaryExpr:
		c.checkExpression(e.Expr)

	case *frontend.NumberExpr, *frontend.StringExpr,
		*frontend.BoolExpr, *frontend.NullExpr:
	}
}

func (c *Checker) addError(errType, message string) {
	c.errors = append(c.errors, SemanticError{
		Type:    errType,
		Message: message,
	})
}
