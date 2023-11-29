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
	input  models.UserInput
}

func (c Turtle) String() string { return "Turtle{}" }
func (c Turtle) Value() any     { return gisp.Nil }

func (c Turtle) getInputs() models.UserInput {
	return c.win.GetCanvas().GetUserInput()
}

func (c Turtle) pressed(k string) bool {
	in := c.getInputs()
	c.input = in

	switch k {
	case "A":
		return in.KeysDown.A
	case "S":
		return in.KeysDown.S

	case "-":
		return in.KeysDown.Minus
	case "=":
		return in.KeysDown.Equal
	case "[":
		return in.KeysDown.OpenSquareBracket
	case "]":
		return in.KeysDown.CloseSquareBracket

	case " ":
		return in.KeysDown.Space
	case "\n":
		return in.KeysDown.Enter
	case "\x1b":
		return in.KeysDown.Escape

	case "1":
		return in.KeysDown.Number1
	case "2":
		return in.KeysDown.Number2
	case "3":
		return in.KeysDown.Number3
	case "4":
		return in.KeysDown.Number4
	case "5":
		return in.KeysDown.Number5
	case "6":
		return in.KeysDown.Number6
	case "7":
		return in.KeysDown.Number7
	case "8":
		return in.KeysDown.Number8
	case "9":
		return in.KeysDown.Number9
	case "0":
		return in.KeysDown.Number0

	case "F1":
		return in.KeysDown.F1
	case "F2":
		return in.KeysDown.F2
	case "F3":
		return in.KeysDown.F3
	case "F4":
		return in.KeysDown.F4
	case "F5":
		return in.KeysDown.F5
	case "F6":
		return in.KeysDown.F6
	case "F7":
		return in.KeysDown.F7
	case "F8":
		return in.KeysDown.F8
	case "F9":
		return in.KeysDown.F9
	case "F10":
		return in.KeysDown.F10
	case "F11":
		return in.KeysDown.F11
	case "F12":
		return in.KeysDown.F12
	}

	return false
}

func (c Turtle) justPressed(k string) bool {
	in := c.getInputs()
	defer func() {
		c.input = in
	}()

	switch k {
	case "A":
		return in.KeysDown.A && !c.input.KeysDown.A
	case "S":
		return in.KeysDown.S && !c.input.KeysDown.S

	case "-":
		return in.KeysDown.Minus && !c.input.KeysDown.Minus
	case "=":
		return in.KeysDown.Equal && !c.input.KeysDown.Equal
	case "[":
		return in.KeysDown.OpenSquareBracket && !c.input.KeysDown.OpenSquareBracket
	case "]":
		return in.KeysDown.CloseSquareBracket && !c.input.KeysDown.CloseSquareBracket

	case " ":
		return in.KeysDown.Space && !c.input.KeysDown.Space
	case "\n":
		return in.KeysDown.Enter && !c.input.KeysDown.Enter
	case "\x1b":
		return in.KeysDown.Escape && !c.input.KeysDown.Escape

	case "1":
		return in.KeysDown.Number1 && !c.input.KeysDown.Number1
	case "2":
		return in.KeysDown.Number2 && !c.input.KeysDown.Number2
	case "3":
		return in.KeysDown.Number3 && !c.input.KeysDown.Number3
	case "4":
		return in.KeysDown.Number4 && !c.input.KeysDown.Number4
	case "5":
		return in.KeysDown.Number5 && !c.input.KeysDown.Number5
	case "6":
		return in.KeysDown.Number6 && !c.input.KeysDown.Number6
	case "7":
		return in.KeysDown.Number7 && !c.input.KeysDown.Number7
	case "8":
		return in.KeysDown.Number8 && !c.input.KeysDown.Number8
	case "9":
		return in.KeysDown.Number9 && !c.input.KeysDown.Number9
	case "0":
		return in.KeysDown.Number0 && !c.input.KeysDown.Number0

	case "F1":
		return in.KeysDown.F1 && !c.input.KeysDown.F1
	case "F2":
		return in.KeysDown.F2 && !c.input.KeysDown.F2
	case "F3":
		return in.KeysDown.F3 && !c.input.KeysDown.F3
	case "F4":
		return in.KeysDown.F4 && !c.input.KeysDown.F4
	case "F5":
		return in.KeysDown.F5 && !c.input.KeysDown.F5
	case "F6":
		return in.KeysDown.F6 && !c.input.KeysDown.F6
	case "F7":
		return in.KeysDown.F7 && !c.input.KeysDown.F7
	case "F8":
		return in.KeysDown.F8 && !c.input.KeysDown.F8
	case "F9":
		return in.KeysDown.F9 && !c.input.KeysDown.F9
	case "F10":
		return in.KeysDown.F10 && !c.input.KeysDown.F10
	case "F11":
		return in.KeysDown.F11 && !c.input.KeysDown.F11
	case "F12":
		return in.KeysDown.F12 && !c.input.KeysDown.F12
	}

	return false
}

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

func callJustPressed(env *gisp.Env, args []any) any {
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

	return gisp.MakeBool(t.justPressed(s))
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

	in := t.getInputs()

	return gisp.MakeList(gisp.MakeInt(in.MouseX), gisp.MakeInt(in.MouseY), gisp.MakeFloat(in.MouseScroll))
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
	gisp.AddPrimitive("pencolor", callPenColor)
	gisp.AddPrimitive("fill", callFill)
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
	gisp.AddPrimitive("pos", callPos)
	gisp.AddPrimitive("pointto", callPointTo)
	gisp.AddPrimitive("circle", callCircle)
	gisp.AddPrimitive("pressed", callPressed)
	gisp.AddPrimitive("justpressed", callJustPressed)
	gisp.AddPrimitive("mousepos", callMousePos)

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
