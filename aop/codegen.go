package aop

import (
	"fmt"
	"reflect"
	"sync"
)

// GeneratedProxyRegistry 代码生成代理注册表
//
// 管理代码生成的代理对象，提供查找和获取功能
type GeneratedProxyRegistry struct {
	proxies sync.Map // beanID -> proxyType
}

// NewGeneratedProxyRegistry 创建代码生成代理注册表
func NewGeneratedProxyRegistry() *GeneratedProxyRegistry {
	return &GeneratedProxyRegistry{}
}

// Register 注册代理类型
func (r *GeneratedProxyRegistry) Register(beanID string, proxyType reflect.Type) {
	r.proxies.Store(beanID, proxyType)
}

// Get 获取代理类型
func (r *GeneratedProxyRegistry) Get(beanID string) (reflect.Type, bool) {
	v, ok := r.proxies.Load(beanID)
	if !ok {
		return nil, false
	}
	return v.(reflect.Type), true
}

// Has 检查是否存在代理
func (r *GeneratedProxyRegistry) Has(beanID string) bool {
	_, ok := r.proxies.Load(beanID)
	return ok
}

// List 列出所有注册的bean ID
func (r *GeneratedProxyRegistry) List() []string {
	var ids []string
	r.proxies.Range(func(key, _ any) bool {
		if s, ok := key.(string); ok {
			ids = append(ids, s)
		}
		return true
	})
	return ids
}

// Clear 清空注册表
func (r *GeneratedProxyRegistry) Clear() {
	r.proxies.Range(func(key, _ any) bool {
		r.proxies.Delete(key)
		return true
	})
}

// GlobalGeneratedProxyRegistry 全局代码生成代理注册表
var GlobalGeneratedProxyRegistry = NewGeneratedProxyRegistry()

// RegisterGeneratedProxy 注册代码生成的代理
func RegisterGeneratedProxy(beanID string, proxyType reflect.Type) {
	GlobalGeneratedProxyRegistry.Register(beanID, proxyType)
}

// GetGeneratedProxy 获取代码生成的代理类型
func GetGeneratedProxy(beanID string) (reflect.Type, bool) {
	return GlobalGeneratedProxyRegistry.Get(beanID)
}

// HasGeneratedProxy 检查是否存在代码生成的代理
func HasGeneratedProxy(beanID string) bool {
	return GlobalGeneratedProxyRegistry.Has(beanID)
}

// GeneratedProxyFactory 代码生成代理工厂
//
// 创建代码生成的代理对象
type GeneratedProxyFactory struct {
	registry *GeneratedProxyRegistry
}

// NewGeneratedProxyFactory 创建代码生成代理工厂
func NewGeneratedProxyFactory() *GeneratedProxyFactory {
	return &GeneratedProxyFactory{
		registry: GlobalGeneratedProxyRegistry,
	}
}

// Create 创建代理对象
func (f *GeneratedProxyFactory) Create(beanID string, target any) (any, error) {
	proxyType, ok := f.registry.Get(beanID)
	if !ok {
		return nil, fmt.Errorf("no generated proxy found for bean: %s", beanID)
	}

	if proxyType.Kind() != reflect.Pointer && proxyType.Kind() != reflect.Interface {
		return nil, fmt.Errorf("proxy type for bean %s is not a pointer or interface: %v", beanID, proxyType.Kind())
	}

	proxyValue := reflect.New(proxyType.Elem())
	proxy := proxyValue.Interface()

	// 设置目标对象
	targetField := proxyValue.Elem().FieldByName("Target")
	if targetField.IsValid() && targetField.CanSet() {
		targetField.Set(reflect.ValueOf(target))
	}

	return proxy, nil
}

// CreateOrFallback 创建代理或回退到运行时代理
func (f *GeneratedProxyFactory) CreateOrFallback(beanID string, target any, fallback Weaver) any {
	proxy, err := f.Create(beanID, target)
	if err != nil {
		// 回退到运行时代理
		if fallback != nil {
			return fallback.Weave(target)
		}
		return target
	}
	return proxy
}

// AspectMetadataExtractor 切面元数据提取器
//
// 从代码生成的代理中提取切面元数据
type AspectMetadataExtractor struct{}

// NewAspectMetadataExtractor 创建切面元数据提取器
func NewAspectMetadataExtractor() *AspectMetadataExtractor {
	return &AspectMetadataExtractor{}
}

// Extract 从代理类型提取切面元数据
func (e *AspectMetadataExtractor) Extract(proxyType reflect.Type) []*AspectMeta {
	if proxyType.Kind() != reflect.Struct {
		return nil
	}

	_, found := proxyType.FieldByName("aspects")
	if !found {
		return nil
	}

	// 这里需要解析代码生成的切面元数据
	// 由于代码生成的结构是固定的，我们可以直接提取
	return nil
}

// ExtractFromBeanID 从bean ID提取切面元数据
func (e *AspectMetadataExtractor) ExtractFromBeanID(beanID string) []*AspectMeta {
	proxyType, ok := GlobalGeneratedProxyRegistry.Get(beanID)
	if !ok {
		return nil
	}
	return e.Extract(proxyType)
}
