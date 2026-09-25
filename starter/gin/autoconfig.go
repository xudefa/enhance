// Package gin 提供 Gin Web 框架自动配置。
package gin

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xudefa/enhance/actuator"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/tracing"
)

var ginAutoConfig = &GinAutoConfiguration{}

func init() {
	boot.RegisterAutoConfigWith(ginAutoConfig,
		boot.WithConditions(
			// 约定优于配置：当 gin.enabled 未配置时，默认为 true（即默认启用）
			condition.OnPropertyOrDefault(GinEnabled, ConditionTrue, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityWebLayer)),
	)
	// 注册为 Starter，使其 Start/Stop 生命周期方法被自动调用
	boot.RegisterStarter(ginAutoConfig)
}

// GinAutoConfiguration Gin Web 框架自动配置类。
type GinAutoConfiguration struct {
	logger     log.Logger
	engine     *gin.Engine
	server     *http.Server
	config     *GinConfig
	tracer     *tracing.Tracer
	mu         sync.Mutex      // 保护 Configure 的并发访问
	configured bool            // 标记是否已配置，防止同一应用上下文重复配置
	ctx        context.Context // 应用上下文
}

// Configure 配置 Gin Web 服务器。
func (c *GinAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 同一应用上下文的 AutoConfig 与 Starter 双注册会调用两次 Configure，直接跳过
	if c.configured && c.ctx == ctx.Context() {
		return nil
	}
	// 新应用上下文（应用重启）时重新配置，更新 ctx/server 等状态
	c.configured = false

	container := ctx.Container()
	env := ctx.Environment()

	c.resolveLogger(container)

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Gin config: %w", err)
	}
	c.config = cfg
	c.applyMode(cfg.Mode)

	c.buildEngine(ctx, container, cfg)

	c.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: c.engine,
	}

	if err := c.registerComponents(ctx, container); err != nil {
		return fmt.Errorf("failed to register Gin components: %w", err)
	}

	c.logger.Info(ctx.Context(), "Gin Web server configured",
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "host", Value: cfg.Host},
		log.KeyValue{Key: "mode", Value: cfg.Mode},
	)

	c.configured = true
	// 存储应用上下文
	c.ctx = ctx.Context()

	return nil
}

// resolveLogger 从容器获取日志记录器，缺省时使用默认实现。
func (c *GinAutoConfiguration) resolveLogger(container core.Container) {
	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}
}

// applyMode 根据配置设置 Gin 的运行模式。
func (c *GinAutoConfiguration) applyMode(mode string) {
	switch mode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}

// buildEngine 获取容器中已存在的 Engine 或创建默认 Engine，并附加中间件与追踪。
func (c *GinAutoConfiguration) buildEngine(ctx boot.ApplicationContext, container core.Container, cfg *GinConfig) {
	if engine, err := core.GetByName[*gin.Engine](container, ""); err == nil {
		c.engine = engine
		c.logger.Info(ctx.Context(), "using existing Gin Engine instance from container")
	} else {
		c.engine = gin.New()
		if cfg.EnableRecover {
			c.engine.Use(gin.Recovery())
		}
		if cfg.EnableLogger {
			c.engine.Use(gin.Logger())
		}
	}

	// 从容器获取 Tracer 并注册 tracing 中间件
	if tracer, err := core.GetByName[*tracing.Tracer](container, ""); err == nil {
		c.tracer = tracer
		c.engine.Use(TracingMiddleware(tracer))
		c.logger.Info(ctx.Context(), "Gin tracing middleware enabled")
	}
}

