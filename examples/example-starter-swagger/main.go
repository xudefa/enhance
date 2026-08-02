// Package main demonstrates the Swagger starter usage.
//
// This example shows how to use the Swagger starter to:
// 1. Auto-configure Swagger documentation
// 2. Generate API documentation
// 3. Serve Swagger UI
//
// Run:
//
//	go run main.go
//
// Test:
//
//	curl http://localhost:8080/swagger/
package main

import (
	"fmt"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/swagger"
)

func main() {
	fmt.Println("=== Swagger Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("swagger-example"),
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

	// Get Swagger config from container
	swaggerConfig, err := core.GetByName[*SwaggerConfig](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get swagger config: %v\n", err)
		return
	}

	fmt.Println("--- Swagger Configuration ---")
	fmt.Printf("Host: %s\n", swaggerConfig.Host)
	fmt.Printf("Port: %d\n", swaggerConfig.Port)
	fmt.Printf("URL: %s\n", swaggerConfig.URL)
	fmt.Printf("Title: %s\n", swaggerConfig.Title)

	fmt.Println("\n--- API Documentation ---")
	fmt.Println("Swagger UI will be available at: http://localhost:8080/swagger/")
	fmt.Println("API documentation will be available at: http://localhost:8080/swagger/doc.json")

	fmt.Println("\n--- Example API Endpoints ---")
	fmt.Println("GET /api/users - List all users")
	fmt.Println("GET /api/users/{id} - Get user by ID")
	fmt.Println("POST /api/users - Create a new user")
	fmt.Println("PUT /api/users/{id} - Update user")
	fmt.Println("DELETE /api/users/{id} - Delete user")

	fmt.Println("\n=== Example completed successfully ===")
	fmt.Println("Server is running on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop")

	// Wait for signal
	app.WaitForSignal()
}

// SwaggerConfig represents swagger configuration
type SwaggerConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	URL     string `json:"url"`
	Title   string `json:"title"`
}
