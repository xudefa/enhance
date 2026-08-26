// Package core 提供了一个类型安全的依赖注入（DI）容器实现，灵感来自 Spring Framework 的 IoC 容器。
//
// # 设计哲学
//
// 本框架遵循 Go 语言哲学，优先使用泛型和函数式 API，最小化反射使用。
// 与 Spring 的注解驱动不同，我们采用显式注册 + 泛型包装的方式，
// 在编译期确保类型安全，避免运行时类型断言错误。
//
// # 设计原则
//
//   - 类型安全：通过泛型 API 在编译期检查类型，避免运行时错误
//   - 零反射注册：用户 API 层完全避免反射，仅在容器内部使用 reflect.Type 存储类型信息
//   - 函数式依赖：通过工厂函数 func(Container) T 显式声明依赖关系
//   - 生命周期管理：支持 Init/Destroy 回调和阶段监听器
//   - 并发安全：使用 sync.Map 优化读多写少的注册场景
//   - 零外部依赖：核心实现仅使用 Go 标准库
//
// # 核心特性
//
//   - 编译期类型安全：通过泛型函数 Register[T]/Get[T] 在编译期检查 Bean 类型
//   - 零反射注册：用户 API 层完全避免反射，仅在容器内部使用 reflect.Type 存储类型信息
//   - 函数式依赖：通过工厂函数 func(Container) T 显式声明依赖关系
//   - 生命周期管理：支持 Init/Destroy 回调和阶段监听器
//   - 作用域支持：Singleton（单例）和 Prototype（原型）两种作用域
//   - 并发安全：使用 sync.Map 优化读多写少的注册场景
//   - 父子容器：支持容器层级关系，子容器可访问父容器中的 Bean
//
// # 架构概览
//
//	core/                          # 核心包
//	├── doc.go                     # 核心接口定义（BeanDef, Container, BeanOption）
//	├── container.go               # Scope 类型定义
//	├── container_impl.go          # 默认容器实现
//	├── generic_api.go             # 泛型 API 包装函数
//	├── errors.go                  # 错误定义
//	│
//	├── scope/                     # 作用域管理
//	│   ├── doc.go                 # Scope, ScopeRegistry 接口
//	│   └── scope_impl.go          # Singleton/Prototype 实现（sync.Map）
//	│
//	├── lifecycle/                 # 生命周期管理
//	│   ├── doc.go                 # LifecycleManager 接口
//	│   └── lifecycle_impl.go      # 生命周期管理器实现
//	│
//	├── binding/                   # 数据绑定
//	│   ├── doc.go                 # Binder, ValueResolver, TypeConverter 接口 + Inject[T]
//	│   ├── binding_impl.go        # 字段注入和配置绑定实现
//	│   └── inject_impl.go         # 泛型注入实现
//	│
//	└── registry/                  # Bean 注册表（内部包）
//	    ├── doc.go                 # BeanRegistry, BeanIDGenerator 接口
//	    └── registry_impl.go       # 注册表实现（sync.Map）
//
// # 快速开始
//
//	// 1. 创建容器
//	container := core.NewContainer()
//
//	// 2. 注册 Bean
//	core.Register(container, "db", func(c core.Container) *Database {
//	    return &Database{DSN: "localhost:3306"}
//	})
//
//	// 3. 注册带依赖的 Bean
//	core.Register(container, "userService", func(c core.Container) *UserService {
//	    db := core.MustGet[*Database](c, "db")
//	    return &UserService{DB: db}
//	})
//
//	// 4. 初始化容器（创建所有非延迟初始化的 Singleton Bean）
//	container.Initialize()
//
//	// 5. 获取 Bean
//	svc := core.MustGet[*UserService](container, "userService")
//
// # 作用域配置
//
//	// 单例作用域（默认）
//	core.Register(container, "cache", func(c core.Container) *Cache {
//	    return NewCache()
//	})
//
//	// 原型作用域（每次获取创建新实例）
//	core.Register(container, "request", func(c core.Container) *Request {
//	    return &Request{}
//	}, core.WithScope[*Request](core.Prototype))
//
//	// 延迟初始化
//	core.Register(container, "expensive", func(c core.Container) *Expensive {
//	    return &Expensive{}
//	}, core.WithLazy[*Expensive](true))
//
// # 生命周期回调
//
//	core.Register(container, "service", func(c core.Container) *Service {
//	    return &Service{}
//	}, core.WithInit(func(s *Service) error {
//	    return s.Start()
//	}), core.WithDestroy(func(s *Service) error {
//	    return s.Stop()
//	}))
//
// # 数据绑定
//
//	// 字段注入
//	type MyBean struct {
//	    DB *Database `inject:"db"`
//	}
//
//	binder := binding.NewBinder()
//	bean := &MyBean{}
//	binder.BindFields(bean, container)
//
//	// 配置值注入
//	type Config struct {
//	    Timeout int `value:"app.timeout"`
//	}
//
//	resolver := binding.ValueResolverFunc(func(key string) (string, bool) {
//	    return "30", true
//	})
//	binder.BindValue(&Config{}, resolver)
//
// # 设计说明
//
// 由于 Go 语言限制（接口方法不能有类型参数），Container 接口使用 reflect.Type 存储类型信息，
// 但通过泛型包装函数（Register[T], Get[T] 等）提供编译期类型安全的 API。
//
// 用户应始终使用泛型 API，避免直接调用 Container 接口的 reflect.Type 方法。
package core

import (
	"reflect"

	"github.com/xudefa/enhance/core/registry"
)

