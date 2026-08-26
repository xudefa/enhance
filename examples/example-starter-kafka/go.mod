module github.com/xudefa/enhance/examples/example-starter-kafka

go 1.25.12

require (
	github.com/segmentio/kafka-go v0.4.47
	github.com/xudefa/enhance v0.0.6
	github.com/xudefa/enhance/starter/kafka v0.0.6
)

require (
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/kafka => ../../starter/kafka
)
