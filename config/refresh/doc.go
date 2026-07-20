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

import "time"

// ConfigChangeEvent 配置变更事件。
type ConfigChangeEvent struct {
	EventType string            // 事件类型："modify"、"delete"、"create"
	Keys      []string          // 变更的配置键列表
	OldValues map[string]any    // 变更前的值
	NewValues map[string]any    // 变更后的值
	Source    string            // 配置源类型（如 "nacos"、"etcd"）
	timestamp time.Time         // 事件发生时间
	Metadata  map[string]string // 额外元数据
}

// Type 返回事件类型标识
func (e *ConfigChangeEvent) Type() string {
	return "ConfigChange"
}

// Timestamp 返回事件发生时间
func (e *ConfigChangeEvent) Timestamp() time.Time {
	return e.timestamp
}

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

// RefreshableBean 可刷新 Bean 接口。
//
// 实现此接口的 Bean 会在配置变更时收到 OnConfigChange 回调。
type RefreshableBean interface {
	OnConfigChange(event ConfigChangeEvent) error
}
