package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/scanner"
)

var (
	ErrInvalid     = fmt.Errorf("invalid-token")
	ErrInvalidType = fmt.Errorf("invalid-argument-type")
	ErrMissing     = fmt.Errorf("missing-argument")
	Verbose        = false

	True = Boolean{value: true}
	Nil  = Boolean{value: false}

	boolstring = map[bool]string{
		true:  "t",
		false: "nil",
	}
)

type Object interface {
	String() string
	Value() any
}

type CanInt interface {
	Int() int64
}

type CanFloat interface {
	Float() float64
}

type CanBool interface {
	Bool() bool
}

type Boolean struct {
	value bool
}

func (o Boolean) String() string { return boolstring[o.value] }
func (o Boolean) Value() any     { return o.value }
func (o Boolean) Bool() bool     { return o.value }

type Symbol struct {
	value string
}

func (o Symbol) String() string { return fmt.Sprintf("%v", o.value) }
func (o Symbol) Value() any     { return o.value }

type Quoted struct {
	value any
}

func (o Quoted) String() string { return fmt.Sprintf("'%v", o.value) }
func (o Quoted) Value() any     { return o.value }

type Op struct {
	value string
}

func (o Op) String() string { return fmt.Sprintf("%q", o.value) }
func (o Op) Value() any     { return o.value }

type Integer struct {
	value int64
}

func (o Integer) String() string { return fmt.Sprint(o.value) }
func (o Integer) Value() any     { return o.value }
func (o Integer) Int() int64     { return o.value }
func (o Integer) Float() float64 { return float64(o.value) }
func (o Integer) Bool() Boolean  { return True }

type Float struct {
	value float64
}

func (o Float) String() string { return fmt.Sprint(o.value) }
func (o Float) Value() any     { return o.value }
func (o Float) Int() int64     { return int64(o.value) }
func (o Float) Float() float64 { return o.value }
func (o Float) Bool() Boolean  { return True }

type String struct {
	value string
}

func (o String) String() string { return o.value }
func (o String) Value() any     { return o.value }
func (o String) Bool() Boolean  { return True }

type List struct {
	items []any
}

func (o List) String() string {
	l := fmt.Sprint(o.items)
	if l[0] == '[' {
		return "(" + l[1:len(l)-1] + ")"
	}
	return l
}
func (o List) Value() any { return o.items }

func (o List) Item(i int) any {
	if i < 0 || i > len(o.items) {
		return nil
	}

	return o.items[i]
}

func (o List) Bool() bool {
	return len(o.items) > 0
}

func ident(v string) Object {
	switch v {
	case "t":
		return True

	case "nil":
		return Nil
	}

	return Symbol{value: v}
}

func quote(v any) any {
	switch v.(type) {
	case Symbol, List:
		if Verbose {
			fmt.Println("Quote", v)
		}
		return Quoted{value: v}
	}

	return v
}

type Parser struct {
	s scanner.Scanner
}

func NewParser(r io.Reader) *Parser {
	var p Parser

	p.s.Init(r)
	p.s.Whitespace = 0
	p.s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings
	return &p
}

func (p *Parser) SepNext() bool {
	switch p.s.Peek() {
	case ' ', '\r', '\n', '(', ')', scanner.EOF:
		return true
	}

	return false
}

func (p *Parser) Parse() (l []any, err error) {
	var neg bool
	var quoted bool

	maybequoted := func(v any) any {
		if quoted {
			quoted = false
			v = quote(v)
		}

		return v
	}

	appendtolist := func(v any) {
		l = append(l, maybequoted(v))
	}

	for tok := p.s.Scan(); tok != scanner.EOF; tok = p.s.Scan() {
		st := p.s.TokenText()

		if Verbose {
			fmt.Printf("%v: %v %q\n", p.s.Position, scanner.TokenString(tok), st)
		}

		switch tok {
		case '(':
			vv, err := p.Parse()
			if err != nil {
				return nil, err
			}

			appendtolist(List{items: vv})

		case ')':
			if quoted {
				appendtolist(Nil)
			}

			return

		case ' ', '\t', '\n', '\r':
			continue

		case scanner.Ident:
			var id string

			for {
				id += st

				if p.SepNext() {
					break
				}

				tok = p.s.Scan()
				st = p.s.TokenText()
			}

			appendtolist(ident(id))

		case scanner.String:
			appendtolist(String{value: st})

		case scanner.Int:
			i, _ := strconv.ParseInt(st, 10, 64)
			if neg {
				i = -i
				neg = false
			}
			appendtolist(Integer{value: i})

		case scanner.Float:
			f, _ := strconv.ParseFloat(st, 64)
			if neg {
				f = -f
				neg = false
			}
			appendtolist(Float{value: f})

		case '\'':
			if Verbose {
				fmt.Println("quote")
			}
			quoted = true

		case '+', '-', '/', '*', '%':
			if tok == '+' || tok == '-' {
				if n := p.s.Peek(); n == '.' || (n >= '0' && n <= '9') { // next token is a number
					neg = tok == '-'
					continue
				}
			}

			appendtolist(Op{value: st})

		default:
			if Verbose {
				fmt.Printf("UNKNOWN %v %q", scanner.TokenString(tok), st)
			}
			return nil, ErrInvalid
		}
	}

	return
}

