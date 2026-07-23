// Package refresh 提供配置刷新功能，用于 enhance 框架。
//
// 该模块支持运行时配置热更新、RefreshScope 作用域管理、配置变更事件通知等动态配置功能。
// 参考 Spring Cloud RefreshScope 的设计理念。
//
// # 架构设计
//
//   - RefreshScope: 刷新作用域接口，管理可刷新的 Bean
//   - RefreshEvent: 刷新事件，表示配置变更
//   - RefreshEndpoint: 刷新端点，提供 HTTP 刷新接口
//   - ConfigWatcher: 配置监听器，监控配置变更
//
// # 核心功能
//
//   - 配置热更新: 支持运行时动态更新配置
//   - RefreshScope: 管理可刷新的 Bean 作用域
//   - 事件通知: 配置变更时发布事件通知
//   - HTTP 端点: 提供 /actuator/refresh 端点触发刷新
//   - 配置监听: 支持监听配置中心变更
//
// # 使用方式
//
// 创建刷新作用域：
//
//	scope := refresh.NewRefreshScope()
//
// 注册可刷新的 Bean：
//
//	scope.Register("myBean", func() any {
//	    return NewMyService(config)
//	})
//
// 触发刷新：
//
//	err := scope.Refresh()
//
// # 刷新端点
//
// 通过 HTTP 端点触发配置刷新：
//
//	POST /actuator/refresh
//
// # 配置变更事件
//
// 配置刷新时会发布事件：
//
//   - EnvironmentChangeEvent: 环境配置变更事件
//   - RefreshScopeEvent: 刷新作用域事件
//   - RefreshEvent: 通用刷新事件
//
// # 使用场景
//
//   - 配置中心集成: 监听配置中心变更并自动刷新
//   - 动态路由: 运行时更新路由配置
//   - 功能开关: 动态启用/禁用功能
package refresh

import (
	"time"

	"github.com/xudefa/enhance/config/environment"
)

// BeanRefreshedEvent Bean 刷新完成事件。
type BeanRefreshedEvent struct {
	BeanID      string    // Bean 标识
	OldVersion  int64     // 刷新前版本号
	NewVersion  int64     // 刷新后版本号
	RefreshTime time.Time // 刷新完成时间
	Success     bool      // 是否刷新成功
	Error       error     // 刷新失败的错误信息
}

// RefreshFailedEvent 刷新失败事件。
type RefreshFailedEvent struct {
	BeanID     string    // 失败的 Bean 标识
	ConfigKeys []string  // 触发刷新的配置键
	Error      error     // 失败原因
	Timestamp  time.Time // 失败时间
}

// ConfigChangeEvent 配置变更事件，别名来自 config/environment 包，保持向后兼容。
type ConfigChangeEvent = environment.ConfigChangeEvent

// RefreshableBean 可刷新 Bean 接口。
//
// 实现此接口的 Bean 会在配置变更时收到 OnConfigChange 回调。
type RefreshableBean interface {
	OnConfigChange(event ConfigChangeEvent) error
}

// RefreshManager 配置刷新管理器接口。
//
// 负责配置的动态刷新，依赖 config/environment 包。
// 依赖方向：config/refresh → config/environment（正确方向）。
type RefreshManager interface {
	// Start 启动刷新管理器。
	Start() error

	// Stop 停止刷新管理器。
	Stop() error

	// Refresh 刷新配置。
	Refresh() error

	// AddRefreshListener 添加刷新监听器。
	AddRefreshListener(listener RefreshListener)
}

// RefreshListener 刷新监听器接口。
type RefreshListener interface {
	// OnRefresh 刷新时回调。
	OnRefresh(event RefreshEvent) error
}

// RefreshEvent 刷新事件。
type RefreshEvent struct {
	// Environment 关联的环境配置。
	Environment *environment.Environment

	// ChangedKeys 变更的配置键列表。
	ChangedKeys []string

	// Timestamp 事件发生时间戳（毫秒）。
	Timestamp int64
}
