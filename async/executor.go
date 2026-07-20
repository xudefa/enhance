// Package async 提供异步执行器功能，用于 enhance 框架。
package async

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AsyncExecutor 异步执行器。
//
// 基于 goroutine 池实现异步任务执行，支持 Future 模式返回值。
type AsyncExecutor struct {
	workerCount int
	queueSize   int
	taskQueue   chan asyncTask
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	running     bool
	started     bool // 标记 worker 是否已启动
}

// asyncTask 异步任务内部结构。
type asyncTask struct {
	fn     func() (any, error)
	future *Future
}

// NewFuture 创建 Future 实例
func NewFuture() *Future {
	return &Future{
		done: make(chan struct{}),
	}
}

// Get 阻塞获取结果
func (f *Future) Get() (any, error) {
	<-f.done
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.result, f.err
}

// GetWithContext 带上下文的阻塞获取
func (f *Future) GetWithContext(ctx context.Context) (any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-f.done:
		f.mu.RLock()
		defer f.mu.RUnlock()
		return f.result, f.err
	}
}

// GetWithTimeout 带超时的阻塞获取
func (f *Future) GetWithTimeout(timeout time.Duration) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return f.GetWithContext(ctx)
}

// IsDone 检查任务是否完成
func (f *Future) IsDone() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

// setResult 设置结果（内部使用）
func (f *Future) setResult(result any, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.result = result
	f.err = err
	close(f.done)
}

// NewAsyncExecutor 创建异步执行器
//
// 参数:
//   - workerCount: 工作线程数
//   - queueSize: 任务队列容量
func NewAsyncExecutor(workerCount, queueSize int) *AsyncExecutor {
	ctx, cancel := context.WithCancel(context.Background())

	executor := &AsyncExecutor{
		workerCount: workerCount,
		queueSize:   queueSize,
		taskQueue:   make(chan asyncTask, queueSize),
		ctx:         ctx,
		cancel:      cancel,
		running:     true,  // 默认允许提交（懒启动模式）
		started:     false, // worker 未启动
	}

	return executor
}

// Start 启动执行器
func (e *AsyncExecutor) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return
	}

	e.running = true
	e.started = true

	// 启动核心工作线程
	for range e.workerCount {
		e.wg.Add(1)
		go e.worker()
	}
}

// worker 工作线程
func (e *AsyncExecutor) worker() {
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			// 完成队列中剩余的任务
			for {
				select {
				case task, ok := <-e.taskQueue:
					if !ok {
						return
					}
					result, err := task.fn()
					if task.future != nil {
						task.future.setResult(result, err)
					}
				default:
					return
				}
			}
		case task, ok := <-e.taskQueue:
			if !ok {
				return
			}

			// 执行任务
			result, err := task.fn()
			if task.future != nil {
				task.future.setResult(result, err)
			}
		}
	}
}

// Submit 提交异步任务
func (e *AsyncExecutor) Submit(fn func() (any, error)) *Future {
	future := NewFuture()

	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		// 执行器已关闭，直接返回错误
		future.setResult(nil, fmt.Errorf("executor is shutdown"))
		return future
	}

	// 懒启动：首次提交时启动 worker
	if !e.started {
		e.started = true
		for range e.workerCount {
			e.wg.Add(1)
			go e.worker()
		}
	}

	task := asyncTask{
		fn:     fn,
		future: future,
	}
	e.mu.Unlock()

	// 使用 defer recover 防止向已关闭的 channel 发送导致 panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				// channel 已关闭，设置错误
				future.setResult(nil, fmt.Errorf("executor is shutdown"))
			}
		}()
		e.taskQueue <- task
	}()

	return future
}

// SubmitVoid 提交无返回值的异步任务
func (e *AsyncExecutor) SubmitVoid(fn func() error) *Future {
	return e.Submit(func() (any, error) {
		return nil, fn()
	})
}

// GetQueueSize 获取当前队列中的任务数
func (e *AsyncExecutor) GetQueueSize() int {
	return len(e.taskQueue)
}

// Shutdown 优雅关闭执行器
func (e *AsyncExecutor) Shutdown() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.running = false
	e.mu.Unlock()

	e.cancel()
	close(e.taskQueue)
	e.wg.Wait()
}

// ShutdownWithTimeout 带超时的优雅关闭
func (e *AsyncExecutor) ShutdownWithTimeout(timeout time.Duration) error {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return nil
	}
	e.running = false
	e.mu.Unlock()

	e.cancel()
	close(e.taskQueue)

	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("关闭超时，等待时间 %v", timeout)
	}
}

// IsRunning 检查执行器是否运行
func (e *AsyncExecutor) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}
