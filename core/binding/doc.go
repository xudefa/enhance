// Package binding 提供了数据绑定和依赖注入的底层支持。
//
// # 设计原则
//
//   - 构造器注入优先：通过工厂函数参数显式声明依赖，编译期类型安全
//   - 字段注入补充：通过 inject 标签绑定，适用于简化代码
//   - 配置绑定：通过 value 标签绑定配置值，支持类型转换
//
// # 注入方式
//
// 构造器注入（推荐）：
//
//	core.Register(container, "userService", func(c core.Container) *UserService {
//	    db := core.MustGet[*Database](c, "db")  // 显式声明依赖
//	    return &UserService{DB: db}
//	})
//
// 字段注入：
//
//	type MyBean struct {
//	    DB *Database `inject:"db"`
//	}
//
//	binder := binding.NewBinder()
//	binder.BindFields(&myBean, container)
//
// 配置绑定：
//
//	type Config struct {
//	    Timeout int    `value:"app.timeout"`
//	    Name    string `value:"app.name"`
//	}
//
//	binder.BindValue(&config, resolver)
//
// # 泛型注入 API
//
// 提供 Inject[T] 和 MustInject[T] 泛型函数，编译期类型安全：
//
//	svc := binding.MustInject[*UserService](container, "userService")
package binding

import (
	"github.com/xudefa/enhance/core"
)

// Binder 数据绑定器接口。
//
// 负责将依赖注入到 Bean 实例中。
type Binder interface {
	// BindFields 将容器中的 Bean 注入到目标对象的字段中。
	// 支持 inject 标签和 @Autowired 注解。
	//
	// 参数:
	//   - target: 目标对象（必须是指针）
	//   - container: IoC 容器
	//
	// 返回:
	//   - error: 错误信息
	BindFields(target any, container core.BeanGet) error

	// BindValue 将配置值绑定到目标对象的字段中。
	// 支持 value 标签和 @Value 注解。
	//
	// 参数:
	//   - target: 目标对象（必须是指针）
	//   - resolver: 配置值解析器
	//
	// 返回:
	//   - error: 错误信息
	BindValue(target any, resolver ValueResolver) error

	// BindAll 执行完整的绑定流程（字段注入 + 配置绑定）。
	//
	// 参数:
	//   - target: 目标对象（必须是指针）
	//   - container: IoC 容器
	//   - resolver: 配置值解析器
	//
	// 返回:
	//   - error: 错误信息
	BindAll(target any, container core.BeanGet, resolver ValueResolver) error
}

// ValueResolver 配置值解析器接口。
//
// 用于解析配置键对应的值。
type ValueResolver interface {
	// Resolve 解析配置键对应的值。
	//
	// 参数:
	//   - key: 配置键
	//
	// 返回:
	//   - string: 配置值
	//   - bool: 是否存在
	Resolve(key string) (string, bool)
}

// ValueResolverFunc 配置值解析器函数类型。
//
// 实现 ValueResolver 接口，方便使用函数式解析器。
type ValueResolverFunc func(key string) (string, bool)

// TypeConverter 类型转换器接口。
//
// 用于将字符串配置值转换为目标类型。
type TypeConverter interface {
	// Convert 将字符串值转换为目标类型。
	//
	// 参数:
	//   - value: 字符串值
	//   - targetType: 目标类型名称
	//
	// 返回:
	//   - any: 转换后的值
	//   - error: 转换错误
	Convert(value string, targetType string) (any, error)
}
