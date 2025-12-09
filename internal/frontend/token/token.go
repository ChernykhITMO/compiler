package token

type TokenType int

const (
	TokenInvalid TokenType = iota
	TokenNumber
	TokenText

	TokenInt
	TokenFloat
	TokenString
	TokenBool
	TokenVoid
	TokenChar

	TokenFunction
	TokenIf
	TokenElse
	TokenWhile
	TokenFor
	TokenReturn
	TokenBreak
	TokenContinue

	TokenNull
	TokenTrue
	TokenFalse
	TokenIdentifier

	TokenAssign
	TokenEqual
	TokenNotEqual
	TokenLess
	TokenLessEqual
	TokenGreater
	TokenGreaterEqual
	TokenNot
	TokenAnd
	TokenOr

	TokenPlus
	TokenMinus
	TokenMultiply
	TokenDivide
	TokenModulo
	TokenPower

	TokenLeftParen
	TokenRightParen
	TokenLeftBrace
	TokenRightBrace
	TokenLeftBracket
	TokenRightBracket
	TokenComma
	TokenNewline
	TokenEnd
	TokenSemicolon
)

type Token struct {
	Type TokenType
	Text string
	Pos  int
}
