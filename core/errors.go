// Package core 错误定义。
//
// 错误分类：
//   - 容器状态错误：已初始化、已销毁
//   - Bean 查找错误：未找到、已存在
//   - 依赖注入错误：依赖未找到、注入失败
//   - 生命周期错误：初始化失败、销毁失败
package core

import "errors"

// 容器状态错误。
var (
	// ErrContainerAlreadyInitialized 容器已初始化，不能再注册新 Bean。
	ErrContainerAlreadyInitialized = errors.New("container already initialized")

	// ErrContainerDestroyed 容器已销毁，不能再使用。
	ErrContainerDestroyed = errors.New("container has been destroyed")
)

// Bean 查找错误。
var (
	// ErrBeanNotFound Bean 未找到。
	ErrBeanNotFound = errors.New("bean not found")

	// ErrBeanAlreadyExists Bean 已存在。
	ErrBeanAlreadyExists = errors.New("bean already exists")

	// ErrInvalidBeanName Bean 名称无效。
	ErrInvalidBeanName = errors.New("invalid bean name")
)

// 依赖注入错误。
var (
	// ErrCircularDependency 循环依赖检测。
	ErrCircularDependency = errors.New("circular dependency detected")

	// ErrDependencyNotFound 依赖的 Bean 未找到。
	ErrDependencyNotFound = errors.New("dependency bean not found")

	// ErrInjectFailed 依赖注入失败。
	ErrInjectFailed = errors.New("failed to inject dependencies")

	// ErrNilFactory 工厂函数不能为 nil。
	ErrNilFactory = errors.New("factory function cannot be nil")
)

// 生命周期错误。
var (
	// ErrInitFailed Bean 初始化失败。
	ErrInitFailed = errors.New("bean initialization failed")

	// ErrDestroyFailed Bean 销毁失败。
	ErrDestroyFailed = errors.New("bean destruction failed")
)
