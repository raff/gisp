package gisp

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/scanner"
	"time"
	"unicode"
)

var (
	ErrEOF         = Error{value: fmt.Errorf("EOF")}
	ErrInvalid     = Error{value: fmt.Errorf("invalid-token")}
	ErrInvalidType = Error{value: fmt.Errorf("invalid-parameter-type")}
	ErrMissing     = Error{value: fmt.Errorf("missing-parameter")}
	Verbose        = false

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

	builtins map[string]Call
)

// Call is the signature for builtin methods
type Call func(env *Env, args []any) any

// AddBuiltin adds a new built-in method.
// Note that it can override existing builtin methods.
func AddBuiltin(name string, value Call) {
	builtins[name] = value
}

// Builtins returns a list of builtin method (names)
// Note that this doesn't return the math and conditional operators.
func Builtins() (l []string) {
	for k, _ := range builtins {
		l = append(l, k)
	}

	return
}

// Object is the interface for all gisp objects.
type Object interface {
	String() string
	Value() any
}

// CanInt is for objects that can cast to an integer (int64) value
type CanInt interface {
	Int() int64
}

// CanFloat is for objects that can cast to a float (float64) value
type CanFloat interface {
	Float() float64
}

// CanBool is for objects that can cast to a boolean (true/false)
type CanBool interface {
	Bool() bool
}

// CanCompare is for objects that can compare with other objects (=, <, <=, >, >=)
type CanCompare interface {
	Eq(v any) bool
	Lt(v any) bool
	Leq(v any) bool
	Gt(v any) bool
	Geq(v any) bool
}

// Error is a primitive object that maps errors
type Error struct {
	value error
}

func (o Error) String() string { return o.value.Error() }
func (o Error) Value() any     { return o.value }
func (o Error) Error() string  { return o.value.Error() }

// Boolean is the boolean primitive object
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

// Symbol is the symbol atom
type Symbol struct {
	value string
}

func (o Symbol) String() string { return o.value }
func (o Symbol) Value() any     { return o.value }

// Quoted is for quoted symbols
type Quoted struct {
	value any
}

func (o Quoted) String() string { return fmt.Sprintf("'%v", o.value) }
func (o Quoted) Value() any     { return o.value }

// Op is for math operators ( +, -, *, / )
type Op struct {
	value string
}

func (o Op) String() string { return fmt.Sprintf("%q", o.value) }
func (o Op) Value() any     { return o.value }

// Cond is for conditional operators ( =, <, <=, >, >= )
type Cond struct {
	value string
}

func (o Cond) String() string { return fmt.Sprintf("%q", o.value) }
func (o Cond) Value() any     { return o.value }

// Integer is the integer primitive type (int64)
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

// Float is the floating point primitive type (float64)
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

// String is the string primitive type
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

// List is the list type
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

func (o List) Items() []any {
	return o.items
}

func (o List) Bool() bool {
	return len(o.items) > 0
}

// Lambda is the anonymous function type
type Lambda struct {
	args []any
	body []any
}

func (o Lambda) String() string { return fmt.Sprintf("(lambda %v %v)", o.args, o.body) }
func (o Lambda) Value() any     { return Nil }

func (o Lambda) Arg(i int) any {
	if i < 0 || i >= len(o.args) {
		return nil
	}

	return o.args[i]
}

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

// Parser can parse a gisp object or program
type Parser struct {
	s scanner.Scanner
}

// NewParser creates a new Parser object that can parse the input Reader
func NewParser(r io.Reader) *Parser {
	var p Parser

	p.s.Init(r)
	p.s.Whitespace = 0
	p.s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanStrings | scanner.ScanRawStrings
	p.s.IsIdentRune = func(ch rune, i int) bool {
		return ch == '_' || ch == '$' || ch == ':' || unicode.IsLetter(ch) || unicode.IsDigit(ch) && i > 0
	}
	return &p
}

// SepNext checks if the next character to parse is a separator between gisp objects
func (p *Parser) SepNext() bool {
	switch p.s.Peek() {
	case ' ', '\t', '\r', '\n', '(', ')', scanner.EOF:
		return true
	}

	return false
}

// Parse parses the input from the Reader until EOF and returns a list of objects
func (p *Parser) Parse() (l []any, err error) {
	return p.parse(false)
}

// ParseOne parses one object from the input
func (p *Parser) ParseOne() (l []any, err error) {
	return p.parse(true)
}

