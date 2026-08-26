module github.com/xudefa/enhance/examples/example-starter-ent

go 1.25.12

require (
	entgo.io/ent v0.14.2
	github.com/xudefa/enhance v0.0.6
	github.com/xudefa/enhance/starter/ent v0.0.6
)

require github.com/google/uuid v1.6.0 // indirect

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/ent => ../../starter/ent
)
