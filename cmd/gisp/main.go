package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/raff/gisp"
)

func main() {
	expr := flag.Bool("e", false, "evaluate expression")
	interactive := flag.Bool("i", false, "interfactive")
	flag.BoolVar(&gisp.Verbose, "v", gisp.Verbose, "verbose")
	flag.Parse()

	var p *gisp.Parser

	if *expr {
		p = gisp.NewParser(strings.NewReader(strings.Join(flag.Args(), " ")))
	} else if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
			return
		}

		p = gisp.NewParser(f)
		defer f.Close()
		/*
			} else if *interactive {
				rl := NewReadLiner("> ", ".gisp_history")
				rl.SetEol("\n\n") // not sure of why
				defer rl.Close()
				p = gisp.NewParser(rl)
		*/
	} else {
		p = gisp.NewParser(os.Stdin)
	}

	env := gisp.NewEnv(nil)

	if *interactive {
		for {
			l, err := p.ParseOne()
			if err != nil {
				fmt.Println(err)
				return
			}

			for _, v := range l {
				fmt.Println(gisp.Eval(env, v))
			}
		}

		return
	}

	l, err := p.Parse()
	if err != nil {
		fmt.Println(err)
		return
	}

	var ret any

	for _, v := range l {
		ret = gisp.Eval(env, v)
	}

	fmt.Println(ret)
}
