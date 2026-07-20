// Package mvc 提供 MVC 控制器支持。
package mvc

import (
	"time"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/web/core"
)

// Config Web 服务器配置。
type Config struct {
	Port    int
	Host    string
	Timeout time.Duration
	Logger  log.Logger
}

// DefaultConfig 返回默认 Web 配置。
func DefaultConfig() Config {
	return Config{
		Port:    8080,
		Host:    "0.0.0.0",
		Timeout: 30 * time.Second,
		Logger:  log.Build(), // 默认服务配置，默认日志记录器
	}
}

// WebConfig 配置别名(向后兼容)。
type WebConfig = Config

// DefaultWebConfig 返回默认 Web 配置(向后兼容)。
func DefaultWebConfig() WebConfig {
	return DefaultConfig()
}

// WithServer 设置 HTTP 服务器实现。
func (s *WebStarter) WithServer(server core.Server) *WebStarter {
	s.server = server
	return s
}

// WithRouter 设置路由器实现。
func (s *WebStarter) WithRouter(router core.Router) *WebStarter {
	s.router = router
	return s
}

// Use 添加中间件。
func (s *WebStarter) Use(middleware core.MiddlewareFunc) *WebStarter {
	s.middlewares = append(s.middlewares, middleware)
	return s
}

// GetRouter 获取路由器。
func (s *WebStarter) GetRouter() core.Router {
	return s.router
}
