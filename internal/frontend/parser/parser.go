package parser

import (
	"fmt"
	"strings"

	"github.com/ChernykhITMO/compiler/internal/frontend/ast"
	"github.com/ChernykhITMO/compiler/internal/frontend/token"
	"github.com/ChernykhITMO/compiler/internal/frontend/types"
)

type Parser struct {
	tokens []token.Token
	pos    int
}

func NewParser(tokens []token.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) current() token.Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *Parser) previous() token.Token {

	return p.tokens[p.pos-1]
}

func (p *Parser) isAtEnd() bool {
	return p.current().Type == token.TokenEnd
}

func (p *Parser) check(tt token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.current().Type == tt
}

func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.pos++
	}
	return p.previous()
}

func (p *Parser) match(tt token.TokenType) bool {
	if p.check(tt) {
		_ = p.advance()
		return true
	}
	return false
}

func (p *Parser) consume(tt token.TokenType, msg string) token.Token {
	if p.check(tt) {
		return p.advance()
	}
	cur := p.current()
	panic(fmt.Errorf("parse error at pos %d: %s", cur.Pos, msg))
}

func (p *Parser) parseTypeName() types.Type {
	var base types.Type
	switch {
	case p.match(token.TokenInt):
		base = types.TypeFromToken(token.TokenInt)
	case p.match(token.TokenFloat):
		base = types.TypeFromToken(token.TokenFloat)
	case p.match(token.TokenString):
		base = types.TypeFromToken(token.TokenString)
	case p.match(token.TokenBool):
		base = types.TypeFromToken(token.TokenBool)
	case p.match(token.TokenChar):
		base = types.TypeFromToken(token.TokenChar)
	case p.match(token.TokenVoid):
		base = types.TypeFromToken(token.TokenVoid)
	default:
		cur := p.current()
		panic(fmt.Errorf("parse error at pos %d: expected types name", cur.Pos))
	}

	for p.match(token.TokenLeftBracket) {
		p.consume(token.TokenRightBracket, "expected ']' after '[' in array type")
		base = types.Type{
			Kind: types.TypeArray,
			Elem: &base,
		}
	}

	return base
}

func (p *Parser) parseBaseTypeName() types.Type {
	switch {
	case p.match(token.TokenInt):
		return types.TypeFromToken(token.TokenInt)
	case p.match(token.TokenFloat):
		return types.TypeFromToken(token.TokenFloat)
	case p.match(token.TokenString):
		return types.TypeFromToken(token.TokenString)
	case p.match(token.TokenBool):
		return types.TypeFromToken(token.TokenBool)
	case p.match(token.TokenChar):
		return types.TypeFromToken(token.TokenChar)
	case p.match(token.TokenVoid):
		return types.TypeFromToken(token.TokenVoid)
	default:
		cur := p.current()
		panic(fmt.Errorf("parse error at pos %d: expected type name", cur.Pos))
	}
}

func (p *Parser) ParseProgram() *ast.Program {
	prog := &ast.Program{}

	for !p.isAtEnd() {
		for p.match(token.TokenNewline) {
		}
		if p.isAtEnd() {
			break
		}
		prog.Functions = append(prog.Functions, p.parseFunction())
	}

	return prog
}

