// Package main demonstrates an integrated web application stack.
//
// This example combines multiple starters to build a complete web application:
// - Gin: HTTP server and routing
// - Zerolog: Structured logging
// - Prometheus: Metrics collection
// - Cron: Scheduled tasks
// - Validator: Request validation
// - Swagger: API documentation
//
// Run:
//
//	go run main.go
//
// Test:
//
//	curl http://localhost:8080/
//	curl http://localhost:8080/api/users
//	curl http://localhost:8080/metrics
//	curl http://localhost:8080/health
package main

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/metrics"

	_ "github.com/xudefa/enhance/starter/cron"
	_ "github.com/xudefa/enhance/starter/gin"
	_ "github.com/xudefa/enhance/starter/prometheus"
	_ "github.com/xudefa/enhance/starter/swagger"
	_ "github.com/xudefa/enhance/starter/validator"
	_ "github.com/xudefa/enhance/starter/zerolog"
)

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=0,lte=130"`
}

// User represents a user entity
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

var (
	users  = []User{}
	nextID = int64(1)
)

// webStackComponents Web 栈核心组件集合，供定时任务与路由注册使用。
type webStackComponents struct {
	logger        log.Logger
	engine        *gin.Engine
	validate      *validator.Validate
	registry      metrics.MeterRegistry
	cronScheduler *cron.Cron
	requestCtr    metrics.Counter
}

func main() {
	fmt.Println("=== Integrated Web Stack Example ===")
	fmt.Println()
	fmt.Println("Starters: gin + zerolog + prometheus + cron + validator + swagger")
	fmt.Println()

	app := buildWebStackApplication()
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	components := setupWebStack(app)
	registerWebStackRoutes(components)
	printWebStackInfo()

	components.cronScheduler.Start()
	app.WaitForSignal()
}

// buildWebStackApplication 创建 Web 栈演示应用。
func buildWebStackApplication() *boot.Boot {
	app, err := boot.NewApplication(
		boot.WithAppName("web-stack-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// setupWebStack 从容器获取组件，注册定时任务与指标。
func setupWebStack(app *boot.Boot) *webStackComponents {
	comp := &webStackComponents{}
	comp.logger, _ = core.GetByName[log.Logger](app.Container(), "")
	comp.engine, _ = core.GetByName[*gin.Engine](app.Container(), "")
	comp.validate, _ = core.GetByName[*validator.Validate](app.Container(), "")
	comp.registry, _ = core.GetByName[metrics.MeterRegistry](app.Container(), "")
	comp.cronScheduler, _ = core.GetByName[*cron.Cron](app.Container(), "")

	// Register cron jobs
	comp.cronScheduler.AddFunc("*/30 * * * * *", func() {
		if comp.logger != nil {
			comp.logger.Info(nil, "Scheduled task: heartbeat check")
		}
	})

	// Register metrics
	comp.requestCtr = comp.registry.Counter("http_requests_total")
	comp.requestCtr.Inc()

	return comp
}

// registerWebStackRoutes 注册 Web 栈的用户管理路由。
func registerWebStackRoutes(comp *webStackComponents) {
	engine := comp.engine

	// Register routes
	engine.GET("/", func(c *gin.Context) {
		comp.requestCtr.Inc()
		c.JSON(http.StatusOK, gin.H{
			"message": "Integrated Web Stack Example",
			"stack":   []string{"gin", "zerolog", "prometheus", "cron", "validator", "swagger"},
		})
	})

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	api := engine.Group("/api")
	{
		api.GET("/users", func(c *gin.Context) {
			comp.requestCtr.Inc()
			c.JSON(http.StatusOK, users)
		})

		api.POST("/users", func(c *gin.Context) {
			comp.requestCtr.Inc()
			var req CreateUserRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err := comp.validate.Struct(req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			user := User{
				ID:    int(atomic.AddInt64(&nextID, 1) - 1),
				Name:  req.Name,
				Email: req.Email,
				Age:   req.Age,
			}
			users = append(users, user)
			if comp.logger != nil {
				comp.logger.Info(nil, "User created",
					log.KeyValue{Key: "user_id", Value: user.ID},
					log.KeyValue{Key: "name", Value: user.Name},
				)
			}
			c.JSON(http.StatusCreated, user)
		})

		api.GET("/users/:id", func(c *gin.Context) {
			comp.requestCtr.Inc()
			id := c.Param("id")
			for _, u := range users {
				if fmt.Sprintf("%d", u.ID) == id {
					c.JSON(http.StatusOK, u)
					return
				}
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		})
	}
}

// printWebStackInfo 输出路由清单与启动信息。
func printWebStackInfo() {
	fmt.Println("Routes:")
	fmt.Println("  GET  /             - Welcome message")
	fmt.Println("  GET  /health       - Health check")
	fmt.Println("  GET  /api/users    - List users")
	fmt.Println("  POST /api/users    - Create user")
	fmt.Println("  GET  /api/users/:id - Get user by ID")
	fmt.Println("  GET  /metrics      - Prometheus metrics")
	fmt.Println()
	fmt.Println("Server is running on http://localhost:8080")
	fmt.Println("Metrics: http://localhost:9090/metrics")
	fmt.Println("Press Ctrl+C to stop")
}
