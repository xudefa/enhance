// Package registry 提供了 Bean 注册表的内部实现。
//
// # 设计原则
//
//   - 内部包：此包仅供容器内部使用，不对外暴露
//   - 并发安全：所有操作使用 sync.Map 保证线程安全
//   - 高效查找：维护类型索引，优化 Bean 查找性能
//
// # Bean ID 格式
//
//	包路径.类型名#自定义名称
//
// 示例：
//   - github.com/example/app/services.UserService（无自定义名称）
//   - github.com/example/app/services.UserService#primary（有自定义名称）
//
// # 注册表结构
//
// 注册表维护以下映射关系：
//   - beanID → BeanDef：Bean 定义映射
//   - beanID → any：Singleton Bean 实例缓存
//   - reflect.Type → []beanID：类型到 Bean ID 列表的映射
//
// # 并发优化
//
// 使用 sync.Map 替代 sync.RWMutex，优化读多写少的注册场景。
package registry

import (
	"reflect"
)

// Scope 定义 Bean 的作用域类型。
type Scope string

const (
	// Singleton 单例作用域
	Singleton Scope = "singleton"
	// Prototype 原型作用域
	Prototype Scope = "prototype"
)

// BeanDef Bean 定义结构。
type BeanDef struct {
	// Name 同一个 Bean 的name唯一, 用于查找 Bean 实例, 可为空
	Name string

	// Type Bean 的类型
	Type reflect.Type

	// Factory 创建 Bean 实例的工厂函数
	Factory func(c ...any) (any, error)

	// Scope Bean 的作用域
	Scope Scope

	// Init 初始化回调
	Init func(bean any) error

	// Destroy 销毁回调
	Destroy func(bean any) error

	// Lazy 是否延迟初始化
	Lazy bool

	// Primary 是否为首选 Bean
	// 用于查找 Bean 实例时，优先返回首选 Bean
	// 仅在作用作用域为 Singleton 时有效
	Primary bool
}

// BeanRegistry Bean 注册表接口。
//
// 内部管理 Bean 定义和实例的注册表。
type BeanRegistry interface {
	// Register 注册 Bean 定义。
	//
	// 参数:
	//   - def: Bean 定义
	//
	// 返回:
	//   - error: 注册错误
	Register(def BeanDef, beanID string) error

	// RegisterInstance 注册一个已存在的 Bean 实例。
	//
	// 参数:
	//   - instance: Bean 实例
	//   - typ: Bean 类型，用于类型检查
	//
	// 返回:
	//   - error: 注册错误
	RegisterInstance(instance any, typ reflect.Type, beanID string) error

	// GetDefinition 获取 Bean 定义。
	//
	// 参数:
	//   - beanID: Bean ID
	//
	// 返回:
	//   - *BeanDef: Bean 定义
	//   - bool: 是否存在
	GetDefinition(beanID string) (*BeanDef, bool)

	// GetInstance 获取 Bean 实例（从缓存）。
	//
	// 参数:
	//   - beanID: Bean ID
	//
	// 返回:
	//   - any: Bean 实例
	//   - bool: 是否存在
	GetInstance(beanID string) (any, bool)

	// SetInstance 设置 Bean 实例（用于缓存）。
	//
	// 参数:
	//   - beanID: Bean ID
	//   - instance: Bean 实例
	SetInstance(beanID string, instance any)

	// GetByType 根据类型获取 Bean ID 列表。
	//
	// 参数:
	//   - typ: Bean 类型
	//
	// 返回:
	//   - []string: Bean ID 列表
	GetByType(typ reflect.Type) []string

	// GetPrimaryByType 根据类型获取首选 Bean ID。
	//
	// 参数:
	//   - typ: Bean 类型
	//
	// 返回:
	//   - string: Bean ID
	//   - bool: 是否存在
	GetPrimaryByType(typ reflect.Type) (string, bool)

	// HasBean 检查 Bean 是否存在。
	//
	// 参数:
	//   - beanID: Bean ID
	//
	// 返回:
	//   - bool: 是否存在
	HasBean(beanID string) bool

	// HasType 检查类型是否存在。
	//
	// 参数:
	//   - typ: Bean 类型
	//
	// 返回:
	//   - bool: 是否存在
	HasType(typ reflect.Type) bool

	// Count 返回已注册的 Bean 数量。
	Count() int

	// CountByType 返回指定类型的 Bean 数量。
	//
	// 参数:
	//   - typ: Bean 类型
	//
	// 返回:
	//   - int: Bean 数量
	CountByType(typ reflect.Type) int

	// Types 返回所有已注册的 Bean 类型。
	//
	// 返回:
	//   - []reflect.Type: Bean 类型列表
	Types() []reflect.Type

	// BeanIDs 返回所有已注册的 Bean ID。
	//
	// 返回:
	//   - []string: Bean ID 列表
	BeanIDs() []string

	// ListBeans 列出所有已注册的Bean信息
	//
	// 返回:
	//   - map[string]*BeanDef: Bean 定义映射
	ListBeans() map[string]*BeanDef

	// ListInstances 列出所有已注册的Bean实例
	//
	// 返回:
	//   - map[string]any: Bean 实例映射
	ListInstances() map[string]any

	// Clear 清空注册表。
	Clear()
}
