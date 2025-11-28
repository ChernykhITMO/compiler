package frontend

import (
	"fmt"
	"strconv"
)

func (p *Parser) parseExpression() Expr {
	return p.parseLogicalOr()
}

func (p *Parser) parseLogicalOr() Expr {
	expr := p.parseLogicalAnd()

	for p.match(TokenOr) {
		op := p.previous()
		right := p.parseLogicalAnd()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseLogicalAnd() Expr {
	expr := p.parseEquality()

	for p.match(TokenAnd) {
		op := p.previous()
		right := p.parseEquality()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseEquality() Expr {
	expr := p.parseComparison()

	for p.match(TokenEqual) || p.match(TokenNotEqual) {
		op := p.previous()
		right := p.parseComparison()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseComparison() Expr {
	expr := p.parseTerm()

	for {
		if p.check(TokenLess) || p.check(TokenLessEqual) ||
			p.check(TokenGreater) || p.check(TokenGreaterEqual) {

			op := p.advance()
			right := p.parseTerm()
			expr = &BinaryExpr{
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

func (p *Parser) parseTerm() Expr {
	expr := p.parseFactor()

	for p.match(TokenPlus) || p.match(TokenMinus) {
		op := p.previous()
		right := p.parseFactor()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseFactor() Expr {
	expr := p.parsePower()

	for p.match(TokenMultiply) || p.match(TokenDivide) || p.match(TokenModulo) {
		op := p.previous()
		right := p.parsePower()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parsePower() Expr {
	expr := p.parseUnary()

	for p.match(TokenPower) {
		op := p.previous()
		right := p.parseUnary()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op.Text,
			Right: right,
		}
	}

	return expr
}

func (p *Parser) parseUnary() Expr {
	if p.match(TokenNot) || p.match(TokenMinus) {
		op := p.previous()
		right := p.parseUnary()
		return &UnaryExpr{
			Op:   op.Text,
			Expr: right,
		}
	}
	return p.parseCall()
}

func (p *Parser) parseCall() Expr {
	expr := p.parsePrimary()

	for {
		if p.match(TokenLeftParen) {
			var args []Expr
			if !p.check(TokenRightParen) {
				for {
					args = append(args, p.parseExpression())
					if !p.match(TokenComma) {
						break
					}
				}
			}
			p.consume(TokenRightParen, "expected ')' after arguments")
			expr = &CallExpr{
				Callee: expr,
				Args:   args,
			}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) parsePrimary() Expr {
	if p.match(TokenNumber) {
		t := p.previous()
		v, err := strconv.ParseFloat(t.Text, 64)
		if err != nil {
			panic(fmt.Errorf("invalid number '%s' at pos %d", t.Text, t.Pos))
		}
		return &NumberExpr{Value: v}
	}

	if p.match(TokenText) {
		t := p.previous()
		return &StringExpr{Value: t.Text}
	}

	if p.match(TokenTrue) {
		return &BoolExpr{Value: true}
	}
	if p.match(TokenFalse) {
		return &BoolExpr{Value: false}
	}

	if p.match(TokenNull) {
		return &NullExpr{}
	}

	if p.match(TokenIdentifier) {
		t := p.previous()
		return &IdentExpr{Name: t.Text}
	}

	if p.match(TokenLeftParen) {
		expr := p.parseExpression()
		p.consume(TokenRightParen, "expected ')' after expression")
		return expr
	}

	panic(fmt.Errorf("parse error at pos %d: expected expression", p.current().Pos))
}
