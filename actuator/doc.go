// Package actuator 提供应用监控端点 Starter，用于 enhance 框架。
//
// 该模块提供多种运维监控端点，包括健康检查、指标收集、环境信息等。
// 支持多种 HTTP 框架集成，如标准库 http、Gin、Hertz 等。
//
// # 架构设计
//
//   - Actuator: 运维端点管理器
//   - health.Aggregator: 健康检查聚合器
//   - metrics.MeterRegistry: 指标注册表
//   - Sanitizer: 敏感信息检测器
//
// # 支持的端点
//
//   - /health: 健康检查
//   - /info: 应用信息
//   - /metrics: 指标暴露
//   - /env: 环境配置查看
//   - /beans: Bean 列表
//   - /admin: 管理端点
//
// # 使用方式
//
// 在 main.go 中引入：
//
//	import _ "github.com/xudefa/enhance/actuator"
//
// # 配置属性
//
//   - actuator.enabled: 是否启用监控端点（默认 true）
//   - actuator.path: 端点路径前缀（默认 /actuator）
//   - actuator.health.enabled: 是否启用健康检查（默认 true）
//   - actuator.metrics.enabled: 是否启用指标收集（默认 true）
//
// # 配置示例
//
// 环境变量：
//
//	export ACTUATOR_ENABLED=true
//	export ACTUATOR_PATH=/actuator
//
// 配置文件（application.json）：
//
//	{
//	  "actuator": {
//	    "enabled": true,
//	    "path": "/actuator",
//	    "health": {
//	      "enabled": true
//	    },
//	    "metrics": {
//	      "enabled": true
//	    }
//	  }
//	}
package actuator

import (
	"net/http"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
)

// AppContext 应用上下文接口。
type AppContext interface {
	// Container 返回 IoC 容器实例。
	Container() core.Container
	// Environment 返回环境配置实例。
	Environment() *environment.Environment
}

// RouteRegistrar 路由注册器接口。
type RouteRegistrar interface {
	// Handle 注册 HTTP 路由处理函数。
	Handle(pattern string, handler http.Handler)
}

// SanitizeStrategy 敏感信息检测策略接口。
type SanitizeStrategy interface {
	IsSensitive(key string, value any) bool
}

// ==================== 配置键常量 ====================

const (
	// 应用配置
	AppName    = "app.name"
	AppVersion = "app.version"

	// Actuator 配置
	ActuatorEnabled = "actuator.enabled"
)

// ==================== 默认值常量 ====================

const (
	// 应用默认值
	DefaultAppName    = "enhance-app"
	DefaultAppVersion = "1.0.0"

	// 条件值常量
	ConditionTrue = "true"
)
