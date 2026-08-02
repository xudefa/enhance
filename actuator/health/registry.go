package health

import (
	"context"
	"maps"
	"os"
	"runtime"
	"sync"
	"time"
)

// IndicatorFunc 健康指标函数类型
//
// 简化版健康指标定义，只需提供一个函数即可注册健康检查。
type IndicatorFunc func(ctx context.Context) Health

// SimpleIndicator 简单健康指标实现
//
// 基于 IndicatorFunc 的健康指标实现。
type SimpleIndicator struct {
	name string
	fn   IndicatorFunc
}

// NewSimpleIndicator 创建简单健康指标
func NewSimpleIndicator(name string, fn IndicatorFunc) *SimpleIndicator {
	return &SimpleIndicator{
		name: name,
		fn:   fn,
	}
}

// Name 返回指标名称
func (s *SimpleIndicator) Name() string {
	return s.name
}

// Health 执行健康检查
func (s *SimpleIndicator) Health(ctx context.Context) Health {
	if s.fn == nil {
		return Health{
			Status:    StatusUnknown,
			Timestamp: time.Now(),
		}
	}
	return s.fn(ctx)
}

// HealthBuilder 健康信息构建器
//
// 提供流式 API 来构建 Health 对象。
type HealthBuilder struct {
	health Health
}

// Up 创建健康状态为 UP 的构建器
func Up() *HealthBuilder {
	return &HealthBuilder{
		health: Health{
			Status:    StatusUp,
			Details:   make(map[string]any),
			Timestamp: time.Now(),
		},
	}
}

// Down 创建健康状态为 DOWN 的构建器
func Down() *HealthBuilder {
	return &HealthBuilder{
		health: Health{
			Status:    StatusDown,
			Details:   make(map[string]any),
			Timestamp: time.Now(),
		},
	}
}

// Degraded 创建健康状态为 DEGRADED 的构建器
func Degraded() *HealthBuilder {
	return &HealthBuilder{
		health: Health{
			Status:    StatusDegraded,
			Details:   make(map[string]any),
			Timestamp: time.Now(),
		},
	}
}

// Outage 创建健康状态为 OUTAGE 的构建器
func Outage() *HealthBuilder {
	return &HealthBuilder{
		health: Health{
			Status:    StatusOutage,
			Details:   make(map[string]any),
			Timestamp: time.Now(),
		},
	}
}

// WithDetail 添加详细信息
func (b *HealthBuilder) WithDetail(key string, value any) *HealthBuilder {
	b.health.Details[key] = value
	return b
}

// WithDetails 批量添加详细信息
func (b *HealthBuilder) WithDetails(details map[string]any) *HealthBuilder {
	maps.Copy(b.health.Details, details)
	return b
}

// WithError 设置错误信息
func (b *HealthBuilder) WithError(err error) *HealthBuilder {
	b.health.Error = err
	return b
}

// Build 构建 Health 对象
func (b *HealthBuilder) Build() Health {
	b.health.Timestamp = time.Now()
	return b.health
}

// IndicatorRegistry 指标注册表
//
// 管理所有健康指标的注册和查询，支持按名称获取。
// 使用 sync.Map 优化读多写少场景的并发性能。
type IndicatorRegistry struct {
	indicators sync.Map // map[string]Indicator
}

// NewIndicatorRegistry 创建指标注册表
func NewIndicatorRegistry() *IndicatorRegistry {
	return &IndicatorRegistry{}
}

// Register 注册健康指标
func (r *IndicatorRegistry) Register(indicator Indicator) {
	r.indicators.Store(indicator.Name(), indicator)
}

// RegisterFunc 注册函数式健康指标
func (r *IndicatorRegistry) RegisterFunc(name string, fn IndicatorFunc) {
	r.Register(NewSimpleIndicator(name, fn))
}

// Get 获取健康指标
func (r *IndicatorRegistry) Get(name string) (Indicator, bool) {
	v, ok := r.indicators.Load(name)
	if !ok {
		return nil, false
	}
	ind, ok := v.(Indicator)
	if !ok {
		return nil, false
	}
	return ind, true
}

// GetAll 获取所有健康指标
func (r *IndicatorRegistry) GetAll() []Indicator {
	var indicators []Indicator
	r.indicators.Range(func(key, value any) bool {
		if ind, ok := value.(Indicator); ok {
			indicators = append(indicators, ind)
		}
		return true
	})
	return indicators
}

// Remove 移除健康指标
func (r *IndicatorRegistry) Remove(name string) {
	r.indicators.Delete(name)
}

