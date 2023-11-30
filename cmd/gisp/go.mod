module github.com/raff/gisp/cmd/gisp

go 1.20

require (
	github.com/raff/gisp v1.0.0
	github.com/raff/readliner v0.0.0-20231130054758-ac9216a4b203
)

require (
	github.com/mattn/go-runewidth v0.0.3 // indirect
	github.com/peterh/liner v1.2.2 // indirect
	golang.org/x/sys v0.0.0-20211117180635-dee7805ff2e1 // indirect
)

replace github.com/raff/gisp v1.0.0 => ../..
