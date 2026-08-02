// Package main demonstrates the Zerolog starter usage.
//
// This example shows how to use the Zerolog starter to:
// 1. Auto-configure Zerolog logger
// 2. Log messages at different levels
// 3. Use structured logging
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"

	_ "github.com/xudefa/enhance/starter/zerolog"
)

func main() {
	fmt.Println("=== Zerolog Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("zerolog-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	// Get the logger from container
	logger, err := core.GetByName[log.Logger](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get logger: %v\n", err)
		return
	}

	// Demonstrate logging at different levels
	fmt.Println("--- Logging at Different Levels ---")

	logger.Info(nil, "This is an info message")
	logger.Warn(nil, "This is a warning message")
	logger.Error(nil, "This is an error message")

	// Demonstrate structured logging
	fmt.Println("\n--- Structured Logging ---")

	logger.Info(nil, "User login",
		log.KeyValue{Key: "user_id", Value: 12345},
		log.KeyValue{Key: "username", Value: "john.doe"},
		log.KeyValue{Key: "ip_address", Value: "192.168.1.100"},
	)

	logger.Info(nil, "Request processed",
		log.KeyValue{Key: "method", Value: "GET"},
		log.KeyValue{Key: "path", Value: "/api/users"},
		log.KeyValue{Key: "status", Value: 200},
		log.KeyValue{Key: "duration_ms", Value: 42},
	)

	// Demonstrate logging with context
	fmt.Println("\n--- Logging with Context ---")

	logger.Info(nil, "Database query executed",
		log.KeyValue{Key: "query", Value: "SELECT * FROM users WHERE id = ?"},
		log.KeyValue{Key: "args", Value: []interface{}{123}},
		log.KeyValue{Key: "rows_affected", Value: 1},
		log.KeyValue{Key: "execution_time_ms", Value: 15},
	)

	// Demonstrate error logging
	fmt.Println("\n--- Error Logging ---")

	logger.Error(nil, "Failed to connect to database",
		log.KeyValue{Key: "error", Value: "connection refused"},
		log.KeyValue{Key: "host", Value: "localhost"},
		log.KeyValue{Key: "port", Value: 5432},
	)

	// Demonstrate direct Zerolog usage
	fmt.Println("\n--- Direct Zerolog Usage ---")

	var zerologLogger *zerolog.Logger
	zerologLogger, err = core.GetByName[*zerolog.Logger](app.Container(), "")
	if err == nil {
		zerologLogger.Info().
			Str("service", "user-service").
			Int("version", 1).
			Bool("debug", false).
			Msg("Direct Zerolog logging")
	}

	fmt.Println("\n=== Example completed successfully ===")
}
