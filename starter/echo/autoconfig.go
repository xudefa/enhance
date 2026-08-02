// Package echo 提供 Echo Web 框架自动配置。
package echo

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/xudefa/enhance/actuator"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/tracing"
)

var echoAutoConfig = &EchoAutoConfiguration{}

func init() {
	boot.RegisterAutoConfigWith(echoAutoConfig,
		boot.WithConditions(
			condition.OnProperty(EchoEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityWebLayer)),
	)
	// 注册为 Starter，使其 Start/Stop 生命周期方法被自动调用
	boot.RegisterStarter(echoAutoConfig)
}

// EchoAutoConfiguration Echo Web 框架自动配置类。
type EchoAutoConfiguration struct {
	logger     log.Logger
	server     *echo.Echo
	config     *EchoConfig
	tracer     *tracing.Tracer
	configured bool // 标记是否已配置，防止重复配置
}

// Configure 配置 Echo Web 服务器。
func (c *EchoAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	// 防止重复配置
	if c.configured {
		return nil
	}
	container := ctx.Container()
	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Echo config: %w", err)
	}

	c.config = cfg

	// 尝试从容器获取已存在的 Echo 实例，如果不存在则创建默认的
	if server, err := core.GetByName[*echo.Echo](container, ""); err == nil {
		c.server = server
		c.logger.Info(ctx.Context(), "using existing Echo Server instance from container")
	} else {
		c.server = echo.New()
		c.server.HideBanner = cfg.HideBanner
		c.server.HidePort = cfg.HidePort

		// 配置中间件
		if cfg.EnableRecover {
			c.server.Use(middleware.Recover())
		}
		if cfg.EnableLogger {
			c.server.Use(middleware.Logger())
		}
		if cfg.EnableCORS {
			c.server.Use(middleware.CORS())
		}
	}

	// 尝试从容器获取 Tracer 并注册 tracing 中间件
	if tracer, err := core.GetByName[*tracing.Tracer](container, ""); err == nil {
		c.tracer = tracer
		c.server.Use(TracingMiddleware(tracer))
		c.logger.Info(ctx.Context(), "Echo tracing middleware enabled")
	}

	// 检查 Server 是否已注册（由外部传入）
	serverAlreadyRegistered := false
	if _, err := core.GetByName[*echo.Echo](container, ""); err == nil {
		serverAlreadyRegistered = true
	}

	if err := container.RegisterInstance(c.config, reflect.TypeFor[*EchoConfig]()); err != nil {
		return fmt.Errorf("failed to register Echo Config: %w", err)
	}

	// 如果 Server 已存在，跳过注册
	if !serverAlreadyRegistered {
		if err := container.RegisterInstance(c.server, reflect.TypeFor[*echo.Echo]()); err != nil {
			return fmt.Errorf("failed to register Echo Server: %w", err)
		}
	}

	// 注册 HttpEndpointRegistry,允许 Actuator 等模块自动挂载端点到 Echo
	endpointRegistry := NewEchoEndpointRegistry(c.server)
	if err := container.RegisterInstance(endpointRegistry, reflect.TypeFor[actuator.HttpEndpointRegistry]()); err != nil {
		c.logger.Warn(ctx.Context(), "failed to register HttpEndpointRegistry, Actuator endpoints will not be mounted automatically",
			log.KeyValue{Key: "error", Value: err.Error()},
		)
	}

	c.logger.Info(ctx.Context(), "Echo Web server configured",
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "host", Value: cfg.Host},
	)

	c.configured = true
	return nil
}

// Start 启动 Echo Web 服务器。
func (c *EchoAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.server == nil || c.config == nil {
		return fmt.Errorf("Echo Web Server not initialized")
	}
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	c.logger.Info(ctx.Context(), "Echo Web server starting",
		log.KeyValue{Key: "addr", Value: addr},
	)

	// 在后台启动服务器，避免阻塞
	go func() {
		if err := c.server.Start(addr); err != nil && err != http.ErrServerClosed {
			c.logger.Error(ctx.Context(), "Echo Web server error",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
		}
	}()

	return nil
}

// Stop 停止 Echo Web 服务器。
func (c *EchoAutoConfiguration) Stop(ctx boot.ApplicationContext) error {
	if c.server == nil {
		return nil
	}
	shutdownCtx, cancel := context.WithTimeout(ctx.Context(), 30*time.Second)
	defer cancel()
	return c.server.Shutdown(shutdownCtx)
}

// Name 返回启动器名称。
func (c *EchoAutoConfiguration) Name() string {
	return "EchoStarter"
}

// Dependencies 返回依赖的其他启动器名称。
func (c *EchoAutoConfiguration) Dependencies() []string {
	return nil
}

// GetCondition 返回启动器条件。
func (c *EchoAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnProperty(EchoEnabled, ConditionTrue)
}

// GetServer 从容器中获取 Echo 服务器实例。
func GetServer(container core.Container) (*echo.Echo, error) {
	return core.GetByName[*echo.Echo](container, "")
}

// EchoConfig Echo Web 服务器配置。
type EchoConfig struct {
	Enabled       bool   `json:"enabled" mapstructure:"enabled"`
	Host          string `json:"host" mapstructure:"host"`
	Port          int    `json:"port" mapstructure:"port"`
	HideBanner    bool   `json:"hide_banner" mapstructure:"hide_banner"`
	HidePort      bool   `json:"hide_port" mapstructure:"hide_port"`
	EnableRecover bool   `json:"enable_recover" mapstructure:"enable_recover"`
	EnableLogger  bool   `json:"enable_logger" mapstructure:"enable_logger"`
	EnableCORS    bool   `json:"enable_cors" mapstructure:"enable_cors"`
}

// 配置常量。
const (
	EchoEnabled   = "echo.enabled"
	ConditionTrue = "true"
)

// loadConfig 从 Environment 加载 Echo 配置。
func (c *EchoAutoConfiguration) loadConfig(env *environment.Environment) (*EchoConfig, error) {
	cfg := &EchoConfig{}

	if err := env.BindPrefix("echo", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Echo config: %w", err)
	}

	return cfg, nil
}
