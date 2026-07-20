// Package audit 提供审计日志功能，用于 enhance 框架。
package audit

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

func init() {
	ErrWriterClosed = errors.New("audit writer is closed")
	ErrChannelFull = errors.New("audit event channel is full")
}

// auditorImpl Auditor 接口的默认实现。
type auditorImpl struct {
	mu         sync.Mutex
	writer     EventWriter
	bufferSize int
	async      bool
	eventChan  chan Event
	wg         sync.WaitGroup
	closed     bool
	idCounter  int64
}

// NewAuditor 创建审计日志器
//
// 支持通过选项函数配置写入器、缓冲区大小和异步模式。
// 默认使用控制台写入器，同步模式，缓冲区大小 1000。
//
// 示例:
//
//	auditor := audit.NewAuditor(
//	    audit.WithWriter(fileWriter),
//	    audit.WithAsync(),
//	    audit.WithBufferSize(500),
//	)
func NewAuditor(opts ...AuditorOption) Auditor {
	auditor := &auditorImpl{
		writer:     NewConsoleWriter(),
		bufferSize: 1000,
		async:      false,
	}

	for _, opt := range opts {
		opt(auditor)
	}

	if auditor.async {
		auditor.eventChan = make(chan Event, auditor.bufferSize)
		auditor.wg.Add(1)
		go auditor.processEvents()
	}

	return auditor
}

// WithWriter 设置事件写入器。
func WithWriter(writer EventWriter) AuditorOption {
	return func(a Auditor) {
		if impl, ok := a.(*auditorImpl); ok {
			impl.writer = writer
		}
	}
}

// WithBufferSize 设置缓冲区大小。
func WithBufferSize(size int) AuditorOption {
	return func(a Auditor) {
		if impl, ok := a.(*auditorImpl); ok {
			impl.bufferSize = size
		}
	}
}

// WithAsync 启用异步写入模式。
func WithAsync() AuditorOption {
	return func(a Auditor) {
		if impl, ok := a.(*auditorImpl); ok {
			impl.async = true
		}
	}
}

func (a *auditorImpl) Log(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if event.ID == "" {
		event.ID = a.generateID()
	}

	if event.Severity == "" {
		event.Severity = SeverityInfo
	}

	if a.async {
		select {
		case a.eventChan <- event:
			return
		default:
			a.writer.Write(event)
		}
		return
	}

	a.writer.Write(event)
}

func (a *auditorImpl) Close() error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.closed = true
	a.mu.Unlock()

	if a.async {
		close(a.eventChan)
		a.wg.Wait()
	}

	return a.writer.Close()
}

func (a *auditorImpl) IsClosed() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.closed
}

// processEvents 处理事件(异步模式)
func (a *auditorImpl) processEvents() {
	defer a.wg.Done()

	for event := range a.eventChan {
		if err := a.writer.Write(event); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write audit event: %v\n", err)
		}
	}
}

// generateID 生成事件 ID
func (a *auditorImpl) generateID() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.idCounter++
	return fmt.Sprintf("audit-%d-%d", time.Now().UnixNano(), a.idCounter)
}
