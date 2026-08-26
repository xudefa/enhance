// Package core 提供了一个类型安全的依赖注入（DI）容器实现，灵感来自 Spring Framework 的 IoC 容器。
//
// 详细文档请参阅 package core 的说明。
package core

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/xudefa/enhance/core/lifecycle"
	"github.com/xudefa/enhance/core/registry"
	"github.com/xudefa/enhance/core/scope"
)

// defaultContainer 默认 IoC 容器实现。
type defaultContainer struct {
	mu            sync.RWMutex
	reg           registry.BeanRegistry
	scopeRegistry scope.ScopeRegistry
	lifecycleMgr  lifecycle.LifecycleManager
	parent        Container
	initialized   bool
	destroyed     bool

	beanCreationLocks sync.Map // beanID -> *sync.Mutex，保护单个 Bean 创建+初始化+缓存序列
	beanCreating      sync.Map // beanID -> goroutineID(string)，用于检测同一 goroutine 的工厂型循环依赖
}

// RegisterBean 注册一个 Bean。
func (c *defaultContainer) RegisterBean(def registry.BeanDef) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if def.Type == nil {
		return fmt.Errorf("type is nil")
	}
	if def.Factory == nil {
		return fmt.Errorf("factory is nil for type %s", def.Type)
	}
	if c.initialized {
		return ErrContainerAlreadyInitialized
	}
	beanID := c.Generate(def.Type, def.Name)
	def.Name = beanID
	if err := c.reg.Register(def, beanID); err != nil {
		if errors.Is(err, registry.ErrBeanAlreadyExists) {
			return fmt.Errorf("%w: %v", ErrBeanAlreadyExists, err)
		}
		return err
	}
	return nil
}

// RegisterInstance 注册一个已存在的 Bean 实例。
func (c *defaultContainer) RegisterInstance(instance any, typ reflect.Type) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return ErrContainerAlreadyInitialized
	}
	return c.reg.RegisterInstance(instance, typ, c.Generate(typ))
}

// Get 获取指定类型的 Bean 实例列表。
func (c *defaultContainer) Get(typ reflect.Type) ([]any, error) {
	c.mu.RLock()

	if c.destroyed {
		c.mu.RUnlock()
		return nil, ErrContainerDestroyed
	}

	beanIDs := c.reg.GetByType(typ)
	if len(beanIDs) == 0 {
		c.mu.RUnlock()
		// 尝试从父容器获取
		if c.parent != nil {
			return c.parent.Get(typ)
		}
		return nil, ErrBeanNotFound
	}

	// 复制beanIDs到本地切片，避免释放锁后被修改
	beanIDsCopy := make([]string, len(beanIDs))
	copy(beanIDsCopy, beanIDs)

	// 首选 Bean 排到最前，保证 GetByName/injectImpl 默认获取首选 Bean
	if primaryID, ok := c.reg.GetPrimaryByType(typ); ok {
		ordered := make([]string, 0, len(beanIDsCopy))
		ordered = append(ordered, primaryID)
		for _, id := range beanIDsCopy {
			if id != primaryID {
				ordered = append(ordered, id)
			}
		}
		beanIDsCopy = ordered
	}

	// 释放读锁，避免在 resolveBean 中死锁
	c.mu.RUnlock()

	instances := make([]any, 0, len(beanIDsCopy))
	for _, beanID := range beanIDsCopy {
		instance, err := c.resolveBean(beanID, typ)
		if err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// GetByTypeAndName 获取指定名称的 Bean 实例。
func (c *defaultContainer) GetByTypeAndName(name string, typ reflect.Type) (any, error) {
	c.mu.RLock()

	if c.destroyed {
		c.mu.RUnlock()
		return nil, ErrContainerDestroyed
	}

	// 检查 Bean 定义是否存在
	def, ok := c.reg.GetDefinition(name)
	if !ok {
		c.mu.RUnlock()
		// 尝试从父容器获取
		if c.parent != nil {
			return c.parent.GetByTypeAndName(name, typ)
		}
		return nil, ErrBeanNotFound
	}

	// 验证类型是否匹配
	if def.Type != typ {
		c.mu.RUnlock()
		return nil, fmt.Errorf("bean %q type mismatch: got %v, want %v", name, def.Type, typ)
	}

	// 释放读锁，避免在 resolveBean 中死锁
	c.mu.RUnlock()

	// 将自定义名称解析为标准 Bean ID，确保命中正确的实例缓存，
	// 避免通过自定义名称创建出重复的单例实例
	canonicalID := c.Generate(typ, name)
	return c.resolveBean(canonicalID, typ)
}

// GetAll 获取所有 Bean 实例列表。
func (c *defaultContainer) GetAll() []any {
	c.mu.RLock()

	if c.destroyed {
		c.mu.RUnlock()
		return []any{}
	}

	mapInstances := c.reg.ListInstances()
	// 复制mapInstances到本地切片，避免释放锁后被修改
	instances := make([]any, 0, len(mapInstances))
	for _, v := range mapInstances {
		instances = append(instances, v)
	}
	c.mu.RUnlock()

	return instances
}

// Has 检查容器中是否存在指定类型和名称组合的 Bean。
func (c *defaultContainer) Has(name string, typ reflect.Type) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// 空名称表示按类型检查，否则 Has[T]() 永远返回 false
	if name == "" {
		return c.reg.HasType(typ)
	}

	if !c.reg.HasBean(name) {
		return false
	}

	// 验证类型是否匹配
	def, ok := c.reg.GetDefinition(name)
	if !ok {
		return false
	}

	return def.Type == typ
}

// HasType 检查容器中是否存在指定类型 Bean。
func (c *defaultContainer) HasType(typ reflect.Type) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.reg.HasType(typ)
}

