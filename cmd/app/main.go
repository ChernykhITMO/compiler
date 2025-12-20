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
	ScenarioFactorial Scenario = "factorial"
	ScenarioSort      Scenario = "sort"
	ScenarioPrimes    Scenario = "primes"
)

func getScenarioSource(s Scenario) string {
	switch s {
	case ScenarioFactorial:
		return srcFactorial
	case ScenarioSort:
		return srcSort
	case ScenarioPrimes:
		return srcPrimes
	default:
		log.Fatalf("unknown scenario %q", s)
		return ""
	}
}

const currentScenario = ScenarioSort

func main() {
	src := getScenarioSource(currentScenario)

	lexer := lexer.NewLexer(src)
	tokens := lexer.Tokenize()

	p := parser.NewParser(tokens)
	prog := p.ParseProgram()

	validator := semantics.NewASTValidator()
	checker := semantics.NewChecker()
	checker.Check(prog)
	errs := validator.Validate(prog)
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Printf("validate: [%s] %s\n", e.Type, e.Message)
		}
		log.Fatal("validation failed")
	}

	comp := backend.NewCompiler()
	mod, err := comp.CompileProgram(prog)
	if err != nil {
		log.Fatalf("compile error: %v", err)
	}

	vm := backend.NewVM(mod, true)

	startCall := time.Now()
	res, err := vm.Call("test", nil)
	endCall := time.Since(startCall)
	fmt.Printf("time vm.Call = %v\n", endCall)

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
}

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

    for (int i = 0; i < 10000; i = i + 1) {
		if (arr[i] != i + 1) {
			return 1
		}
		else {
			return 0
    	}
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
