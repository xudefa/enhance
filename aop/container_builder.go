package aop

import (
	"fmt"
	"reflect"

	"github.com/xudefa/enhance/core"
)

// AopContainerBuilder AOP容器构建器
//
// 提供流式API构建AOP容器
type AopContainerBuilder struct {
	baseContainer core.Container
	config        *AopConfig
	aspects       []*AspectMeta
	beans         []*AopBeanDefinition
}

// NewAopContainerBuilder 创建AOP容器构建器
func NewAopContainerBuilder() *AopContainerBuilder {
	return &AopContainerBuilder{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
		beans:   make([]*AopBeanDefinition, 0),
	}
}

// WithBaseContainer 设置基础容器
func (b *AopContainerBuilder) WithBaseContainer(container core.Container) *AopContainerBuilder {
	b.baseContainer = container
	return b
}

// WithConfig 设置配置
func (b *AopContainerBuilder) WithConfig(config *AopConfig) *AopContainerBuilder {
	b.config = config
	return b
}

// WithAopMode 设置AOP模式
func (b *AopContainerBuilder) WithAopMode(mode AopMode) *AopContainerBuilder {
	b.config.Mode = mode
	return b
}

// WithAspect 添加切面
func (b *AopContainerBuilder) WithAspect(aspect *AspectMeta) *AopContainerBuilder {
	b.aspects = append(b.aspects, aspect)
	return b
}

// WithAspects 批量添加切面
func (b *AopContainerBuilder) WithAspects(aspects ...*AspectMeta) *AopContainerBuilder {
	b.aspects = append(b.aspects, aspects...)
	return b
}

// WithBean 添加Bean
func (b *AopContainerBuilder) WithBean(beanDef *AopBeanDefinition) *AopContainerBuilder {
	b.beans = append(b.beans, beanDef)
	return b
}

// WithBeanWithID 添加Bean（指定ID）
func (b *AopContainerBuilder) WithBeanWithID(beanID string, beanType reflect.Type, target any) *AopContainerBuilder {
	beanDef := NewAopBeanDefinition(beanID, beanType)
	b.beans = append(b.beans, beanDef)
	return b
}

// WithBeanWithAspects 添加Bean（带切面）
func (b *AopContainerBuilder) WithBeanWithAspects(beanID string, beanType reflect.Type, target any, aspects ...*AspectMeta) *AopContainerBuilder {
	beanDef := NewAopBeanDefinition(beanID, beanType)
	beanDef.WithAspects(aspects...)
	b.beans = append(b.beans, beanDef)
	return b
}

// Build 构建AOP容器
func (b *AopContainerBuilder) Build() (*AopContainer, error) {
	container := NewAopContainer(b.baseContainer)
	container.integration.config = b.config

	// 注册切面
	container.RegisterAspects(b.aspects...)

	// 注册Bean
	for _, beanDef := range b.beans {
		// 从基础容器获取目标对象，如果基础容器不存在或对象不存在，则创建默认实例
		var target any
		if b.baseContainer != nil && beanDef.TargetType != nil {
			if objs, err := b.baseContainer.Get(beanDef.TargetType); err == nil && len(objs) > 0 {
				target = objs[0]
			}
		}

		// 如果仍未获取到目标对象，且定义中有具体类型，则创建零值
		if target == nil && beanDef.TargetType != nil {
			target = reflect.New(beanDef.TargetType).Interface()
		}

		// 注册AOP Bean
		if err := container.RegisterAopBean(beanDef, target); err != nil {
			return nil, fmt.Errorf("failed to register AOP bean %s: %w", beanDef.BeanID, err)
		}
	}

	return container, nil
}

// BuildOrPanic 构建AOP容器（失败时panic）
func (b *AopContainerBuilder) BuildOrPanic() *AopContainer {
	container, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build AOP container: %v", err))
	}
	return container
}

// CreateAopContainer 创建AOP容器的便捷函数
func CreateAopContainer() *AopContainer {
	return NewAopContainer(nil)
}

// CreateAopContainerWithConfig 创建AOP容器（指定配置）
func CreateAopContainerWithConfig(config *AopConfig) *AopContainer {
	container := NewAopContainer(nil)
	container.integration.config = config
	return container
}

// RegisterAopBeanToGlobal 注册AOP Bean到全局容器
func RegisterAopBeanToGlobal(beanID string, beanType reflect.Type, target any) error {
	beanDef := NewAopBeanDefinition(beanID, beanType)
	_, err := GlobalAopBeanFactory.CreateBean(beanID, beanDef, target)
	return err
}

// RegisterAspectToGlobalContainer 注册切面到全局容器
func RegisterAspectToGlobalContainer(aspect *AspectMeta) {
	GetGlobalAopIntegration().RegisterAspect(aspect)
}
