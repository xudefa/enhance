package main

import (
	"log"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v2"
	_ "github.com/xudefa/enhance/actuator" // 触发 Actuator 自动配置注册
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	_ "github.com/xudefa/enhance/starter/fiber" // 触发 Fiber 自动配置注册
	"github.com/xudefa/enhance/tracing"
)

func main() {
	// 创建 Fiber App（不注册中间件，由自动配置处理）
	fiberApp := fiber.New()

	app, err := boot.NewApplication(
		boot.WithAppName("fiber-tracing-example"),
		boot.WithProperty("fiber.enabled", "true"),
		boot.WithProperty("fiber.port", "8082"),
		boot.WithProperty("tracing.enabled", "true"),
		boot.WithProperty("tracing.service_name", "fiber-demo"),
		boot.WithProperty("tracing.sampling_rate", "1.0"),
		boot.WithProperty("actuator.enabled", "true"),
		boot.WithProperty("actuator.path", "/actuator"),
	)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	defer app.Stop()

	// 将 Fiber App 注册到容器中，自动配置会复用这个实例
	ctx := app.Context()
	if err := ctx.Container().RegisterInstance(fiberApp, reflect.TypeFor[*fiber.App]()); err != nil {
		log.Fatalf("Failed to register fiber app: %v", err)
	}

	// 启动应用（执行自动配置和启动器）
	// TracingAutoConfiguration 会创建并注册 Tracer
	// FiberAutoConfiguration 会从容器获取 Tracer 并注册 TracingMiddleware
	// 注意：Configure() 阶段会自动添加 TracingMiddleware 到 fiberApp
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}

	// 启动后获取 Tracer（由 TracingAutoConfiguration 创建）
	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		log.Fatal("Tracer not found, please ensure tracing starter is enabled")
	}

	// 注册查看链路数据的端点
	fiberApp.Get("/api/spans", func(c *fiber.Ctx) error {
		spans := tracer.GetSpans()
		return c.JSON(fiber.Map{
			"total_spans": len(spans),
		})
	})

	// 注册业务路由
	fiberApp.Get("/api/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello from Fiber Tracing Example!",
		})
	})

	fiberApp.Get("/api/error", func(c *fiber.Ctx) error {
		return c.Status(500).JSON(fiber.Map{
			"error": "This is a test error",
		})
	})

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	log.Println("========================================")
	log.Println("Fiber Tracing Example with Actuator")
	log.Println("========================================")
	log.Println("Business endpoints:")
	log.Println("  GET /api/hello    - Hello endpoint")
	log.Println("  GET /api/error    - Error test endpoint")
	log.Println("  GET /api/spans    - View tracing spans")
	log.Println("Actuator monitoring endpoints:")
	log.Println("  GET /actuator/health  - Health check")
	log.Println("  GET /actuator/metrics - Application metrics")
	log.Println("  GET /actuator/env     - Environment info")
	log.Println("  GET /actuator/beans   - Bean list")
	log.Println("  GET /actuator/info    - Application info")
	log.Println("========================================")

	app.WaitForSignal()
}
