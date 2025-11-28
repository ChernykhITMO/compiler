package main

import (
	"fmt"

	"github.com/ChernykhITMO/compiler/internal/frontend"
	"github.com/ChernykhITMO/compiler/internal/frontend/parser"
)

func main() {
	src := `
function main() void {
    int x = 2 * 3
    if (x > 5) {
        x = 7
    }
}
`
	lexer := frontend.NewLexer(src)
	tokens := lexer.Tokenize()

	p := parser.NewParser(tokens)
	prog := p.ParseProgram()
	parser.PrintProgram(prog)
	fmt.Println("Total statements:", CountStatements(prog))
}

func CountStatements(prog *frontend.Program) int {
	total := 0
	for _, fn := range prog.Functions {
		total += countStmtBlock(fn.Body)
	}
	return total
}

func countStmtBlock(block *frontend.BlockStmt) int {
	if block == nil {
		return 0
	}
	total := 0

	for _, stmt := range block.Statements {
		total++

		switch s := stmt.(type) {

		case *frontend.IfStmt:
			total += countStmtBlock(s.ThenBlock)
			total += countStmtBlock(s.ElseBlock)

		case *frontend.WhileStmt:
			total += countStmtBlock(s.Body)

		}
	}
	return total
}
