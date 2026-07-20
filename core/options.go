package core

import (
	"reflect"

	"github.com/xudefa/enhance/core/registry"
)

// WithName 设置 Bean 名称的选项函数。
func WithName[T any](name string) BeanOption {
	return func(def *registry.BeanDef) {
		def.Name = name
	}
}

// WithType 设置 Bean 类型的选项函数。
func WithType[T any](typ reflect.Type) BeanOption {
	return func(def *registry.BeanDef) {
		def.Type = typ
	}
}

// WithFactory 设置创建 Bean 实例的工厂函数的选项函数。
func WithFactory[T any](factory func(c ...any) (any, error)) BeanOption {
	return func(def *registry.BeanDef) {
		def.Factory = factory
	}
}

// WithScope 设置作用域的选项函数。
func WithScope[T any](scope registry.Scope) BeanOption {
	return func(def *registry.BeanDef) {
		def.Scope = scope
	}
}

// WithInit 设置初始化回调的选项函数。
func WithInit[T any](init func(bean any) error) BeanOption {
	return func(def *registry.BeanDef) {
		def.Init = init
	}
}

// WithDestroy 设置销毁回调的选项函数。
func WithDestroy[T any](destroy func(bean any) error) BeanOption {
	return func(def *registry.BeanDef) {
		def.Destroy = destroy
	}
}

// WithLazy 设置延迟初始化的选项函数。
func WithLazy[T any](lazy bool) BeanOption {
	return func(def *registry.BeanDef) {
		def.Lazy = lazy
	}
}

// WithPrimary 设置首选 Bean 的选项函数。
func WithPrimary[T any](primary bool) BeanOption {
	return func(def *registry.BeanDef) {
		def.Primary = primary
	}
}
