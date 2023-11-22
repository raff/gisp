package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/scanner"
	"time"
)

var (
	ErrInvalid     = fmt.Errorf("invalid-token")
	ErrInvalidType = fmt.Errorf("invalid-parameter-type")
	ErrMissing     = fmt.Errorf("missing-parameter")
	Verbose, _     = strconv.ParseBool(os.Getenv("VERBOSE"))

	True = Boolean{value: true}
	Nil  = Boolean{value: false}

	boolstring = map[bool]string{
		false: "nil",
		true:  "true",
	}

	boolint = map[bool]int{
		false: 0,
		true:  1,
	}

	primitives map[string]Call
)

type Call func(env *Env, args []any) any

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

type CanCond interface {
	Eq(v any) bool
	Lt(v any) bool
	Leq(v any) bool
	Gt(v any) bool
	Geq(v any) bool
}

type Boolean struct {
	value bool
}

func (o Boolean) String() string { return boolstring[o.value] }
func (o Boolean) Value() any     { return o.value }
func (o Boolean) Bool() bool     { return o.value }

func (o Boolean) Eq(v any) bool {
	if b, ok := v.(CanBool); ok {
		return o.value == b.Bool()
	}

	return !o.value
}

func (o Boolean) Lt(v any) bool {
	if b, ok := v.(CanBool); ok {
		return boolint[o.value] < boolint[b.Bool()]
	}

	return false
}

func (o Boolean) Leq(v any) bool {
	if b, ok := v.(CanBool); ok {
		return boolint[o.value] <= boolint[b.Bool()]
	}

	return boolint[o.value] <= boolint[false]
}

func (o Boolean) Gt(v any) bool {
	if b, ok := v.(CanBool); ok {
		return boolint[o.value] > boolint[b.Bool()]
	}

	return o.value
}

func (o Boolean) Geq(v any) bool {
	if b, ok := v.(CanBool); ok {
		return boolint[o.value] >= boolint[b.Bool()]
	}

	return true
}

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

type Cond struct {
	value string
}

func (o Cond) String() string { return fmt.Sprintf("%q", o.value) }
func (o Cond) Value() any     { return o.value }

type Integer struct {
	value int64
}

func (o Integer) String() string { return fmt.Sprint(o.value) }
func (o Integer) Value() any     { return o.value }
func (o Integer) Int() int64     { return o.value }
func (o Integer) Float() float64 { return float64(o.value) }
func (o Integer) Bool() bool     { return true }

func (o Integer) Eq(v any) bool {
	if i, ok := v.(CanInt); ok {
		return o.value == i.Int()
	}

	return false
}

func (o Integer) Lt(v any) bool {
	if i, ok := v.(CanInt); ok {
		return o.value < i.Int()
	}

	return false
}

func (o Integer) Leq(v any) bool {
	if i, ok := v.(CanInt); ok {
		return o.value <= i.Int()
	}

	return false
}

func (o Integer) Gt(v any) bool {
	if i, ok := v.(CanInt); ok {
		return o.value > i.Int()
	}

	return false
}

func (o Integer) Geq(v any) bool {
	if i, ok := v.(CanInt); ok {
		return o.value >= i.Int()
	}

	return false
}

type Float struct {
	value float64
}

func (o Float) String() string { return fmt.Sprint(o.value) }
func (o Float) Value() any     { return o.value }
func (o Float) Int() int64     { return int64(o.value) }
func (o Float) Float() float64 { return o.value }
func (o Float) Bool() bool     { return true }

func (o Float) Eq(v any) bool {
	if f, ok := v.(CanFloat); ok {
		return o.value == f.Float()
	}

	return false
}

func (o Float) Lt(v any) bool {
	if f, ok := v.(CanFloat); ok {
		return o.value < f.Float()
	}

	return false
}

func (o Float) Leq(v any) bool {
	if f, ok := v.(CanFloat); ok {
		return o.value <= f.Float()
	}

	return false
}

func (o Float) Gt(v any) bool {
	if f, ok := v.(CanFloat); ok {
		return o.value > f.Float()
	}

	return false
}

func (o Float) Geq(v any) bool {
	if f, ok := v.(CanFloat); ok {
		return o.value >= f.Float()
	}

	return false
}

type String struct {
	value string
}

func (o String) String() string { return o.value }
func (o String) Value() any     { return o.value }
func (o String) Bool() bool     { return true }

func (o String) Eq(v any) bool {
	if s, ok := v.(String); ok {
		return o.value == s.value
	}

	return false
}

func (o String) Lt(v any) bool {
	if s, ok := v.(String); ok {
		return o.value < s.value
	}

	return false
}

func (o String) Leq(v any) bool {
	if s, ok := v.(String); ok {
		return o.value <= s.value
	}

	return false
}

func (o String) Gt(v any) bool {
	if s, ok := v.(String); ok {
		return o.value > s.value
	}

	return false
}

func (o String) Geq(v any) bool {
	if s, ok := v.(String); ok {
		return o.value >= s.value
	}

	return false
}

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

