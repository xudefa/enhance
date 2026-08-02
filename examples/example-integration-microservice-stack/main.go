// Package main demonstrates a microservice stack integration.
//
// Combines starters for a production-ready microservice:
// - Gin: HTTP server
// - Zerolog: Structured logging
// - Redis: Caching
// - Prometheus: Metrics
// - OpenTelemetry: Distributed tracing
// - JWT: Authentication
// - Cron: Background tasks
//
// Run:
//
//	go run main.go
//
// Note: Some starters require external services (Redis, etc.)
// The example will start and log warnings for unavailable services.
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/metrics"

	_ "github.com/xudefa/enhance/starter/cron"
	_ "github.com/xudefa/enhance/starter/gin"
	_ "github.com/xudefa/enhance/starter/jwt"
	_ "github.com/xudefa/enhance/starter/otel"
	_ "github.com/xudefa/enhance/starter/prometheus"
	_ "github.com/xudefa/enhance/starter/redis"
	_ "github.com/xudefa/enhance/starter/zerolog"
)

func main() {
	fmt.Println("=== Microservice Stack Integration ===")
	fmt.Println()
	fmt.Println("Starters: gin + zerolog + redis + prometheus + otel + jwt + cron")
	fmt.Println()

	app, err := boot.NewApplication(
		boot.WithAppName("microservice-stack"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	logger, _ := core.GetByName[log.Logger](app.Container(), "")
	engine, _ := core.GetByName[*gin.Engine](app.Container(), "")
	registry, _ := core.GetByName[metrics.MeterRegistry](app.Container(), "")
	cronScheduler, _ := core.GetByName[*cron.Cron](app.Container(), "")

	requestCounter := registry.Counter("microservice_requests_total")
	errorCounter := registry.Counter("microservice_errors_total")
	latencyHistogram := registry.Histogram("microservice_request_duration_seconds")

	// Background task
	cronScheduler.AddFunc("*/60 * * * * *", func() {
		if logger != nil {
			logger.Info(nil, "Scheduled task: cache cleanup")
		}
	})

	// Middleware
	engine.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start).Seconds()
		latencyHistogram.Record(latency)
		requestCounter.Inc()
	})

	// Routes
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "microservice-stack",
			"version": "1.0.0",
			"stack":   []string{"gin", "zerolog", "redis", "prometheus", "otel", "jwt", "cron"},
		})
	})

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Protected routes group
	protected := engine.Group("/api")
	{
		protected.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":   "running",
				"uptime":   time.Since(time.Now()).String(),
				"services": "all systems operational",
			})
		})

		protected.POST("/cache/set", func(c *gin.Context) {
			var req struct {
				Key   string `json:"key" binding:"required"`
				Value string `json:"value" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				errorCounter.Inc()
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if logger != nil {
				logger.Info(nil, "Cache SET",
					log.KeyValue{Key: "key", Value: req.Key},
				)
			}
			c.JSON(http.StatusOK, gin.H{"message": "cached", "key": req.Key})
		})

		protected.GET("/cache/:key", func(c *gin.Context) {
			key := c.Param("key")
			if logger != nil {
				logger.Info(nil, "Cache GET",
					log.KeyValue{Key: "key", Value: key},
				)
			}
			c.JSON(http.StatusOK, gin.H{"key": key, "value": "cached-value"})
		})
	}

	// Metrics endpoint
	engine.GET("/metrics", func(c *gin.Context) {
		allMetrics := registry.Collect()
		result := make(map[string]any)
		for _, m := range allMetrics {
			result[m.Name] = m.Value
		}
		c.JSON(http.StatusOK, result)
	})

	fmt.Println("Routes:")
	fmt.Println("  GET  /              - Service info")
	fmt.Println("  GET  /health        - Health check")
	fmt.Println("  GET  /api/status    - Service status")
	fmt.Println("  POST /api/cache/set - Set cache value")
	fmt.Println("  GET  /api/cache/:key - Get cache value")
	fmt.Println("  GET  /metrics       - Prometheus metrics")
	fmt.Println()
	fmt.Println("Server: http://localhost:8080")
	fmt.Println("Metrics: http://localhost:9090")
	fmt.Println("Press Ctrl+C to stop")

	cronScheduler.Start()
	app.WaitForSignal()
}
