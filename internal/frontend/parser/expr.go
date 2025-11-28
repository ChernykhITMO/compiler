package parser

import (
	"fmt"
	"strconv"

	"github.com/ChernykhITMO/compiler/internal/frontend"
)

func (p *Parser) parseExpression() frontend.Expr {
	return p.parseLogicalOr()
}

func (p *Parser) parseLogicalOr() frontend.Expr {
	expr := p.parseLogicalAnd()

	for p.match(frontend.TokenOr) {
		op := p.previous()
		right := p.parseLogicalAnd()
		expr = &frontend.BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseLogicalAnd() frontend.Expr {
	expr := p.parseEquality()

	for p.match(frontend.TokenAnd) {
		op := p.previous()
		right := p.parseEquality()
		expr = &frontend.BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseEquality() frontend.Expr {
	expr := p.parseComparison()

	for p.match(frontend.TokenEqual) || p.match(frontend.TokenNotEqual) {
		op := p.previous()
		right := p.parseComparison()
		expr = &frontend.BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseComparison() frontend.Expr {
	expr := p.parseTerm()

	for {
		if p.check(frontend.TokenLess) || p.check(frontend.TokenLessEqual) ||
			p.check(frontend.TokenGreater) || p.check(frontend.TokenGreaterEqual) {

			op := p.advance()
			right := p.parseTerm()
			expr = &frontend.BinaryExpr{
				Left:  expr,
				Op:    op.Text,
				Right: right,
			}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) parseTerm() frontend.Expr {
	expr := p.parseFactor()

	for p.match(frontend.TokenPlus) || p.match(frontend.TokenMinus) {
		op := p.previous()
		right := p.parseFactor()
		expr = &frontend.BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseFactor() frontend.Expr {
	expr := p.parsePower()

	for p.match(frontend.TokenMultiply) ||
		p.match(frontend.TokenDivide) ||
		p.match(frontend.TokenModulo) {

		op := p.previous()
		right := p.parsePower()
		expr = &frontend.BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parsePower() frontend.Expr {
	expr := p.parseUnary()

	for p.match(frontend.TokenPower) {
		op := p.previous()
		right := p.parseUnary()
		expr = &frontend.BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseUnary() frontend.Expr {
	if p.match(frontend.TokenNot) || p.match(frontend.TokenMinus) {
		op := p.previous()
		right := p.parseUnary()
		return &frontend.UnaryExpr{
			Op:   op.Text,
			Expr: right,
		}
	}
	return p.parseCall()
}

func (p *Parser) parseCall() frontend.Expr {
	expr := p.parsePrimary()

	for {
		if p.match(frontend.TokenLeftParen) {
			var args []frontend.Expr
			if !p.check(frontend.TokenRightParen) {
				for {
					args = append(args, p.parseExpression())
					if !p.match(frontend.TokenComma) {
						break
					}
				}
			}
			p.consume(frontend.TokenRightParen, "expected ')' after arguments")
			expr = &frontend.CallExpr{
				Callee: expr,
				Args:   args,
			}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) parsePrimary() frontend.Expr {
	if p.match(frontend.TokenNumber) {
		t := p.previous()
		v, err := strconv.ParseFloat(t.Text, 64)
		if err != nil {
			panic(fmt.Errorf("invalid number '%s' at pos %d", t.Text, t.Pos))
		}
		return &frontend.NumberExpr{Value: v}
	}

	if p.match(frontend.TokenText) {
		t := p.previous()
		return &frontend.StringExpr{Value: t.Text}
	}

	if p.match(frontend.TokenTrue) {
		return &frontend.BoolExpr{Value: true}
	}
	if p.match(frontend.TokenFalse) {
		return &frontend.BoolExpr{Value: false}
	}

	if p.match(frontend.TokenNull) {
		return &frontend.NullExpr{}
	}

	if p.match(frontend.TokenIdentifier) {
		t := p.previous()
		return &frontend.IdentExpr{Name: t.Text}
	}

	if p.match(frontend.TokenLeftParen) {
		expr := p.parseExpression()
		p.consume(frontend.TokenRightParen, "expected ')' after expression")
		return expr
	}

	panic(fmt.Errorf("parse error at pos %d: expected expression", p.current().Pos))
}
