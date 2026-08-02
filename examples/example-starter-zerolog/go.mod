module github.com/xudefa/enhance/examples/example-starter-zerolog

go 1.25.12

require (
	github.com/rs/zerolog v1.34.0
	github.com/xudefa/enhance v0.0.4
	github.com/xudefa/enhance/starter/zerolog v0.0.4
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/zerolog => ../../starter/zerolog
)