func (p *Parser) parse(one bool) (l []any, err error) {
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

	if p.s.Peek() == scanner.EOF {
		return nil, ErrEOF
	}

	for tok := p.s.Scan(); tok != scanner.EOF; tok = p.s.Scan() {
		st := p.s.TokenText()

		if Verbose {
			fmt.Printf("%v: %v %q\n", p.s.Position, scanner.TokenString(tok), st)
		}

		switch tok {
		case '(':
			vv, err := p.parse(false)
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
			if Verbose {
				fmt.Printf("separator: %d", tok)
			}
			if one {
				return
			}
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

		case scanner.String, scanner.RawString:
			st, _ = strconv.Unquote(st)
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

func invalidType(v any) error {
	if Verbose {
		fmt.Printf("invalid (%T) %#v", v, v)
	}

	return ErrInvalidType
}

func init() {
	// primitive functions

	builtins = map[string]Call{
		//
		// print args
		//
		"print": func(env *Env, args []any) any {
			args = env.GetList(args)

			fmt.Print(args...)

			if len(args) > 0 {
				return args[len(args)-1]
			}

			return Nil
		},

		//
		// println args...
		//
		"println": func(env *Env, args []any) any {
			args = env.GetList(args)

			fmt.Println(args...)

			if len(args) > 0 {
				return args[len(args)-1]
			}

			return Nil
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
				return invalidType(f)
			}

			return fmt.Sprintf(sfmt.String(), args...)
		},

		//
		// readfile [filename]
		//
		"readfile": func(env *Env, args []any) any {
			fin := os.Stdin

			if len(args) > 0 {
				fname, ok := env.Get(args[0]).(String)
				if !ok {
					return invalidType(args[0])
				}

				f, err := os.Open(fname.value)
				if err != nil {
					return MakeError(err)
				}

				defer f.Close()
				fin = f
			}

			content, err := io.ReadAll(fin)
			if err != nil {
				return MakeError(err)
			}

			return String{value: string(content)}
		},

		//
		// readlines [filename]
		//
		"readlines": func(env *Env, args []any) any {
			fin := os.Stdin

			if len(args) > 0 {
				fname, ok := env.Get(args[0]).(String)
				if !ok {
					return invalidType(args[0])
				}

				f, err := os.Open(fname.value)
				if err != nil {
					return MakeError(err)
				}

				defer f.Close()
				fin = f
			}

			var lines []any

			scanner := bufio.NewScanner(fin)
			for scanner.Scan() {
				lines = append(lines, String{value: scanner.Text()})
			}
			if err := scanner.Err(); err != nil {
				return MakeError(err)
			}

			return List{items: lines}
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

			return invalidType(v)
		},

		//
		// (rand) -> float
		// (rand n) -> 0..n
		// (rand a b c d) -> one of a b c d
		"rand": func(env *Env, args []any) any {
			switch len(args) {
			case 0:
				return rand.Float64()

			case 1:
				v := env.Get(args[0])
				if v, ok := v.(CanInt); ok {
					return rand.Int63n(v.Int())
				}

				return invalidType(v)

			default:
				n := rand.Intn(len(args))
				return args[n]
			}

			return Nil
		},

		//
		// find needle haysstack
		//
		"find": func(env *Env, args []any) any {
			if len(args) != 2 {
				return ErrMissing
			}

			n := env.Get(args[0])
			s := env.Get(args[1])

			switch t := s.(type) {
			case String:
				if ss, ok := n.(String); ok {
					if p := strings.Index(t.String(), ss.String()); p > 0 {
						return p
					}

					return Nil
				}

			case List:
				if p := slices.Index(t.items, n); p > 0 {
					return p
				}

				return Nil
			}

			return ErrInvalidType
		},

		//
		// contains needle haysstack
		//
		"contains": func(env *Env, args []any) any {
			if len(args) != 2 {
				return ErrMissing
			}

			n := env.Get(args[0])
			s := env.Get(args[1])

			switch t := s.(type) {
			case String:
				if ss, ok := n.(String); ok {
					return Boolean{value: strings.Contains(t.String(), ss.String())}
				}

			case List:
				return Boolean{value: slices.Contains(t.items, n)}
			}

			return ErrInvalidType
		},

		//
		// append string string... | list list...
		//
		"append": func(env *Env, args []any) any {
			if len(args) == 0 {
				return List{}
			}

			args = env.GetList(args)

			switch t := args[0].(type) {
			case String:
				var sb strings.Builder
				sb.WriteString(t.value)

				for _, v := range args[1:] {
					s, ok := v.(String)
					if !ok {
						return ErrInvalidType
					}

					sb.WriteString(s.value)
				}

				return String{value: sb.String()}

			case List:
				for _, v := range args[1:] {
					ll, ok := v.(List)
					if !ok {
						return ErrInvalidType
					}

					t.items = append(t.items, ll.items...)
				}

				return t
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
				barg, args = env.Get(args[0]), args[1:] // cond, expr

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
				case 0, 1: // skip if expr
					return Nil

				case 2: // return else expr
					return env.Get(args[1])
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
				return invalidType(locals)
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
		// eval form
		//
		"eval": func(env *Env, args []any) any {
			if len(args) == 0 {
				return ErrMissing
			}

			e := env.Get(args[0])
			return Eval(env, e)
		},

		//
		// lambda (args) stmt...
		//
		"lambda": func(env *Env, args []any) any {
			if len(args) == 0 {
				return ErrMissing
			}

			params, args := args[0], args[1:]
			pparams, ok := params.(List)
			if !ok {
				return invalidType(params)
			}

			return Lambda{args: pparams.items, body: args}
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

			l, ok := env.Get(args[0]).(List)
			if !ok {
				return invalidType(args[0])
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

			l, ok := env.Get(args[0]).(List)
			if !ok {
				return invalidType(args[0])
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

			n, ok := env.Get(args[0]).(CanInt)
			if !ok {
				return invalidType(args[0])
			}

			l, ok := env.Get(args[1]).(List)
			if !ok {
				return invalidType(args[1])
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

			l, ok := env.Get(args[0]).(List)
			if !ok {
				return invalidType(args[0])
			}

			if len(l.items) == 0 {
				return Nil
			}

			return List{items: l.items[1:]}
		},
	}
}

// CallLambda call a lambda function, passing the local enviroment and some input parameters
func CallLambda(l Lambda, env *Env, args []any) (ret any) {
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
				return invalidType(a)
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
				return invalidType(a)
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

	return invalidType(first)
}

func callcond(op Cond, env *Env, args []any) any {
	if len(args) == 0 {
		return True
	}

	c1, ok := env.Get(args[0]).(CanCompare)
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

		c1, ok = c2.(CanCompare)
		if !ok {
			break
		}
	}

	return True
}

// Env stores the current environments (collection of variables)
type Env struct {
	vars map[string]any
	next *Env
}

// NewEnv creates a new enviroment.
// The root environment should have prev=nil, local environment will link to the previous (parent) one.
func NewEnv(prev *Env) *Env {
	return &Env{vars: map[string]any{}, next: prev}
}

func getname(o any) (string, error) {
	switch t := o.(type) {
	case Symbol:
		return t.value, nil

	case string:
		return t, nil
	}

	return "", ErrInvalidType
}

// PutLocal creates or update a variable in the local environment
func (e *Env) PutLocal(o, value any) any {
	name, err := getname(o)
	if err != nil {
		return err
	}

	e.vars[name] = value
	return value
}

// Put update a variable with the same name, starting from the local environment.
// If the variable doesn't already exist, it will be created in the global environment.
func (e *Env) Put(o, value any) any {
	name, err := getname(o)
	if err != nil {
		return err
	}

	if _, ok := e.vars[name]; ok || e.next == nil {
		e.vars[name] = value
	} else {
		e.next.Put(o, value)
	}

	return value
}

// Get tries to resolve to an existing variable or evaluate the input.
func (e *Env) Get(o any) any {
	name, err := getname(o)
	if err != nil {
		return Eval(e, o)
	}

	if v, ok := e.vars[name]; ok {
		return v
	}

	if e.next != nil {
		return e.next.Get(o)
	}

	return Nil
}

// Get resolves all input as variables or evaluates them.
func (e *Env) GetList(l []any) (el []any) {
	for _, v := range l {
		el = append(el, e.Get(v))
	}

	return
}

// GetValues returns the primitive value for the input variables or evaluated objects.
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

// AsBool converts the input object to a boolean, if possible or return the default value.
func AsBool(o any, def bool) bool {
	if i, ok := o.(CanBool); ok {
		return i.Bool()
	}

	return def
}

// AsInt converts the input object to an integer, if possible or return the default value.
func AsInt(o any, def int64) int64 {
	if i, ok := o.(CanInt); ok {
		return i.Int()
	}

	return def
}

// AsFloat converts the input object to a float, if possible or return the default value.
func AsFloat(o any, def float64) float64 {
	if i, ok := o.(CanFloat); ok {
		return i.Float()
	}

	return def
}

// AsString returns the input representation for the input object, or the default.
func AsString(o any, def string) string {
	switch t := o.(type) {
	case string:
		return t

	case Quoted:
		return t.value.(Object).String()

	case Object:
		return t.String()
	}

	return def
}

// MakeBool creates a Boolean object from a bool
func MakeBool(v bool) Boolean {
	return Boolean{value: v}
}

// MakeInt creates an Integer object from an int
func MakeInt[T int8 | int | int16 | int64](v T) Integer {
	return Integer{value: int64(v)}
}

// MakeFloat creates a Float object from a float64
func MakeFloat[T float32 | float64](v T) Float {
	return Float{value: float64(v)}
}

// MakeString creates a String object from a string
func MakeString(v string) String {
	return String{value: v}
}

// MakeList creates a List object from a list of objects
func MakeList(items ...any) List {
	return List{items: items}
}

// MakeList creates an Error object from a go error
func MakeError(e error) Error {
	return Error{value: e}
}

// Eval evaluates the current object
func Eval(env *Env, v any) any {
	if Verbose {
		fmt.Println("eval", v)
	}

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
			if f, ok := builtins[i.value]; ok {
				return f(env, t.items[1:])
			}
			v := env.Get(i)
			if l, ok := v.(Lambda); ok {
				return CallLambda(l, env, t.items[1:])
			}

			return v

		case Op:
			return callop(i, env, t.items[1:])

		case Cond:
			return callcond(i, env, t.items[1:])
		}
	}

	return v
}
