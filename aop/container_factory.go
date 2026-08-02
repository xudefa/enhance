package aop

import (
	"github.com/xudefa/enhance/core"
)

// AopBeanFactory AOP Bean工厂
//
// 创建AOP代理Bean
type AopBeanFactory struct {
	integration *AopIntegration
	processor   *AopBeanPostProcessor
}

// NewAopBeanFactory 创建AOP Bean工厂
func NewAopBeanFactory(integration *AopIntegration) *AopBeanFactory {
	if integration == nil {
		integration = GetGlobalAopIntegration()
	}
	return &AopBeanFactory{
		integration: integration,
		processor:   NewAopBeanPostProcessor(integration),
	}
}

// CreateBean 创建Bean
func (f *AopBeanFactory) CreateBean(beanID string, beanDef *AopBeanDefinition, target any) (any, error) {
	if beanDef == nil || !beanDef.EnableAop {
		return target, nil
	}

	// 注册切面
	if len(beanDef.Aspects) > 0 {
		f.integration.RegisterAspects(beanDef.Aspects...)
	}

	// 创建代理
	proxy := f.integration.CreateProxy(beanID, target)
	if proxy == nil {
		return target, nil
	}

	return proxy, nil
}

// RegisterBean 注册Bean到容器
func (f *AopBeanFactory) RegisterBean(container core.Container, beanDef *AopBeanDefinition, target any) error {
	proxy, err := f.CreateBean(beanDef.BeanID, beanDef, target)
	if err != nil {
		return err
	}

	// 注册到容器
	return container.RegisterInstance(proxy, beanDef.ProxyType)
}

// GetProcessor 获取后置处理器
func (f *AopBeanFactory) GetProcessor() *AopBeanPostProcessor {
	return f.processor
}

// GlobalAopBeanFactory 全局AOP Bean工厂
var GlobalAopBeanFactory = NewAopBeanFactory(nil)
