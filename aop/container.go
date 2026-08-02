// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"reflect"

	"github.com/xudefa/enhance/core"
)

// ==================== AopContainer 结构体 ====================

// AopContainer AOP 容器。
//
// 集成 AOP 功能的 IoC 容器。
type AopContainer struct {
	core.Container
	integration *AopIntegration
	factory     *AopBeanFactory
	processor   *AopBeanPostProcessor
}

// NewAopContainer 创建AOP容器。
func NewAopContainer(baseContainer core.Container) *AopContainer {
	if baseContainer == nil {
		baseContainer = core.NewContainer()
	}

	// 每个容器使用独立的集成器，保证配置互不影响。
	// 集成器内部的 manager 仍为全局共享（NewAopIntegration 使用 GlobalAopManager），
	// 因此切面仍可在容器间共享。
	integration := NewAopIntegration(DefaultAopConfig())
	factory := NewAopBeanFactory(integration)
	processor := factory.GetProcessor()

	return &AopContainer{
		Container:   baseContainer,
		integration: integration,
		factory:     factory,
		processor:   processor,
	}
}

// RegisterAopBean 注册AOP Bean
func (c *AopContainer) RegisterAopBean(beanDef *AopBeanDefinition, target any) error {
	return c.factory.RegisterBean(c.Container, beanDef, target)
}

// RegisterAopBeanWithID 注册AOP Bean（指定ID）
func (c *AopContainer) RegisterAopBeanWithID(beanID string, beanType reflect.Type, target any) error {
	beanDef := NewAopBeanDefinition(beanID, beanType)
	beanDef.BeanID = beanID
	return c.RegisterAopBean(beanDef, target)
}

// RegisterAopBeanWithAspects 注册AOP Bean（带切面）
func (c *AopContainer) RegisterAopBeanWithAspects(beanID string, beanType reflect.Type, target any, aspects ...*AspectMeta) error {
	beanDef := NewAopBeanDefinition(beanID, beanType)
	beanDef.WithAspects(aspects...)
	return c.RegisterAopBean(beanDef, target)
}

// RegisterAspect 注册切面
func (c *AopContainer) RegisterAspect(aspect *AspectMeta) {
	c.integration.RegisterAspect(aspect)
}

// RegisterAspects 批量注册切面
func (c *AopContainer) RegisterAspects(aspects ...*AspectMeta) {
	c.integration.RegisterAspects(aspects...)
}

// GetAspects 获取所有切面
func (c *AopContainer) GetAspects() []*AspectMeta {
	return c.integration.GetAspects()
}

// GetIntegration 获取AOP集成器
func (c *AopContainer) GetIntegration() *AopIntegration {
	return c.integration
}

// GetFactory 获取Bean工厂
func (c *AopContainer) GetFactory() *AopBeanFactory {
	return c.factory
}

// GetProcessor 获取后置处理器
func (c *AopContainer) GetProcessor() *AopBeanPostProcessor {
	return c.processor
}

// EnableAop 启用AOP
func (c *AopContainer) EnableAop() {
	c.processor.Enable()
}

// DisableAop 禁用AOP
func (c *AopContainer) DisableAop() {
	c.processor.Disable()
}

// IsAopEnabled 检查AOP是否启用
func (c *AopContainer) IsAopEnabled() bool {
	return c.processor.IsEnabled()
}

// registerProxyType 注册代理类型（供扫描器使用）
func (c *AopContainer) registerProxyType(typeName string, filePath string) {
	c.integration.RegisterProxyType(typeName, filePath)
}
