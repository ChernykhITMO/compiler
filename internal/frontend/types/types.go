package types

import "github.com/ChernykhITMO/compiler/internal/frontend/token"

type BasicType int

const (
	TypeInvalid BasicType = iota
	TypeInt
	TypeFloat
	TypeBool
	TypeString
	TypeChar
	TypeVoid
	TypeNull
)

type Type struct {
	Kind BasicType
}

func (t Type) String() string {
	switch t.Kind {
	case TypeInt:
		return "int"
	case TypeFloat:
		return "float"
	case TypeBool:
		return "bool"
	case TypeString:
		return "string"
	case TypeChar:
		return "char"
	case TypeVoid:
		return "void"
	case TypeNull:
		return "null"
	default:
		return "invalid"
	}
}

func TypeFromToken(tt token.TokenType) Type {
	switch tt {
	case token.TokenInt:
		return Type{Kind: TypeInt}
	case token.TokenFloat:
		return Type{Kind: TypeFloat}
	case token.TokenBool:
		return Type{Kind: TypeBool}
	case token.TokenString:
		return Type{Kind: TypeString}
	case token.TokenChar:
		return Type{Kind: TypeChar}
	case token.TokenVoid:
		return Type{Kind: TypeVoid}
	default:
		return Type{Kind: TypeInvalid}
	}
}
