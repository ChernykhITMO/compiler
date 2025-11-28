package frontend

import (
	"fmt"
)

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) current() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return p.tokens[len(p.tokens)-1]
}

func (p *Parser) previous() Token {
	return p.tokens[p.pos-1]
}

func (p *Parser) isAtEnd() bool {
	return p.current().Type == TokenEnd
}

func (p *Parser) check(tt TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.current().Type == tt
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.pos++
	}
	return p.previous()
}

func (p *Parser) match(tt TokenType) bool {
	if p.check(tt) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) consume(tt TokenType, msg string) Token {
	if p.check(tt) {
		return p.advance()
	}
	panic(fmt.Errorf("parse error at pos %d: %s", p.current().Pos, msg))
}

func (p *Parser) parseTypeName() string {
	switch {
	case p.match(TokenInt):
		return "int"
	case p.match(TokenFloat):
		return "float"
	case p.match(TokenString):
		return "string"
	case p.match(TokenBool):
		return "bool"
	case p.match(TokenChar):
		return "char"
	case p.match(TokenVoid):
		return "void"
	default:
		panic(fmt.Errorf("parse error at pos %d: expected type name", p.current().Pos))
	}
}

func (p *Parser) ParseProgram() *Program {
	prog := &Program{}

	for !p.isAtEnd() {
		for p.match(TokenNewline) {
		}
		if p.isAtEnd() {
			break
		}
		prog.Functions = append(prog.Functions, p.parseFunction())
	}

	return prog
}

func (p *Parser) parseFunction() *FunctionDecl {
	p.consume(TokenFunction, "expected 'function'")
	nameTok := p.consume(TokenIdentifier, "expected function name")

	fn := &FunctionDecl{Name: nameTok.Text}

	p.consume(TokenLeftParen, "expected '(' after function name")

	if !p.check(TokenRightParen) {
		for {
			paramType := p.parseTypeName()
			paramNameTok := p.consume(TokenIdentifier, "expected parameter name")
			fn.Params = append(fn.Params, Param{TypeName: paramType, Name: paramNameTok.Text})

			if !p.match(TokenComma) {
				break
			}
		}
	}
	p.consume(TokenRightParen, "expected ')' after parameters")

	if p.check(TokenInt) || p.check(TokenFloat) ||
		p.check(TokenString) || p.check(TokenBool) ||
		p.check(TokenChar) || p.check(TokenVoid) {
		fn.ReturnType = p.parseTypeName()
	} else {
		fn.ReturnType = "void"
	}

	fn.Body = p.parseBlock()
	p.match(TokenNewline) // опциональный \n после функции

	return fn
}

func (p *Parser) parseBlock() *BlockStmt {
	p.consume(TokenLeftBrace, "expected '{' to start block")
	block := &BlockStmt{}

	for !p.check(TokenRightBrace) && !p.isAtEnd() {
		if p.match(TokenNewline) {
			continue
		}
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}

	p.consume(TokenRightBrace, "expected '}' to end block")
	p.match(TokenNewline) // опциональный \n после блока

	return block
}

func (p *Parser) parseStatement() Stmt {
	if p.match(TokenIf) {
		return p.parseIfStmt()
	}
	if p.match(TokenWhile) {
		return p.parseWhileStmt()
	}
	if p.match(TokenFor) {
		return p.parseForStmt()
	}
	if p.match(TokenReturn) {
		return p.parseReturnStmt()
	}

	if p.check(TokenInt) || p.check(TokenFloat) ||
		p.check(TokenString) || p.check(TokenBool) ||
		p.check(TokenChar) {
		return p.parseVarDeclOrExprStmt()
	}

	if p.check(TokenIdentifier) {
		nextType := TokenEnd
		if p.pos+1 < len(p.tokens) {
			nextType = p.tokens[p.pos+1].Type
		}
		if nextType == TokenAssign {
			nameTok := p.consume(TokenIdentifier, "expected identifier")
			p.consume(TokenAssign, "expected '='")
			value := p.parseExpression()
			p.match(TokenNewline)

			target := &IdentExpr{Name: nameTok.Text}
			return &AssignStmt{Target: target, Value: value}
		}
	}

	expr := p.parseExpression()
	p.match(TokenNewline)
	return &ExprStmt{Expr: expr}
}

func (p *Parser) parseVarDeclOrExprStmt() Stmt {
	typeName := p.parseTypeName()
	nameTok := p.consume(TokenIdentifier, "expected variable name")

	var init Expr
	if p.match(TokenAssign) {
		init = p.parseExpression()
	}
	p.match(TokenNewline)

	return &VarDeclStmt{
		TypeName: typeName,
		Name:     nameTok.Text,
		Init:     init,
	}
}

func (p *Parser) parseReturnStmt() Stmt {
	if p.check(TokenNewline) || p.check(TokenRightBrace) {
		p.match(TokenNewline)
		return &ReturnStmt{}
	}
	val := p.parseExpression()
	p.match(TokenNewline)
	return &ReturnStmt{Value: val}
}

func (p *Parser) parseIfStmt() Stmt {
	hasParen := p.match(TokenLeftParen)
	cond := p.parseExpression()
	if hasParen {
		p.consume(TokenRightParen, "expected ')' after if condition")
	}

	thenBlock := p.parseBlock()

	var elseBlock *BlockStmt
	if p.match(TokenElse) {
		elseBlock = p.parseBlock()
	}

	return &IfStmt{
		Condition: cond,
		ThenBlock: thenBlock,
		ElseBlock: elseBlock,
	}
}

func (p *Parser) parseWhileStmt() Stmt {
	hasParen := p.match(TokenLeftParen)
	cond := p.parseExpression()
	if hasParen {
		p.consume(TokenRightParen, "expected ')' after while condition")
	}

	body := p.parseBlock()
	return &WhileStmt{
		Condition: cond,
		Body:      body,
	}
}

func (p *Parser) parseForStmt() Stmt {
	panic("for-statement parsing not implemented yet")
}
