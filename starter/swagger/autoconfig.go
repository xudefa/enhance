// Package swagger 提供 Swagger API 文档自动配置。
//
// Swagger 是最流行的 API 文档生成工具，支持 OpenAPI 规范。
// 本模块提供自动配置支持，无需手动初始化即可使用 Swagger UI。
//
// 功能特性：
//   - 自动配置 Swagger UI
//   - 支持 OpenAPI 规范
//   - 交互式 API 文档
//   - API 测试支持
//
// 配置示例：
//
//	{
//	  "swagger": {
//	    "enabled": true,
//	    "host": "localhost",
//	    "port": 8080,
//	    "url": "/swagger/*",
//	    "title": "My API"
//	  }
//	}
//
// 使用示例：
//
//	swagger := core.MustGetBean[*swagger.SwaggerAutoConfiguration](app.Container())
//	engine.GET("/swagger/*", gin.WrapH(swagger.GetHandler()))
package swagger

import (
	"fmt"
	"net/http"
	"reflect"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

// init 注册 Swagger 自动配置类。
// 当配置 swagger.enabled=true 时自动触发配置。
func init() {
	boot.RegisterAutoConfigWith(&SwaggerAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(SwaggerEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityWebLayer)),
	)
}

// SwaggerAutoConfiguration Swagger API 文档自动配置类。
// 负责从配置中心加载 Swagger 配置并初始化相关组件。
type SwaggerAutoConfiguration struct {
	logger log.Logger     // 日志记录器
	config *SwaggerConfig // Swagger 配置信息
}

// Configure 配置 Swagger。
// 从 ApplicationContext 中获取环境和日志器，加载配置并记录配置信息。
func (c *SwaggerAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	// 获取日志记录器，如果不存在则使用默认日志器
	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	// 加载配置
	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Swagger config: %w", err)
	}

	c.config = cfg

	// 注册 Swagger 配置到容器
	if err := ctx.Container().RegisterInstance(cfg, reflect.TypeFor[*SwaggerConfig]()); err != nil {
		return fmt.Errorf("failed to register SwaggerConfig: %w", err)
	}

	// 记录配置信息
	c.logger.Info(ctx.Context(), "Swagger API documentation configured",
		log.KeyValue{Key: "host", Value: cfg.Host},
		log.KeyValue{Key: "port", Value: cfg.Port},
		log.KeyValue{Key: "url", Value: cfg.URL},
		log.KeyValue{Key: "title", Value: cfg.Title},
	)

	return nil
}

// GetHandler 获取 Swagger HTTP Handler。
// 返回的 Handler 可以直接注册到 Web 框架的路由中。
//
// 使用示例：
//
//	handler := swagger.GetHandler()
//	http.Handle("/swagger/", handler)
func (c *SwaggerAutoConfiguration) GetHandler() http.Handler {
	return httpSwagger.WrapHandler
}

// SwaggerConfig Swagger API 文档配置。
// 包含 Swagger UI 的所有可配置参数。
type SwaggerConfig struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"` // 是否启用 Swagger
	Host    string `json:"host" mapstructure:"host"`       // 服务器主机地址
	Port    int    `json:"port" mapstructure:"port"`       // 服务器端口
	URL     string `json:"url" mapstructure:"url"`         // Swagger UI 访问路径
	Title   string `json:"title" mapstructure:"title"`     // API 文档标题
}

// 配置常量。
const (
	SwaggerEnabled      = "swagger.enabled" // 启用条件配置键
	DefaultSwaggerHost  = "localhost"       // 默认主机地址
	DefaultSwaggerPort  = 8080              // 默认端口
	DefaultSwaggerURL   = "/swagger/*"      // 默认访问路径
	DefaultSwaggerTitle = "Enhance API"     // 默认文档标题
	ConditionTrue       = "true"            // 条件真值
)

// loadConfig 从 Environment 加载 Swagger 配置。
// 使用默认值初始化配置，然后从配置中心绑定用户自定义值。
func (c *SwaggerAutoConfiguration) loadConfig(env *environment.Environment) (*SwaggerConfig, error) {
	cfg := &SwaggerConfig{
		Host:  DefaultSwaggerHost,
		Port:  DefaultSwaggerPort,
		URL:   DefaultSwaggerURL,
		Title: DefaultSwaggerTitle,
	}

	if err := env.BindPrefix("swagger", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Swagger config: %w", err)
	}

	return cfg, nil
}