// 全局指标注册表
var globalIndicatorRegistry = NewIndicatorRegistry()

// GlobalIndicatorRegistry 返回全局指标注册表
func GlobalIndicatorRegistry() *IndicatorRegistry {
	return globalIndicatorRegistry
}

// RegisterIndicator 注册健康指标到全局注册表
func RegisterIndicator(name string, fn IndicatorFunc) {
	globalIndicatorRegistry.RegisterFunc(name, fn)
}

// RuntimeHealthIndicator 运行时健康指标
//
// 提供 Go 运行时相关的健康指标，包括:
// - Goroutine 数量
// - GC 统计信息
// - 内存使用情况
type RuntimeHealthIndicator struct{}

// NewRuntimeHealthIndicator 创建运行时健康指标
func NewRuntimeHealthIndicator() *RuntimeHealthIndicator {
	return &RuntimeHealthIndicator{}
}

// Name 返回指标名称
func (r *RuntimeHealthIndicator) Name() string {
	return "runtime"
}

// Health 执行健康检查
func (r *RuntimeHealthIndicator) Health(ctx context.Context) Health {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	return Health{
		Status: StatusUp,
		Details: map[string]any{
			"goroutines":    runtime.NumGoroutine(),
			"heap_alloc":    stats.HeapAlloc,
			"heap_sys":      stats.HeapSys,
			"heap_idle":     stats.HeapIdle,
			"heap_inuse":    stats.HeapInuse,
			"heap_released": stats.HeapReleased,
			"gc_pause_ns":   stats.PauseTotalNs,
			"gc_count":      stats.NumGC,
			"num_cpu":       runtime.NumCPU(),
			"go_version":    runtime.Version(),
		},
		Timestamp: time.Now(),
	}
}

// SystemHealthIndicator 系统健康指标
//
// 提供系统级别的指标，包括:
// - CPU 使用率
// - 内存使用率
// - 磁盘使用率
type SystemHealthIndicator struct{}

// NewSystemHealthIndicator 创建系统健康指标
func NewSystemHealthIndicator() *SystemHealthIndicator {
	return &SystemHealthIndicator{}
}

// Name 返回指标名称
func (s *SystemHealthIndicator) Name() string {
	return "system"
}

// Health 执行健康检查
func (s *SystemHealthIndicator) Health(ctx context.Context) Health {
	return Health{
		Status: StatusUp,
		Details: map[string]any{
			"os":       runtime.GOOS,
			"arch":     runtime.GOARCH,
			"num_cpu":  runtime.NumCPU(),
			"hostname": getHostname(),
		},
		Timestamp: time.Now(),
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

// HealthCheckService 健康检查服务
//
// 提供统一的健康检查入口，聚合所有注册的指标。
type HealthCheckService struct {
	registry   *IndicatorRegistry
	aggregator *Aggregator
}

// NewHealthCheckService 创建健康检查服务
func NewHealthCheckService() *HealthCheckService {
	service := &HealthCheckService{
		registry:   NewIndicatorRegistry(),
		aggregator: NewAggregator(),
	}

	// 注册内置指标
	service.registry.Register(NewRuntimeHealthIndicator())
	service.registry.Register(NewSystemHealthIndicator())

	// 将注册表的指标添加到聚合器
	for _, ind := range service.registry.GetAll() {
		service.aggregator.AddIndicator(ind)
	}

	return service
}

// Check 执行健康检查
func (s *HealthCheckService) Check(ctx context.Context) Health {
	return s.aggregator.Aggregate(ctx)
}

// RegisterIndicator 注册自定义健康指标
func (s *HealthCheckService) RegisterIndicator(name string, fn IndicatorFunc) {
	indicator := NewSimpleIndicator(name, fn)
	s.registry.Register(indicator)
	s.aggregator.AddIndicator(indicator)
}

// GetIndicator 获取健康指标
func (s *HealthCheckService) GetIndicator(name string) (Indicator, bool) {
	return s.registry.Get(name)
}

// GetAllIndicators 获取所有健康指标
func (s *HealthCheckService) GetAllIndicators() []Indicator {
	return s.registry.GetAll()
}

// DefaultHealthCheckService 默认健康检查服务
var DefaultHealthCheckService = NewHealthCheckService()

// CheckHealth 执行默认健康检查
func CheckHealth(ctx context.Context) Health {
	return DefaultHealthCheckService.Check(ctx)
}

// RegisterCustomIndicator 注册自定义健康指标到默认服务
func RegisterCustomIndicator(name string, fn IndicatorFunc) {
	DefaultHealthCheckService.RegisterIndicator(name, fn)
}
