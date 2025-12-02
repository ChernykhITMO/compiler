package main

import (
	"fmt"

	"github.com/ChernykhITMO/compiler/internal/frontend"
	"github.com/ChernykhITMO/compiler/internal/frontend/parser"
)

func main() {
	src := `
function main() void {
    for (int i = 0; i < 10; i = i + 1) {
        if (i == 5) {
            break
        }
        print(i)
    }
    
    int x = 0
    while (x < 5) {
        x = x + 1
        if (x == 3) {
            continue
        }
        print(x)
    }
}
`
	lexer := frontend.NewLexer(src)
	tokens := lexer.Tokenize()

	p := parser.NewParser(tokens)
	prog := p.ParseProgram()

	fmt.Println("AST:")
	parser.PrintProgram(prog)

	fmt.Println("\nValidation:")
	validator := frontend.NewASTValidator()
	errors := validator.Validate(prog)

	if len(errors) == 0 {
		fmt.Println("No validation errors found")
	} else {
		fmt.Println("Validation errors:")
		for _, err := range errors {
			fmt.Printf("  [%s] %s\n", err.Type, err.Message)
		}
	}

	fmt.Println("\nTotal statements:", CountStatements(prog))
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
