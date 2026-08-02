module github.com/xudefa/enhance/examples/example-starter-redis

go 1.25.12

require (
	github.com/xudefa/enhance v0.0.4
	github.com/xudefa/enhance/starter/redis v0.0.4
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/redis/go-redis/v9 v9.5.1 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/redis => ../../starter/redis
)