// ListBeans 列出所有已注册的Bean信息
func (c *defaultContainer) ListBeans() map[string]*registry.BeanDef {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.reg.ListBeans()
}

// Initialize 初始化容器，创建所有 Singleton Bean 并调用 Init 回调。
//
// 初始化完成后，容器会标记为已初始化状态。
// 如果初始化过程中发生错误，容器会重置状态以允许重试。
func (c *defaultContainer) Initialize() error {
	c.mu.Lock()
	if c.initialized || c.destroyed {
		c.mu.Unlock()
		return ErrContainerAlreadyInitialized
	}
	// 使用 initializing 标记防止重入，不对外暴露为 initialized
	c.initialized = true
	c.mu.Unlock()

	// 获取所有 Bean ID
	beanIDs := c.reg.BeanIDs()

	// 预创建所有非延迟初始化的 Singleton Bean
	var initErr error
	for _, beanID := range beanIDs {
		def, ok := c.reg.GetDefinition(beanID)
		if !ok {
			continue
		}

		// 跳过延迟初始化的 Bean
		if def.Lazy {
			continue
		}

		// 只处理 Singleton Bean（空作用域默认为 Singleton）
		if def.Scope != registry.Singleton && def.Scope != "" {
			continue
		}

		// 创建并初始化 Bean
		if _, err := c.createAndInitialize(beanID, def); err != nil {
			initErr = err
			break
		}
	}

	// 初始化失败时重置状态，允许重试
	if initErr != nil {
		c.mu.Lock()
		c.initialized = false
		c.mu.Unlock()
		return initErr
	}

	return nil
}

// Destroy 销毁容器，调用所有 Singleton Bean 的 Destroy 回调并清理资源。
func (c *defaultContainer) Destroy() error {
	c.mu.Lock()
	if c.destroyed {
		c.mu.Unlock()
		return ErrContainerDestroyed
	}
	c.destroyed = true
	c.mu.Unlock()

	// 销毁所有 Bean
	return c.lifecycleMgr.DestroyAll()
}

// SetParent 设置父容器，子容器可以获取父容器中的 Bean。
func (c *defaultContainer) SetParent(parent Container) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.parent = parent
}

// GetParent 获取父容器，如果没有父容器则返回 nil。
func (c *defaultContainer) GetParent() Container {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.parent
}

// Types 返回容器中所有已注册的 Bean 类型列表。
func (c *defaultContainer) Types() []reflect.Type {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reg.Types()
}

// BeanCount 返回容器中已注册的 Bean 数量。
func (c *defaultContainer) BeanCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reg.Count()
}

// BeanCountType 返回容器中指定类型 Bean 数量。
func (c *defaultContainer) BeanCountType(typ reflect.Type) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reg.CountByType(typ)
}

// CreateBean 创建指定 ID 的 Bean 实例。
//
// 该方法实现了 BeanCreator 接口，用于在运行时动态创建 Bean 实例。
func (c *defaultContainer) CreateBean(beanID string) (any, error) {
	c.mu.RLock()

	if c.destroyed {
		c.mu.RUnlock()
		return nil, ErrContainerDestroyed
	}

	// 获取 Bean 定义
	def, ok := c.reg.GetDefinition(beanID)
	if !ok {
		c.mu.RUnlock()
		return nil, ErrBeanNotFound
	}

	// 释放读锁
	c.mu.RUnlock()

	// 创建并初始化 Bean
	return c.createAndInitialize(beanID, def)
}

