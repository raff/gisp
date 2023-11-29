# turtle
a `gisp` REPL with turtle abilities

This is an example of adding new `built-in` methods to gisp, based on https://github.com/GaryBrownEEngr/turtle (and [Ebitengine](https://ebitengine.org/),

## Additional built-ins

    (color r g b [a])                             ; return a color
    (turtle [ (width height show) ] drawFunction) ; create a turtle window and run `drawFunction`
    (clear t [ color ])                           ; clear window (and, if present, fill with color `color`
    (show [bool|'turtle|'arrow|'imagefile)        ; show/hide cursor and set shape (turtle, arrow or from image file)
    (speed t pixelsPerSecond)                     ; set speed of drawing
    (scale t number)                              ; scale cursor
    (pencolor t color)                            ; set pen color
    (pendown t)                                   ; pen down (draw line when moving)
    (pendown t)                                   ; pen up (move without drawing)
    (fill t color)                                ; fill drawing with color
    (size t n)                                    ; line width (`n` pixels)
    (dot t n)                                     ; draw a dot of radius `n`
    (angle t angle)                               ; rotate to angle
    (left t angle)                                ; rotate left of `angle` degrees
    (right t angle)                               ; rotate right of `angle` degrees
    (panleft t distance)
    (panright t distance)
    (forward t distance)                          ; move forward of `distance` pixels
    (backward t distance)                         ; move backward of `distance` pixels
    (goto t x y)                                  ; go to `x, y` coordinates
    (pos t)                                       ; return current position (x, y)
    (pointto t x y)                               ; point to `x, y` coordinates
    (circle t radius angle steps)                 ; draw a circle of radious `radius` 

See the file `turtle.gisp` for a working example

    > go run . turtle.gisp
