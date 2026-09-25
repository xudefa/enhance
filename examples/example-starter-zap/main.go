// Package main demonstrates the Zap starter usage.
//
// This example shows how to use the Zap starter to:
// 1. Auto-configure Zap logger
// 2. Log messages at different levels
// 3. Use structured logging
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"go.uber.org/zap"

	_ "github.com/xudefa/enhance/starter/zap"
)

func main() {
	fmt.Println("=== Zap Starter Example ===")
	fmt.Println()

	app := newApp("zap-example")
	defer app.Stop()

	if !startApp(app) {
		return
	}
	logger, ok := getLogger(app)
	if !ok {
		return
	}
	logLevelsDemo(logger)
	logDirectZapDemo(app)

	fmt.Println("\n=== Example completed successfully ===")
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

// startApp 启动应用，触发 Zap 自动配置。
func startApp(app *boot.Boot) bool {
	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return false
	}
	return true
}

// getLogger 从容器中获取增强框架日志器。
func getLogger(app *boot.Boot) (log.Logger, bool) {
	// Get the logger from container
	logger, err := core.GetByName[log.Logger](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get logger: %v\n", err)
		return nil, false
	}
	return logger, true
}

// logLevelsDemo 演示不同级别与结构化日志输出。
func logLevelsDemo(logger log.Logger) {
	// Demonstrate logging at different levels
	fmt.Println("--- Logging at Different Levels ---")

	logger.Info(nil, "This is an info message")
	logger.Warn(nil, "This is a warning message")
	logger.Error(nil, "This is an error message")

	// Demonstrate structured logging
	fmt.Println("\n--- Structured Logging ---")

	logger.Info(nil, "User login",
		log.KeyValue{Key: "user_id", Value: 12345},
		log.KeyValue{Key: "username", Value: "john.doe"},
		log.KeyValue{Key: "ip_address", Value: "192.168.1.100"},
	)

	logger.Info(nil, "Request processed",
		log.KeyValue{Key: "method", Value: "GET"},
		log.KeyValue{Key: "path", Value: "/api/users"},
		log.KeyValue{Key: "status", Value: 200},
		log.KeyValue{Key: "duration_ms", Value: 42},
	)

	// Demonstrate logging with context
	fmt.Println("\n--- Logging with Context ---")

	logger.Info(nil, "Database query executed",
		log.KeyValue{Key: "query", Value: "SELECT * FROM users WHERE id = ?"},
		log.KeyValue{Key: "args", Value: []interface{}{123}},
		log.KeyValue{Key: "rows_affected", Value: 1},
		log.KeyValue{Key: "execution_time_ms", Value: 15},
	)

	// Demonstrate error logging
	fmt.Println("\n--- Error Logging ---")

	logger.Error(nil, "Failed to connect to database",
		log.KeyValue{Key: "error", Value: "connection refused"},
		log.KeyValue{Key: "host", Value: "localhost"},
		log.KeyValue{Key: "port", Value: 5432},
	)
}

// logDirectZapDemo 演示直接使用 Zap 日志器输出。
func logDirectZapDemo(app *boot.Boot) {
	// Demonstrate direct Zap usage
	fmt.Println("\n--- Direct Zap Usage ---")

	var zapLogger *zap.Logger
	zapLogger, err := core.GetByName[*zap.Logger](app.Container(), "")
	if err == nil {
		zapLogger.Info("Direct Zap logging",
			zap.String("service", "user-service"),
			zap.Int("version", 1),
			zap.Bool("debug", false),
		)
	}
}
