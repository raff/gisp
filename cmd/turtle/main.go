package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"

	_ "image/gif"
	_ "image/png"

	"github.com/gary23b/turtle"
	"github.com/gary23b/turtle/turtlemodel"
	"github.com/raff/gisp"
)

type Color struct {
	value color.Color
}

func (c Color) String() string { return fmt.Sprintf("Color%v", c.value) }
func (c Color) Value() any     { return c.value }

type Turtle struct {
	win    turtle.Window
	turtle turtlemodel.Turtle
	input  chan *turtlemodel.UserInput
}

func (c Turtle) String() string { return "Turtle{}" }
func (c Turtle) Value() any     { return gisp.Nil }

func (c Turtle) pressed(k string) bool {
	in := c.win.GetCanvas().PressedUserInput()
	return in.IsPressedByName(k)
}

func (c Turtle) justPressed() *turtlemodel.UserInput {
	return turtle.GetNewestJustPressedFromChan(c.input)
}

var namedcolors = map[string]color.Color{
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
		t := Turtle{win: w, turtle: w.NewTurtle()}
		t.input = t.win.GetCanvas().SubscribeToJustPressedUserInput()
		gisp.CallLambda(ldraw, env, []any{t})
	})

	return gisp.Nil
}

// (exit t)
func callExit(env *gisp.Env, args []any) any {
	if len(args) != 1 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.win.GetCanvas().Exit()
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

	var bg color.Color = turtle.Black

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
			if f, err := os.Open(s); err == nil { // it's a file
				ima, _, err := image.Decode(f)
				f.Close()

				if err == nil {
					t.turtle.ShapeAsImage(ima)
				} else {
					return gisp.ErrInvalidType
				}
			} else {
				show = gisp.AsBool(args[1], true)
			}
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

// (pencolor t color)
func callPenColor(env *gisp.Env, args []any) any {
	if len(args) < 1 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	if len(args) == 1 {
		c := t.turtle.GetColor()
		return Color{value: c}
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
	if len(args) < 2 {
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
	if len(args) < 1 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	if len(args) == 1 {
		s := t.turtle.GetSize()
		return gisp.MakeFloat(s)
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
	if len(args) < 1 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	if len(args) == 1 {
		a := t.turtle.GetAngle()
		return gisp.MakeFloat(a)
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

// (pos t)
func callPos(env *gisp.Env, args []any) any {
	if len(args) != 1 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	x, y := t.turtle.GetPos()
	return gisp.MakeList(gisp.MakeFloat(x), gisp.MakeFloat(y))
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

// (circle t radius angle steps)
func callCircle(env *gisp.Env, args []any) any {
	if len(args) != 4 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	r, ok := env.Get(args[1]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	a, ok := env.Get(args[2]).(gisp.CanFloat)
	if !ok {
		return gisp.ErrInvalidType
	}

	s, ok := env.Get(args[3]).(gisp.CanInt)
	if !ok {
		return gisp.ErrInvalidType
	}

	t.turtle.Circle(r.Float(), a.Float(), int(s.Int()))
	return nil
}

// (pressed key)
func callPressed(env *gisp.Env, args []any) any {
	if len(args) != 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	s := gisp.AsString(args[1], "")
	if len(s) == 0 {
		return gisp.ErrInvalidType
	}

	return gisp.MakeBool(t.pressed(s))
}

// (justpressed key key...)
func callJustPressed(env *gisp.Env, args []any) any {
	if len(args) < 2 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	var l []any

	in := t.justPressed()
	if in == nil {
		return gisp.MakeList(l...)
	}

	for _, v := range args[1:] {
		k := gisp.AsString(v, "")
		if len(k) == 0 {
			return gisp.ErrInvalidType
		}

		if in.IsPressedByName(k) {
			l = append(l, gisp.MakeString(k))
		}
	}

	return gisp.MakeList(l...)
}

// (mousepos t)
func callMousePos(env *gisp.Env, args []any) any {
	if len(args) != 1 {
		return gisp.ErrMissing
	}

	t, ok := env.Get(args[0]).(Turtle)
	if !ok {
		return gisp.ErrInvalidType
	}

	in := t.win.GetCanvas().PressedUserInput()
	return gisp.MakeList(gisp.MakeInt(in.Mouse.MouseX), gisp.MakeInt(in.Mouse.MouseY), gisp.MakeFloat(in.Mouse.MouseScroll))
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

	gisp.AddBuiltin("color", callColor)
	gisp.AddBuiltin("turtle", callTurtle)
	gisp.AddBuiltin("exit", callExit)
	gisp.AddBuiltin("clear", callClear)
	gisp.AddBuiltin("show", callShow)
	gisp.AddBuiltin("scale", callScale)
	gisp.AddBuiltin("pendown", callPenDown)
	gisp.AddBuiltin("penup", callPenUp)
	gisp.AddBuiltin("speed", callSpeed)
	gisp.AddBuiltin("pencolor", callPenColor)
	gisp.AddBuiltin("fill", callFill)
	gisp.AddBuiltin("size", callSize)
	gisp.AddBuiltin("dot", callDot)
	gisp.AddBuiltin("angle", callAngle)
	gisp.AddBuiltin("left", callLeft)
	gisp.AddBuiltin("right", callRight)
	gisp.AddBuiltin("panleft", callPanLeft)
	gisp.AddBuiltin("panright", callPanRight)
	gisp.AddBuiltin("backward", callBackward)
	gisp.AddBuiltin("forward", callForward)
	gisp.AddBuiltin("goto", callGoTo)
	gisp.AddBuiltin("pos", callPos)
	gisp.AddBuiltin("pointto", callPointTo)
	gisp.AddBuiltin("circle", callCircle)
	gisp.AddBuiltin("pressed", callPressed)
	gisp.AddBuiltin("justpressed", callJustPressed)
	gisp.AddBuiltin("mousepos", callMousePos)

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
