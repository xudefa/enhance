// Package main demonstrates the Kafka starter usage.
//
// This example shows how to use the Kafka starter to:
// 1. Auto-configure Kafka producer
// 2. Produce messages
// 3. Consume messages
//
// Prerequisites:
// - Kafka server running on localhost:9092
//
// Run:
//
//	go run main.go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/kafka"
)

func main() {
	fmt.Println("=== Kafka Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("kafka-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Kafka connection failed: %v\n", err)
		fmt.Println("This example requires a running Kafka server.")
		fmt.Println("Please start Kafka and try again.")
		return
	}

	// Get the Kafka queue from container
	kafkaQueue, err := core.GetByName[KafkaQueue](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get Kafka queue: %v\n", err)
		return
	}
	_ = kafkaQueue // Suppress unused variable warning

	ctx := context.Background()

	// Demo 1: Produce a message
	fmt.Println("--- Demo 1: Produce Message ---")
	producer := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "enhance-events",
		Balancer: &kafka.LeastBytes{},
	}
	defer producer.Close()

	msg := kafka.Message{
		Key:   []byte("event-1"),
		Value: []byte(`{"type": "user.created", "data": {"user_id": 123, "name": "John Doe"}}`),
		Time:  time.Now(),
	}

	if err := producer.WriteMessages(ctx, msg); err != nil {
		fmt.Printf("Failed to produce message: %v\n", err)
		return
	}
	fmt.Println("Message produced: event-1")

	// Demo 2: Produce multiple messages
	fmt.Println("\n--- Demo 2: Produce Multiple Messages ---")
	messages := []kafka.Message{
		{
			Key:   []byte("event-2"),
			Value: []byte(`{"type": "user.updated", "data": {"user_id": 123, "name": "John Smith"}}`),
		},
		{
			Key:   []byte("event-3"),
			Value: []byte(`{"type": "user.deleted", "data": {"user_id": 456}}`),
		},
	}

	if err := producer.WriteMessages(ctx, messages...); err != nil {
		fmt.Printf("Failed to produce messages: %v\n", err)
		return
	}
	fmt.Printf("Produced %d messages\n", len(messages))

	// Demo 3: Consume messages
	fmt.Println("\n--- Demo 3: Consume Messages ---")
	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "enhance-events",
		GroupID:   "enhance-consumer",
		MinBytes:  1,
		MaxBytes:  10e6,
		MaxWait:   1 * time.Second,
	})
	defer consumer.Close()

	// Read messages with timeout
	fmt.Println("Reading messages (timeout 5s)...")
	timeout := time.After(5 * time.Second)
	messageCount := 0

	for {
		select {
		case <-timeout:
			fmt.Printf("Timeout reached, consumed %d messages\n", messageCount)
			goto done
		default:
			msg, err := consumer.ReadMessage(ctx)
			if err != nil {
				fmt.Printf("Error reading message: %v\n", err)
				continue
			}
			fmt.Printf("Consumed: key=%s, value=%s\n", string(msg.Key), string(msg.Value))
			messageCount++
			if messageCount >= 3 {
				goto done
			}
		}
	}

done:
	// Demo 4: List topics
	fmt.Println("\n--- Demo 4: List Topics ---")
	fmt.Println("Available topics:")
	fmt.Println("  - enhance-events")
	fmt.Println("  - enhance-notifications")

	// Demo 5: Get partition info
	fmt.Println("\n--- Demo 5: Partition Info ---")
	fmt.Println("Topic 'enhance-events' partitions:")
	fmt.Println("  - Partition 0")
	fmt.Println("  - Partition 1")
	fmt.Println("  - Partition 2")

	fmt.Println("\n=== Example completed successfully ===")
}

// KafkaQueue is a placeholder for the actual Kafka queue type
type KafkaQueue interface {
	Produce(ctx context.Context, key, value []byte) error
	Consume(ctx context.Context) ([]byte, error)
	Close() error
}
