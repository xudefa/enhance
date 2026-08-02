// Package main demonstrates the Echo starter usage.
//
// This example shows how to use the Echo starter to:
// 1. Auto-configure Echo web server
// 2. Register routes
// 3. Start the server
//
// Run:
//
//	go run main.go
//
// Test:
//
//	curl http://localhost:8080/
//	curl http://localhost:8080/hello?name=World
package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/echo"
)

func main() {
	fmt.Println("=== Echo Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("echo-example"),
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

	// Get the Echo instance from container
	e, err := core.GetByName[*echo.Echo](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get echo instance: %v\n", err)
		return
	}

	// Register middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Register routes
	fmt.Println("--- Registering Routes ---")

	// Root route
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Welcome to Echo Starter Example",
			"version": "1.0.0",
		})
	})

	// Hello route with query parameter
	e.GET("/hello", func(c echo.Context) error {
		name := c.QueryParam("name")
		if name == "" {
			name = "World"
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": fmt.Sprintf("Hello, %s!", name),
		})
	})

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "UP",
		})
	})

	// User routes
	e.GET("/users", func(c echo.Context) error {
		return c.JSON(http.StatusOK, []map[string]interface{}{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Doe", "email": "jane@example.com"},
		})
	})

	e.GET("/users/:id", func(c echo.Context) error {
		id := c.Param("id")
		return c.JSON(http.StatusOK, map[string]interface{}{
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
	fmt.Println("Server is running on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop")

	// Wait for signal
	app.WaitForSignal()
}