// isDestroyed 检查容器是否已销毁（内部使用，可安全在释放锁后调用）。
func (c *defaultContainer) isDestroyed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.destroyed
}

// resolveBean 解析 Bean 实例（内部使用）。
func (c *defaultContainer) resolveBean(beanID string, _ reflect.Type) (any, error) {
	if c.isDestroyed() {
		return nil, ErrContainerDestroyed
	}

	// 尝试从缓存获取
	if instance, ok := c.reg.GetInstance(beanID); ok {
		return instance, nil
	}

	// 获取 Bean 定义
	def, ok := c.reg.GetDefinition(beanID)
	if !ok {
		return nil, ErrBeanNotFound
	}

	// 创建并初始化 Bean（内部有双重检查缓存）
	return c.createAndInitialize(beanID, def)
}

// createAndInitialize 创建 Bean 实例并调用初始化回调。
func (c *defaultContainer) createAndInitialize(beanID string, def *registry.BeanDef) (any, error) {
	if c.isDestroyed() {
		return nil, ErrContainerDestroyed
	}

	// 双重检查缓存
	if instance, ok := c.reg.GetInstance(beanID); ok {
		return instance, nil
	}

	// 检测工厂型循环依赖：仅当同一 goroutine 重入时才判定为循环依赖。
	// 并发访问（不同 goroutine）应继续阻塞在下面的 per-beanID 锁上，
	// 等待创建完成后命中缓存，而非直接报错。
	if creating, ok := c.beanCreating.Load(beanID); ok {
		if id := currentGoroutineID(); id != "" && creating.(string) == id {
			return nil, fmt.Errorf("%w: bean %q is being created", ErrCircularDependency, beanID)
		}
	}

	// 每个 beanID 一把锁，保证 创建+初始化+缓存 整体只执行一次，
	// 避免并发 Get 时 Init 被多次调用、产生重复销毁记录
	lock, _ := c.beanCreationLocks.LoadOrStore(beanID, &sync.Mutex{})
	mu, ok := lock.(*sync.Mutex)
	if !ok {
		mu = &sync.Mutex{}
		c.beanCreationLocks.Store(beanID, mu)
	}
	mu.Lock()
	defer mu.Unlock()

	// 获取锁后再次检查缓存
	if instance, ok := c.reg.GetInstance(beanID); ok {
		return instance, nil
	}

	c.beanCreating.Store(beanID, currentGoroutineID())
	defer c.beanCreating.Delete(beanID)

	// 根据作用域获取实例
	scopeImpl := c.scopeRegistry.Get(string(def.Scope))
	if scopeImpl == nil {
		scopeImpl = c.scopeRegistry.Get(scope.SingletonScope)
	}

	instance, err := scopeImpl.Get(beanID, def.Factory)
	if err != nil {
		return nil, err
	}

	// 调用初始化回调（支持函数式回调和 LifecycleBean 接口）
	if err := c.lifecycleMgr.InvokeInit(beanID, instance, def.Init); err != nil {
		return nil, err
	}

	// 缓存 Singleton 实例（空作用域默认为 Singleton）
	if def.Scope == registry.Singleton || def.Scope == "" {
		c.reg.SetInstance(beanID, instance)
	}

	// 仅 Singleton Bean 注册销毁回调，避免原型 Bean 内存泄漏
	if def.Scope == registry.Singleton || def.Scope == "" {
		// 检查是否实现了 LifecycleBean 接口或提供了销毁回调
		_, hasLifecycleBean := instance.(lifecycle.LifecycleBean)
		if hasLifecycleBean || def.Destroy != nil {
			c.lifecycleMgr.RegisterBean(beanID, instance, def.Destroy)
		}
	}

	return instance, nil
}

// currentGoroutineID 返回当前 goroutine 的 ID 字符串。
//
// Go 标准库未公开 goroutine ID，这里通过解析 runtime.Stack 输出的首行获取。
// 仅在 Bean 创建发生冲突（in-progress 命中）的慢速路径调用，热路径不受影响。
func currentGoroutineID() string {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	fields := strings.Fields(string(buf[:n]))
	if len(fields) >= 2 {
		return fields[1]
	}
	return ""
}

// NewContainer 创建一个新的 IoC 容器实例。
func NewContainer() Container {
	return &defaultContainer{
		reg:           registry.NewBeanRegistry(),
		scopeRegistry: scope.NewScopeRegistry(),
		lifecycleMgr:  lifecycle.NewLifecycleManager(),
	}
}
