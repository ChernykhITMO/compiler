package lexer

import (
	"unicode"

	"github.com/ChernykhITMO/compiler/internal/frontend/token"
)

type Lexer struct {
	code     string
	position int
}

func NewLexer(src string) *Lexer {
	return &Lexer{
		code:     src,
		position: 0}
}

func (l *Lexer) currentChar() byte {
	if l.position < len(l.code) {
		return l.code[l.position]
	}
	return 0
}

func (l *Lexer) nextChar() byte {
	if l.position+1 < len(l.code) {
		return l.code[l.position+1]
	}
	return 0
}

func (l *Lexer) skipChar() {
	if l.position < len(l.code) {
		l.position++
	}
}

func (l *Lexer) skipWhitespace() {
	for {
		c := l.currentChar()
		if c == ' ' || c == '\t' {
			l.skipChar()
		} else {
			break
		}
	}
}

func (l *Lexer) readNumber() token.Token {
	start := l.position
	var buf []byte

	for c := l.currentChar(); unicode.IsDigit(rune(c)); c = l.currentChar() {
		buf = append(buf, c)
		l.skipChar()
	}

	if l.currentChar() == '.' && unicode.IsDigit(rune(l.nextChar())) {
		buf = append(buf, l.currentChar())
		l.skipChar()
		for c := l.currentChar(); unicode.IsDigit(rune(c)); c = l.currentChar() {
			buf = append(buf, c)
			l.skipChar()
		}
	}

	return token.Token{Type: token.TokenNumber, Text: string(buf), Pos: start}
}

func (l *Lexer) readString() token.Token {
	start := l.position
	l.skipChar()

	var buf []byte
	for {
		c := l.currentChar()
		if c == 0 || c == '\n' || c == '"' {
			break
		}
		buf = append(buf, c)
		l.skipChar()
	}

	if l.currentChar() == '"' {
		l.skipChar()
		return token.Token{Type: token.TokenText, Text: string(buf), Pos: start}
	}
	return token.Token{Type: token.TokenInvalid, Text: string(buf), Pos: start}
}

func (l *Lexer) readIdentifier() token.Token {
	start := l.position
	var buf []byte

	c := l.currentChar()
	if !unicode.IsLetter(rune(c)) && c != '_' {
		return token.Token{Type: token.TokenInvalid, Text: string(c), Pos: start}
	}
	buf = append(buf, c)
	l.skipChar()

	for {
		c = l.currentChar()
		if unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '_' {
			buf = append(buf, c)
			l.skipChar()
		} else {
			break
		}
	}

	ident := string(buf)

	switch ident {
	case "int":
		return token.Token{Type: token.TokenInt, Text: ident, Pos: start}
	case "float":
		return token.Token{Type: token.TokenFloat, Text: ident, Pos: start}
	case "string":
		return token.Token{Type: token.TokenString, Text: ident, Pos: start}
	case "bool":
		return token.Token{Type: token.TokenBool, Text: ident, Pos: start}
	case "void":
		return token.Token{Type: token.TokenVoid, Text: ident, Pos: start}
	case "char":
		return token.Token{Type: token.TokenChar, Text: ident, Pos: start}
	case "function":
		return token.Token{Type: token.TokenFunction, Text: ident, Pos: start}
	case "if":
		return token.Token{Type: token.TokenIf, Text: ident, Pos: start}
	case "else":
		return token.Token{Type: token.TokenElse, Text: ident, Pos: start}
	case "while":
		return token.Token{Type: token.TokenWhile, Text: ident, Pos: start}
	case "for":
		return token.Token{Type: token.TokenFor, Text: ident, Pos: start}
	case "return":
		return token.Token{Type: token.TokenReturn, Text: ident, Pos: start}
	case "null":
		return token.Token{Type: token.TokenNull, Text: ident, Pos: start}
	case "true":
		return token.Token{Type: token.TokenTrue, Text: ident, Pos: start}
	case "false":
		return token.Token{Type: token.TokenFalse, Text: ident, Pos: start}
	case "break":
		return token.Token{Type: token.TokenBreak, Text: ident, Pos: start}
	case "continue":
		return token.Token{Type: token.TokenContinue, Text: ident, Pos: start}
	default:
		return token.Token{Type: token.TokenIdentifier, Text: ident, Pos: start}
	}
}

