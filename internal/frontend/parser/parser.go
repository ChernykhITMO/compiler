package parser

import (
	"fmt"
	"strings"

	"github.com/ChernykhITMO/compiler/internal/frontend"
)

type Parser struct {
	tokens []frontend.Token
	pos    int
}

func NewParser(tokens []frontend.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) current() frontend.Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *Parser) previous() frontend.Token {
	return p.tokens[p.pos-1]
}

func (p *Parser) isAtEnd() bool {
	return p.current().Type == frontend.TokenEnd
}

func (p *Parser) check(tt frontend.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.current().Type == tt
}

func (p *Parser) advance() frontend.Token {
	if !p.isAtEnd() {
		p.pos++
	}
	return p.previous()
}

func (p *Parser) match(tt frontend.TokenType) bool {
	if p.check(tt) {
		_ = p.advance()
		return true
	}
	return false
}

func (p *Parser) consume(tt frontend.TokenType, msg string) frontend.Token {
	if p.check(tt) {
		return p.advance()
	}
	cur := p.current()
	panic(fmt.Errorf("parse error at pos %d: %s", cur.Pos, msg))
}

func (p *Parser) parseTypeName() string {
	switch {
	case p.match(frontend.TokenInt):
		return "int"
	case p.match(frontend.TokenFloat):
		return "float"
	case p.match(frontend.TokenString):
		return "string"
	case p.match(frontend.TokenBool):
		return "bool"
	case p.match(frontend.TokenChar):
		return "char"
	case p.match(frontend.TokenVoid):
		return "void"
	default:
		cur := p.current()
		panic(fmt.Errorf("parse error at pos %d: expected type name", cur.Pos))
	}
}

func (p *Parser) ParseProgram() *frontend.Program {
	prog := &frontend.Program{}

	for !p.isAtEnd() {
		for p.match(frontend.TokenNewline) {
		}
		if p.isAtEnd() {
			break
		}
		prog.Functions = append(prog.Functions, p.parseFunction())
	}

	return prog
}

func (p *Parser) parseFunction() *frontend.FunctionDecl {
	p.consume(frontend.TokenFunction, "expected 'function'")
	nameTok := p.consume(frontend.TokenIdentifier, "expected function name")

	fn := &frontend.FunctionDecl{Name: nameTok.Text}

	p.consume(frontend.TokenLeftParen, "expected '(' after function name")

	if !p.check(frontend.TokenRightParen) {
		for {
			paramType := p.parseTypeName()
			paramNameTok := p.consume(frontend.TokenIdentifier, "expected parameter name")
			fn.Params = append(fn.Params, frontend.Param{
				TypeName: paramType,
				Name:     paramNameTok.Text,
			})

			if !p.match(frontend.TokenComma) {
				break
			}
		}
	}
	p.consume(frontend.TokenRightParen, "expected ')' after parameters")

	if p.check(frontend.TokenInt) || p.check(frontend.TokenFloat) ||
		p.check(frontend.TokenString) || p.check(frontend.TokenBool) ||
		p.check(frontend.TokenChar) || p.check(frontend.TokenVoid) {
		fn.ReturnType = p.parseTypeName()
	} else {
		fn.ReturnType = "void" // Надо подумать, убрать ли в конце функции тип
		// (UPD: НЕ УБИРАЙ!! В ВАЛИДАТОРЕ ИДЕТ ПРОВЕРКА)
		// Каждая фукнкция если не задана на тип будет войдовской
	}

	fn.Body = p.parseBlock()
	p.match(frontend.TokenNewline) // опциональный \n после функции

	return fn
}

