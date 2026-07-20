// Package async 提供异步执行器功能，用于 enhance 框架。
//
// 该模块提供异步任务执行、线程池管理、异步结果获取等异步编程支持。
// 适用于需要并发执行任务的场景。
//
// # 架构设计
//
//   - AsyncExecutor: 异步任务执行器，基于 goroutine 池实现
//   - Future: 异步结果，支持阻塞获取、超时获取、上下文控制
//   - ExecutorOption: 执行器配置选项函数
//   - RejectHandler: 任务拒绝处理函数
//
// # 核心功能
//
//   - 异步任务执行: 基于 goroutine 池的异步任务执行
//   - Future 模式: 支持异步结果获取
//   - 超时控制: 支持带超时的结果获取
//   - 上下文控制: 支持通过上下文取消任务
//   - 优雅关闭: 支持执行器的优雅关闭
//
// # 使用方式
//
//	// 创建执行器
//	executor := async.NewAsyncExecutor(10, 100)
//	executor.Start()
//
//	// 提交异步任务
//	future := executor.Submit(func() (any, error) {
//	    // 执行耗时操作
//	    return result, nil
//	})
//
//	// 获取结果
//	result, err := future.Get()
//
//	// 带超时获取
//	result, err := future.GetWithTimeout(5 * time.Second)
//
//	// 优雅关闭
//	executor.Shutdown()
//
// # 配置选项
//
//   - WithPoolSize: 设置线程池大小
//   - WithQueueSize: 设置任务队列大小
//   - WithRejectHandler: 设置任务拒绝策略
package async

import (
	"sync"
)

// Future 异步任务结果。
//
// 提供阻塞获取结果、超时获取、检查是否完成等功能。
type Future struct {
	done   chan struct{}
	result any
	err    error
	mu     sync.RWMutex
}

// ExecutorOption 执行器配置选项函数。
type ExecutorOption func(*AsyncExecutor)

// RejectHandler 任务拒绝处理函数。
//
// 当任务队列满时调用，用于处理任务被拒绝的情况。
type RejectHandler func(task func() (any, error))
