package parser

import (
	"github.com/ChernykhITMO/compiler/internal/frontend/ast"
	"github.com/ChernykhITMO/compiler/internal/frontend/token"
)

func (p *Parser) parseStatement() ast.Stmt {
	if p.match(token.TokenIf) {
		return p.parseIfStmt()
	}
	if p.match(token.TokenWhile) {
		return p.parseWhileStmt()
	}
	if p.match(token.TokenFor) {
		return p.parseForStmt()
	}
	if p.match(token.TokenReturn) {
		return p.parseReturnStmt()
	}
	if p.match(token.TokenBreak) {
		p.match(token.TokenNewline)
		return &ast.BreakStmt{}
	}
	if p.match(token.TokenContinue) {
		p.match(token.TokenNewline)
		return &ast.ContinueStmt{}
	}

	if p.check(token.TokenInt) || p.check(token.TokenFloat) ||
		p.check(token.TokenString) || p.check(token.TokenBool) ||
		p.check(token.TokenChar) {
		return p.parseVarDeclOrExprStmt()
	}

	if p.check(token.TokenIdentifier) {
		nextType := token.TokenEnd
		if p.pos+1 < len(p.tokens) {
			nextType = p.tokens[p.pos+1].Type
		}
		if nextType == token.TokenAssign {
			nameTok := p.consume(token.TokenIdentifier, "expected identifier")
			p.consume(token.TokenAssign, "expected '='")
			value := p.parseExpression()
			p.match(token.TokenNewline)

			target := &ast.IdentExpr{Name: nameTok.Text}
			return &ast.AssignStmt{Target: target, Value: value}
		}
	}

	expr := p.parseExpression()
	p.match(token.TokenNewline)
	return &ast.ExprStmt{Expr: expr}
}

func (p *Parser) parseVarDeclOrExprStmt() ast.Stmt {
	typeName := p.parseTypeName()
	nameTok := p.consume(token.TokenIdentifier, "expected variable name")

	var init ast.Expr
	if p.match(token.TokenAssign) {
		init = p.parseExpression()
	}
	p.match(token.TokenNewline)

	return &ast.VarDeclStmt{
		Type: typeName,
		Name: nameTok.Text,
		Init: init,
	}
}

func (p *Parser) parseReturnStmt() ast.Stmt {
	if p.check(token.TokenNewline) || p.check(token.TokenRightBrace) {
		p.match(token.TokenNewline)
		return &ast.ReturnStmt{}
	}
	val := p.parseExpression()
	p.match(token.TokenNewline)
	return &ast.ReturnStmt{Value: val}
}

func (p *Parser) parseIfStmt() ast.Stmt {
	hasParen := p.match(token.TokenLeftParen)
	cond := p.parseExpression()
	if hasParen {
		p.consume(token.TokenRightParen, "expected ')' after if condition")
	}

	thenBlock := p.parseBlock()

	var elseBlock *ast.BlockStmt
	if p.match(token.TokenElse) {
		elseBlock = p.parseBlock()
	}

	return &ast.IfStmt{
		Condition: cond,
		ThenBlock: thenBlock,
		ElseBlock: elseBlock,
	}
}

func (p *Parser) parseWhileStmt() ast.Stmt {
	hasParen := p.match(token.TokenLeftParen)
	cond := p.parseExpression()
	if hasParen {
		p.consume(token.TokenRightParen, "expected ')' after while condition")
	}

	body := p.parseBlock()
	return &ast.WhileStmt{
		Condition: cond,
		Body:      body,
	}
}

func (p *Parser) parseForStmt() ast.Stmt {
	hasParen := p.match(token.TokenLeftParen)

	var init ast.Stmt
	if !p.check(token.TokenSemicolon) {
		init = p.parseStatement()
	}
	p.consume(token.TokenSemicolon, "expected ';' after for init")

	var cond ast.Expr
	if !p.check(token.TokenSemicolon) {
		cond = p.parseExpression()
	}
	p.consume(token.TokenSemicolon, "expected ';' after for condition")

	var incr ast.Stmt
	if !p.check(token.TokenRightParen) {
		incr = p.parseStatement()
	}

	if hasParen {
		p.consume(token.TokenRightParen, "expected ')' after for clauses")
	}

	body := p.parseBlock()

	return &ast.ForStmt{
		Init:      init,
		Condition: cond,
		Increment: incr,
		Body:      body,
	}
}
