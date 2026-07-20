// Package resilience 提供弹性容错功能，用于 enhance 框架。
//
// 该模块提供熔断器、负载均衡、限流等弹性容错机制，提升系统的稳定性和可用性。
// 参考 Resilience4j 的设计理念。
//
// # 架构设计
//
//   - Breaker: 熔断器接口，防止级联故障
//   - State: 熔断器状态枚举
//   - BreakerOption: 熔断器配置选项函数
//   - Selector: 负载均衡选择器接口
//   - Registry: 服务注册中心接口
//   - InstanceInfo: 服务实例信息
//   - HealthStatus: 健康状态枚举
//   - ServiceInstance: 服务实例结构
//   - Balancer: 负载均衡接口
//
// # 核心功能
//
//   - 熔断器: 支持 OPEN、HALF_OPEN、CLOSED 三种状态
//   - 负载均衡: 支持轮询、随机、最少连接等策略
//   - 限流: 支持自适应限流，动态调整速率
//   - 健康检查: 支持主动和被动健康检查
//   - 服务注册: 支持服务注册和发现
//
// # 使用方式
//
// 使用熔断器：
//
//	cb := resilience.NewBreaker(
//	    resilience.WithErrorThreshold(0.5),
//	    resilience.WithWaitDuration(30 * time.Second),
//	)
//
//	if err := cb.Allow(); err != nil {
//	    return err // 熔断器打开，快速失败
//	}
//	cb.RecordSuccess()
//
// 使用负载均衡：
//
//	sel := resilience.NewRoundRobinSelector()
//	instance, err := sel.Select(instances)
//
// # 熔断器状态
//
//   - CLOSED: 正常状态，请求正常通过
//   - OPEN: 熔断状态，请求直接失败
//   - HALF_OPEN: 半开状态，允许部分请求通过进行测试
package resilience

import (
	"context"
	"time"
)

// 默认熔断器配置常量。
const (
	// DefaultMaxRequests 半开状态下允许的最大试探请求数。
	DefaultMaxRequests = 10
	// DefaultErrorThreshold 错误率阈值（50%）。
	DefaultErrorThreshold = 0.5
	// DefaultWaitDuration 打开状态等待时间。
	DefaultWaitDuration = 30 * time.Second
)

// State 熔断器状态。
//
// 熔断器有三种状态：
//   - Closed（关闭）：正常处理请求，统计错误率
//   - Open（打开）：快速失败，拒绝所有请求
//   - HalfOpen（半开）：允许少量请求试探，评估是否恢复
type State int32

// 熔断器状态常量。
const (
	// StateClosed 关闭状态，正常处理请求。
	StateClosed State = iota
	// StateOpen 打开状态，快速失败。
	StateOpen
	// StateHalfOpen 半开状态，尝试恢复。
	StateHalfOpen
)

// Breaker 熔断器接口。
//
// 熔断器模式用于防止级联故障。当依赖服务不可用时，
// 快速失败而不是等待超时，保护系统资源。
//
// 使用示例：
//
//	breaker := resilience.NewBreaker(
//	    resilience.WithErrorThreshold(0.5),
//	    resilience.WithWaitDuration(30 * time.Second),
//	)
//
//	if err := breaker.Allow(); err != nil {
//	    return err // 熔断器打开，快速失败
//	}
//	// 执行业务逻辑
//	breaker.RecordSuccess() // 或 RecordFailure()
type Breaker interface {
	// Allow 检查是否允许请求通过。
	Allow() error
	// RecordSuccess 记录成功请求。
	RecordSuccess()
	// RecordFailure 记录失败请求。
	RecordFailure()
	// State 获取当前状态。
	State() State
}

// BreakerOption 熔断器配置选项函数。
type BreakerOption func(breaker Breaker)

// Selector 负载均衡选择器接口。
//
// 从一组服务实例中按策略选出一个目标实例。
//
// 使用示例：
//
//	sel := resilience.NewRoundRobinSelector()
//	inst, err := sel.Select(instances)
type Selector interface {
	// Select 选择服务实例。
	Select(instances []InstanceInfo) (InstanceInfo, error)
}

// Registry 服务注册中心接口。
//
// 提供服务实例的注册、注销、发现和监听功能。
//
// 使用示例：
//
//	var reg resilience.Registry
//	err := reg.Register(ctx, resilience.InstanceInfo{
//	    ServiceName: "user-service",
//	    ID:          "192.168.1.1:8080",
//	    Host:        "192.168.1.1",
//	    Port:        8080,
//	})
//	instances, err := reg.Discover(ctx, "user-service")
type Registry interface {
	// Register 注册服务实例到注册中心。
	Register(ctx context.Context, info InstanceInfo) error
	// Deregister 从注册中心注销服务实例。
	Deregister(ctx context.Context, info InstanceInfo) error
	// Discover 发现指定服务的所有实例。
	Discover(ctx context.Context, serviceName string) ([]InstanceInfo, error)
	// Watch 监听服务实例变更，返回实例列表变更通道。
	Watch(ctx context.Context, serviceName string) (<-chan []InstanceInfo, error)
}

// InstanceInfo 注册中心中的服务实例信息。
//
// 包含服务名、实例 ID、网络地址、权重、健康状态和元数据。
type InstanceInfo struct {
	// ServiceName 服务名称。
	ServiceName string
	// ID 实例唯一标识。
	ID string
	// Host 主机地址。
	Host string
	// Port 端口号。
	Port int
	// Weight 负载均衡权重。
	Weight int
	// Healthy 是否健康。
	Healthy bool
	// Metadata 扩展元数据。
	Metadata map[string]string
}

// HealthStatus 健康状态。
type HealthStatus int

// 健康状态常量。
const (
	// HealthUp 健康。
	HealthUp HealthStatus = iota
	// HealthDown 不健康。
	HealthDown
	// HealthUnknown 未知。
	HealthUnknown
)

// ServiceInstance 服务实例。
type ServiceInstance struct {
	// ID 服务实例 ID。
	ID string
	// URL 服务地址。
	URL string
	// Weight 权重。
	Weight int
	// Metadata 元数据。
	Metadata map[string]string
	// Health 健康状态。
	Health HealthStatus
	// Active 活跃连接数。
	Active int64
}

// Balancer 负载均衡接口。
type Balancer interface {
	// Next 选择下一个服务实例。
	Next(backends []*ServiceInstance) (*ServiceInstance, error)
}

// Breaker errors.
var (
	// ErrCircuitOpen 熔断器处于打开状态。
	ErrCircuitOpen error
	// ErrCircuitHalfOpen 熔断器处于半开状态且已达到最大请求数。
	ErrCircuitHalfOpen error
	// ErrNoInstances 没有可用实例。
	ErrNoInstances error
	// ErrNoBackends 没有可用后端服务。
	ErrNoBackends error
)