type BeanGet interface {
	// Get 获取指定类型的 Bean 实例列表。
	//
	// 参数:
	//   - typ: Bean 类型，用于类型检查
	//
	// 返回:
	//   - []any: Bean 实例列表
	//   - error: 错误信息
	Get(typ reflect.Type) ([]any, error)

	// GetByTypeAndName 获取指定名称的 Bean 实例。
	//
	// 参数:
	//   - name: Bean 名称，可以为空字符串
	//   - typ: Bean 类型，用于类型检查
	//
	// 返回:
	//   - any: Bean 实例
	//   - error: 错误信息
	GetByTypeAndName(name string, typ reflect.Type) (any, error)

	// GetAll 获取所有 Bean 实例列表。
	//
	// 参数:
	//   - typ: Bean 类型，用于类型检查
	//
	// 返回:
	//   - []any: Bean 实例列表
	GetAll() []any

	// Has 检查容器中是否存在指定类型和名称组合的 Bean。
	//
	// 参数:
	//   - name: Bean 名称，可以为空字符串
	//   - typ: Bean 类型，用于类型检查
	//
	// 返回:
	//   - bool: 是否存在
	Has(name string, typ reflect.Type) bool

	// HasType 检查容器中是否存在指定类型 Bean。
	//
	// 参数:
	//   - typ: Bean 类型，用于类型检查
	//
	// 返回:
	//   - bool: 是否存在
	HasType(typ reflect.Type) bool

	// Types 返回容器中所有已注册的 Bean 类型列表。
	Types() []reflect.Type

	// ListBeans 列出所有已注册的Bean信息
	//
	// 返回:
	//   - map[string]*registry.BeanDef: Bean 定义映射
	ListBeans() map[string]*registry.BeanDef
}

// BeanRegister 容器注册接口。
//
// Bean ID 格式：包路径.类型名#自定义名称
//   - 包路径：Bean 类型所属的包路径
//   - 类型名：Bean 类型的名称
//   - 自定义名称：Bean 实例的自定义名称，为空时自动生成
type BeanRegister interface {
	// RegisterBean 注册一个 Bean。
	//
	// 参数:
	//   - def: Bean 定义，包含类型信息和工厂函数
	//
	// 返回:
	//   - error: 错误信息
	RegisterBean(def registry.BeanDef) error

	// RegisterInstance 注册一个已存在的 Bean 实例。
	//
	// 参数:
	//   - instance: Bean 实例
	//   - typ: Bean 类型，用于类型检查
	//
	//
	// 返回:
	//   - error: 错误信息
	RegisterInstance(instance any, typ reflect.Type) error
}

// BeanIDGenerator Bean ID 生成器。
//
// 负责根据类型和名称生成标准格式的 Bean ID。
type BeanIDGenerator interface {
	// Generate 生成 Bean ID。
	//
	// 参数:
	//   - typ: Bean 类型，用于类型检查
	//   - customName: 自定义名称（可选）
	//
	// 返回:
	//   - string: Bean ID
	Generate(typ reflect.Type, customName ...string) string

	// Parse 解析 Bean ID。
	//
	// 参数:
	//   - beanID: Bean ID
	//
	// 返回:
	//   - pkgPath: 包路径
	//   - typeName: 类型名称
	//   - customName: 自定义名称
	Parse(beanID string) (pkgPath, typeName, customName string)
}

// Container IoC 容器接口。
//
// 设计说明：
//   - 接口方法不使用泛型（Go 语法限制）
//   - 通过泛型包装函数实现编译期类型安全
//   - 内部使用 reflect.Type 存储类型信息
type Container interface {
	BeanGet
	BeanRegister
	BeanIDGenerator
	BeanCreator

	// Initialize 初始化容器，创建所有 Singleton Bean 并调用 Init 回调。
	Initialize() error

	// Destroy 销毁容器，调用所有 Singleton Bean 的 Destroy 回调并清理资源。
	Destroy() error
}

// ContainerExt 容器扩展接口，提供高级功能如子容器、条件注册等。
type ContainerExt interface {
	// SetParent 设置父容器，子容器可以获取父容器中的 Bean。
	SetParent(parent Container)

	// GetParent 获取父容器，如果没有父容器则返回 nil。
	GetParent() Container

	// Types 返回容器中所有已注册的 Bean 类型列表。
	Types() []reflect.Type

	// BeanCount 返回容器中已注册的 Bean 数量。
	BeanCount() int

	// BeanCountType 返回容器中指定类型 Bean 数量。
	// 参数:
	//   - typ: Bean 类型，用于类型检查
	// 返回:
	//   - int: Bean 数量
	BeanCountType(typ reflect.Type) int

	// Validate 验证所有已注册Bean的依赖是否可解析，检测循环依赖。
	// 应在 Initialize() 之前调用，提前发现配置错误。
	//
	// 返回:
	//   - error: 验证失败时返回错误，验证通过返回 nil
	Validate() error
}

// BeanOption Bean 注册选项函数类型。
//
// 用于函数式配置 Bean 的生命周期等属性。
//
// 可用的选项函数：
//   - WithInit: 设置初始化回调
//   - WithDestroy: 设置销毁回调
//   - WithLazy: 设置延迟初始化
//   - WithScope: 设置作用域
type BeanOption func(*registry.BeanDef)

// BeanCreator Bean 创建器接口。
//
// 用于在运行时动态创建 Bean 实例，通常用于刷新作用域、原型作用域等场景，必须在容器中已经存在对应的 Bean 定义。
type BeanCreator interface {
	// CreateBean 创建指定 ID 的 Bean 实例。
	//
	// 参数:
	//   - beanID: Bean ID
	//
	// 返回:
	//   - any: Bean 实例
	//   - error: 错误信息
	CreateBean(beanID string) (any, error)
}
