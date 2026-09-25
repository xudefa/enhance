// Package main demonstrates the RabbitMQ starter usage.
//
// This example shows how to use the RabbitMQ starter to:
// 1. Auto-configure RabbitMQ connection
// 2. Publish messages
// 3. Consume messages
//
// Prerequisites:
// - RabbitMQ server running on localhost:5672
//
// Run:
//
//	go run main.go
package main

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/rabbitmq"
)

func main() {
	fmt.Println("=== RabbitMQ Starter Example ===")
	fmt.Println()

	app := newApp()
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Warning: RabbitMQ connection failed: %v\n", err)
		fmt.Println("This example requires a running RabbitMQ server.")
		fmt.Println("Please start RabbitMQ and try again.")
		return
	}

	ch, err := getChannel(app)
	if err != nil {
		return
	}
	defer ch.Close()

	ctx := context.Background()

	queueName, err := demoDeclareAndPublish(ctx, ch)
	if err != nil {
		return
	}
	if err := demoConsume(ctx, ch, queueName); err != nil {
		return
	}
	if err := demoQueueInfo(ch); err != nil {
		return
	}

	fmt.Println("\n=== Example completed successfully ===")
}

// newApp 创建应用实例，配置名称与激活的 Profile。
func newApp() *boot.Boot {
	app, err := boot.NewApplication(
		boot.WithAppName("rabbitmq-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// getChannel 从容器的 Bean 中获取 RabbitMQ 通道。
func getChannel(app *boot.Boot) (*amqp.Channel, error) {
	ch, err := core.GetByName[*amqp.Channel](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get RabbitMQ channel: %v\n", err)
		return nil, fmt.Errorf("Failed to get RabbitMQ channel: %w", err)
	}
	return ch, nil
}

// demoDeclareAndPublish 演示声明队列并发布单条与多条消息（Demo 1-3）。
func demoDeclareAndPublish(ctx context.Context, ch *amqp.Channel) (string, error) {
	// Demo 1: Declare a queue
	fmt.Println("--- Demo 1: Declare Queue ---")
	queue, err := ch.QueueDeclare(
		"enhance-queue", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		fmt.Printf("Failed to declare queue: %v\n", err)
		return "", fmt.Errorf("Failed to declare queue: %w", err)
	}
	fmt.Printf("Queue declared: %s (messages: %d, consumers: %d)\n", queue.Name, queue.Messages, queue.Consumers)

	// Demo 2: Publish a message
	fmt.Println("\n--- Demo 2: Publish Message ---")
	body := `{"type": "user.created", "data": {"user_id": 123, "name": "John Doe"}}`
	err = ch.PublishWithContext(ctx,
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Body:         []byte(body),
		})
	if err != nil {
		fmt.Printf("Failed to publish message: %v\n", err)
		return "", fmt.Errorf("Failed to publish message: %w", err)
	}
	fmt.Println("Message published")

	// Demo 3: Publish multiple messages
	fmt.Println("\n--- Demo 3: Publish Multiple Messages ---")
	messages := []string{
		`{"type": "user.updated", "data": {"user_id": 123, "name": "John Smith"}}`,
		`{"type": "user.deleted", "data": {"user_id": 456}}`,
		`{"type": "order.created", "data": {"order_id": 789, "amount": 99.99}}`,
	}

	for i, msg := range messages {
		err = ch.PublishWithContext(ctx,
			"",
			queue.Name,
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Body:         []byte(msg),
			})
		if err != nil {
			fmt.Printf("Failed to publish message %d: %v\n", i+1, err)
			continue
		}
		fmt.Printf("Message %d published\n", i+1)
	}
	return queue.Name, nil
}

// demoConsume 演示在超时时间内的消息消费（Demo 4）。
func demoConsume(ctx context.Context, ch *amqp.Channel, queueName string) error {
	// Demo 4: Consume messages
	fmt.Println("\n--- Demo 4: Consume Messages ---")
	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		fmt.Printf("Failed to consume messages: %v\n", err)
		return fmt.Errorf("Failed to consume messages: %w", err)
	}

	// Read messages with timeout
	fmt.Println("Reading messages (timeout 5s)...")
	timeout := time.After(5 * time.Second)
	messageCount := 0

	for {
		select {
		case <-timeout:
			fmt.Printf("Timeout reached, consumed %d messages\n", messageCount)
			goto done
		case msg, ok := <-msgs:
			if !ok {
				goto done
			}
			fmt.Printf("Consumed: body=%s\n", string(msg.Body))
			messageCount++
			if messageCount >= 3 {
				goto done
			}
		}
	}

done:
	return nil
}

// demoQueueInfo 演示获取队列信息、清空与删除队列（Demo 5-7）。
func demoQueueInfo(ch *amqp.Channel) error {
	// Demo 5: Get queue info
	fmt.Println("\n--- Demo 5: Queue Info ---")
	queue, err := ch.QueueInspect("enhance-queue")
	if err != nil {
		fmt.Printf("Failed to get queue info: %v\n", err)
		return fmt.Errorf("Failed to get queue info: %w", err)
	}
	fmt.Printf("Queue: %s\n", queue.Name)
	fmt.Printf("Messages: %d\n", queue.Messages)
	fmt.Printf("Consumers: %d\n", queue.Consumers)

	// Demo 6: Purge queue
	fmt.Println("\n--- Demo 6: Purge Queue ---")
	purged, err := ch.QueuePurge("enhance-queue", false)
	if err != nil {
		fmt.Printf("Failed to purge queue: %v\n", err)
		return fmt.Errorf("Failed to purge queue: %w", err)
	}
	fmt.Printf("Purged %d messages\n", purged)

	// Demo 7: Delete queue
	fmt.Println("\n--- Demo 7: Delete Queue ---")
	_, err = ch.QueueDelete("enhance-queue", false, false, false)
	if err != nil {
		fmt.Printf("Failed to delete queue: %v\n", err)
		return fmt.Errorf("Failed to delete queue: %w", err)
	}
	fmt.Println("Queue deleted")
	return nil
}
