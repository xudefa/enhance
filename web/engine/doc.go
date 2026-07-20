// Package engine 提供网络引擎工厂和注册表。
//
// 该模块管理不同 HTTP 框架的引擎实现，支持运行时切换框架。
// 通过 Factory 接口创建特定框架的组件实例。
//
// # 核心接口
//
//   - Factory: 引擎工厂接口，用于创建路由器和服务器
//   - Registry: 引擎注册表，管理所有已注册的引擎工厂
//   - ServerOption: 服务器配置选项函数类型
//
// # 支持的引擎
//
//   - StdLib: 标准库 net/http
//
// # 使用方式
//
// 注册引擎：
//
//	engine.GlobalRegistry.Register(myFactory)
//
// 创建路由器：
//
//	router, err := engine.GlobalRegistry.CreateRouter()
//
// 创建服务器：
//
//	server, err := engine.GlobalRegistry.CreateServer(
//	    engine.WithHost("0.0.0.0"),
//	    engine.WithPort(8080),
//	)
package engine

import (
	"github.com/xudefa/enhance/web/core"
)

// ==================== 引擎类型 ====================

// Type 网络引擎类型标识。
type Type string

const (
	// StdLib 标准库 net/http。
	StdLib Type = "stdlib"
)

// ==================== 核心接口 ====================

// Factory 网络引擎工厂接口。
//
// 实现此接口可以创建特定网络框架的组件实例。
type Factory interface {
	// Type 返回引擎类型标识。
	Type() Type

	// CreateRouter 创建路由器实例。
	CreateRouter() (core.Router, error)

	// CreateServer 创建服务器实例。
	CreateServer(opts ...ServerOption) (core.Server, error)
}

// ==================== 配置选项 ====================

// ServerOption 服务器配置选项函数类型。
type ServerOption func(*ServerConfig)

// ServerConfig 服务器配置。
type ServerConfig struct {
	// Host 监听地址。
	Host string
	// Port 监听端口。
	Port int
	// ReadTimeout 读取超时时间(秒)。
	ReadTimeout int
	// WriteTimeout 写入超时时间(秒)。
	WriteTimeout int
	// IdleTimeout 空闲超时时间(秒)。
	IdleTimeout int
	// TLSCertFile TLS 证书文件路径。
	TLSCertFile string
	// TLSKeyFile TLS 密钥文件路径。
	TLSKeyFile string
}

// ==================== 注册表 ====================
