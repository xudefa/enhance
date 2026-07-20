package scope

import (
	"sync"
)

// singletonScope 单例作用域实现。
// 使用 sync.Map 优化读多写少场景：
// - 注册阶段（启动时）：写入 Bean 实例
// - 运行阶段：高频读取，极少写入
type singletonScope struct {
	instances sync.Map
	mu        sync.Map // 保护每个 beanID 的创建过程
}

// Get 获取 Bean 实例（单例模式）。
func (s *singletonScope) Get(beanID string, factory func(c ...any) (any, error)) (any, error) {
	// 快速路径：直接使用传入的标准 Bean ID 作为 key
	if instance, ok := s.instances.Load(beanID); ok {
		return instance, nil
	}

	// 慢速路径：使用 per-beanID 锁避免并发创建
	lock, _ := s.mu.LoadOrStore(beanID, &sync.Mutex{})
	mu := lock.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	// 双重检查：获取锁后再次检查缓存
	if instance, ok := s.instances.Load(beanID); ok {
		return instance, nil
	}

	// 创建实例
	instance, err := factory()
	if err != nil {
		return nil, err
	}

	// Store 是并发安全的，使用标准 Bean ID 存储
	s.instances.Store(beanID, instance)
	return instance, nil
}

// Remove 移除 Bean 实例。
func (s *singletonScope) Remove(beanID string) {
	s.instances.Delete(beanID)
}

// Clear 清空所有实例。
func (s *singletonScope) Clear() {
	s.instances.Range(func(key, value any) bool {
		s.instances.Delete(key)
		return true
	})
}

// prototypeScope 原型作用域实现。
type prototypeScope struct {
}

// Get 获取 Bean 实例（每次创建新实例）。
func (s *prototypeScope) Get(beanID string, factory func(c ...any) (any, error)) (any, error) {
	// 原型作用域：每次创建新实例，不缓存
	return factory()
}

// Remove 原型作用域无需移除。
func (s *prototypeScope) Remove(beanID string) {
	// 原型作用域每次创建新实例，无需移除
}

// Clear 原型作用域无需清空。
func (s *prototypeScope) Clear() {}

// defaultScopeRegistry 默认作用域注册表实现。
// 使用 sync.Map 优化读多写少场景。
type defaultScopeRegistry struct {
	scopes sync.Map
}

// Register 注册自定义作用域。
func (r *defaultScopeRegistry) Register(name string, scope Scope) {
	r.scopes.Store(name, scope)
}

// Get 获取指定名称的作用域。
func (r *defaultScopeRegistry) Get(name string) Scope {
	if scope, ok := r.scopes.Load(name); ok {
		return scope.(Scope)
	}
	return nil
}

// Has 检查是否存在指定名称的作用域。
func (r *defaultScopeRegistry) Has(name string) bool {
	_, exists := r.scopes.Load(name)
	return exists
}

// NewSingletonScope 创建单例作用域实例。
func NewSingletonScope() Scope {
	return &singletonScope{}
}

// NewPrototypeScope 创建原型作用域实例。
func NewPrototypeScope() Scope {
	return &prototypeScope{}
}

// NewScopeRegistry 创建作用域注册表实例。
func NewScopeRegistry() ScopeRegistry {
	registry := &defaultScopeRegistry{}
	// 注册内置作用域
	registry.scopes.Store(SingletonScope, NewSingletonScope())
	registry.scopes.Store(PrototypeScope, NewPrototypeScope())
	return registry
}
