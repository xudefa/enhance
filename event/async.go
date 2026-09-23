package event

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// AsyncPublisher 异步事件发布器
//
// 提供异步事件发布功能，支持上下文超时控制和错误处理。
// 使用工作协程池处理事件，避免阻塞发布者。
//
// 使用示例：
//
//	bus := event.NewEventBusWithOrdering()
//	publisher := event.NewAsyncPublisher(bus,
//	    event.WithWorkerCount(5),
//	    event.WithWorkerQueueSize(100),
//	    event.WithErrorHandler(func(err error, e event.ApplicationEvent) {
//	        log.Printf("event error: %v", err)
//	    }),
//	)
//	defer publisher.Close()
//
//	ctx := context.Background()
//	publisher.Publish(ctx, &event.BaseEvent{EventType: "MyEvent"})
type AsyncPublisher struct {
	bus         AsyncPublisherBus
	worker      chan func()
	done        chan struct{}
	closed      atomic.Bool
	wg          sync.WaitGroup
	taskWg      sync.WaitGroup
	closeMu     sync.Mutex
	errHandler  func(error, ApplicationEvent)
	workerCount int // 工作协程数量
	queueSize   int // 工作队列缓冲大小
}

// AsyncPublisherOption 异步发布器选项函数
type AsyncPublisherOption func(*AsyncPublisher)

// WithWorkerCount 设置工作协程池大小
//
// 参数:
//   - n: 工作协程数量
//
// 返回:
//   - AsyncPublisherOption: 选项函数
func WithWorkerCount(n int) AsyncPublisherOption {
	return func(p *AsyncPublisher) {
		if n > 0 {
			p.workerCount = n
		}
	}
}

// WithWorkerQueueSize 设置工作队列缓冲大小
//
// 独立于 workerCount 配置，允许设置更大的队列缓冲以应对瞬时高峰。
//
// 参数:
//   - n: 队列缓冲大小
//
// 返回:
//   - AsyncPublisherOption: 选项函数
func WithWorkerQueueSize(n int) AsyncPublisherOption {
	return func(p *AsyncPublisher) {
		if n > 0 {
			p.queueSize = n
		}
	}
}

// WithErrorHandler 设置错误处理器
//
// 参数:
//   - handler: 错误处理函数，接收错误和事件作为参数
//
// 返回:
//   - AsyncPublisherOption: 选项函数
func WithErrorHandler(handler func(error, ApplicationEvent)) AsyncPublisherOption {
	return func(p *AsyncPublisher) {
		p.errHandler = handler
	}
}

// NewAsyncPublisher 创建异步事件发布器
//
// 参数:
//   - bus: 事件发布器接口（支持 EventBus、EventBusWithOrdering 等）
//   - opts: 可选配置项
//
// 返回:
//   - *AsyncPublisher: 异步发布器实例
func NewAsyncPublisher(bus AsyncPublisherBus, opts ...AsyncPublisherOption) *AsyncPublisher {
	publisher := &AsyncPublisher{
		bus:         bus,
		done:        make(chan struct{}),
		workerCount: 1,  // 默认 1 个工作协程
		queueSize:   10, // 默认缓冲 10
	}
	for _, opt := range opts {
		opt(publisher)
	}

	// 创建工作队列
	publisher.worker = make(chan func(), publisher.queueSize)

	// 启动工作协程池
	for range publisher.workerCount {
		publisher.wg.Add(1)
		go publisher.run()
	}

	return publisher
}

// run 工作协程主循环
func (p *AsyncPublisher) run() {
	defer p.wg.Done()
	defer func() {
		if rec := recover(); rec != nil {
			_, _ = fmt.Fprintf(os.Stderr, "panic in async publisher worker: %v\n", rec)
		}
	}()
	for {
		select {
		case fn, ok := <-p.worker:
			if !ok {
				return
			}
			fn()
		case <-p.done:
			// 排空 channel 中剩余的任务
			for task := range p.worker {
				task()
			}
			return
		}
	}
}

