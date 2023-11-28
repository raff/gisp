# turtle
a `gisp` REPL with turtle abilities

This is an example of adding new `built-in` methods to gisp, based on https://github.com/GaryBrownEEngr/turtle (and [Ebitengine](https://ebitengine.org/),

## Additional built-ins

    (color r g b [a])
    (turtle [ (width height show) ] drawFunction)
    (clear t [ color ])
    (show [bool|'turtle|'arrow|'imagefile)
    (scale t number)
    (pendown t)
    (pendown t)
    (speed t pixelsPerSecond)
    (pencolor t color)
    (fill t color)
    (size t n)
    (dot t n)
    (angle t angle)
    (left t angle)
    (right t angle)
    (panleft t distance)
    (panright t distance)
    (forward t angle)
    (backward t angle)
    (goto t x y)
    (pos t)
    (pointto t x y)
    (circle t radius angle steps)

See the file `turtle.gisp` for a working example
