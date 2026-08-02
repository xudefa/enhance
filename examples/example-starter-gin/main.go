// Package main demonstrates the Gin starter usage.
//
// This example shows how to use the Gin starter to:
// 1. Auto-configure Gin web server
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

	"github.com/gin-gonic/gin"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/gin"
)

func main() {
	fmt.Println("=== Gin Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("gin-example"),
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

	// Get the Gin engine from container
	engine, err := core.GetByName[*gin.Engine](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get gin engine: %v\n", err)
		return
	}

	// Register routes
	fmt.Println("--- Registering Routes ---")

	// Root route
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Gin Starter Example",
			"version": "1.0.0",
		})
	})

	// Hello route with query parameter
	engine.GET("/hello", func(c *gin.Context) {
		name := c.DefaultQuery("name", "World")
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Hello, %s!", name),
		})
	})

	// Health check endpoint
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})

	fmt.Println("Routes registered:")
	fmt.Println("  GET / - Welcome message")
	fmt.Println("  GET /hello?name=World - Hello message")
	fmt.Println("  GET /health - Health check")
	fmt.Println()
	fmt.Println("Server is running on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop")

	// Wait for signal
	app.WaitForSignal()
}
