package gisp

import (
	"fmt"
	"strings"
	"testing"
)

// parseForms parses source text and fatals on error.
func parseForms(tb testing.TB, src string) []any {
	tb.Helper()
	p := NewParser(strings.NewReader(src))
	forms, err := p.Parse()
	if err != nil {
		tb.Fatalf("parse error: %v", err)
	}
	return forms
}

// prepEnv parses and evaluates src into a fresh env.
func prepEnv(tb testing.TB, src string) *Env {
	tb.Helper()
	env := NewEnv(nil)
	for _, f := range parseForms(tb, src) {
		Eval(env, f)
	}
	return env
}

// --- Tests ------------------------------------------------------------------

// fib(n): 1-indexed Fibonacci: fib(1)=0, fib(2)=1, fib(3)=1, fib(10)=34, fib(20)=4181
const fibDef = `
(setq fib (lambda (n)
  (if (= n 1) 0
      (= n 2) 1
      (+ (fib (- n 1)) (fib (- n 2))))))
`

// sum(n) = 1+2+...+n  (iterative, while loop)
const sumDef = `
(setq sum (lambda (n)
  (let (i acc)
    (setq i 1)
    (setq acc 0)
    (while (<= i n)
      (setq acc (+ acc i))
      (setq i (+ i 1)))
    acc)))
`

func TestFib(t *testing.T) {
	env := prepEnv(t, fibDef)
	cases := []struct {
		n    int
		want int64
	}{
		{1, 0}, {2, 1}, {3, 1}, {4, 2}, {5, 3}, {6, 5}, {10, 34}, {15, 377}, {20, 4181},
	}
	for _, tc := range cases {
		call := parseForms(t, fmt.Sprintf("(fib %d)", tc.n))
		got := AsInt(Eval(env, call[0]), -1)
		if got != tc.want {
			t.Errorf("fib(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestArithLoop(t *testing.T) {
	env := prepEnv(t, sumDef)
	call := parseForms(t, "(sum 100)")
	got := AsInt(Eval(env, call[0]), -1)
	if got != 5050 {
		t.Errorf("sum(100) = %d, want 5050", got)
	}
}

func TestCallLambda(t *testing.T) {
	env := prepEnv(t, "(setq add (lambda (a b) (+ a b)))")
	call := parseForms(t, "(add 3 4)")
	got := AsInt(Eval(env, call[0]), -1)
	if got != 7 {
		t.Errorf("add(3, 4) = %d, want 7", got)
	}
}

func TestEnvLookupChain(t *testing.T) {
	root := NewEnv(nil)
	root.PutLocal(Symbol{value: "x"}, Integer{value: 99})
	inner := NewEnv(NewEnv(root))
	got := AsInt(inner.Get(Symbol{value: "x"}), -1)
	if got != 99 {
		t.Errorf("env chain lookup: got %d, want 99", got)
	}
}

func TestStringAppend(t *testing.T) {
	env := NewEnv(nil)
	call := parseForms(t, `(append "foo" "bar" "baz")`)
	got := AsString(Eval(env, call[0]), "")
	if got != "foobarbaz" {
		t.Errorf("append = %q, want %q", got, "foobarbaz")
	}
}

// --- Benchmarks -------------------------------------------------------------

func BenchmarkFib15(b *testing.B) {
	env := prepEnv(b, fibDef)
	call := parseForms(b, "(fib 15)")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(env, call[0])
	}
}

func BenchmarkFib20(b *testing.B) {
	env := prepEnv(b, fibDef)
	call := parseForms(b, "(fib 20)")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(env, call[0])
	}
}

// BenchmarkLoopSum exercises while, setq, arithmetic, and let.
func BenchmarkLoopSum1000(b *testing.B) {
	env := prepEnv(b, sumDef)
	call := parseForms(b, "(sum 1000)")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(env, call[0])
	}
}

// BenchmarkCallLambda isolates the cost of a single trivial lambda call.
func BenchmarkCallLambda(b *testing.B) {
	env := prepEnv(b, "(setq id (lambda (x) x))")
	call := parseForms(b, "(id 42)")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(env, call[0])
	}
}

// BenchmarkEnvLookup measures variable lookup through 3 nested env frames.
func BenchmarkEnvLookup(b *testing.B) {
	root := NewEnv(nil)
	root.PutLocal(Symbol{value: "x"}, Integer{value: 42})
	inner := NewEnv(NewEnv(root))
	xSym := Symbol{value: "x"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inner.Get(xSym)
	}
}

// BenchmarkParse measures parsing cost alone (no evaluation).
func BenchmarkParse(b *testing.B) {
	src := fibDef + sumDef
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := NewParser(strings.NewReader(src))
		if _, err := p.Parse(); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkStringAppend measures string concatenation via the append builtin.
func BenchmarkStringAppend(b *testing.B) {
	env := NewEnv(nil)
	call := parseForms(b, `(append "hello" " " "world")`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(env, call[0])
	}
}

// BenchmarkArith measures a repeated arithmetic expression.
func BenchmarkArith(b *testing.B) {
	env := NewEnv(nil)
	call := parseForms(b, "(+ 1 2 3 4 5 6 7 8 9 10)")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(env, call[0])
	}
}
