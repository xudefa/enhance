module github.com/xudefa/enhance/examples/example-starter-ratelimiter

go 1.25.12

require (
	github.com/xudefa/enhance v0.0.3
	github.com/xudefa/enhance/starter/ratelimiter v0.0.3
)

require golang.org/x/time v0.5.0 // indirect

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/ratelimiter => ../../starter/ratelimiter
)