func (p *Parser) parseFunction() *ast.FunctionDecl {
	p.consume(token.TokenFunction, "expected 'function'")
	nameTok := p.consume(token.TokenIdentifier, "expected function name")

	fn := &ast.FunctionDecl{Name: nameTok.Text}

	p.consume(token.TokenLeftParen, "expected '(' after function name")

	if !p.check(token.TokenRightParen) {
		for {
			parseType := p.parseTypeName()
			paramNameTok := p.consume(token.TokenIdentifier, "expected parameter name")
			fn.Params = append(fn.Params, ast.Param{
				Name: paramNameTok.Text,
				Type: parseType,
			})

			if !p.match(token.TokenComma) {
				break
			}
		}
	}
	p.consume(token.TokenRightParen, "expected ')' after parameters")

	if p.check(token.TokenInt) || p.check(token.TokenFloat) ||
		p.check(token.TokenString) || p.check(token.TokenBool) ||
		p.check(token.TokenChar) || p.check(token.TokenVoid) {
		fn.ReturnType = p.parseTypeName()
	} else {
		fn.ReturnType = types.Type{Kind: types.TypeVoid} // Надо подумать, убрать ли в конце функции тип
		// (UPD: НЕ УБИРАЙ!! В ВАЛИДАТОРЕ ИДЕТ ПРОВЕРКА)
		// Каждая фукнкция если не задана на тип будет войдовской
	}

	fn.Body = p.parseBlock()
	p.match(token.TokenNewline) // опциональный \n после функции

	return fn
}

