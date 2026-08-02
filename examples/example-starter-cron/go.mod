module github.com/xudefa/enhance/examples/example-starter-cron

go 1.25.12

require (
	github.com/robfig/cron/v3 v3.0.1
	github.com/xudefa/enhance v0.0.5
	github.com/xudefa/enhance/starter/cron v0.0.5
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/cron => ../../starter/cron
)