// registerComponents 将 config、engine、server、endpointRegistry 注册到容器。
func (c *GinAutoConfiguration) registerComponents(ctx boot.ApplicationContext, container core.Container) error {
	if err := container.RegisterInstance(c.config, reflect.TypeFor[*GinConfig]()); err != nil {
		return fmt.Errorf("failed to register Gin Config: %w", err)
	}

	// 如果 Engine 已存在，跳过注册
	if _, err := core.GetByName[*gin.Engine](container, ""); err != nil {
		if err := container.RegisterInstance(c.engine, reflect.TypeFor[*gin.Engine]()); err != nil {
			return fmt.Errorf("failed to register Gin Engine: %w", err)
		}
	}

	if err := container.RegisterInstance(c.server, reflect.TypeFor[*http.Server]()); err != nil {
		return fmt.Errorf("failed to register HTTP Server: %w", err)
	}

	// 注册 HttpEndpointRegistry,允许 Actuator 等模块自动挂载端点到 Gin
	endpointRegistry := NewGinEndpointRegistry(c.engine)
	if err := container.RegisterInstance(endpointRegistry, reflect.TypeFor[actuator.HttpEndpointRegistry]()); err != nil {
		c.logger.Warn(ctx.Context(), "failed to register HttpEndpointRegistry, Actuator endpoints will not be mounted automatically",
			log.KeyValue{Key: "error", Value: err.Error()},
		)
	}
	return nil
}

// Start 启动 Gin Web 服务器。
func (c *GinAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.server == nil {
		return fmt.Errorf("Gin HTTP Server not initialized")
	}
	c.logger.Info(ctx.Context(), "Gin Web server starting",
		log.KeyValue{Key: "addr", Value: c.server.Addr},
	)

	go c.runGinServer()

	return nil
}

// runGinServer 在独立 goroutine 中运行 Gin 服务器。
func (c *GinAutoConfiguration) runGinServer() {
	if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		c.logger.Error(c.ctx, "Gin Web server error",
			log.KeyValue{Key: "error", Value: err.Error()},
		)
	}
}

// Stop 停止 Gin Web 服务器。
func (c *GinAutoConfiguration) Stop(ctx boot.ApplicationContext) error {
	if c.server == nil {
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(ctx.Context(), 30*time.Second)
	defer cancel()
	return c.server.Shutdown(shutdownCtx)
}

// Name 返回启动器名称。
func (c *GinAutoConfiguration) Name() string {
	return "GinStarter"
}

// Dependencies 返回依赖的其他启动器名称。
func (c *GinAutoConfiguration) Dependencies() []string {
	return nil
}

// GetCondition 返回启动器条件。
func (c *GinAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnPropertyOrDefault(GinEnabled, ConditionTrue, ConditionTrue)
}

// GetEngine 从容器中获取 Gin 引擎实例。
func GetEngine(container core.Container) (*gin.Engine, error) {
	return core.GetByName[*gin.Engine](container, "")
}

// GetServer 从容器中获取 HTTP 服务器实例。
func GetServer(container core.Container) (*http.Server, error) {
	return core.GetByName[*http.Server](container, "")
}

// GinConfig Gin Web 服务器配置。
type GinConfig struct {
	Enabled       bool   `json:"enabled" mapstructure:"enabled"`
	Host          string `json:"host" mapstructure:"host"`
	Port          int    `json:"port" mapstructure:"port"`
	Mode          string `json:"mode" mapstructure:"mode"`
	EnableRecover bool   `json:"enable_recover" mapstructure:"enable_recover"`
	EnableLogger  bool   `json:"enable_logger" mapstructure:"enable_logger"`
}

// 配置常量。
const (
	GinEnabled    = "gin.enabled"
	ConditionTrue = "true"

	// 默认值
	DefaultGinHost          = "0.0.0.0"
	DefaultGinPort          = 8080
	DefaultGinMode          = "debug"
	DefaultGinEnableRecover = true
	DefaultGinEnableLogger  = true
)

// loadConfig 从 Environment 加载 Gin 配置。
func (c *GinAutoConfiguration) loadConfig(env *environment.Environment) (*GinConfig, error) {
	cfg := &GinConfig{
		Host:          DefaultGinHost,
		Port:          DefaultGinPort,
		Mode:          DefaultGinMode,
		EnableRecover: DefaultGinEnableRecover,
		EnableLogger:  DefaultGinEnableLogger,
	}

	if err := env.BindPrefix("gin", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Gin config: %w", err)
	}

	return cfg, nil
}
