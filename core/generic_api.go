// Package core 提供泛型 API 包装函数，用于编译期类型安全的 Bean 注册和获取。
//
// 用户应优先使用本文件的泛型 API，而非直接调用 Container 接口的反射方法。
package core

import (
	"reflect"

	"github.com/xudefa/enhance/core/registry"
)

// Register 使用泛型注册 Bean 定义，提供编译期类型安全。
//
// 参数:
//   - container: IoC 容器
//   - opts: Bean 选项（工厂函数、作用域、生命周期回调等）
//
// 示例:
//
//	core.Register(container, func(c core.Container) *UserService {
//	    db := core.MustGet[*Database](c, "db")
//	    return &UserService{DB: db}
//	})
func Register[T any](container Container, opts ...BeanOption) error {
	def := registry.BeanDef{
		Type: reflect.TypeOf((*T)(nil)).Elem(),
	}
	for _, opt := range opts {
		opt(&def)
	}

	if def.Factory == nil {
		def.Factory = func(c ...any) (any, error) {
			return reflect.New(def.Type.Elem()).Interface(), nil
		}
	}

	if def.Scope == "" {
		def.Scope = registry.Singleton
	}

	return container.RegisterBean(def)
}

// GetByName 泛型获取函数，提供编译期类型安全的 Bean 获取。
//
// 参数:
//   - container: IoC 容器
//   - name: Bean 名称（可选）
//
// 返回:
//   - T: Bean 实例
//   - error: 错误信息
func GetByName[T any](container Container, name string) (T, error) {
	var zero T
	typ := reflect.TypeOf((*T)(nil)).Elem()

	if name != "" {
		instance, err := container.GetByTypeAndName(name, typ)
		if err != nil {
			return zero, err
		}
		return instance.(T), nil
	}

	instances, err := container.Get(typ)
	if err != nil {
		return zero, err
	}

	if len(instances) == 0 {
		return zero, ErrBeanNotFound
	}

	return instances[0].(T), nil
}

// MustGet 泛型获取函数，失败时 panic。
//
// 参数:
//   - container: IoC 容器
//   - name: Bean 名称（可选）
//
// 返回:
//   - T: Bean 实例
func MustGet[T any](container Container, name string) T {
	instance, err := GetByName[T](container, name)
	if err != nil {
		panic(err)
	}
	return instance
}

// Has 泛型检查函数，检查容器中是否存在指定类型的 Bean。
//
// 参数:
//   - container: IoC 容器
//   - name: Bean 名称（可选）
//
// 返回:
//   - bool: 是否存在
func Has[T any](container Container, name string) bool {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	return container.Has(name, typ)
}