type Lambda struct {
	args []any
	body []any
}

func (o Lambda) String() string { return fmt.Sprintf("(lambda %v %v)", o.args, o.body) }
func (o Lambda) Value() any     { return Nil }

func ident(v string) Object {
	switch v {
	case "true":
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

		case '<':
			if p.s.Peek() == '=' {
				p.s.Next()
				appendtolist(Cond{value: "<="})
			} else {
				appendtolist(Cond{value: "<"})
			}

		case '>':
			if p.s.Peek() == '=' {
				p.s.Next()
				appendtolist(Cond{value: ">="})
			} else {
				appendtolist(Cond{value: ">"})
			}

		case '=':
			appendtolist(Cond{value: "="})

		default:
			if Verbose {
				fmt.Printf("UNKNOWN %v %q", scanner.TokenString(tok), st)
			}
			return nil, ErrInvalid
		}
	}

	return
}

func init() {
	// primitive functions

	primitives = map[string]Call{
		//
		// print args
		//
		"print": func(env *Env, args []any) any {
			n, _ := fmt.Print(env.GetList(args)...)
			return n
		},

		//
		// println args...
		//
		"println": func(env *Env, args []any) any {
			n, _ := fmt.Println(env.GetList(args)...)
			return n
		},

		//
		// format fmt args...
		//
		"format": func(env *Env, args []any) any {
			if len(args) == 0 {
				return ErrMissing
			}

			f, args := args[0], env.GetValues(args[1:])

			sfmt, ok := f.(String)
			if !ok {
				return ErrInvalidType
			}

			return fmt.Sprintf(sfmt.String(), args...)
		},

		//
		// sleep ms
		//
		"sleep": func(env *Env, args []any) any {
			if len(args) == 0 {
				return ErrMissing
			}

			v := env.Get(args[0])

			if tm, ok := v.(CanInt); ok {
				time.Sleep(time.Millisecond * time.Duration(tm.Int()))
				return tm
			}

			return ErrInvalidType
		},

		//
		// quote symbol
		//
		"quote": func(env *Env, args []any) any {
			if len(args) == 0 {
				return Nil
			}

			return quote(args[0])
		},

		//
		// setq name value
		//
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

		//
		// not bool
		//
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

		//
		// or bool...
		//
		"or": func(env *Env, args []any) any {
			for _, arg := range args {
				v := env.Get(arg)
				if b, ok := v.(CanBool); ok {
					if b.Bool() {
						return True
					}
				}
			}

			return Nil
		},

		//
		// and bool...
		//
		"and": func(env *Env, args []any) any {
			for _, arg := range args {
				v := env.Get(arg)
				if b, ok := v.(CanBool); ok {
					if !b.Bool() {
						return Nil
					}
				}
			}

			return True
		},

		//
		// if cond then [cond then...] else
		//
		"if": func(env *Env, args []any) any {
			if len(args) == 0 {
				return Nil
			}

			var barg any

			for {
				barg, args = env.Get(args[0]), args[1:]
				bval, ok := barg.(CanBool)
				if !ok {
					return barg
				}

				// if
				if bval.Bool() {
					if len(args) == 0 {
						return barg
					}

					return env.Get(args[0])
				}

				// else
				switch len(args) {
				case 0:
					return Nil

				case 1:
					return env.Get(args[0])
				}

				// else if
				//   shift and continue
				args = args[1:]
			}
		},

		//
		// while cond block
		//
		"while": func(env *Env, args []any) (ret any) {
			if len(args) == 0 {
				return Nil
			}

			cond, args := args[0], args[1:]

			for {
				bval, ok := env.Get(cond).(CanBool)
				if Verbose {
					fmt.Println(cond, bval)
				}

				if !ok || !bval.Bool() {
					break
				}

				for _, v := range args {
					if Verbose {
						fmt.Println("  ", v)
					}
					ret = Eval(env, v)
				}
			}

			return
		},

		//
		// begin stmt...
		//
		"begin": func(env *Env, args []any) (ret any) {
			for _, v := range args {
				if Verbose {
					fmt.Println("  ", v)
				}
				ret = Eval(env, v)
			}

			return
		},

		//
		// let (locals) stmt...
		//
		"let": func(env *Env, args []any) (ret any) {
			if len(args) == 0 {
				return ErrMissing
			}

			locals, args := args[0], args[1:]
			llocals, ok := locals.(List)
			if !ok {
				return ErrInvalidType
			}

			env = NewEnv(env)

			for _, n := range llocals.items {
				env.PutLocal(n, nil)
			}

			for _, v := range args {
				if Verbose {
					fmt.Println("  ", v)
				}
				ret = Eval(env, v)
			}

			return
		},

		//
		// lambda (args) stmt...
		//
		"lambda": func(env *Env, args []any) any {
			if len(args) == 0 {
				return ErrMissing
			}

			locals, args := args[0], args[1:]
			llocals, ok := locals.(List)
			if !ok {
				return ErrInvalidType
			}

			return Lambda{args: llocals.items, body: args}
		},

		//
		// list items...
		//
		"list": func(env *Env, args []any) any {
			return List{items: args}
		},

		//
		// first list
		//
		"first": func(env *Env, args []any) any {
			if len(args) == 0 {
				return ErrMissing
			}

			l, ok := args[0].(List)
			if !ok {
				return ErrInvalidType
			}

			if len(l.items) == 0 {
				return Nil
			}

			return l.items[0]
		},

		//
		// last list
		//
		"last": func(env *Env, args []any) any {
			if len(args) == 0 {
				return ErrMissing
			}

			l, ok := args[0].(List)
			if !ok {
				return ErrInvalidType
			}

			if len(l.items) == 0 {
				return Nil
			}

			return l.items[len(l.items)-1]
		},

		//
		// nth list
		//
		"nth": func(env *Env, args []any) any {
			if len(args) < 2 {
				return ErrMissing
			}

			n, ok := args[0].(CanInt)
			if !ok {
				return ErrInvalidType
			}

			l, ok := args[1].(List)
			if !ok {
				return ErrInvalidType
			}

			nn := int(n.Int())

			if nn < 0 || nn >= len(l.items) {
				return Nil
			}

			return l.items[nn]
		},

		//
		// rest list
		//
		"rest": func(env *Env, args []any) any {
			if len(args) == 0 {
				return ErrMissing
			}

			l, ok := args[0].(List)
			if !ok {
				return ErrInvalidType
			}

			if len(l.items) == 0 {
				return Nil
			}

			return List{items: l.items[1:]}
		},
	}
}

