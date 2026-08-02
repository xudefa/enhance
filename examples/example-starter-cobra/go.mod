module github.com/xudefa/enhance/examples/example-starter-cobra

go 1.25.12

require (
	github.com/spf13/cobra v1.8.0
	github.com/xudefa/enhance v0.0.4
	github.com/xudefa/enhance/starter/cobra v0.0.4
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/cobra => ../../starter/cobra
)
