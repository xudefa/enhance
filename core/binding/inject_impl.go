package binding

import (
	"fmt"
	"reflect"

	"github.com/xudefa/enhance/core"
)

// InjectOption 注入选项函数类型。
type InjectOption func(*injectConfig)

// injectConfig 注入配置。
type injectConfig struct {
	// Required 是否为必需依赖。
	Required bool

	// SkipNotFound 是否跳过未找到的依赖。
	SkipNotFound bool
}

// WithRequired 设置依赖为必需。
func WithRequired() InjectOption {
	return func(cfg *injectConfig) {
		cfg.Required = true
	}
}

// WithOptional 设置依赖为可选。
func WithOptional() InjectOption {
	return func(cfg *injectConfig) {
		cfg.Required = false
	}
}

// Inject 泛型注入函数，从容器中获取指定类型的 Bean 并注入。
//
// 这是用户 API 层使用的函数，提供编译期类型安全。
// 具体实现在 binding_impl.go 中。
//
// 参数:
//   - container: IoC 容器
//   - beanName: Bean 名称（可选）
//
// 返回:
//   - T: Bean 实例
//   - error: 错误信息
func Inject[T any](container core.BeanGet, beanName string) (T, error) {
	return injectImpl[T](container, beanName)
}

// MustInject 泛型注入函数，失败时 panic。
//
// 参数:
//   - container: IoC 容器
//   - beanName: Bean 名称（可选）
//
// 返回:
//   - T: Bean 实例
func MustInject[T any](container core.BeanGet, beanName string) T {
	instance, err := injectImpl[T](container, beanName)
	if err != nil {
		panic(err)
	}
	return instance
}

// injectImpl 注入实现函数。
func injectImpl[T any](container core.BeanGet, beanName string) (T, error) {
	var zero T
	typ := reflect.TypeOf((*T)(nil)).Elem()

	if beanName != "" {
		instance, err := container.GetByTypeAndName(beanName, typ)
		if err != nil {
			return zero, err
		}
		if instance == nil {
			return zero, fmt.Errorf("no bean found with name '%s' and type %v", beanName, typ)
		}
		result, ok := instance.(T)
		if !ok {
			return zero, fmt.Errorf("bean %q type mismatch: got %T, want %v", beanName, instance, typ)
		}
		return result, nil
	}

	instances, err := container.Get(typ)
	if err != nil {
		return zero, err
	}

	if len(instances) == 0 {
		return zero, fmt.Errorf("no bean found for type %v", typ)
	}

	result, ok := instances[0].(T)
	if !ok {
		return zero, fmt.Errorf("bean type mismatch: got %T, want %v", instances[0], typ)
	}
	return result, nil
}
