module github.com/xudefa/enhance/examples/example-starter-elasticsearch

go 1.25.12

require (
	github.com/elastic/go-elasticsearch/v8 v8.13.0
	github.com/xudefa/enhance v0.0.3
	github.com/xudefa/enhance/starter/elasticsearch v0.0.3
)

require (
	github.com/elastic/elastic-transport-go/v8 v8.5.0 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/otel v1.24.0 // indirect
	go.opentelemetry.io/otel/metric v1.24.0 // indirect
	go.opentelemetry.io/otel/trace v1.24.0 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/elasticsearch => ../../starter/elasticsearch
)
