package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"strings"

	"github.com/GaryBrownEEngr/turtle"
	"github.com/GaryBrownEEngr/turtle/models"
	"github.com/raff/gisp"
)

type Color struct {
	value color.RGBA
}

func (c Color) String() string { return fmt.Sprintf("Color%v", c.value) }
func (c Color) Value() any     { return c.value }

type Turtle struct {
	win    turtle.Window
	turtle models.Turtle
}

func (c Turtle) String() string { return "Turtle{}" }
func (c Turtle) Value() any     { return gisp.Nil }

var namedcolors = map[string]color.RGBA{
	"black": turtle.Black,
	"white": turtle.White,
	"red":   turtle.Red,
	"lime":  turtle.Lime,
	"blue":  turtle.Blue,

	"yellow":  turtle.Yellow,
	"aqua":    turtle.Aqua,
	"magenta": turtle.Magenta,

	"orange": turtle.Orange,
	"green":  turtle.Green,
	"purple": turtle.Purple,
	"indigo": turtle.Indigo,
	"violet": turtle.Violet,
}

// (color r g b [a])
func callColor(env *gisp.Env, args []any) any {
	args = env.GetList(args) // evaluate the arguments
	n := len(args)

	switch n {
	case 1:
		name := gisp.AsString(args[0], "")

		if c, ok := namedcolors[name]; ok {
			return Color{value: c}
		}

		return gisp.ErrInvalidType

	case 3, 4:
		r := uint8(gisp.AsInt(args[0], 0))
		g := uint8(gisp.AsInt(args[1], 0))
		b := uint8(gisp.AsInt(args[2], 0))
		a := uint8(255)

		if n == 4 {
			a = uint8(gisp.AsInt(args[3], 255))
		}

		return Color{value: color.RGBA{r, g, b, a}}
	}

	return gisp.ErrMissing
}

// (turtle [ (width height show) ] drawFunction)
func callTurtle(env *gisp.Env, args []any) any {
	args = env.GetList(args) // evaluate the arguments
	n := len(args)

	if n == 0 {
		return gisp.ErrMissing
	}

	params := turtle.Params{Width: 800, Height: 800}

	if n > 1 {
		l, ok := args[0].(gisp.List)
		if !ok {
			return gisp.ErrInvalidType
		}

		lp := env.GetList(l.Items())
		ln := len(lp)

		if ln > 0 {
			params.Width = int(gisp.AsInt(lp[0], int64(params.Width)))
		}

		if ln > 1 {
			params.Height = int(gisp.AsInt(lp[1], int64(params.Height)))
		}

		if ln > 2 {
			params.ShowFPS = gisp.AsBool(lp[2], params.ShowFPS)
		}

		args = args[1:]
	}

	ldraw, ok := args[0].(gisp.Lambda)
	if !ok {
		return gisp.ErrInvalidType
	}

	turtle.Start(params, func(w turtle.Window) {
		gisp.CallLambda(ldraw, env, []any{Turtle{win: w, turtle: w.NewTurtle()}})
	})

	return gisp.Nil
}

// (clear t [ color ])
func callClear(env *gisp.Env, args []any) any {
	args = env.GetList(args) // evaluate the arguments
	n := len(args)

	if n == 0 {
		return gisp.ErrMissing
	}

	t, ok := args[0].(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	bg := turtle.Black

	if n > 1 {
		c, ok := args[1].(Color)
		if !ok {
			return gisp.ErrInvalidType
		}

		bg = c.value
	}

	t.win.GetCanvas().ClearScreen(bg)
	return gisp.Nil
}

// (show bool)
func callShow(env *gisp.Env, args []any) any {
	args = env.GetList(args) // evaluate the arguments
	n := len(args)

	if n == 0 {
		return gisp.ErrMissing
	}

	t, ok := args[0].(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	show := true

	if len(args) > 1 {
		s := gisp.AsString(args[1], "")

		switch s {
		case "turtle":
			t.turtle.ShapeAsTurtle()

		case "arrow":
			t.turtle.ShapeAsArrow()

		default:
			show = gisp.AsBool(args[1], true)
		}
	}

	if show {
		t.turtle.ShowTurtle()
	} else {
		t.turtle.HideTurtle()
	}

	return show
}

// (scale t number)
func callScale(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	v, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.ShapeScale(v.Float())
	return nil
}

// (pendown t)
func callPenDown(env *gisp.Env, args []any) any {
	if len(args) == 0 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.PenDown()
	return nil
}

// (pendown t)
func callPenUp(env *gisp.Env, args []any) any {
	if len(args) == 0 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.PenUp()
	return nil
}

// (speed t pixelsPerSecond)
func callSpeed(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	a, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Speed(a.Float())
	return nil
}

// (setcolor t color)
func callSetColor(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	c, ok := env.Get(args[1]).(Color)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Color(c.value)
	return nil
}

// (fill t color)
func callFill(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	c, ok := env.Get(args[1]).(Color)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Fill(c.value)
	return nil
}

// (size t n)
func callSize(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	a, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Size(a.Float())
	return nil
}

// (dot t n)
func callDot(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	a, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Dot(a.Float())
	return nil
}

// (angle t angle)
func callAngle(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	a, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Angle(a.Float())
	return nil
}

// (left t angle)
func callLeft(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	a, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Left(a.Float())
	return nil
}

// (right t angle)
func callRight(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	a, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Right(a.Float())
	return nil
}

// (panleft t distance)
func callPanLeft(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	d, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.PanLeftward(d.Float())
	return nil
}

// (panright t distance)
func callPanRight(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	d, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.PanRightward(d.Float())
	return nil
}

// (forward t angle)
func callForward(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	a, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Forward(a.Float())
	return nil
}

// (backward t angle)
func callBackward(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	a, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Backward(a.Float())
	return nil
}

// (goto t x y)
func callGoTo(env *gisp.Env, args []any) any {
	if len(args) != 3 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	x, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	y, ok := env.Get(args[2]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.GoTo(x.Float(), y.Float())
	return nil
}

// (pointto t x y)
func callPointTo(env *gisp.Env, args []any) any {
	if len(args) != 3 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	x, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	y, ok := env.Get(args[2]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.PointToward(x.Float(), y.Float())
	return nil
}

func main() {
	expr := flag.Bool("e", false, "evaluate expression")
	interactive := flag.Bool("i", false, "interactive")
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
	} else {
		p = gisp.NewParser(os.Stdin)
	}

	gisp.AddPrimitive("color", callColor)
	gisp.AddPrimitive("turtle", callTurtle)
	gisp.AddPrimitive("clear", callClear)
	gisp.AddPrimitive("show", callShow)
	gisp.AddPrimitive("scale", callScale)
	gisp.AddPrimitive("pendown", callPenDown)
	gisp.AddPrimitive("penup", callPenUp)
	gisp.AddPrimitive("speed", callSpeed)
	gisp.AddPrimitive("setcolor", callSetColor)
	gisp.AddPrimitive("fill", callSetColor)
	gisp.AddPrimitive("size", callSize)
	gisp.AddPrimitive("dot", callDot)
	gisp.AddPrimitive("angle", callAngle)
	gisp.AddPrimitive("left", callLeft)
	gisp.AddPrimitive("right", callRight)
	gisp.AddPrimitive("panleft", callPanLeft)
	gisp.AddPrimitive("panright", callPanRight)
	gisp.AddPrimitive("backward", callBackward)
	gisp.AddPrimitive("forward", callForward)
	gisp.AddPrimitive("goto", callGoTo)
	gisp.AddPrimitive("pointto", callPointTo)

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
