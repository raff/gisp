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
	ErrInvalid = fmt.Errorf("invalid-token")
)

type Object interface {
	String() string
	Value() any
}

type Nil struct{}

func (o Nil) String() string { return `nil` }
func (o Nil) Value() any     { return nil }

type True struct{}

func (o True) String() string { return `true` }
func (o True) Value() any     { return true }

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

type Float struct {
	value float64
}

func (o Float) String() string { return fmt.Sprint(o.value) }
func (o Float) Value() any     { return o.value }

type String struct {
	value string
}

func (o String) String() string { return o.value }
func (o String) Value() any     { return o.value }

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

func ident(v string) Object {
	switch v {
	case "t":
		return True{}

	case "nil":
		return Nil{}
	}

	return Symbol{value: v}
}

type Parser struct {
	s scanner.Scanner
	//curr any
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

			switch v.(type) {
			case Symbol, List:
				fmt.Println("Quote", v)
				v = Quoted{value: v}
			}
		}

		return v
	}

	appendtolist := func(v any) {
		l = append(l, maybequoted(v))
	}

	for tok := p.s.Scan(); tok != scanner.EOF; tok = p.s.Scan() {
		st := p.s.TokenText()

		fmt.Printf("%v: %v %q\n", p.s.Position, scanner.TokenString(tok), st)

		switch tok {
		case '(':
			vv, err := p.Parse()
			if err != nil {
				return nil, err
			}

			appendtolist(List{items: vv})

		case ')':
			if quoted {
				appendtolist(Nil{})
			}

			break

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
			fmt.Println("quote")
			quoted = true

		case '+', '-', '/', '*':
			if tok == '+' || tok == '-' {
				if n := p.s.Peek(); n == '.' || (n >= '0' && n <= '9') { // next token is a number
					neg = tok == '-'
					continue
				}
			}

			appendtolist(Op{value: st})

		default:
			fmt.Printf("UNKNOWN %v %q", scanner.TokenString(tok), st)
			return nil, ErrInvalid
		}
	}

	return
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

	if len(l) != 1 {
		fmt.Println(l)
	} else {
		fmt.Println(l[0])
	}
}
