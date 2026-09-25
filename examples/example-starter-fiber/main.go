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

	app := newApp("fiber-example")
	defer app.Stop()

	if !startApp(app) {
		return
	}
	fiberApp, ok := getFiberApp(app)
	if !ok {
		return
	}
	registerRoutes(fiberApp)
	waitForSignal(app)
}

// newApp 创建 enhance 应用，创建失败时 panic。
func newApp(name string) *boot.Boot {
	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName(name),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// startApp 启动应用，触发 Fiber 自动配置。
func startApp(app *boot.Boot) bool {
	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return false
	}
	return true
}

// getFiberApp 从容器中获取 Fiber 应用实例。
func getFiberApp(app *boot.Boot) (*fiber.App, bool) {
	// Get the Fiber app from container
	fiberApp, err := core.GetByName[*fiber.App](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get fiber app: %v\n", err)
		return nil, false
	}
	return fiberApp, true
}

// registerRoutes 注册 HTTP 路由。
func registerRoutes(f *fiber.App) {
	// Register routes
	fmt.Println("--- Registering Routes ---")

	// Root route
	f.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Fiber Starter Example",
			"version": "1.0.0",
		})
	})

	// Hello route with query parameter
	f.Get("/hello", func(c *fiber.Ctx) error {
		name := c.Query("name", "World")
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Hello, %s!", name),
		})
	})

	// Health check endpoint
	f.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "UP",
		})
	})

	// User routes
	f.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON([]fiber.Map{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Doe", "email": "jane@example.com"},
		})
	})

	f.Get("/users/:id", func(c *fiber.Ctx) error {
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
}

// waitForSignal 阻塞等待退出信号。
func waitForSignal(app *boot.Boot) {
	app.WaitForSignal()
}
