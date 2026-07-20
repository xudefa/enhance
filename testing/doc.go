// Package testing 提供测试工具支持，用于 enhance 框架。
//
// 该模块提供测试辅助函数、模拟对象创建、测试断言等测试相关功能，简化单元测试和集成测试的编写。
//
// # 架构设计
//
//   - TestingT: 测试接口，兼容 testing.T
//   - TestRunner: 测试运行器接口，用于运行测试
//   - TestContext: 测试上下文接口，提供测试环境
//   - Mock: Mock 对象接口，用于设置期望和验证调用
//
// # 核心功能
//
//   - 断言函数: 提供 Equal、NotNil、True、False 等常用断言
//   - 模拟对象: 支持创建模拟依赖对象
//   - 测试辅助: 提供临时文件创建、随机数据生成等辅助函数
//   - 测试超时: 支持设置测试超时时间
//
// # 使用方式
//
// 使用断言：
//
//	func TestUserService(t *testing.T) {
//	    user := service.GetUser(1)
//	    testing.AssertNotNil(t, user, "user should not be nil")
//	    testing.AssertEqual(t, "John", user.Name, "name should match")
//	}
//
// 使用测试辅助：
//
//	func TestFileProcessing(t *testing.T) {
//	    tmpFile := testing.CreateTempFile(t, "test content")
//	    defer tmpFile.Close()
//	    // 测试文件处理逻辑
//	}
//
// # 断言函数
//
//   - AssertEqual: 断言两个值相等
//   - AssertNotEqual: 断言两个值不相等
//   - AssertNil: 断言值为 nil
//   - AssertNotNil: 断言值不为 nil
//   - AssertTrue: 断言条件为 true
//   - AssertFalse: 断言条件为 false
//   - AssertContains: 断言字符串包含子串
//   - AssertPanics: 断言函数会 panic
package testing

import "reflect"

// TestingT 测试接口，兼容 testing.T。
type TestingT interface {
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Helper()
}

// TestContext 测试上下文接口。
//
// 提供测试环境，包括 IoC 容器、测试配置、清理函数等。
type TestContext interface {
	// T 获取底层测试对象。
	T() TestingT

	// GetByType 从容器按类型获取 Bean，如果获取失败则测试失败。
	GetByType(t reflect.Type) any

	// Register 向容器注册 Bean。
	Register(name string, bean any)

	// SetProperty 设置测试属性。
	SetProperty(key string, value any)

	// GetProperty 获取测试属性。
	GetProperty(key string) any

	// AddCleanup 添加测试清理函数。
	AddCleanup(fn func())

	// Cleanup 执行所有清理函数。
	Cleanup()

	// Close 关闭测试上下文。
	Close()

	// Container 获取 IoC 容器。
	Container() any
}

// Mock 模拟对象接口。
//
// 用于设置方法调用期望和验证调用是否满足。
type Mock interface {
	// Expect 设置方法调用期望，默认期望调用 1 次。
	Expect(method string, args []any, result any, err error) Mock

	// ExpectTimes 设置方法调用期望，指定期望调用次数。
	ExpectTimes(method string, args []any, result any, err error, times int) Mock

	// Call 模拟方法调用，返回匹配的期望结果。
	Call(method string, args ...any) (any, error)

	// Verify 验证所有期望是否满足。
	Verify() error

	// Reset 重置 Mock 对象的所有状态。
	Reset()
}
