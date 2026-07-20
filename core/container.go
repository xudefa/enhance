// Package core 提供了一个类型安全的依赖注入（DI）容器实现，灵感来自 Spring Framework 的 IoC 容器。
//
// 详细文档请参阅 package core 的说明。
package core

import (
	"fmt"
	"reflect"
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
}

// RegisterBean 注册一个 Bean。
func (c *defaultContainer) RegisterBean(def registry.BeanDef) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if def.Type == nil {
		return fmt.Errorf("type is nil")
	}
	if c.initialized {
		return ErrContainerAlreadyInitialized
	}
	beanID := c.Generate(def.Type, def.Name)
	def.Name = beanID
	return c.reg.Register(def, beanID)
}

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

	// 释放读锁，避免在 resolveBean 中死锁
	c.mu.RUnlock()

	var instances []any
	for _, beanID := range beanIDs {
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
	if !c.reg.HasBean(name) {
		c.mu.RUnlock()
		// 尝试从父容器获取
		if c.parent != nil {
			return c.parent.GetByTypeAndName(name, typ)
		}
		return nil, ErrBeanNotFound
	}

	// 释放读锁，避免在 resolveBean 中死锁
	c.mu.RUnlock()

	// 通过 resolveBean 获取（支持懒加载）
	return c.resolveBean(name, typ)
}

func (c *defaultContainer) GetAll() []any {
	c.mu.RLock()

	if c.destroyed {
		c.mu.RUnlock()
		return []any{}
	}

	mapInstances := c.reg.ListInstances()
	c.mu.RUnlock()

	var instances []any
	for _, v := range mapInstances {
		instances = append(instances, v)
	}
	return instances
}

// Has 检查容器中是否存在指定类型和名称组合的 Bean。
func (c *defaultContainer) Has(name string, typ reflect.Type) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

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

func (c *defaultContainer) Generate(typ reflect.Type, customName ...string) string {
	// 处理指针类型，获取实际类型
	actualType := typ
	if typ.Kind() == reflect.Ptr {
		actualType = typ.Elem()
	}

	// idGenerator ID 格式：pkgPath.typeName，id 生成器一定存在，不需要判断空指针
	prefix := actualType.PkgPath() + "." + actualType.Name()

	// 没有提供自定义名称
	if len(customName) == 0 || customName[0] == "" {
		return prefix
	}

	// 如果已经提供了标准格式的 Name，直接使用
	if strings.HasPrefix(customName[0], prefix) {
		return customName[0]
	}

	// 如果没有提供标准格式的 Name，使用自定义名称，格式为：pkgPath.typeName#customName
	return prefix + "#" + customName[0]
}

func (c *defaultContainer) Parse(beanID string) (pkgPath, typeName, customName string) {
	// 解析自定义名称
	parts := strings.SplitN(beanID, "#", 2)
	mainPart := parts[0]
	if len(parts) > 1 {
		customName = parts[1]
	}

	// 解析包路径和类型名
	lastDot := strings.LastIndex(mainPart, ".")
	if lastDot == -1 {
		return "", mainPart, customName
	}

	pkgPath = mainPart[:lastDot]
	typeName = mainPart[lastDot+1:]
	return pkgPath, typeName, customName
}

// Initialize 初始化容器，创建所有 Singleton Bean 并调用 Init 回调。
func (c *defaultContainer) Initialize() error {
	c.mu.Lock()
	if c.initialized || c.destroyed {
		c.mu.Unlock()
		return ErrContainerAlreadyInitialized
	}
	c.initialized = true
	c.mu.Unlock()

	// 获取所有 Bean ID
	beanIDs := c.reg.BeanIDs()

	// 预创建所有非延迟初始化的 Singleton Bean
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
			return err
		}
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
	c.parent = parent
}

// GetParent 获取父容器，如果没有父容器则返回 nil。
func (c *defaultContainer) GetParent() Container {
	return c.parent
}

// Types 返回容器中所有已注册的 Bean 类型列表。
func (c *defaultContainer) Types() []reflect.Type {
	return c.reg.Types()
}

// BeanCount 返回容器中已注册的 Bean 数量。
func (c *defaultContainer) BeanCount() int {
	return c.reg.Count()
}

