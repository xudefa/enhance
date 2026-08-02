// Package main demonstrates the Fiber starter usage.
//
// This example shows how to use the Fiber starter to:
// 1. Auto-configure Fiber web server
// 2. Register routes
// 3. Start the server
//
// Run:
//
//	go run main.go
//
// Test:
//
//	curl http://localhost:3000/
//	curl http://localhost:3000/hello?name=World
package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/fiber"
)

func main() {
	fmt.Println("=== Fiber Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("fiber-example"),
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

	// Get the Fiber app from container
	fiberApp, err := core.GetByName[*fiber.App](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get fiber app: %v\n", err)
		return
	}

	// Register routes
	fmt.Println("--- Registering Routes ---")

	// Root route
	fiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Fiber Starter Example",
			"version": "1.0.0",
		})
	})

	// Hello route with query parameter
	fiberApp.Get("/hello", func(c *fiber.Ctx) error {
		name := c.Query("name", "World")
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Hello, %s!", name),
		})
	})

	// Health check endpoint
	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "UP",
		})
	})

	// User routes
	fiberApp.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON([]fiber.Map{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Doe", "email": "jane@example.com"},
		})
	})

	fiberApp.Get("/users/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		return c.JSON(fiber.Map{
			"id":    id,
			"name":  "John Doe",
			"email": "john@example.com",
		})
	})

	fmt.Println("Routes registered:")
	fmt.Println("  GET / - Welcome message")
	fmt.Println("  GET /hello?name=World - Hello message")
	fmt.Println("  GET /health - Health check")
	fmt.Println("  GET /users - List users")
	fmt.Println("  GET /users/:id - Get user by ID")
	fmt.Println()
	fmt.Println("Server is running on http://localhost:3000")
	fmt.Println("Press Ctrl+C to stop")

	// Wait for signal
	app.WaitForSignal()
}
