// Package fiber 提供 Fiber Web 框架自动配置。
package fiber

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/xudefa/enhance/actuator"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/tracing"
)

var fiberAutoConfig = &FiberAutoConfiguration{}

func init() {
	boot.RegisterAutoConfigWith(fiberAutoConfig,
		boot.WithConditions(
			condition.OnProperty(FiberEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityWebLayer)),
	)
	// 注册为 Starter，使其 Start/Stop 生命周期方法被自动调用
	boot.RegisterStarter(fiberAutoConfig)
}

// FiberAutoConfiguration Fiber Web 框架自动配置类。
type FiberAutoConfiguration struct {
	logger     log.Logger
	app        *fiber.App
	config     *FiberConfig
	tracer     *tracing.Tracer
	configured bool // 标记是否已配置，防止重复配置
}

// Configure 配置 Fiber Web 服务器。
func (c *FiberAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
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
		return fmt.Errorf("加载 Fiber 配置失败: %w", err)
	}

	c.config = cfg

	// 尝试从容器获取已存在的 Fiber App 实例，如果不存在则创建默认的
	if app, err := core.GetByName[*fiber.App](container, ""); err == nil {
		c.app = app
		c.logger.Info(context.Background(), "使用容器中已存在的 Fiber App 实例")
	} else {
		c.app = fiber.New(fiber.Config{
			Prefork:       cfg.Prefork,
			BodyLimit:     cfg.BodyLimit,
			Concurrency:   cfg.Concurrency,
			ReadTimeout:   time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:  time.Duration(cfg.WriteTimeout) * time.Second,
			IdleTimeout:   time.Duration(cfg.IdleTimeout) * time.Second,
			ServerHeader:  cfg.ServerHeader,
			AppName:       cfg.AppName,
			CaseSensitive: cfg.CaseSensitive,
			StrictRouting: cfg.StrictRouting,
		})
	}

	// 尝试从容器获取 Tracer 并注册 tracing 中间件
	if tracer, err := core.GetByName[*tracing.Tracer](container, ""); err == nil {
		c.tracer = tracer
		c.app.Use(TracingMiddleware(tracer))
		c.logger.Info(context.Background(), "Fiber Tracing 中间件已启用")
	}

	// 检查 App 是否已注册（由外部传入）
	appAlreadyRegistered := false
	if _, err := core.GetByName[*fiber.App](container, ""); err == nil {
		appAlreadyRegistered = true
	}

	if err := container.RegisterInstance(c.config, reflect.TypeFor[*FiberConfig]()); err != nil {
		return fmt.Errorf("注册 Fiber Config 失败: %w", err)
	}

	// 如果 App 已存在，跳过注册
	if !appAlreadyRegistered {
		if err := container.RegisterInstance(c.app, reflect.TypeFor[*fiber.App]()); err != nil {
			return fmt.Errorf("注册 Fiber App 失败: %w", err)
		}
	}

	// 注册 HttpEndpointRegistry,允许 Actuator 等模块自动挂载端点到 Fiber
	endpointRegistry := NewFiberEndpointRegistry(c.app)
	if err := container.RegisterInstance(endpointRegistry, reflect.TypeFor[actuator.HttpEndpointRegistry]()); err != nil {
		c.logger.Warn(context.Background(), "注册 HttpEndpointRegistry 失败,Actuator 端点将无法自动挂载",
			log.KeyValue{Key: "error", Value: err.Error()},
		)
	}

	c.logger.Info(context.Background(), "Fiber Web 服务器已配置",
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "host", Value: cfg.Host},
	)

	c.configured = true
	return nil
}

// Start 启动 Fiber Web 服务器。
func (c *FiberAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.app == nil || c.config == nil {
		return fmt.Errorf("Fiber Web Server 未初始化")
	}
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	c.logger.Info(context.Background(), "Fiber Web 服务器启动中",
		log.KeyValue{Key: "addr", Value: addr},
	)

	// 在后台启动服务器，避免阻塞
	go func() {
		if err := c.app.Listen(addr); err != nil {
			c.logger.Error(context.Background(), "Fiber Web 服务器错误",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
		}
	}()

	return nil
}

// Stop 停止 Fiber Web 服务器。
func (c *FiberAutoConfiguration) Stop(ctx boot.ApplicationContext) error {
	if c.app == nil {
		return nil
	}
	return c.app.Shutdown()
}

// Name 返回启动器名称。
func (c *FiberAutoConfiguration) Name() string {
	return "FiberStarter"
}

// Dependencies 返回依赖的其他启动器名称。
func (c *FiberAutoConfiguration) Dependencies() []string {
	return nil
}

// GetCondition 返回启动器条件。
func (c *FiberAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnProperty(FiberEnabled, ConditionTrue)
}

// GetApp 从容器中获取 Fiber App 实例。
func GetApp(container core.Container) (*fiber.App, error) {
	return core.GetByName[*fiber.App](container, "")
}

// FiberConfig Fiber Web 服务器配置。
type FiberConfig struct {
	Enabled       bool   `json:"enabled" value:"${fiber.enabled:false}"`
	Host          string `json:"host" value:"${fiber.host:0.0.0.0}"`
	Port          int    `json:"port" value:"${fiber.port:3000}"`
	Prefork       bool   `json:"prefork" value:"${fiber.prefork:false}"`
	BodyLimit     int    `json:"body-limit" value:"${fiber.body-limit:4194304}"`
	Concurrency   int    `json:"concurrency" value:"${fiber.concurrency:262144}"`
	ReadTimeout   int    `json:"read-timeout" value:"${fiber.read-timeout:0}"`
	WriteTimeout  int    `json:"write-timeout" value:"${fiber.write-timeout:0}"`
	IdleTimeout   int    `json:"idle-timeout" value:"${fiber.idle-timeout:0}"`
	ServerHeader  string `json:"server-header" value:"${fiber.server-header:}"`
	AppName       string `json:"app-name" value:"${fiber.app-name:}"`
	CaseSensitive bool   `json:"case-sensitive" value:"${fiber.case-sensitive:false}"`
	StrictRouting bool   `json:"strict-routing" value:"${fiber.strict-routing:false}"`
}

// 配置常量。
const (
	FiberEnabled  = "fiber.enabled"
	ConditionTrue = "true"
)

// loadConfig 从 Environment 加载 Fiber 配置。
func (c *FiberAutoConfiguration) loadConfig(env *environment.Environment) (*FiberConfig, error) {
	cfg := &FiberConfig{}

	if err := env.BindProperties(cfg); err != nil {
		return nil, fmt.Errorf("绑定 Fiber 配置失败: %w", err)
	}

	return cfg, nil
}