type Call func(env *Env, args []any) any

var functions = map[string]Call{
	"print": func(env *Env, args []any) any {
		n, _ := fmt.Print(env.GetList(args)...)
		return n
	},
	"println": func(env *Env, args []any) any {
		n, _ := fmt.Println(env.GetList(args)...)
		return n
	},
	"quote": func(env *Env, args []any) any {
		if len(args) == 0 {
			return Nil
		}

		return quote(args[0])
	},
	"setq": func(env *Env, args []any) (ret any) {
		l := len(args)
		if l == 0 || l%2 != 0 {
			return ErrMissing
		}

		for i := 0; i < l; i += 2 {
			name, value := args[i+0], env.Get(args[i+1])
			ret = env.Put(name, value)
		}

		return
	},
	"not": func(env *Env, args []any) any {
		if len(args) == 0 {
			return True
		}

		v := env.Get(args[0])
		if b, ok := v.(CanBool); ok {
			return Boolean{value: !b.Bool()}
		}

		return Nil
	},
}

func callop(op Op, env *Env, args []any) any {
	if len(args) == 0 {
		if op.value == "+" {
			return 0
		}

		return ErrMissing
	}

	first := env.Get(args[0])

	if i, ok := first.(Integer); ok {
		v := i.value

		for _, a := range args[1:] {
			a = env.Get(a)

			ii, ok := a.(CanInt)
			if !ok {
				return ErrInvalidType
			}

			switch op.value {
			case "+":
				v += ii.Int()
			case "-":
				v -= ii.Int()
			case "*":
				v *= ii.Int()
			case "/":
				v /= ii.Int()
			case "%":
				v %= ii.Int()
			}
		}

		return Integer{value: v}
	} else if f, ok := first.(Float); ok {
		v := f.value

		for _, a := range args[1:] {
			a = env.Get(a)

			ii, ok := a.(CanFloat)
			if !ok {
				return ErrInvalidType
			}

			switch op.value {
			case "+":
				v += ii.Float()
			case "-":
				v -= ii.Float()
			case "*":
				v *= ii.Float()
			case "/":
				v /= ii.Float()
			case "%":
				v = float64(int64(v) % int64(ii.Float()))
			}
		}

		return Float{value: v}
	}

	return ErrInvalidType
}

type Env struct {
	vars map[string]any
	next *Env
}

func NewEnv() *Env {
	return &Env{vars: map[string]any{}}
}

func (e *Env) Put(o, value any) any {
	if s, ok := o.(Symbol); ok {
		e.vars[s.value] = value
	}

	return value
}

func (e *Env) Get(o any) any {
	if s, ok := o.(Symbol); ok {
		if v, ok := e.vars[s.value]; ok {
			return v
		}

		if e.next != nil {
			return e.next.Get(o)
		}

		return Nil
	}

	return o
}

func (e *Env) GetList(l []any) (el []any) {
	for _, v := range l {
		el = append(el, e.Get(v))
	}

	return
}

func Eval(v any, env *Env) any {
	switch t := v.(type) {
	case String:
		return t

	case Integer:
		return t

	case Float:
		return t

	case Boolean:
		return t

	case Quoted:
		return t.value

	case Symbol:
		return env.Get(t)

	case List:
		if len(t.items) == 0 {
			return Nil
		}
		switch i := t.items[0].(type) {
		case Symbol:
			if f, ok := functions[i.value]; ok {
				return f(env, t.items[1:])
			}
			return env.Get(i.value)

		case Op:
			return callop(i, env, t.items[1:])
		}
	}

	return nil
}

func main() {
	var expr string

	if len(os.Args) > 1 {
		expr = strings.Join(os.Args[1:], " ")
	} else {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println(err)
			return
		}

		expr = string(b)
	}

	p := NewParser(strings.NewReader(expr))

	l, _ := p.Parse()

	fmt.Println()

	env := NewEnv()

	for _, v := range l {
		fmt.Println(Eval(v, env))
	}
}
