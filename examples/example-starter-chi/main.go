// Package main demonstrates the Chi starter usage.
//
// This example shows how to use the Chi starter to:
// 1. Auto-configure Chi router
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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/chi"
)

func main() {
	fmt.Println("=== Chi Starter Example ===")
	fmt.Println()

	app := newApp("chi-example")
	defer app.Stop()

	if !startApp(app) {
		return
	}
	router, ok := getChiRouter(app)
	if !ok {
		return
	}
	registerRoutes(router)
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

// startApp 启动应用，触发 Chi 自动配置。
func startApp(app *boot.Boot) bool {
	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return false
	}
	return true
}

// getChiRouter 从容器中获取 Chi 路由器。
func getChiRouter(app *boot.Boot) (*chi.Mux, bool) {
	// Get the Chi router from container
	chiRouter, err := core.GetByName[*chi.Mux](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get chi router: %v\n", err)
		return nil, false
	}
	return chiRouter, true
}

// registerRoutes 注册中间件与 HTTP 路由。
func registerRoutes(router *chi.Mux) {
	// Use middleware
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)

	// Register routes
	fmt.Println("--- Registering Routes ---")

	// Root route
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to Chi Starter Example", "version": "1.0.0"}`))
	})

	// Hello route with query parameter
	router.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "World"
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"message": "Hello, %s!"}`, name)))
	})

	// Health check endpoint
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "UP"}`))
	})

	// User routes with sub-router
	router.Route("/users", func(subRouter chi.Router) {
		subRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"id": 1, "name": "John Doe"}, {"id": 2, "name": "Jane Doe"}]`))
		})

		subRouter.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"id": "%s", "name": "John Doe"}`, id)))
		})
	})

	fmt.Println("Routes registered:")
	fmt.Println("  GET / - Welcome message")
	fmt.Println("  GET /hello?name=World - Hello message")
	fmt.Println("  GET /health - Health check")
	fmt.Println("  GET /users - List users")
	fmt.Println("  GET /users/{id} - Get user by ID")
	fmt.Println()
	fmt.Println("Server is running on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop")
}

// waitForSignal 阻塞等待退出信号。
func waitForSignal(app *boot.Boot) {
	app.WaitForSignal()
}
