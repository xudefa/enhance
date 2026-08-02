module github.com/xudefa/enhance/examples/example-starter-casbin

go 1.25.12

require (
	github.com/casbin/casbin/v2 v2.123.0
	github.com/xudefa/enhance v0.0.5
	github.com/xudefa/enhance/starter/casbin v0.0.5
)

require (
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/casbin/govaluate v1.3.0 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/casbin => ../../starter/casbin
)
