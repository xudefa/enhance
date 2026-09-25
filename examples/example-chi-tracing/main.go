package main

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/xudefa/enhance/actuator" // 触发 Actuator 自动配置注册
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	_ "github.com/xudefa/enhance/starter/chi" // 触发 Chi 自动配置注册
	"github.com/xudefa/enhance/tracing"
)

func main() {
	// 创建 Chi Router（不注册中间件，由自动配置处理）
	router := chi.NewRouter()

	app := newTracingApplication()
	defer app.Stop()

	// 将 Router 注册到容器中，自动配置会复用这个实例
	ctx := app.Context()
	registerRouter(ctx, router)

	// 启动应用（执行自动配置和启动器）
	// TracingAutoConfiguration 会创建并注册 Tracer
	// ChiAutoConfiguration 会从容器获取 Tracer 并注册 TracingMiddleware
	// 注意：Configure() 阶段会自动添加 TracingMiddleware 到 router
	startApplication(app)

	// 启动后获取 Tracer（由 TracingAutoConfiguration 创建）
	tracer := getTracer(ctx)

	// 注册查看链路数据和业务端点
	registerChiRoutes(router, tracer)

	// 打印启动信息并等待退出信号
	printStartupBanner()
	app.WaitForSignal()
}

// newTracingApplication 创建带链路追踪与 Actuator 配置的 Chi 应用，失败时直接退出。
func newTracingApplication() *boot.Boot {
	app, err := boot.NewApplication(
		boot.WithAppName("chi-tracing-example"),
		boot.WithProperty("chi.enabled", "true"),
		boot.WithProperty("chi.port", "8084"),
		boot.WithProperty("tracing.enabled", "true"),
		boot.WithProperty("tracing.service_name", "chi-demo"),
		boot.WithProperty("tracing.sampling_rate", "1.0"),
		boot.WithProperty("actuator.enabled", "true"),
		boot.WithProperty("actuator.path", "/actuator"),
	)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}
	return app
}

// registerRouter 将 Chi Router 注册到 IoC 容器中，自动配置会复用该实例。
func registerRouter(ctx boot.ApplicationContext, router chi.Router) {
	if err := ctx.Container().RegisterInstance(router, reflect.TypeFor[*chi.Mux]()); err != nil {
		log.Fatalf("Failed to register chi router: %v", err)
	}
}

// startApplication 启动应用，执行自动配置与启动器，失败时直接退出。
func startApplication(app *boot.Boot) {
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}

// getTracer 从容器中获取由 TracingAutoConfiguration 创建的 Tracer。
func getTracer(ctx boot.ApplicationContext) *tracing.Tracer {
	tracer, err := core.GetByName[*tracing.Tracer](ctx.Container(), "")
	if err != nil {
		log.Fatal("Tracer not found, please ensure tracing starter is enabled")
	}
	return tracer
}

// registerChiRoutes 注册查看链路数据与业务测试端点。
func registerChiRoutes(router chi.Router, tracer *tracing.Tracer) {
	// 注册查看链路数据的端点
	router.Get("/api/spans", func(w http.ResponseWriter, r *http.Request) {
		spans := tracer.GetSpans()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"total_spans": len(spans),
		})
	})

	// 注册业务路由
	router.Get("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello from Chi Tracing Example!",
		})
	})

	router.Get("/api/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "This is a test error",
		})
	})
}

// printStartupBanner 打印服务启动横幅与可用端点列表。
func printStartupBanner() {
	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	log.Println("========================================")
	log.Println("Chi Tracing Example with Actuator")
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
}
