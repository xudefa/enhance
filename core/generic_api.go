// Package core 提供泛型 API 包装函数，用于编译期类型安全的 Bean 注册和获取。
//
// 用户应优先使用本文件的泛型 API，而非直接调用 Container 接口的反射方法。
package core

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/xudefa/enhance/core/registry"
)

// defaultFactories 缓存各类型对应的默认工厂函数。
// 相同类型始终返回同一个函数实例，用于重复注册的幂等性检测。
var defaultFactories sync.Map // reflect.Type -> func(c ...any) (any, error)

// getDefaultFactory 返回指定类型的默认工厂函数。
func getDefaultFactory(defType reflect.Type) func(c ...any) (any, error) {
	if f, ok := defaultFactories.Load(defType); ok {
		return f.(func(c ...any) (any, error))
	}
	f := func(c ...any) (any, error) {
		if defType.Kind() == reflect.Ptr {
			return reflect.New(defType.Elem()).Interface(), nil
		}
		return reflect.Zero(defType).Interface(), nil
	}
	actual, _ := defaultFactories.LoadOrStore(defType, f)
	return actual.(func(c ...any) (any, error))
}

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
		def.Factory = getDefaultFactory(def.Type)
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
		result, ok := instance.(T)
		if !ok {
			return zero, fmt.Errorf("bean %q type mismatch: got %T, want %v", name, instance, typ)
		}
		return result, nil
	}

	instances, err := container.Get(typ)
	if err != nil {
		return zero, err
	}

	if len(instances) == 0 {
		return zero, ErrBeanNotFound
	}

	result, ok := instances[0].(T)
	if !ok {
		return zero, fmt.Errorf("bean type mismatch: got %T, want %v", instances[0], typ)
	}
	return result, nil
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
