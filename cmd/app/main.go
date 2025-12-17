package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ChernykhITMO/compiler/internal/backend"
	"github.com/ChernykhITMO/compiler/internal/frontend/lexer"
	"github.com/ChernykhITMO/compiler/internal/frontend/parser"
	"github.com/ChernykhITMO/compiler/internal/frontend/semantics"
)

type Scenario string

const (
	ScenarioGC               Scenario = "gc"
	ScenarioFactorial        Scenario = "factorial"
	ScenarioSort             Scenario = "sort"
	ScenarioPrimes           Scenario = "primes"
	ScenarioForBreakContinue Scenario = "forBreakContinue"
)

func getScenarioSource(s Scenario) string {
	switch s {
	case ScenarioGC:
		return srcGC
	case ScenarioFactorial:
		return srcFactorial
	case ScenarioSort:
		return srcSort
	case ScenarioPrimes:
		return srcPrimes
	case ScenarioForBreakContinue:
		return forBreakContinue
	default:
		log.Fatalf("unknown scenario %q", s)
		return ""
	}
}

const currentScenario = ScenarioPrimes

func main() {
	start := time.Now()

	src := getScenarioSource(currentScenario)

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
	vm.JitEnabled = true

	// test() без аргументов
	res, err := vm.Call("test", nil)
	if err != nil {
		log.Fatalf("vm error: %v", err)
	}

	switch currentScenario {
	case ScenarioFactorial:
		const want = 2432902008176640000
		if res.I != want {
			log.Fatalf("wrong result: want %d got %d", want, res.I)
		} else {
			fmt.Println("factorial success")
		}
	}

	fmt.Println("OK, test() returned", res.I)

	elapsed := time.Since(start)
	fmt.Println("time run:", elapsed.Seconds())
}

const forBreakContinue = `
function test() int{
	int a = 0
	int b = 5
	while (a<10){
		a = a + 1
		if a == 9{
			continue
		}
	}
	return a
}

function main() void{
}
`

// Стресс-тест gc
const srcGC = `
function main() void {
    int i = 0
    while (i < 100000) {
        int[] arr
        arr = new int[1000]
        arr[0] = i
        i = i + 1
    }
}

function test() int {
    main()
    return 0
}
`

// факториал
const srcFactorial = `
function main() void {
}

function fac(int n) int {
    if (n == 0) {
        return 1
    }
    else {
        return n * fac(n - 1)
    }
}

function test() int {
    int x = fac(20)
    return x
}
`

// сортировка
const srcSort = `
function main() void {
}

function bubbleSort(int[] arr, int n) void {
    int i = 0
    while (i < n) {
        int j = 0
        while (j < n - 1) {
            if (arr[j] > arr[j + 1]) {
                int tmp = arr[j]
                arr[j] = arr[j + 1]
                arr[j + 1] = tmp
            }
            j = j + 1
        }
        i = i + 1
    }
}

function test() int {
    int n = 10000
    int[] arr
    arr = new int[n]

    int i = 0
    while (i < n) {
        arr[i] = n - i
        i = i + 1
    }

    bubbleSort(arr, n)

    if (arr[0] == 1 && arr[n - 1] == n) {
        return 1
    }
    else {
        return 0
    }
}
`

// Простые числа
const srcPrimes = `
function main() void {
}

function test() int {
    int N = 100000

    bool[] isPrime
    isPrime = new bool[N + 1]

    int i = 0
    while (i <= N) {
        isPrime[i] = true
        i = i + 1
    }

    isPrime[0] = false
    isPrime[1] = false

    int p = 2
    while (p * p <= N) {
        if (isPrime[p]) {
            int multiple = p * p
            while (multiple <= N) {
                isPrime[multiple] = false
                multiple = multiple + p
            }
        }
        p = p + 1
    }

    int count = 0
    int k = 2
    while (k <= N) {
        if (isPrime[k]) {
            count = count + 1
        }
        k = k + 1
    }

    return count
}
`
