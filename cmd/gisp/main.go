package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/raff/gisp"
	"github.com/raff/readliner"
)

// (with-html (:html (:head (:title "Hello World")) (:body (:h1 "Hello World")))
func builtinHtml(env *gisp.Env, args []any) any {
	var sb = new(strings.Builder)
	processTags(sb, env, args)
	return gisp.MakeString(sb.String())
}

func processTags(sb *strings.Builder, env *gisp.Env, tags []any) []any {
	for len(tags) > 0 {
		if l, ok := tags[0].(gisp.List); ok && strings.HasPrefix(l.Item(0).(gisp.Object).String(), ":") {
			processTags(sb, env, l.Items())
			tags = tags[1:]
			continue
		}

		if tag, ok := tags[0].(gisp.Symbol); ok && strings.HasPrefix(tag.String(), ":") {
			tags = tags[1:]
			tagname := tag.String()[1:]

			sb.WriteString("<" + tagname)
			tags = processAttrs(sb, env, tags)
			if len(tags) > 0 {
				sb.WriteString(">\n")

				tags = processTags(sb, env, tags)
				sb.WriteString("</" + tagname + ">\n")
			} else {
				sb.WriteString("/>\n")
			}

			continue
		}

		sb.WriteString(fmt.Sprint(gisp.Eval(env, tags[0])) + "\n")
		tags = tags[1:]
	}

	return tags
}

func processAttrs(sb *strings.Builder, env *gisp.Env, tags []any) []any {
	for len(tags) > 0 {
		if tag, ok := tags[0].(gisp.Symbol); ok && strings.HasPrefix(tag.String(), ":") {
			sb.WriteString(" " + tag.String()[1:])
			tags = tags[1:]
		} else {
			break
		}

		if len(tags) > 0 {
			if tag, ok := tags[0].(gisp.String); ok {
				sb.WriteString("=" + strconv.Quote(tag.String()))
				tags = tags[1:]
			}
		}
	}

	return tags
}

func main() {
	expr := flag.Bool("e", false, "evaluate expression")
	interactive := flag.Bool("i", false, "interfactive")
	flag.BoolVar(&gisp.Verbose, "v", gisp.Verbose, "verbose")
	flag.Parse()

	var p *gisp.Parser
	var rl *readliner.ReadLiner

	gisp.AddBuiltin("with-html", builtinHtml)

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
	} else if *interactive {
		rl = readliner.New("> ", ".gisp_history")
		rl.SetContPrompt(": ")
		rl.SetCompletions(gisp.Builtins(), false)
		defer rl.Close()
		p = gisp.NewParser(rl)
	} else {
		p = gisp.NewParser(os.Stdin)
	}

	env := gisp.NewEnv(nil)

	if *interactive {
		for {
			rl.Newline()
			l, err := p.ParseOne()
			if err != nil {
				fmt.Println(err)
				return
			}

			for _, v := range l {
				v = env.Get(v)
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
