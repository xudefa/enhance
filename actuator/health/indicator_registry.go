package health

import "sync"

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

// GlobalIndicatorRegistry 返回全局指标注册表
var globalIndicatorRegistry = NewIndicatorRegistry()

// GlobalIndicatorRegistry 返回全局指标注册表
func GlobalIndicatorRegistry() *IndicatorRegistry {
	return globalIndicatorRegistry
}

// RegisterIndicator 注册健康指标到全局注册表
func RegisterIndicator(name string, fn IndicatorFunc) {
	globalIndicatorRegistry.RegisterFunc(name, fn)
}
