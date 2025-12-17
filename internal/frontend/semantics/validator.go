package semantics

import (
	"fmt"

	"github.com/ChernykhITMO/compiler/internal/frontend/ast"
	"github.com/ChernykhITMO/compiler/internal/frontend/types"
)

const (
	duplicateFunction  = "Duplicate function"
	duplicateParameter = "Duplicate parameter"
	mainSignature      = "Main signature"
	mainReturnType     = "Main return types"
	noMainFunction     = "No main function"
	invalidReturn      = "Invalid return"
	missingReturnValue = "Missing return value"
)

type ValidationError struct {
	Type    string
	Message string
}

type ASTValidator struct {
	functions map[string]struct{}
	errors    []ValidationError
}

func NewASTValidator() *ASTValidator {
	return &ASTValidator{
		functions: make(map[string]struct{}),
		errors:    make([]ValidationError, 0),
	}
}

func (v *ASTValidator) Validate(program *ast.Program) []ValidationError {
	v.errors = []ValidationError{}

	for _, fn := range program.Functions {
		v.validateFunction(fn)
	}

	v.validateMainFunction(program)

	for _, fn := range program.Functions {
		v.validateBlock(fn.Body, fn.Name)
	}

	return v.errors
}

func (v *ASTValidator) validateFunction(fun *ast.FunctionDecl) {
	if _, ok := v.functions[fun.Name]; ok {
		v.addError(duplicateFunction,
			fmt.Sprintf("Function '%s' is already defined", fun.Name))
	}
	v.functions[fun.Name] = struct{}{}

	paramNames := make(map[string]struct{})

	for _, p := range fun.Params {
		if _, ok := paramNames[p.Name]; ok {
			v.addError(duplicateParameter,
				fmt.Sprintf("Duplicate parameter name '%s' in function '%s'", p.Name, fun.Name))
		}
		paramNames[p.Name] = struct{}{}
	}

	v.validateReturnStatements(fun.Body, fun.ReturnType, fun.Name)
}

func (v *ASTValidator) validateMainFunction(program *ast.Program) {
	hasMain := false
	mainCount := 0

	for _, fun := range program.Functions {
		if fun.Name != "main" {
			continue
		}
		mainCount++

		if mainCount == 1 {
			hasMain = true
			if len(fun.Params) != 0 {
				v.addError(mainSignature,
					fmt.Sprintf("main doesn't have parameters"))
			}

			if fun.ReturnType.Kind != types.TypeVoid {
				v.addError(mainReturnType,
					fmt.Sprintf("main should be void types"))
			}
		}
	}
	if !hasMain {
		v.addError(noMainFunction,
			fmt.Sprintf("program must have 'main' function"))
	}
}

func (v *ASTValidator) validateBlock(block *ast.BlockStmt, context string) {
	if block == nil {
		return
	}

	for _, stmt := range block.Statements {
		v.validateStatement(stmt, context)
	}
}

func (v *ASTValidator) validateStatement(stmt ast.Stmt, context string) {
	switch s := stmt.(type) {
	case *ast.VarDeclStmt:
		if s.Init != nil {
			v.validateExpression(s.Init, context)
		}

	case *ast.AssignStmt:
		v.validateExpression(s.Target, context)
		v.validateExpression(s.Value, context)

	case *ast.ReturnStmt:
		if s.Value != nil {
			v.validateExpression(s.Value, context)
		}

	case *ast.IfStmt:
		v.validateExpression(s.Condition, context)
		v.validateBlock(s.ThenBlock, context)
		v.validateBlock(s.ElseBlock, context)

	case *ast.WhileStmt:
		v.validateExpression(s.Condition, context)
		v.validateBlock(s.Body, context)

	case *ast.ExprStmt:
		v.validateExpression(s.Expr, context)

	case *ast.ForStmt:
		if s.Init != nil {
			v.validateStatement(s.Init, context)
		}
		if s.Condition != nil {
			v.validateExpression(s.Condition, context)
		}
		if s.Increment != nil {
			v.validateStatement(s.Increment, context)
		}
		v.validateBlock(s.Body, context)
	}

}

func (v *ASTValidator) validateExpression(expr ast.Expr, context string) {
	switch e := expr.(type) {

	case *ast.BinaryExpr:
		v.validateExpression(e.Left, context)
		v.validateExpression(e.Right, context)

	case *ast.UnaryExpr:
		v.validateExpression(e.Expr, context)

	case *ast.CallExpr:
		v.validateExpression(e.Callee, context)
		for _, arg := range e.Args {
			v.validateExpression(arg, context)
		}

	case *ast.IndexExpr:
		v.validateExpression(e.Array, context)
		v.validateExpression(e.Index, context)

	case *ast.NewArrayExpr:
		v.validateExpression(e.Length, context)

	case *ast.IdentExpr, *ast.LiteralExpr:
		return
	}
}

func (v *ASTValidator) addError(errType, message string) {
	v.errors = append(v.errors, ValidationError{
		Type:    errType,
		Message: message,
	})
}

func (v *ASTValidator) validateReturnStatements(block *ast.BlockStmt, expectedType types.Type, funcName string) {
	if block == nil {
		return
	}

	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.ReturnStmt:
			if expectedType.Kind == types.TypeVoid && s.Value != nil {
				v.addError(invalidReturn,
					fmt.Sprintf("Function '%s' returns value but declared as void", funcName))
			}
			if expectedType.Kind != types.TypeVoid && s.Value == nil {
				v.addError(missingReturnValue,
					fmt.Sprintf("Function '%s' must return a value of types %s",
						funcName, expectedType))
			}
		case *ast.IfStmt:
			v.validateReturnStatements(s.ThenBlock, expectedType, funcName)
			v.validateReturnStatements(s.ElseBlock, expectedType, funcName)
		case *ast.WhileStmt:
			v.validateReturnStatements(s.Body, expectedType, funcName)
		case *ast.ForStmt:
			v.validateReturnStatements(s.Body, expectedType, funcName)
		}
	}
}
