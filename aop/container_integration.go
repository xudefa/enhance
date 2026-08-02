package aop

import (
	"reflect"
	"sync"
)

// AopBeanPostProcessor AOP Bean后置处理器
//
// 在Bean创建后自动应用AOP代理
type AopBeanPostProcessor struct {
	mu          sync.RWMutex
	integration *AopIntegration
	enabled     bool
}

// NewAopBeanPostProcessor 创建AOP Bean后置处理器
func NewAopBeanPostProcessor(integration *AopIntegration) *AopBeanPostProcessor {
	if integration == nil {
		integration = GetGlobalAopIntegration()
	}
	return &AopBeanPostProcessor{
		integration: integration,
		enabled:     true,
	}
}

// ProcessBean 处理Bean
func (p *AopBeanPostProcessor) ProcessBean(beanID string, bean any) any {
	p.mu.RLock()
	enabled := p.enabled
	integration := p.integration
	p.mu.RUnlock()

	if !enabled || bean == nil {
		return bean
	}

	// 检查是否需要AOP代理
	if p.needsProxy(bean) {
		proxy := integration.CreateProxy(beanID, bean)
		if proxy != nil && proxy != bean {
			return proxy
		}
	}

	return bean
}

// needsProxy 检查是否需要代理
func (p *AopBeanPostProcessor) needsProxy(bean any) bool {
	if bean == nil {
		return false
	}

	// 检查是否有匹配的切面
	beanType := reflect.TypeOf(bean)
	aspects := p.integration.GetManager().MatchAspectsForType(beanType)
	return len(aspects) > 0
}

// Enable 启用处理器
func (p *AopBeanPostProcessor) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

// Disable 禁用处理器
func (p *AopBeanPostProcessor) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

// IsEnabled 检查是否启用
func (p *AopBeanPostProcessor) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// GlobalAopBeanPostProcessor 全局AOP Bean后置处理器
var GlobalAopBeanPostProcessor = NewAopBeanPostProcessor(nil)

// AopBeanDefinition AOP Bean定义
//
// 扩展标准Bean定义，添加AOP相关配置
type AopBeanDefinition struct {
	BeanID     string
	EnableAop  bool
	ProxyMode  AopMode
	TargetType reflect.Type
	ProxyType  reflect.Type
	Aspects    []*AspectMeta
}

// NewAopBeanDefinition 创建AOP Bean定义
func NewAopBeanDefinition(beanID string, beanType reflect.Type) *AopBeanDefinition {
	return &AopBeanDefinition{
		BeanID:     beanID,
		EnableAop:  true,
		ProxyMode:  AopModeMixed,
		TargetType: beanType,
	}
}

// WithAopEnabled 设置启用AOP
func (d *AopBeanDefinition) WithAopEnabled(enabled bool) *AopBeanDefinition {
	d.EnableAop = enabled
	return d
}

// WithProxyMode 设置代理模式
func (d *AopBeanDefinition) WithProxyMode(mode AopMode) *AopBeanDefinition {
	d.ProxyMode = mode
	return d
}

// WithAspects 设置切面
func (d *AopBeanDefinition) WithAspects(aspects ...*AspectMeta) *AopBeanDefinition {
	d.Aspects = append(d.Aspects, aspects...)
	return d
}

// WithProxyType 设置代理类型
func (d *AopBeanDefinition) WithProxyType(proxyType reflect.Type) *AopBeanDefinition {
	d.ProxyType = proxyType
	return d
}
