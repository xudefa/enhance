package event

import (
	"context"
	"log/slog"
	"sync"
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
	wg          sync.WaitGroup
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
	p := &AsyncPublisher{
		bus:         bus,
		done:        make(chan struct{}),
		workerCount: 1,  // 默认 1 个工作协程
		queueSize:   10, // 默认缓冲 10
	}
	for _, opt := range opts {
		opt(p)
	}

	// 创建工作队列
	p.worker = make(chan func(), p.queueSize)

	// 启动工作协程池
	for range p.workerCount {
		p.wg.Add(1)
		go p.run()
	}

	return p
}

// run 工作协程主循环
func (p *AsyncPublisher) run() {
	defer p.wg.Done()
	for {
		select {
		case fn := <-p.worker:
			fn()
		case <-p.done:
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
	// 先检查上下文是否已经完成
	select {
	case <-ctx.Done():
		if p.errHandler != nil {
			p.errHandler(ctx.Err(), event)
		}
		return
	default:
	}

	// 在发送前增加 wg 计数，避免 Close() 在 wg.Add 和 select 之间调用导致 goroutine 泄漏
	p.wg.Add(1)
	select {
	case p.worker <- func() {
		defer p.wg.Done()
		p.publishEvent(event)
	}:
	case <-ctx.Done():
		// 等待期间上下文取消，取消本次发布
		p.wg.Done()
		if p.errHandler != nil {
			p.errHandler(ctx.Err(), event)
		}
	}
}

// publishEvent 发布单个事件，包含 panic 恢复逻辑
func (p *AsyncPublisher) publishEvent(event ApplicationEvent) {
	defer func() {
		if r := recover(); r != nil {
			if p.errHandler != nil {
				p.errHandler(nil, event)
			}
			slog.Error("event handler panic", "event", event.Type(), "recover", r)
		}
	}()
	p.bus.Publish(event)
}

// Close 关闭异步发布器
//
// 等待所有待处理的事件处理完成后返回。
func (p *AsyncPublisher) Close() {
	close(p.done)
	// 等待工作协程退出
	p.wg.Wait()
}
