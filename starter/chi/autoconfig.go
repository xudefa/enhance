// Package chi 提供 Chi HTTP 路由器自动配置。
package chi

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/xudefa/enhance/actuator"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/tracing"
)

var chiAutoConfig = &ChiAutoConfiguration{}

func init() {
	boot.RegisterAutoConfigWith(chiAutoConfig,
		boot.WithConditions(
			condition.OnProperty(ChiEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityWebLayer)),
	)
	// 注册为 Starter，使其 Start/Stop 生命周期方法被自动调用
	boot.RegisterStarter(chiAutoConfig)
}

// ChiAutoConfiguration Chi HTTP 路由器自动配置类。
type ChiAutoConfiguration struct {
	logger     log.Logger
	router     *chi.Mux
	server     *http.Server
	config     *ChiConfig
	tracer     *tracing.Tracer
	configured bool // 标记是否已配置，防止重复配置
}

// Configure 配置 Chi HTTP 路由器。
func (c *ChiAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
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
		return fmt.Errorf("加载 Chi 配置失败: %w", err)
	}

	c.config = cfg

	// 尝试从容器获取已存在的 Router 实例，如果不存在则创建默认的
	if router, err := core.GetByName[*chi.Mux](container, ""); err == nil {
		c.router = router
		c.logger.Info(context.Background(), "使用容器中已存在的 Chi Router 实例")
	} else {
		c.router = chi.NewRouter()
		if cfg.EnableRecover {
			c.router.Use(middleware.Recoverer)
		}
		if cfg.EnableLogger {
			c.router.Use(middleware.Logger)
		}
		if cfg.EnableRequestID {
			c.router.Use(middleware.RequestID)
		}
		if cfg.EnableRealIP {
			c.router.Use(middleware.RealIP)
		}
	}

	// 尝试从容器获取 Tracer 并注册 tracing 中间件
	if tracer, err := core.GetByName[*tracing.Tracer](container, ""); err == nil {
		c.tracer = tracer
		c.router.Use(TracingMiddleware(tracer))
		c.logger.Info(context.Background(), "Chi Tracing 中间件已启用")
	}

	c.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: c.router,
	}

	// 检查 Router 是否已注册（由外部传入）
	routerAlreadyRegistered := false
	if _, err := core.GetByName[*chi.Mux](container, ""); err == nil {
		routerAlreadyRegistered = true
	}

	if err := container.RegisterInstance(c.config, reflect.TypeFor[*ChiConfig]()); err != nil {
		return fmt.Errorf("注册 Chi Config 失败: %w", err)
	}

	// 如果 Router 已存在，跳过注册
	if !routerAlreadyRegistered {
		if err := container.RegisterInstance(c.router, reflect.TypeFor[*chi.Mux]()); err != nil {
			return fmt.Errorf("注册 Chi Router 失败: %w", err)
		}
	}

	if err := container.RegisterInstance(c.server, reflect.TypeFor[*http.Server]()); err != nil {
		return fmt.Errorf("注册 HTTP Server 失败: %w", err)
	}

	// 注册 HttpEndpointRegistry,允许 Actuator 等模块自动挂载端点到 Chi
	endpointRegistry := NewChiEndpointRegistry(c.router)
	if err := container.RegisterInstance(endpointRegistry, reflect.TypeFor[actuator.HttpEndpointRegistry]()); err != nil {
		c.logger.Warn(context.Background(), "注册 HttpEndpointRegistry 失败,Actuator 端点将无法自动挂载",
			log.KeyValue{Key: "error", Value: err.Error()},
		)
	}

	c.logger.Info(context.Background(), "Chi HTTP 路由器已配置",
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "host", Value: cfg.Host},
	)

	c.configured = true
	return nil
}

// Start 启动 HTTP 服务器。
func (c *ChiAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.server == nil {
		return fmt.Errorf("Chi HTTP Server 未初始化")
	}
	c.logger.Info(context.Background(), "Chi HTTP 服务器启动中",
		log.KeyValue{Key: "addr", Value: c.server.Addr},
	)

	// 在后台启动服务器，避免阻塞
	go func() {
		if err := c.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			c.logger.Error(context.Background(), "Chi HTTP 服务器错误",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
		}
	}()

	return nil
}

// Stop 停止 HTTP 服务器。
func (c *ChiAutoConfiguration) Stop(ctx boot.ApplicationContext) error {
	if c.server == nil {
		return nil
	}
	return c.server.Shutdown(context.Background())
}

// Name 返回启动器名称。
func (c *ChiAutoConfiguration) Name() string {
	return "ChiStarter"
}

// Dependencies 返回依赖的其他启动器名称。
func (c *ChiAutoConfiguration) Dependencies() []string {
	return nil
}

// GetCondition 返回启动器条件。
func (c *ChiAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnProperty(ChiEnabled, ConditionTrue)
}

// GetRouter 从容器中获取 Chi 路由器实例。
func GetRouter(container core.Container) (*chi.Mux, error) {
	return core.GetByName[*chi.Mux](container, "")
}

// GetServer 从容器中获取 HTTP 服务器实例。
func GetServer(container core.Container) (*http.Server, error) {
	return core.GetByName[*http.Server](container, "")
}

// ChiConfig Chi HTTP 路由器配置。
type ChiConfig struct {
	Enabled         bool   `json:"enabled" value:"${chi.enabled:false}"`
	Host            string `json:"host" value:"${chi.host:0.0.0.0}"`
	Port            int    `json:"port" value:"${chi.port:8080}"`
	EnableRecover   bool   `json:"enable_recover" value:"${chi.enable_recover:true}"`
	EnableLogger    bool   `json:"enable_logger" value:"${chi.enable_logger:true}"`
	EnableRequestID bool   `json:"enable_request_id" value:"${chi.enable_request_id:true}"`
	EnableRealIP    bool   `json:"enable_real_ip" value:"${chi.enable_real_ip:false}"`
}

// 配置常量。
const (
	ChiEnabled    = "chi.enabled"
	ConditionTrue = "true"
)

// loadConfig 从 Environment 加载 Chi 配置。
func (c *ChiAutoConfiguration) loadConfig(env *environment.Environment) (*ChiConfig, error) {
	cfg := &ChiConfig{}

	if err := env.BindProperties(cfg); err != nil {
		return nil, fmt.Errorf("绑定 Chi 配置失败: %w", err)
	}

	return cfg, nil
}