// Publish 异步发布事件
//
// 将事件发布到工作队列，由工作协程异步处理。
// 支持上下文超时控制，超时后调用错误处理器。
//
// 参数:
//   - ctx: 上下文，用于超时控制
//   - event: 要发布的事件
func (p *AsyncPublisher) Publish(ctx context.Context, event ApplicationEvent) {
	// 已关闭或上下文已完成时快速失败
	if p.closed.Load() {
		p.reportError(fmt.Errorf("event: publisher is closed"), event)
		return
	}

	// 先检查上下文是否已经完成
	select {
	case <-ctx.Done():
		p.reportError(ctx.Err(), event)
		return
	default:
	}

	// 关闭与入队之间的竞态保护：加锁后再次检查
	p.closeMu.Lock()
	if p.closed.Load() {
		p.closeMu.Unlock()
		p.reportError(fmt.Errorf("event: publisher is closed"), event)
		return
	}
	p.wg.Add(1)
	p.taskWg.Add(1)
	p.closeMu.Unlock()

	p.enqueueOrFallback(ctx, event)
}

// reportError 通过错误处理器上报发布错误。
func (p *AsyncPublisher) reportError(err error, event ApplicationEvent) {
	if p.errHandler != nil {
		p.errHandler(err, event)
	}
}

// enqueueOrFallback 将发布任务写入工作队列，队列满或超时则回退为阻塞发布。
func (p *AsyncPublisher) enqueueOrFallback(ctx context.Context, event ApplicationEvent) {
	select {
	case p.worker <- func() {
		defer p.wg.Done()
		defer p.taskWg.Done()
		p.publishEvent(event)
	}:
	case <-ctx.Done():
		p.wg.Done()
		p.taskWg.Done()
		p.reportError(ctx.Err(), event)
	default:
		p.reportError(fmt.Errorf("event: async worker queue full"), event)
		p.fallbackPublish(event)
	}
}

// fallbackPublish 队列满时在独立协程中阻塞发布，带超时保护。
func (p *AsyncPublisher) fallbackPublish(event ApplicationEvent) {
	go p.runFallbackPublish(event)
}

// runFallbackPublish 后备发布协程（WaitGroup 保护）。
func (p *AsyncPublisher) runFallbackPublish(event ApplicationEvent) {
	defer p.wg.Done()
	defer p.taskWg.Done()
	done := make(chan struct{}, 1)
	go p.publishWithDone(event, done)
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		slog.Error("fallback event publish timeout", "event", event.Type())
	}
}

// publishWithDone 发布事件并在完成后关闭 done 通道。
func (p *AsyncPublisher) publishWithDone(event ApplicationEvent, done chan struct{}) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("fallback event handler panic", "event", event.Type(), "recover", rec)
		}
		close(done)
	}()
	p.bus.Publish(event)
}

// publishEvent 发布单个事件，包含 panic 恢复逻辑
func (p *AsyncPublisher) publishEvent(event ApplicationEvent) {
	defer func() {
		if rec := recover(); rec != nil {
			var err error
			if e, ok := rec.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("event handler panic: %v", rec)
			}
			if p.errHandler != nil {
				p.errHandler(err, event)
			}
			slog.Error("event handler panic", "event", event.Type(), "recover", rec)
		}
	}()
	p.bus.Publish(event)
}

// Close 关闭异步发布器
//
// 先通知工作协程停止接收新任务，等待所有正在执行的任务完成，
// 再关闭工作通道让排空循环退出，最后等待所有工作协程退出。
func (p *AsyncPublisher) Close() {
	p.closeMu.Lock()
	if p.closed.Load() {
		p.closeMu.Unlock()
		return
	}
	p.closed.Store(true)
	close(p.done)
	p.closeMu.Unlock()

	p.taskWg.Wait()
	close(p.worker)
	p.wg.Wait()
}