func (p *Parser) parseBlock() *frontend.BlockStmt {
	p.consume(frontend.TokenLeftBrace, "expected '{' to start block")
	block := &frontend.BlockStmt{}

	for !p.check(frontend.TokenRightBrace) && !p.isAtEnd() {
		if p.match(frontend.TokenNewline) {
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	p.consume(frontend.TokenRightBrace, "expected '}' to end block")
	p.match(frontend.TokenNewline)

	return block
}

func PrintProgram(prog *frontend.Program) {
	for _, fn := range prog.Functions {
		printFunction(fn, 0)
		fmt.Println()
	}
}

func printFunction(fn *frontend.FunctionDecl, indent int) {
	ind := strings.Repeat("  ", indent)
	fmt.Printf("%sFunction %s(", ind, fn.Name)
	for i, p := range fn.Params {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s %s", p.TypeName, p.Name)
	}
	fmt.Printf(") %s\n", fn.ReturnType)
	printBlock(fn.Body, indent+1)
}

func printBlock(block *frontend.BlockStmt, indent int) {
	if block == nil {
		return
	}
	ind := strings.Repeat("  ", indent)
	fmt.Printf("%sBlock:\n", ind)
	for _, stmt := range block.Statements {
		printStmt(stmt, indent+1)
	}
}

func printStmt(s frontend.Stmt, indent int) {
	ind := strings.Repeat("  ", indent)

	switch st := s.(type) {
	case *frontend.VarDeclStmt:
		fmt.Printf("%sVarDecl %s %s\n", ind, st.TypeName, st.Name)
		if st.Init != nil {
			fmt.Printf("%s  Init:\n", ind)
			printExpr(st.Init, indent+2)
		}
	case *frontend.ForStmt:
		fmt.Printf("%sFor:\n", ind)
		if st.Init != nil {
			fmt.Printf("%s  Init:\n", ind)
			printStmt(st.Init, indent+2)
		}
		if st.Condition != nil {
			fmt.Printf("%s  Condition:\n", ind)
			printExpr(st.Condition, indent+2)
		}
		if st.Increment != nil {
			fmt.Printf("%s  Increment:\n", ind)
			printStmt(st.Increment, indent+2)
		}
		fmt.Printf("%s  Body:\n", ind)
		printBlock(st.Body, indent+2)

	case *frontend.AssignStmt:
		fmt.Printf("%sAssign:\n", ind)
		fmt.Printf("%s  Target:\n", ind)
		printExpr(st.Target, indent+2)
		fmt.Printf("%s  Value:\n", ind)
		printExpr(st.Value, indent+2)

	case *frontend.ExprStmt:
		fmt.Printf("%sExprStmt:\n", ind)
		printExpr(st.Expr, indent+1)

	case *frontend.ReturnStmt:
		fmt.Printf("%sReturn\n", ind)
		if st.Value != nil {
			fmt.Printf("%s  Value:\n", ind)
			printExpr(st.Value, indent+2)
		}

	case *frontend.IfStmt:
		fmt.Printf("%sIf:\n", ind)
		fmt.Printf("%s  Condition:\n", ind)
		printExpr(st.Condition, indent+2)
		fmt.Printf("%s  Then:\n", ind)
		printBlock(st.ThenBlock, indent+2)
		if st.ElseBlock != nil {
			fmt.Printf("%s  Else:\n", ind)
			printBlock(st.ElseBlock, indent+2)
		}

	case *frontend.WhileStmt:
		fmt.Printf("%sWhile:\n", ind)
		fmt.Printf("%s  Condition:\n", ind)
		printExpr(st.Condition, indent+2)
		fmt.Printf("%s  Body:\n", ind)
		printBlock(st.Body, indent+2)
	case *frontend.BreakStmt:
		fmt.Printf("%sBreak\n", ind)

	case *frontend.ContinueStmt:
		fmt.Printf("%sContinue\n", ind)

	default:
		fmt.Printf("%s<unknown stmt %T>\n", ind, st)
	}
}

func printExpr(e frontend.Expr, indent int) {
	ind := strings.Repeat("  ", indent)

	switch ex := e.(type) {
	case *frontend.NumberExpr:
		fmt.Printf("%sNumber(%v)\n", ind, ex.Value)

	case *frontend.StringExpr:
		fmt.Printf("%sString(%q)\n", ind, ex.Value)

	case *frontend.BoolExpr:
		fmt.Printf("%sBool(%v)\n", ind, ex.Value)

	case *frontend.NullExpr:
		fmt.Printf("%sNull\n", ind)

	case *frontend.IdentExpr:
		fmt.Printf("%sIdent(%s)\n", ind, ex.Name)

	case *frontend.UnaryExpr:
		fmt.Printf("%sUnary(%s):\n", ind, ex.Op)
		printExpr(ex.Expr, indent+1)

	case *frontend.BinaryExpr:
		fmt.Printf("%sBinary(%s):\n", ind, ex.Op)
		fmt.Printf("%s  Left:\n", ind)
		printExpr(ex.Left, indent+2)
		fmt.Printf("%s  Right:\n", ind)
		printExpr(ex.Right, indent+2)

	case *frontend.CallExpr:
		fmt.Printf("%sCall:\n", ind)
		fmt.Printf("%s  Callee:\n", ind)
		printExpr(ex.Callee, indent+2)
		fmt.Printf("%s  Args:\n", ind)
		for _, a := range ex.Args {
			printExpr(a, indent+2)
		}

	default:
		fmt.Printf("%s<unknown expr %T>\n", ind, ex)
	}
}

func printInlineStmt(s frontend.Stmt) {
	switch st := s.(type) {
	case *frontend.VarDeclStmt:
		fmt.Printf("%s %s", st.Name, st.TypeName)
		if st.Init != nil {
			fmt.Print(" = ")
			printInlineExpr(st.Init)
		}

	case *frontend.AssignStmt:
		printInlineExpr(st.Target)
		fmt.Print(" = ")
		printInlineExpr(st.Value)

	case *frontend.ExprStmt:
		printInlineExpr(st.Expr)

	default:
		fmt.Printf("<stmt %T>", st)
	}
}

func printInlineExpr(e frontend.Expr) {
	switch ex := e.(type) {
	case *frontend.NumberExpr:
		fmt.Printf("%v", ex.Value)

	case *frontend.StringExpr:
		fmt.Printf("%q", ex.Value)

	case *frontend.BoolExpr:
		fmt.Printf("%v", ex.Value)

	case *frontend.NullExpr:
		fmt.Print("null")

	case *frontend.IdentExpr:
		fmt.Printf("%s", ex.Name)

	case *frontend.UnaryExpr:
		fmt.Printf("%s", ex.Op)
		printInlineExpr(ex.Expr)

	case *frontend.BinaryExpr:
		printInlineExpr(ex.Left)
		fmt.Printf(" %s ", ex.Op)
		printInlineExpr(ex.Right)

	case *frontend.CallExpr:
		printInlineExpr(ex.Callee)
		fmt.Print("(")
		for i, arg := range ex.Args {
			if i > 0 {
				fmt.Print(", ")
			}
			printInlineExpr(arg)
		}
		fmt.Print(")")

	default:
		fmt.Printf("<expr %T>", ex)
	}
}
