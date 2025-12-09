package parser

import (
	"fmt"
	"strings"

	"github.com/ChernykhITMO/compiler/internal/frontend/ast"
	"github.com/ChernykhITMO/compiler/internal/frontend/token"
	"github.com/ChernykhITMO/compiler/internal/frontend/types"
)

func (p *Parser) parseExpression() ast.Expr {
	return p.parseLogicalOr()
}

func (p *Parser) parseLogicalOr() ast.Expr {
	expr := p.parseLogicalAnd()

	for p.match(token.TokenOr) {
		op := p.previous()
		right := p.parseLogicalAnd()
		expr = &ast.BinaryExpr{
			Left:  expr,
			Op:    op.Type,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseLogicalAnd() ast.Expr {
	expr := p.parseEquality()

	for p.match(token.TokenAnd) {
		op := p.previous()
		right := p.parseEquality()
		expr = &ast.BinaryExpr{
			Left:  expr,
			Op:    op.Type,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseEquality() ast.Expr {
	expr := p.parseComparison()

	for p.match(token.TokenEqual) || p.match(token.TokenNotEqual) {
		op := p.previous()
		right := p.parseComparison()
		expr = &ast.BinaryExpr{
			Left:  expr,
			Op:    op.Type,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseComparison() ast.Expr {
	expr := p.parseTerm()

	for {
		if p.check(token.TokenLess) || p.check(token.TokenLessEqual) ||
			p.check(token.TokenGreater) || p.check(token.TokenGreaterEqual) {

			op := p.advance()
			right := p.parseTerm()
			expr = &ast.BinaryExpr{
				Left:  expr,
				Op:    op.Type,
				Right: right,
			}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) parseTerm() ast.Expr {
	expr := p.parseFactor()

	for p.match(token.TokenPlus) || p.match(token.TokenMinus) {
		op := p.previous()
		right := p.parseFactor()
		expr = &ast.BinaryExpr{
			Left:  expr,
			Op:    op.Type,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseFactor() ast.Expr {
	expr := p.parsePower()

	for p.match(token.TokenMultiply) ||
		p.match(token.TokenDivide) ||
		p.match(token.TokenModulo) {

		op := p.previous()
		right := p.parsePower()
		expr = &ast.BinaryExpr{
			Left:  expr,
			Op:    op.Type,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parsePower() ast.Expr {
	expr := p.parseUnary()

	for p.match(token.TokenPower) {
		op := p.previous()
		right := p.parseUnary()
		expr = &ast.BinaryExpr{
			Left:  expr,
			Op:    op.Type,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseUnary() ast.Expr {
	if p.match(token.TokenNot) || p.match(token.TokenMinus) {
		op := p.previous()
		right := p.parseUnary()
		return &ast.UnaryExpr{
			Op:   op.Type,
			Expr: right,
		}
	}
	return p.parseCall()
}

func (p *Parser) parseCall() ast.Expr {
	expr := p.parsePrimary()

	for {
		if p.match(token.TokenLeftParen) {
			var args []ast.Expr
			if !p.check(token.TokenRightParen) {
				for {
					args = append(args, p.parseExpression())
					if !p.match(token.TokenComma) {
						break
					}
				}
			}
			p.consume(token.TokenRightParen, "expected ')' after arguments")
			expr = &ast.CallExpr{
				Callee: expr,
				Args:   args,
			}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) parsePrimary() ast.Expr {
	if p.match(token.TokenNumber) {
		t := p.previous()

		litType := types.Type{Kind: types.TypeInt}

		if strings.Contains(t.Text, ".") {
			litType = types.Type{Kind: types.TypeFloat}
		}
		return &ast.LiteralExpr{
			Lexeme: t.Text,
			Token:  t.Type,
			Type:   litType,
		}
	}
	if p.match(token.TokenText) {
		t := p.previous()
		return &ast.LiteralExpr{
			Lexeme: t.Text,
			Token:  t.Type,
			Type:   types.TypeFromToken(token.TokenString),
		}
	}

	if p.match(token.TokenTrue) {
		t := p.previous()
		return &ast.LiteralExpr{
			Lexeme: t.Text,
			Token:  t.Type,
			Type:   types.TypeFromToken(token.TokenBool),
		}
	}

	if p.match(token.TokenFalse) {
		t := p.previous()
		return &ast.LiteralExpr{
			Lexeme: t.Text,
			Token:  t.Type,
			Type:   types.TypeFromToken(token.TokenBool),
		}
	}

	if p.match(token.TokenNull) {
		t := p.previous()
		return &ast.LiteralExpr{
			Lexeme: t.Text,
			Token:  t.Type,
			Type:   types.Type{Kind: types.TypeNull},
		}
	}

	if p.match(token.TokenIdentifier) {
		t := p.previous()
		return &ast.IdentExpr{Name: t.Text}
	}

	if p.match(token.TokenLeftParen) {
		expr := p.parseExpression()
		p.consume(token.TokenRightParen, "expected ')'")
		return expr
	}

	panic(fmt.Errorf("parse error at pos %d: expected expression", p.current().Pos))
}
