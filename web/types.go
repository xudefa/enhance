// Package web 提供 Web 框架类型定义和辅助函数。
package web

import (
	"github.com/xudefa/enhance/web/engine"
)

// 引擎类型常量。
const (
	EngineStdLib = engine.StdLib
)

// 全局引擎注册表。
var GlobalEngineRegistry = engine.GlobalRegistry

// NewEngineRegistry 创建引擎注册表。
func NewEngineRegistry() *engine.Registry {
	return engine.NewRegistry()
}

// DefaultServerConfig 返回默认服务器配置。
func DefaultServerConfig() *engine.ServerConfig {
	return engine.DefaultServerConfig()
}

// WithHost 设置服务器监听地址。
func WithHost(host string) engine.ServerOption {
	return engine.WithHost(host)
}

// WithPort 设置服务器监听端口。
func WithPort(port int) engine.ServerOption {
	return engine.WithPort(port)
}

// WithReadTimeout 设置读取超时时间(秒)。
func WithReadTimeout(timeout int) engine.ServerOption {
	return engine.WithReadTimeout(timeout)
}

// WithWriteTimeout 设置写入超时时间(秒)。
func WithWriteTimeout(timeout int) engine.ServerOption {
	return engine.WithWriteTimeout(timeout)
}

// WithIdleTimeout 设置空闲超时时间(秒)。
func WithIdleTimeout(timeout int) engine.ServerOption {
	return engine.WithIdleTimeout(timeout)
}

// WithTLS 设置 TLS 证书和密钥文件。
func WithTLS(certFile, keyFile string) engine.ServerOption {
	return engine.WithTLS(certFile, keyFile)
}
