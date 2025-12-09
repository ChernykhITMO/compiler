package semantics

import (
	"fmt"

	"github.com/ChernykhITMO/compiler/internal/frontend/ast"
)

const (
	duplicateVar       = "duplicateVariable"
	undeclaredVariable = "UndeclaredVariable"
)

type SemanticError struct {
	Type    string
	Message string
}

type Checker struct {
	functions map[string]struct{}
	errors    []SemanticError
	scopes    []map[string]struct{}
}

func NewChecker() *Checker {
	return &Checker{
		functions: make(map[string]struct{}),
		errors:    make([]SemanticError, 0),
	}
}

func (c *Checker) Check(program *ast.Program) []SemanticError {
	c.errors = []SemanticError{}

	for _, fn := range program.Functions {
		c.functions[fn.Name] = struct{}{}
	}

	for _, fn := range program.Functions {
		c.checkFunction(fn)
	}

	return c.errors
}

func (c *Checker) checkFunction(fn *ast.FunctionDecl) {
	c.pushScope()
	defer c.popScope()

	for _, param := range fn.Params {
		c.declareVar(param.Name)
	}

	c.checkBlock(fn.Body)
}

func (c *Checker) checkBlock(block *ast.BlockStmt) {
	if block == nil {
		return
	}
	c.pushScope()
	defer c.popScope()

	for _, stmt := range block.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkStatement(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		c.declareVar(s.Name)
		if s.Init != nil {
			c.checkExpression(s.Init)
		}

	case *ast.AssignStmt:
		c.checkExpression(s.Target)
		c.checkExpression(s.Value)

	case *ast.ExprStmt:
		c.checkExpression(s.Expr)

	case *ast.ReturnStmt:
		if s.Value != nil {
			c.checkExpression(s.Value)
		}

	case *ast.IfStmt:
		c.checkExpression(s.Condition)
		c.checkBlock(s.ThenBlock)
		if s.ElseBlock != nil {
			c.checkBlock(s.ElseBlock)
		}

	case *ast.WhileStmt:
		c.checkExpression(s.Condition)
		c.checkBlock(s.Body)

	case *ast.ForStmt:
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

func (c *Checker) checkExpression(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.IdentExpr:
		if !c.isVarDeclared(e.Name) {
			c.addError(undeclaredVariable,
				fmt.Sprintf("variable '%s' is not declared", e.Name))
		}

	case *ast.CallExpr:
		if ident, ok := e.Callee.(*ast.IdentExpr); ok {
			if _, ok := c.functions[ident.Name]; !ok {
				c.addError("UndeclaredFunction",
					fmt.Sprintf("Function '%s' is not declared", ident.Name))
			}
		}
		for _, arg := range e.Args {
			c.checkExpression(arg)
		}

	case *ast.BinaryExpr:
		c.checkExpression(e.Left)
		c.checkExpression(e.Right)

	case *ast.UnaryExpr:
		c.checkExpression(e.Expr)

	case *ast.LiteralExpr:
	}
}

func (c *Checker) addError(errType, message string) {
	c.errors = append(c.errors, SemanticError{
		Type:    errType,
		Message: message,
	})
}

func (c *Checker) pushScope() {
	c.scopes = append(c.scopes, make(map[string]struct{}))
}

func (c *Checker) popScope() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *Checker) declareVar(name string) {
	scope := c.scopes[len(c.scopes)-1]
	if _, ok := scope[name]; ok {
		c.addError(duplicateVar,
			fmt.Sprintf("variable '%s already exists in this scope", name))
		return
	}
	scope[name] = struct{}{}
}

func (c *Checker) isVarDeclared(name string) bool {
	for _, scope := range c.scopes {
		if _, ok := scope[name]; ok {
			return true
		}
	}
	return false
}
