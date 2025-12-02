package frontend

import "fmt"

const (
	duplicateFunction  = "Duplicate function"
	duplicateParameter = "Duplicate parameter"
	mainSignature      = "Main signature"
	mainReturnType     = "Main return type"
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

func (v *ASTValidator) Validate(program *Program) []ValidationError {
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

func (v *ASTValidator) validateFunction(fun *FunctionDecl) {
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

func (v *ASTValidator) validateMainFunction(program *Program) {
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

			if fun.ReturnType != "void" {
				v.addError(mainReturnType,
					fmt.Sprintf("main should be void type"))
			}
		}
	}
	if !hasMain {
		v.addError(noMainFunction,
			fmt.Sprintf("program must have 'main' function"))
	}
}

func (v *ASTValidator) validateBlock(block *BlockStmt, context string) {
	if block == nil {
		return
	}

	for _, stmt := range block.Statements {
		v.validateStatement(stmt, context)
	}
}

func (v *ASTValidator) validateStatement(stmt Stmt, context string) {
	switch s := stmt.(type) {
	case *VarDeclStmt:
		if s.Init != nil {
			v.validateExpression(s.Init, context)
		}

	case *AssignStmt:
		v.validateExpression(s.Target, context)
		v.validateExpression(s.Value, context)

	case *ReturnStmt:
		if s.Value != nil {
			v.validateExpression(s.Value, context)
		}

	case *IfStmt:
		v.validateExpression(s.Condition, context)
		v.validateBlock(s.ThenBlock, context)
		v.validateBlock(s.ElseBlock, context)

	case *WhileStmt:
		v.validateExpression(s.Condition, context)
		v.validateBlock(s.Body, context)

	case *ExprStmt:
		v.validateExpression(s.Expr, context)

	case *ForStmt:
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
	case *BreakStmt, *ContinueStmt:
		// позже допишу, пока нет идеи что проверять
	}

}

func (v *ASTValidator) validateExpression(expr Expr, context string) {
	switch e := expr.(type) {
	case *BinaryExpr:
		v.validateExpression(e.Left, context)
		v.validateExpression(e.Right, context)

	case *UnaryExpr:
		v.validateExpression(e.Expr, context)

	case *CallExpr:
		v.validateExpression(e.Callee, context)
		for _, arg := range e.Args {
			v.validateExpression(arg, context)
		}
	case *IdentExpr, *NumberExpr, *StringExpr, *BoolExpr, *NullExpr:
		return
	}
}

func (v *ASTValidator) addError(errType, message string) {
	v.errors = append(v.errors, ValidationError{
		Type:    errType,
		Message: message,
	})
}

func (v *ASTValidator) validateReturnStatements(block *BlockStmt, expectedType string, funcName string) {
	if block == nil {
		return
	}

	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ReturnStmt:
			if expectedType == "void" && s.Value != nil {
				v.addError(invalidReturn,
					fmt.Sprintf("Function '%s' returns value but declared as void", funcName))
			}
			if expectedType != "void" && s.Value == nil {
				v.addError(missingReturnValue,
					fmt.Sprintf("Function '%s' must return a value of type %s",
						funcName, expectedType))
			}
		case *IfStmt:
			v.validateReturnStatements(s.ThenBlock, expectedType, funcName)
			v.validateReturnStatements(s.ElseBlock, expectedType, funcName)
		case *WhileStmt:
			v.validateReturnStatements(s.Body, expectedType, funcName)
		case *ForStmt:
			v.validateReturnStatements(s.Body, expectedType, funcName)
		}
	}
}
