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
			condition.OnProperty(GinEnabled, ConditionTrue),
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

	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Gin config: %w", err)
	}

	c.config = cfg

	if cfg.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.Mode == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 尝试从容器获取已存在的 Engine 实例，如果不存在则创建默认的
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

	// 尝试从容器获取 Tracer 并注册 tracing 中间件
	if tracer, err := core.GetByName[*tracing.Tracer](container, ""); err == nil {
		c.tracer = tracer
		c.engine.Use(TracingMiddleware(tracer))
		c.logger.Info(ctx.Context(), "Gin tracing middleware enabled")
	}

	c.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: c.engine,
	}

	// 检查 Engine 是否已注册（由外部传入）
	engineAlreadyRegistered := false
	if _, err := core.GetByName[*gin.Engine](container, ""); err == nil {
		engineAlreadyRegistered = true
	}

	if err := container.RegisterInstance(c.config, reflect.TypeFor[*GinConfig]()); err != nil {
		return fmt.Errorf("failed to register Gin Config: %w", err)
	}

	// 如果 Engine 已存在，跳过注册
	if !engineAlreadyRegistered {
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

// Start 启动 Gin Web 服务器。
func (c *GinAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.server == nil {
		return fmt.Errorf("Gin HTTP Server not initialized")
	}
	c.logger.Info(ctx.Context(), "Gin Web server starting",
		log.KeyValue{Key: "addr", Value: c.server.Addr},
	)

	// 在后台启动服务器，避免阻塞
	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.logger.Error(c.ctx, "Gin Web server error",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
		}
	}()

	return nil
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
	return condition.OnProperty(GinEnabled, ConditionTrue)
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
