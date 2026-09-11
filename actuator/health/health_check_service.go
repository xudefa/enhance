package health

import "context"

// HealthCheckService 健康检查服务
//
// 提供统一的健康检查入口，聚合所有注册的指标。
type HealthCheckService struct {
	registry   *IndicatorRegistry
	aggregator *Aggregator
}

// DefaultHealthCheckService 默认健康检查服务
var DefaultHealthCheckService = NewHealthCheckService()

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

// CheckHealth 执行默认健康检查
func CheckHealth(ctx context.Context) Health {
	return DefaultHealthCheckService.Check(ctx)
}

// RegisterCustomIndicator 注册自定义健康指标到默认服务
func RegisterCustomIndicator(name string, fn IndicatorFunc) {
	DefaultHealthCheckService.RegisterIndicator(name, fn)
}
