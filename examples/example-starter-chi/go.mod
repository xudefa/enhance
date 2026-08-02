module github.com/xudefa/enhance/examples/example-starter-chi

go 1.25.12

require (
	github.com/go-chi/chi/v5 v5.0.12
	github.com/xudefa/enhance v0.0.4
	github.com/xudefa/enhance/starter/chi v0.0.4
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/chi => ../../starter/chi
)
