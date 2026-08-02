module github.com/xudefa/enhance/examples/example-starter-prometheus

go 1.25.12

require (
	github.com/xudefa/enhance v0.0.5
	github.com/xudefa/enhance/starter/prometheus v0.0.5
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/prometheus/client_golang v1.19.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/prometheus => ../../starter/prometheus
)