func (p *Parser) parseBlock() *ast.BlockStmt {
	p.consume(token.TokenLeftBrace, "expected '{' to start block")
	block := &ast.BlockStmt{}

	for !p.check(token.TokenRightBrace) && !p.isAtEnd() {
		if p.match(token.TokenNewline) {
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	p.consume(token.TokenRightBrace, "expected '}' to end block")
	p.match(token.TokenNewline)

	return block
}

func PrintProgram(prog *ast.Program) {
	for _, fn := range prog.Functions {
		printFunction(fn, 0)
		fmt.Println()
	}
}

func printFunction(fn *ast.FunctionDecl, indent int) {
	ind := strings.Repeat("  ", indent)
	fmt.Printf("%sFunction %s(", ind, fn.Name)
	for i, p := range fn.Params {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s %s", p.Type, p.Name)
	}
	fmt.Printf(") %s\n", fn.ReturnType)
	printBlock(fn.Body, indent+1)
}

func printBlock(block *ast.BlockStmt, indent int) {
	if block == nil {
		return
	}
	ind := strings.Repeat("  ", indent)
	fmt.Printf("%sBlock:\n", ind)
	for _, stmt := range block.Statements {
		printStmt(stmt, indent+1)
	}
}

func printStmt(s ast.Stmt, indent int) {
	ind := strings.Repeat("  ", indent)

	switch st := s.(type) {
	case *ast.VarDeclStmt:
		fmt.Printf("%sVarDecl %s %s\n", ind, st.Type, st.Name)
		if st.Init != nil {
			fmt.Printf("%s  Init:\n", ind)
			printExpr(st.Init, indent+2)
		}
	case *ast.ForStmt:
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

	case *ast.AssignStmt:
		fmt.Printf("%sAssign:\n", ind)
		fmt.Printf("%s  Target:\n", ind)
		printExpr(st.Target, indent+2)
		fmt.Printf("%s  Value:\n", ind)
		printExpr(st.Value, indent+2)

	case *ast.ExprStmt:
		fmt.Printf("%sExprStmt:\n", ind)
		printExpr(st.Expr, indent+1)

	case *ast.ReturnStmt:
		fmt.Printf("%sReturn\n", ind)
		if st.Value != nil {
			fmt.Printf("%s  Value:\n", ind)
			printExpr(st.Value, indent+2)
		}

	case *ast.IfStmt:
		fmt.Printf("%sIf:\n", ind)
		fmt.Printf("%s  Condition:\n", ind)
		printExpr(st.Condition, indent+2)
		fmt.Printf("%s  Then:\n", ind)
		printBlock(st.ThenBlock, indent+2)
		if st.ElseBlock != nil {
			fmt.Printf("%s  Else:\n", ind)
			printBlock(st.ElseBlock, indent+2)
		}

	case *ast.WhileStmt:
		fmt.Printf("%sWhile:\n", ind)
		fmt.Printf("%s  Condition:\n", ind)
		printExpr(st.Condition, indent+2)
		fmt.Printf("%s  Body:\n", ind)
		printBlock(st.Body, indent+2)
	case *ast.BreakStmt:
		fmt.Printf("%sBreak\n", ind)

	case *ast.ContinueStmt:
		fmt.Printf("%sContinue\n", ind)

	default:
		fmt.Printf("%s<unknown stmt %T>\n", ind, st)
	}
}

func printExpr(e ast.Expr, indent int) {
	ind := strings.Repeat("  ", indent)

	switch ex := e.(type) {

	case *ast.LiteralExpr:
		fmt.Printf("%sLiteral(%s : %s)\n", ind, ex.Lexeme, ex.Type)

	case *ast.IdentExpr:
		fmt.Printf("%sIdent(%s)\n", ind, ex.Name)

	case *ast.UnaryExpr:
		fmt.Printf("%sUnary(%v):\n", ind, ex.Op)
		printExpr(ex.Expr, indent+1)

	case *ast.BinaryExpr:
		fmt.Printf("%sBinary(%v):\n", ind, ex.Op)
		fmt.Printf("%s  Left:\n", ind)
		printExpr(ex.Left, indent+2)
		fmt.Printf("%s  Right:\n", ind)
		printExpr(ex.Right, indent+2)

	case *ast.CallExpr:
		fmt.Printf("%sCall:\n", ind)
		fmt.Printf("%s  Callee:\n", ind)
		printExpr(ex.Callee, indent+2)
		fmt.Printf("%s  Args:\n", ind)
		for _, a := range ex.Args {
			printExpr(a, indent+2)
		}

	case *ast.IndexExpr:
		fmt.Printf("%sIndex:\n", ind)
		fmt.Printf("%s  Array:\n", ind)
		printExpr(ex.Array, indent+2)
		fmt.Printf("%s  Index:\n", ind)
		printExpr(ex.Index, indent+2)

	case *ast.NewArrayExpr:
		fmt.Printf("%sNewArray(%s):\n", ind, ex.ElementType.String())
		fmt.Printf("%s  Length:\n", ind)
		printExpr(ex.Length, indent+2)

	default:
		fmt.Printf("%s<unknown expr %T>\n", ind, ex)
	}
}

func printInlineStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.VarDeclStmt:
		fmt.Printf("%s %s", st.Type, st.Name)
		if st.Init != nil {
			fmt.Print(" = ")
			printInlineExpr(st.Init)
		}

	case *ast.AssignStmt:
		printInlineExpr(st.Target)
		fmt.Print(" = ")
		printInlineExpr(st.Value)

	case *ast.ExprStmt:
		printInlineExpr(st.Expr)

	default:
		fmt.Printf("<stmt %T>", st)
	}
}

func printInlineExpr(e ast.Expr) {
	switch ex := e.(type) {

	case *ast.LiteralExpr:
		fmt.Print(ex.Lexeme)

	case *ast.IdentExpr:
		fmt.Printf("%s", ex.Name)

	case *ast.UnaryExpr:
		fmt.Printf("%v", ex.Op)
		printInlineExpr(ex.Expr)

	case *ast.BinaryExpr:
		printInlineExpr(ex.Left)
		fmt.Printf(" %v ", ex.Op)
		printInlineExpr(ex.Right)

	case *ast.CallExpr:
		printInlineExpr(ex.Callee)
		fmt.Print("(")
		for i, arg := range ex.Args {
			if i > 0 {
				fmt.Print(", ")
			}
			printInlineExpr(arg)
		}
		fmt.Print(")")

	case *ast.IndexExpr:
		printInlineExpr(ex.Array)
		fmt.Print("[")
		printInlineExpr(ex.Index)
		fmt.Print("]")

	case *ast.NewArrayExpr:
		fmt.Printf("new %s[", ex.ElementType.String())
		printInlineExpr(ex.Length)
		fmt.Print("]")

	default:
		fmt.Printf("<expr %T>", ex)
	}
}
