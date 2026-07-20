// Package echo 提供 Echo Web 框架自动配置。
package echo

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

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
		return fmt.Errorf("加载 Echo 配置失败: %w", err)
	}

	c.config = cfg

	// 尝试从容器获取已存在的 Echo 实例，如果不存在则创建默认的
	if server, err := core.GetByName[*echo.Echo](container, ""); err == nil {
		c.server = server
		c.logger.Info(context.Background(), "使用容器中已存在的 Echo Server 实例")
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
		c.logger.Info(context.Background(), "Echo Tracing 中间件已启用")
	}

	// 检查 Server 是否已注册（由外部传入）
	serverAlreadyRegistered := false
	if _, err := core.GetByName[*echo.Echo](container, ""); err == nil {
		serverAlreadyRegistered = true
	}

	if err := container.RegisterInstance(c.config, reflect.TypeFor[*EchoConfig]()); err != nil {
		return fmt.Errorf("注册 Echo Config 失败: %w", err)
	}

	// 如果 Server 已存在，跳过注册
	if !serverAlreadyRegistered {
		if err := container.RegisterInstance(c.server, reflect.TypeFor[*echo.Echo]()); err != nil {
			return fmt.Errorf("注册 Echo Server 失败: %w", err)
		}
	}

	// 注册 HttpEndpointRegistry,允许 Actuator 等模块自动挂载端点到 Echo
	endpointRegistry := NewEchoEndpointRegistry(c.server)
	if err := container.RegisterInstance(endpointRegistry, reflect.TypeFor[actuator.HttpEndpointRegistry]()); err != nil {
		c.logger.Warn(context.Background(), "注册 HttpEndpointRegistry 失败,Actuator 端点将无法自动挂载",
			log.KeyValue{Key: "error", Value: err.Error()},
		)
	}

	c.logger.Info(context.Background(), "Echo Web 服务器已配置",
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "host", Value: cfg.Host},
	)

	c.configured = true
	return nil
}

// Start 启动 Echo Web 服务器。
func (c *EchoAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.server == nil || c.config == nil {
		return fmt.Errorf("Echo Web Server 未初始化")
	}
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	c.logger.Info(context.Background(), "Echo Web 服务器启动中",
		log.KeyValue{Key: "addr", Value: addr},
	)

	// 在后台启动服务器，避免阻塞
	go func() {
		if err := c.server.Start(addr); err != nil && err != http.ErrServerClosed {
			c.logger.Error(context.Background(), "Echo Web 服务器错误",
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
	return c.server.Shutdown(context.Background())
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
	Enabled       bool   `json:"enabled" value:"${echo.enabled:false}"`
	Host          string `json:"host" value:"${echo.host:0.0.0.0}"`
	Port          int    `json:"port" value:"${echo.port:8080}"`
	HideBanner    bool   `json:"hide_banner" value:"${echo.hide_banner:false}"`
	HidePort      bool   `json:"hide_port" value:"${echo.hide_port:false}"`
	EnableRecover bool   `json:"enable_recover" value:"${echo.enable_recover:true}"`
	EnableLogger  bool   `json:"enable_logger" value:"${echo.enable_logger:true}"`
	EnableCORS    bool   `json:"enable_cors" value:"${echo.enable_cors:false}"`
}

// 配置常量。
const (
	EchoEnabled   = "echo.enabled"
	ConditionTrue = "true"
)

// loadConfig 从 Environment 加载 Echo 配置。
func (c *EchoAutoConfiguration) loadConfig(env *environment.Environment) (*EchoConfig, error) {
	cfg := &EchoConfig{}

	if err := env.BindProperties(cfg); err != nil {
		return nil, fmt.Errorf("绑定 Echo 配置失败: %w", err)
	}

	return cfg, nil
}
