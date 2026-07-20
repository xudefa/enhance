// Package gin 提供 Gin Web 框架自动配置。
package gin

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

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
	configured bool // 标记是否已配置，防止重复配置
}

// Configure 配置 Gin Web 服务器。
func (c *GinAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
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
		return fmt.Errorf("加载 Gin 配置失败: %w", err)
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
		c.logger.Info(context.Background(), "使用容器中已存在的 Gin Engine 实例")
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
		c.logger.Info(context.Background(), "Gin Tracing 中间件已启用")
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
		return fmt.Errorf("注册 Gin Config 失败: %w", err)
	}

	// 如果 Engine 已存在，跳过注册
	if !engineAlreadyRegistered {
		if err := container.RegisterInstance(c.engine, reflect.TypeFor[*gin.Engine]()); err != nil {
			return fmt.Errorf("注册 Gin Engine 失败: %w", err)
		}
	}

	if err := container.RegisterInstance(c.server, reflect.TypeFor[*http.Server]()); err != nil {
		return fmt.Errorf("注册 HTTP Server 失败: %w", err)
	}

	// 注册 HttpEndpointRegistry,允许 Actuator 等模块自动挂载端点到 Gin
	endpointRegistry := NewGinEndpointRegistry(c.engine)
	if err := container.RegisterInstance(endpointRegistry, reflect.TypeFor[actuator.HttpEndpointRegistry]()); err != nil {
		c.logger.Warn(context.Background(), "注册 HttpEndpointRegistry 失败,Actuator 端点将无法自动挂载",
			log.KeyValue{Key: "error", Value: err.Error()},
		)
	}

	c.logger.Info(context.Background(), "Gin Web 服务器已配置",
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "host", Value: cfg.Host},
		log.KeyValue{Key: "mode", Value: cfg.Mode},
	)

	c.configured = true
	return nil
}

// Start 启动 Gin Web 服务器。
func (c *GinAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.server == nil {
		return fmt.Errorf("Gin HTTP Server 未初始化")
	}
	c.logger.Info(context.Background(), "Gin Web 服务器启动中",
		log.KeyValue{Key: "addr", Value: c.server.Addr},
	)

	// 在后台启动服务器，避免阻塞
	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.logger.Error(context.Background(), "Gin Web 服务器错误",
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
	return c.server.Shutdown(context.Background())
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
	Enabled       bool   `json:"enabled" value:"${gin.enabled:false}"`
	Host          string `json:"host" value:"${gin.host:0.0.0.0}"`
	Port          int    `json:"port" value:"${gin.port:8080}"`
	Mode          string `json:"mode" value:"${gin.mode:debug}"`
	EnableRecover bool   `json:"enable_recover" value:"${gin.enable_recover:true}"`
	EnableLogger  bool   `json:"enable_logger" value:"${gin.enable_logger:true}"`
}

// 配置常量。
const (
	GinEnabled    = "gin.enabled"
	ConditionTrue = "true"
)

// loadConfig 从 Environment 加载 Gin 配置。
func (c *GinAutoConfiguration) loadConfig(env *environment.Environment) (*GinConfig, error) {
	cfg := &GinConfig{}

	if err := env.BindProperties(cfg); err != nil {
		return nil, fmt.Errorf("绑定 Gin 配置失败: %w", err)
	}

	return cfg, nil
}
