// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"reflect"
	"sync"
)

// weaver 织入器内部实现。
type weaver struct {
	mu      sync.RWMutex
	aspects []*AspectMeta
	factory *ProxyFactory
}

// NewWeaver 创建织入器
//
// 返回值:
//   - Weaver: 织入器实例
//
// 示例:
//
//	weaver := aop.NewWeaver()
//	weaver.AddAspects(aspectMeta)
//	proxy := weaver.Weave(&UserService{})
func NewWeaver() Weaver {
	return &weaver{
		aspects: make([]*AspectMeta, 0),
		factory: nil,
	}
}

// AddAspects 添加切面
func (w *weaver) AddAspects(aspects ...*AspectMeta) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.aspects = append(w.aspects, aspects...)
}

// Weave 织入目标对象
func (w *weaver) Weave(target any) any {
	if target == nil {
		return nil
	}

	w.mu.RLock()
	if len(w.aspects) == 0 {
		w.mu.RUnlock()
		return target
	}
	aspects := make([]*AspectMeta, len(w.aspects))
	copy(aspects, w.aspects)
	w.mu.RUnlock()

	factory := NewProxyFactory(target)
	factory.SetAspects(aspects)
	return factory.GetProxy()
}

// AopRegistry AOP 注册表
//
// 管理所有切面和织入器的注册中心。
// 用于在 IoC 容器中集成 AOP 功能。
type AopRegistry struct {
	mu      sync.RWMutex
	aspects []*AspectMeta
	weavers map[string]Weaver
}

// NewAopRegistry 创建 AOP 注册表
//
// 返回值:
//   - *AopRegistry: 注册表实例
func NewAopRegistry() *AopRegistry {
	return &AopRegistry{
		aspects: make([]*AspectMeta, 0),
		weavers: make(map[string]Weaver),
	}
}

// RegisterAspect 注册切面
func (r *AopRegistry) RegisterAspect(aspect *AspectMeta) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.aspects = append(r.aspects, aspect)
}

// GetAspects 获取所有切面
func (r *AopRegistry) GetAspects() []*AspectMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*AspectMeta, len(r.aspects))
	copy(result, r.aspects)
	return result
}

// RegisterWeaver 注册织入器
func (r *AopRegistry) RegisterWeaver(beanID string, weaver Weaver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.weavers[beanID] = weaver
}

// GetWeaver 获取织入器
func (r *AopRegistry) GetWeaver(beanID string) (Weaver, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.weavers[beanID]
	return w, ok
}

// WeaveIfNeeded 按需织入
//
// 如果指定 beanID 有对应的织入器，则织入目标对象。
func (r *AopRegistry) WeaveIfNeeded(beanID string, target any) any {
	r.mu.RLock()
	weaver, ok := r.weavers[beanID]
	r.mu.RUnlock()
	if !ok {
		return target
	}
	return weaver.Weave(target)
}

// MatchAspectsForType 为类型匹配切面
//
// 根据类型匹配所有适用的切面，并按 Order 排序。
func (r *AopRegistry) MatchAspectsForType(t reflect.Type) []*AspectMeta {
	r.mu.RLock()
	aspects := make([]*AspectMeta, len(r.aspects))
	copy(aspects, r.aspects)
	r.mu.RUnlock()

	var matched []*AspectMeta
	for _, a := range aspects {
		if a.PointCut == nil {
			continue
		}
		for i := 0; i < t.NumMethod(); i++ {
			m := t.Method(i)
			if a.PointCut.MatchMethod(m) {
				matched = append(matched, a)
				break
			}
		}
	}
	SortAspectsByOrder(matched)
	return matched
}