func (l *Lexer) Tokenize() []token.Token {
	var tokens []token.Token

	for {
		l.skipWhitespace()

		c := l.currentChar()
		if c == 0 {
			break
		}

		if unicode.IsDigit(rune(c)) {
			tokens = append(tokens, l.readNumber())
			continue
		}

		if c == '"' {
			tokens = append(tokens, l.readString())
			continue
		}

		if unicode.IsLetter(rune(c)) || c == '_' {
			tokens = append(tokens, l.readIdentifier())
			continue
		}

		switch c {
		case '=':
			l.skipChar()
			if l.currentChar() == '=' {
				l.skipChar()
				tokens = append(tokens, token.Token{Type: token.TokenEqual, Text: "==", Pos: l.position - 2})
			} else {
				tokens = append(tokens, token.Token{Type: token.TokenAssign, Text: "=", Pos: l.position - 1})
			}
		case ';':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenSemicolon, Text: ";", Pos: l.position - 1})
		case '!':
			l.skipChar()
			if l.currentChar() == '=' {
				l.skipChar()
				tokens = append(tokens, token.Token{Type: token.TokenNotEqual, Text: "!=", Pos: l.position - 2})
			} else {
				tokens = append(tokens, token.Token{Type: token.TokenNot, Text: "!", Pos: l.position - 1})
			}
		case '<':
			l.skipChar()
			if l.currentChar() == '=' {
				l.skipChar()
				tokens = append(tokens, token.Token{Type: token.TokenLessEqual, Text: "<=", Pos: l.position - 2})
			} else {
				tokens = append(tokens, token.Token{Type: token.TokenLess, Text: "<", Pos: l.position - 1})
			}
		case '>':
			l.skipChar()
			if l.currentChar() == '=' {
				l.skipChar()
				tokens = append(tokens, token.Token{Type: token.TokenGreaterEqual, Text: ">=", Pos: l.position - 2})
			} else {
				tokens = append(tokens, token.Token{Type: token.TokenGreater, Text: ">", Pos: l.position - 1})
			}
		case '&':
			l.skipChar()
			if l.currentChar() == '&' {
				l.skipChar()
				tokens = append(tokens, token.Token{Type: token.TokenAnd, Text: "&&", Pos: l.position - 2})
			} else {
				tokens = append(tokens, token.Token{Type: token.TokenInvalid, Text: "&", Pos: l.position - 1})
			}
		case '|':
			l.skipChar()
			if l.currentChar() == '|' {
				l.skipChar()
				tokens = append(tokens, token.Token{Type: token.TokenOr, Text: "||", Pos: l.position - 2})
			} else {
				tokens = append(tokens, token.Token{Type: token.TokenInvalid, Text: "|", Pos: l.position - 1})
			}
		case '+':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenPlus, Text: "+", Pos: l.position - 1})
		case '-':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenMinus, Text: "-", Pos: l.position - 1})
		case '*':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenMultiply, Text: "*", Pos: l.position - 1})
		case '/':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenDivide, Text: "/", Pos: l.position - 1})
		case '%':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenModulo, Text: "%", Pos: l.position - 1})
		case '^':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenPower, Text: "^", Pos: l.position - 1})
		case '(':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenLeftParen, Text: "(", Pos: l.position - 1})
		case ')':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenRightParen, Text: ")", Pos: l.position - 1})
		case '{':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenLeftBrace, Text: "{", Pos: l.position - 1})
		case '}':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenRightBrace, Text: "}", Pos: l.position - 1})
		case '[':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenLeftBracket, Text: "[", Pos: l.position - 1})
		case ']':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenRightBracket, Text: "]", Pos: l.position - 1})
		case ',':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenComma, Text: ",", Pos: l.position - 1})
		case '\n':
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenNewline, Text: "\\n", Pos: l.position - 1})
		default:
			l.skipChar()
			tokens = append(tokens, token.Token{Type: token.TokenInvalid, Text: string(c), Pos: l.position - 1})
		}
	}

	tokens = append(tokens, token.Token{Type: token.TokenEnd, Text: "", Pos: l.position})
	return tokens
}
