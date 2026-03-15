module github.com/grimdork/mew

require (
	github.com/alecthomas/chroma v0.10.0
	github.com/grimdork/climate v0.24.1
	github.com/muesli/termenv v0.16.0
	github.com/yuin/goldmark v1.7.16
)

replace github.com/grimdork/climate => ../climate

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/lucasb-eyer/go-colorful v1.3.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)

go 1.25.0
