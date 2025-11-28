package parser

import (
	"fmt"

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
		p.advance()
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

// ===== типы =====

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

// ===== программа / функции =====

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
		fn.ReturnType = "void"
	}

	fn.Body = p.parseBlock()
	p.match(frontend.TokenNewline) // опциональный \n после функции

	return fn
}

// ===== блок =====

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
	p.match(frontend.TokenNewline) // опциональный \n после блока

	return block
}
