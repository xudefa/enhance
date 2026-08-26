// Package main demonstrates the enhance web REST API:
// HTTP server setup with middleware, route registration,
// request handling, error handling, and graceful shutdown.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/xudefa/enhance/web/core"
	"github.com/xudefa/enhance/web/engine"
	"github.com/xudefa/enhance/web/engine/stdlib"
	"github.com/xudefa/enhance/web/server"
)

func main() {
	fmt.Println("=== enhance Web REST API Example ===")
	fmt.Println()

	// ---- 1. Create router ----
	router := server.NewRouter()

	// ---- 2. Add global middleware ----
	router.Use(func(ctx core.Context) {
		start := time.Now()
		fmt.Printf("  [middleware] %s %s - started\n", ctx.RequestMethod(), ctx.RequestURI())
		ctx.Next()
		fmt.Printf("  [middleware] %s %s - completed in %v\n",
			ctx.RequestMethod(), ctx.RequestURI(), time.Since(start))
	})

	// Recovery middleware
	router.Use(func(ctx core.Context) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("  [recovery] Panic recovered: %v\n", r)
				ctx.JSON(http.StatusInternalServerError, map[string]any{
					"error": "internal server error",
				})
			}
		}()
		ctx.Next()
	})

	// ---- 3. Register routes ----
	router.GET("/", func(ctx core.Context) {
		ctx.JSON(http.StatusOK, map[string]any{
			"message": "Welcome to enhance REST API",
			"version": "1.0.0",
		})
	})

	// GET /users - list users
	router.GET("/users", func(ctx core.Context) {
		users := []map[string]any{
			{"id": 1, "name": "Alice", "email": "alice@example.com"},
			{"id": 2, "name": "Bob", "email": "bob@example.com"},
			{"id": 3, "name": "Charlie", "email": "charlie@example.com"},
		}
		ctx.JSON(http.StatusOK, map[string]any{
			"users": users,
			"total": len(users),
		})
	})

	// GET /users/:id - get user by ID
	router.GET("/users/{id}", func(ctx core.Context) {
		id := ctx.PathParam("id")
		if id == "" {
			ctx.JSON(http.StatusBadRequest, map[string]any{"error": "id is required"})
			return
		}
		ctx.JSON(http.StatusOK, map[string]any{
			"id":    id,
			"name":  "Alice",
			"email": "alice@example.com",
		})
	})

	// POST /users - create user
	router.POST("/users", func(ctx core.Context) {
		var body struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := ctx.BindJSON(&body); err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"error":   "invalid request body",
				"details": err.Error(),
			})
			return
		}
		if body.Name == "" {
			ctx.JSON(http.StatusBadRequest, map[string]any{"error": "name is required"})
			return
		}
		ctx.JSON(http.StatusCreated, map[string]any{
			"id":      4,
			"name":    body.Name,
			"email":   body.Email,
			"message": "user created",
		})
	})

	// GET /health - health check
	router.GET("/health", func(ctx core.Context) {
		ctx.JSON(http.StatusOK, map[string]any{
			"status": "UP",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// GET /error - error handling demo
	router.GET("/error", func(ctx core.Context) {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"error":   "something went wrong",
			"code":    "INTERNAL_ERROR",
			"details": "simulated error for demo",
		})
	})

	// ---- 4. Create server ----
	srv := stdlib.NewServer(
		engine.WithHost("127.0.0.1"),
		engine.WithPort(9999),
		engine.WithReadTimeout(10),
		engine.WithWriteTimeout(10),
	)
	srv.SetHandler(router)

	// ---- 5. Start server in background ----
	go func() {
		fmt.Println("  Server starting on http://127.0.0.1:9999")
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}()

	// Wait for server to be ready
	time.Sleep(100 * time.Millisecond)

	// ---- 6. Make test requests ----
	fmt.Println()
	fmt.Println("--- Test Requests ---")

	testRequests()

	// ---- 7. Graceful shutdown ----
	fmt.Println()
	fmt.Println("--- Graceful Shutdown ---")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("  Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "  Shutdown error: %v\n", err)
	}
	fmt.Println("  Server stopped gracefully")

	fmt.Println()
	fmt.Println("=== Example completed successfully ===")
}

// testRequests makes HTTP requests to the running server.
func testRequests() {
	client := &http.Client{Timeout: 5 * time.Second}

	tests := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/", ""},
		{"GET", "/users", ""},
		{"GET", "/users/1", ""},
		{"POST", "/users", `{"name":"Dave","email":"dave@example.com"}`},
		{"GET", "/health", ""},
		{"GET", "/error", ""},
	}

	for _, tt := range tests {
		var req *http.Request
		var err error

		if tt.body != "" {
			req, err = http.NewRequest(tt.method, "http://127.0.0.1:9999"+tt.path,
				strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req, err = http.NewRequest(tt.method, "http://127.0.0.1:9999"+tt.path, nil)
		}
		if err != nil {
			fmt.Printf("  [%s %s] Request error: %v\n", tt.method, tt.path, err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("  [%s %s] HTTP error: %v\n", tt.method, tt.path, err)
			continue
		}
		_ = resp.Body.Close()

		status := "OK"
		if resp.StatusCode >= 400 {
			status = "ERROR"
		}
		fmt.Printf("  [%s %s] -> %d %s\n", tt.method, tt.path, resp.StatusCode, status)
	}
}
