package main

import (
	"fmt"
	"github.com/ChernykhITMO/compiler/internal/backend"
	"github.com/ChernykhITMO/compiler/internal/bytecode"
	"github.com/ChernykhITMO/compiler/internal/frontend/ast"
	"github.com/ChernykhITMO/compiler/internal/frontend/lexer"
	"github.com/ChernykhITMO/compiler/internal/frontend/parser"
	"github.com/ChernykhITMO/compiler/internal/frontend/semantics"
	"log"
	"math"
)

func main() {
	src := `
function main() void {
}

function test() int{
	int a = fac(20)
	return a
}

function fac(int a) int{
	if a == 1 {
		return a
	}
	else {
		return a * fac(a - 1)
	}
}
`

	// 1) лексер + парсер
	lexer := lexer.NewLexer(src)
	tokens := lexer.Tokenize()

	p := parser.NewParser(tokens)
	prog := p.ParseProgram()

	// 2) валидация (по желанию выведи ошибки)
	validator := semantics.NewASTValidator()
	errs := validator.Validate(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Printf("validate: [%s] %s\n", e.Type, e.Message)
		}
		log.Fatal("validation failed")
	}

	// 3) компиляция
	comp := backend.NewCompiler()
	mod, err := comp.CompileProgram(prog)
	if err != nil {
		log.Fatalf("compile error: %v", err)
	}

	// 4) запуск VM и вызов test()
	vm := backend.NewVM(mod)

	// test() без аргументов
	res, err := vm.Call("test", nil)
	if err != nil {
		log.Fatalf("vm error: %v", err)
	}

	// 5) проверяем, что вернулось 5
	if res.Kind != bytecode.ValInt {
		log.Fatalf("expected float, got kind=%v", res.Kind)
	}
	if math.Abs(float64(res.I)) > 1e-9 {
		log.Fatalf("wrong result: want 1, got %v", res.I)
	}

	fmt.Println("OK, test() returned", res.I)
}

func dumpFunction(fn *bytecode.FunctionInfo) {
	ch := &fn.Chunk
	fmt.Printf("  code bytes: %v\n", ch.Code)
	fmt.Printf("  consts (%d):\n", len(ch.Constants))
	for i, c := range ch.Constants {
		fmt.Printf("    %d: %+v\n", i, c)
	}
}

func CountStatements(prog *ast.Program) int {
	total := 0
	for _, fn := range prog.Functions {
		total += countStmtBlock(fn.Body)
	}
	return total
}

func countStmtBlock(block *ast.BlockStmt) int {
	if block == nil {
		return 0
	}
	total := 0

	for _, stmt := range block.Statements {
		total++

		switch s := stmt.(type) {

		case *ast.IfStmt:
			total += countStmtBlock(s.ThenBlock)
			total += countStmtBlock(s.ElseBlock)

		case *ast.WhileStmt:
			total += countStmtBlock(s.Body)

		}
	}
	return total
}