func callambda(l Lambda, env *Env, args []any) (ret any) {
	lenv := NewEnv(env)

	for i, n := range l.args {
		var v any = nil

		if i < len(args) {
			v = lenv.PutLocal(n, env.Get(args[i]))
		}

		lenv.PutLocal(n, v)
	}

	for _, v := range l.body {
		if Verbose {
			fmt.Println("  ", v)
		}
		ret = Eval(lenv, v)
	}

	return
}

func callop(op Op, env *Env, args []any) any {
	if len(args) == 0 {
		if op.value == "+" {
			return 0
		}

		return ErrMissing
	}

	first := env.Get(args[0])

	switch t := first.(type) {
	case Integer:
		v := t.value

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

	case Float:
		v := t.value

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

func callcond(op Cond, env *Env, args []any) any {
	if len(args) == 0 {
		return True
	}

	c1, ok := env.Get(args[0]).(CanCond)
	if !ok {
		return True
	}

	for _, a := range args[1:] {
		var cond bool
		c2 := env.Get(a)

		switch op.value {
		case "=":
			cond = c1.Eq(c2)

		case "<":
			cond = c1.Lt(c2)

		case "<=":
			cond = c1.Leq(c2)

		case ">":
			cond = c1.Gt(c2)

		case ">=":
			cond = c1.Geq(c2)
		}

		if !cond {
			return Nil
		}

		c1, ok = c2.(CanCond)
		if !ok {
			break
		}
	}

	return True
}

type Env struct {
	vars map[string]any
	next *Env
}

func NewEnv(prev *Env) *Env {
	return &Env{vars: map[string]any{}, next: prev}
}

func (e *Env) PutLocal(o, value any) any {
	if s, ok := o.(Symbol); ok {
		e.vars[s.value] = value
	}

	return value
}

func (e *Env) Put(o, value any) any {
	if s, ok := o.(Symbol); ok {
		if _, ok := e.vars[s.value]; ok || e.next == nil {
			e.vars[s.value] = value
		} else {
			e.next.Put(o, value)
		}
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

	return Eval(e, o)
}

func (e *Env) GetList(l []any) (el []any) {
	for _, v := range l {
		el = append(el, e.Get(v))
	}

	return
}

func (e *Env) GetValues(l []any) (el []any) {
	for _, v := range l {
		v = e.Get(v)
		if o, ok := v.(Object); ok {
			v = o.Value()
		}

		el = append(el, v)
	}

	return
}

func Eval(env *Env, v any) any {
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
			if f, ok := primitives[i.value]; ok {
				return f(env, t.items[1:])
			}
			v := env.Get(i)
			if l, ok := v.(Lambda); ok {
				return callambda(l, env, t.items[1:])
			}

			return v

		case Op:
			return callop(i, env, t.items[1:])

		case Cond:
			return callcond(i, env, t.items[1:])
		}
	}

	return nil
}

func main() {
	expr := flag.Bool("e", false, "evaluate expression")
	flag.BoolVar(&Verbose, "v", Verbose, "verbose")
	flag.Parse()

	var p *Parser

	if *expr {
		p = NewParser(strings.NewReader(strings.Join(flag.Args(), " ")))
	} else if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
			return
		}

		p = NewParser(f)
		defer f.Close()
	} else {
		p = NewParser(os.Stdin)
	}

	l, err := p.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}

	if Verbose {
		fmt.Println(l)
	}

	fmt.Println()

	env := NewEnv(nil)

	for _, v := range l {
		fmt.Println(Eval(env, v))
	}
}
