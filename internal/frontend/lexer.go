package frontend

import (
	"unicode"
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

func (l *Lexer) readNumber() Token {
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

	return Token{Type: TokenNumber, Text: string(buf), Pos: start}
}

func (l *Lexer) readString() Token {
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
		return Token{Type: TokenText, Text: string(buf), Pos: start}
	}
	return Token{Type: TokenInvalid, Text: string(buf), Pos: start}
}

func (l *Lexer) readIdentifier() Token {
	start := l.position
	var buf []byte

	c := l.currentChar()
	if !unicode.IsLetter(rune(c)) && c != '_' {
		return Token{Type: TokenInvalid, Text: string(c), Pos: start}
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
		return Token{Type: TokenInt, Text: ident, Pos: start}
	case "float":
		return Token{Type: TokenFloat, Text: ident, Pos: start}
	case "string":
		return Token{Type: TokenString, Text: ident, Pos: start}
	case "bool":
		return Token{Type: TokenBool, Text: ident, Pos: start}
	case "void":
		return Token{Type: TokenVoid, Text: ident, Pos: start}
	case "char":
		return Token{Type: TokenChar, Text: ident, Pos: start}
	case "function":
		return Token{Type: TokenFunction, Text: ident, Pos: start}
	case "if":
		return Token{Type: TokenIf, Text: ident, Pos: start}
	case "else":
		return Token{Type: TokenElse, Text: ident, Pos: start}
	case "while":
		return Token{Type: TokenWhile, Text: ident, Pos: start}
	case "for":
		return Token{Type: TokenFor, Text: ident, Pos: start}
	case "return":
		return Token{Type: TokenReturn, Text: ident, Pos: start}
	case "null":
		return Token{Type: TokenNull, Text: ident, Pos: start}
	case "true":
		return Token{Type: TokenTrue, Text: ident, Pos: start}
	case "false":
		return Token{Type: TokenFalse, Text: ident, Pos: start}
	default:
		return Token{Type: TokenIdentifier, Text: ident, Pos: start}
	}
}

func (l *Lexer) Tokenize() []Token {
	var tokens []Token

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
				tokens = append(tokens, Token{Type: TokenEqual, Text: "==", Pos: l.position - 2})
			} else {
				tokens = append(tokens, Token{Type: TokenAssign, Text: "=", Pos: l.position - 1})
			}
		case '!':
			l.skipChar()
			if l.currentChar() == '=' {
				l.skipChar()
				tokens = append(tokens, Token{Type: TokenNotEqual, Text: "!=", Pos: l.position - 2})
			} else {
				tokens = append(tokens, Token{Type: TokenNot, Text: "!", Pos: l.position - 1})
			}
		case '<':
			l.skipChar()
			if l.currentChar() == '=' {
				l.skipChar()
				tokens = append(tokens, Token{Type: TokenLessEqual, Text: "<=", Pos: l.position - 2})
			} else {
				tokens = append(tokens, Token{Type: TokenLess, Text: "<", Pos: l.position - 1})
			}
		case '>':
			l.skipChar()
			if l.currentChar() == '=' {
				l.skipChar()
				tokens = append(tokens, Token{Type: TokenGreaterEqual, Text: ">=", Pos: l.position - 2})
			} else {
				tokens = append(tokens, Token{Type: TokenGreater, Text: ">", Pos: l.position - 1})
			}
		case '&':
			l.skipChar()
			if l.currentChar() == '&' {
				l.skipChar()
				tokens = append(tokens, Token{Type: TokenAnd, Text: "&&", Pos: l.position - 2})
			} else {
				tokens = append(tokens, Token{Type: TokenInvalid, Text: "&", Pos: l.position - 1})
			}
		case '|':
			l.skipChar()
			if l.currentChar() == '|' {
				l.skipChar()
				tokens = append(tokens, Token{Type: TokenOr, Text: "||", Pos: l.position - 2})
			} else {
				tokens = append(tokens, Token{Type: TokenInvalid, Text: "|", Pos: l.position - 1})
			}
		case '+':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenPlus, Text: "+", Pos: l.position - 1})
		case '-':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenMinus, Text: "-", Pos: l.position - 1})
		case '*':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenMultiply, Text: "*", Pos: l.position - 1})
		case '/':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenDivide, Text: "/", Pos: l.position - 1})
		case '%':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenModulo, Text: "%", Pos: l.position - 1})
		case '^':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenPower, Text: "^", Pos: l.position - 1})
		case '(':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenLeftParen, Text: "(", Pos: l.position - 1})
		case ')':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenRightParen, Text: ")", Pos: l.position - 1})
		case '{':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenLeftBrace, Text: "{", Pos: l.position - 1})
		case '}':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenRightBrace, Text: "}", Pos: l.position - 1})
		case '[':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenLeftBracket, Text: "[", Pos: l.position - 1})
		case ']':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenRightBracket, Text: "]", Pos: l.position - 1})
		case ',':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenComma, Text: ",", Pos: l.position - 1})
		case '\n':
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenNewline, Text: "\\n", Pos: l.position - 1})
		default:
			l.skipChar()
			tokens = append(tokens, Token{Type: TokenInvalid, Text: string(c), Pos: l.position - 1})
		}
	}

	tokens = append(tokens, Token{Type: TokenEnd, Text: "", Pos: l.position})
	return tokens
}
