# gisp
My attempt to write a minimal Lisp interpreter (in Go)

## Supported types
- boolean (true, nil)
- integer 64 bits
- float 64 bits
- string
- symbol

## Supported primitives

- +, -, *, /, % : arithmetic operators
- =, <, <=, >, >= : conditionals

- quote
- setq
- let
- not, or, and
- if
- while
- begin
- lambda

- list, first, last, nth, rest

- print, println, format, sleep, rand

## Examples
- cmd/gisp : a REPL for gisp (can run single expressions, programs from file or expressions interactively)
- cmd/turtle : a REPL for gisp with turtle abilities (it shows how to add new types and primitives to gisp)
