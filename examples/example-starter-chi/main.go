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

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("chi-example"),
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

	// Get the Chi router from container
	r, err := core.GetByName[*chi.Mux](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get chi router: %v\n", err)
		return
	}

	// Use middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Register routes
	fmt.Println("--- Registering Routes ---")

	// Root route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "Welcome to Chi Starter Example", "version": "1.0.0"}`))
	})

	// Hello route with query parameter
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "World"
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"message": "Hello, %s!"}`, name)))
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "UP"}`))
	})

	// User routes with sub-router
	r.Route("/users", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`[{"id": 1, "name": "John Doe"}, {"id": 2, "name": "Jane Doe"}]`))
		})

		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
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

	// Wait for signal
	app.WaitForSignal()
}