// BeanCountType 返回容器中指定类型 Bean 数量。
func (c *defaultContainer) BeanCountType(typ reflect.Type) int {
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

// resolveBean 解析 Bean 实例（内部使用）。
func (c *defaultContainer) resolveBean(beanID string, _ reflect.Type) (any, error) {
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
	// 双重检查缓存
	if instance, ok := c.reg.GetInstance(beanID); ok {
		return instance, nil
	}

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

// Validate 验证所有已注册Bean的依赖是否可解析，检测循环依赖。
func (c *defaultContainer) Validate() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.initialized {
		return fmt.Errorf("container already initialized, cannot validate")
	}

	// 1. 检查所有Bean的依赖是否已注册
	if err := c.validateDependencies(); err != nil {
		return err
	}

	// 2. 检测循环依赖
	if err := c.detectCircularDependencies(); err != nil {
		return err
	}

	return nil
}

// validateDependencies 检查所有Bean的依赖是否已注册
func (c *defaultContainer) validateDependencies() error {
	types := c.Types()
	typeSet := make(map[reflect.Type]bool)
	for _, typ := range types {
		typeSet[typ] = true
	}

	for _, typ := range types {
		if err := c.validateTypeDependencies(typ, typeSet); err != nil {
			return err
		}
	}

	return nil
}

// validateTypeDependencies 检查指定类型的依赖
func (c *defaultContainer) validateTypeDependencies(typ reflect.Type, typeSet map[reflect.Type]bool) error {
	actualType := typ
	if typ.Kind() == reflect.Ptr {
		actualType = typ.Elem()
	}

	if actualType.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < actualType.NumField(); i++ {
		field := actualType.Field(i)
		if !field.IsExported() {
			continue
		}

		// 检查是否有 inject 标签（包括空标签）
		if _, ok := field.Tag.Lookup("inject"); !ok {
			continue
		}

		fieldType := field.Type
		if typeSet[fieldType] {
			continue
		}

		if c.parent == nil {
			return fmt.Errorf("dependency not found: field '%s' of type '%s' requires '%s'",
				field.Name, typ.String(), fieldType.String())
		}

		ext, ok := c.parent.(ContainerExt)
		if !ok {
			return fmt.Errorf("dependency not found: field '%s' of type '%s' requires '%s'",
				field.Name, typ.String(), fieldType.String())
		}

		parentTypes := ext.Types()
		found := false
		for _, pt := range parentTypes {
			if pt == fieldType {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("dependency not found: field '%s' of type '%s' requires '%s'",
				field.Name, typ.String(), fieldType.String())
		}
	}

	return nil
}

// detectCircularDependencies 检测循环依赖
func (c *defaultContainer) detectCircularDependencies() error {
	types := c.Types()
	visited := make(map[reflect.Type]bool)
	recStack := make(map[reflect.Type]bool)

	for _, typ := range types {
		if !visited[typ] {
			if err := c.detectCircularDFS(typ, visited, recStack, []string{}); err != nil {
				return err
			}
		}
	}

	return nil
}

// detectCircularDFS 深度优先搜索检测循环依赖
func (c *defaultContainer) detectCircularDFS(typ reflect.Type, visited, recStack map[reflect.Type]bool, path []string) error {
	visited[typ] = true
	recStack[typ] = true
	path = append(path, typ.String())

	actualType := typ
	if typ.Kind() == reflect.Ptr {
		actualType = typ.Elem()
	}

	if actualType.Kind() == reflect.Struct {
		for i := 0; i < actualType.NumField(); i++ {
			field := actualType.Field(i)
			if !field.IsExported() {
				continue
			}

			if _, ok := field.Tag.Lookup("inject"); ok {
				fieldType := field.Type

				if recStack[fieldType] {
					path = append(path, fieldType.String())
					return fmt.Errorf("circular dependency detected: %s", strings.Join(path, " -> "))
				}

				if !visited[fieldType] {
					if err := c.detectCircularDFS(fieldType, visited, recStack, path); err != nil {
						return err
					}
				}
			}
		}
	}

	recStack[typ] = false
	return nil
}

// NewContainer 创建一个新的 IoC 容器实例。
func NewContainer() Container {
	return &defaultContainer{
		reg:           registry.NewBeanRegistry(),
		scopeRegistry: scope.NewScopeRegistry(),
		lifecycleMgr:  lifecycle.NewLifecycleManager(),
	}
}
