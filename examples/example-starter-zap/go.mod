module github.com/xudefa/enhance/examples/example-starter-zap

go 1.25.12

require (
	github.com/xudefa/enhance v0.0.5
	github.com/xudefa/enhance/starter/zap v0.0.5
	go.uber.org/zap v1.27.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/zap => ../../starter/zap
)
