module github.com/xudefa/enhance/examples/example-starter-rabbitmq

go 1.25.12

require (
	github.com/rabbitmq/amqp091-go v1.9.0
	github.com/xudefa/enhance v0.0.6
	github.com/xudefa/enhance/starter/rabbitmq v0.0.6
)

replace (
	github.com/xudefa/enhance => ../../
	github.com/xudefa/enhance/starter/rabbitmq => ../../starter/rabbitmq
)
