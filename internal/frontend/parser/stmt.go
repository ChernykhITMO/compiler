package parser

import "github.com/ChernykhITMO/compiler/internal/frontend"

func (p *Parser) parseStatement() frontend.Stmt {
	if p.match(frontend.TokenIf) {
		return p.parseIfStmt()
	}
	if p.match(frontend.TokenWhile) {
		return p.parseWhileStmt()
	}
	if p.match(frontend.TokenFor) {
		return p.parseForStmt()
	}
	if p.match(frontend.TokenReturn) {
		return p.parseReturnStmt()
	}

	// объявление переменной
	if p.check(frontend.TokenInt) || p.check(frontend.TokenFloat) ||
		p.check(frontend.TokenString) || p.check(frontend.TokenBool) ||
		p.check(frontend.TokenChar) {
		return p.parseVarDeclOrExprStmt()
	}

	// присваивание: Identifier '=' expr
	if p.check(frontend.TokenIdentifier) {
		nextType := frontend.TokenEnd
		if p.pos+1 < len(p.tokens) {
			nextType = p.tokens[p.pos+1].Type
		}
		if nextType == frontend.TokenAssign {
			nameTok := p.consume(frontend.TokenIdentifier, "expected identifier")
			p.consume(frontend.TokenAssign, "expected '='")
			value := p.parseExpression()
			p.match(frontend.TokenNewline)

			target := &frontend.IdentExpr{Name: nameTok.Text}
			return &frontend.AssignStmt{Target: target, Value: value}
		}
	}

	expr := p.parseExpression()
	p.match(frontend.TokenNewline)
	return &frontend.ExprStmt{Expr: expr}
}

func (p *Parser) parseVarDeclOrExprStmt() frontend.Stmt {
	typeName := p.parseTypeName()
	nameTok := p.consume(frontend.TokenIdentifier, "expected variable name")

	var init frontend.Expr
	if p.match(frontend.TokenAssign) {
		init = p.parseExpression()
	}
	p.match(frontend.TokenNewline)

	return &frontend.VarDeclStmt{
		TypeName: typeName,
		Name:     nameTok.Text,
		Init:     init,
	}
}

func (p *Parser) parseReturnStmt() frontend.Stmt {
	if p.check(frontend.TokenNewline) || p.check(frontend.TokenRightBrace) {
		p.match(frontend.TokenNewline)
		return &frontend.ReturnStmt{}
	}
	val := p.parseExpression()
	p.match(frontend.TokenNewline)
	return &frontend.ReturnStmt{Value: val}
}

func (p *Parser) parseIfStmt() frontend.Stmt {
	hasParen := p.match(frontend.TokenLeftParen)
	cond := p.parseExpression()
	if hasParen {
		p.consume(frontend.TokenRightParen, "expected ')' after if condition")
	}

	thenBlock := p.parseBlock()

	var elseBlock *frontend.BlockStmt
	if p.match(frontend.TokenElse) {
		elseBlock = p.parseBlock()
	}

	return &frontend.IfStmt{
		Condition: cond,
		ThenBlock: thenBlock,
		ElseBlock: elseBlock,
	}
}

func (p *Parser) parseWhileStmt() frontend.Stmt {
	hasParen := p.match(frontend.TokenLeftParen)
	cond := p.parseExpression()
	if hasParen {
		p.consume(frontend.TokenRightParen, "expected ')' after while condition")
	}

	body := p.parseBlock()
	return &frontend.WhileStmt{
		Condition: cond,
		Body:      body,
	}
}

func (p *Parser) parseForStmt() frontend.Stmt {
	panic("for-statement parsing not implemented yet")
}
