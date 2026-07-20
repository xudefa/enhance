// Package schedule 提供定时任务调度功能，用于 enhance 框架。
package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/xudefa/enhance/log"
)

// SchedulerBuilder 调度器构建器，支持链式配置。
type SchedulerBuilder struct {
	poolSize     int
	errorHandler func(taskName string, err error)
	logger       log.Logger
	tasks        []Task
}

// NewSchedulerBuilder 创建调度器构建器。
func NewSchedulerBuilder() *SchedulerBuilder {
	return &SchedulerBuilder{
		poolSize: DefaultSchedulePoolSize,
		tasks:    make([]Task, 0),
	}
}

// PoolSize 设置任务执行池大小。
func (b *SchedulerBuilder) PoolSize(size int) *SchedulerBuilder {
	b.poolSize = size
	return b
}

// ErrorHandler 设置错误处理函数。
func (b *SchedulerBuilder) ErrorHandler(fn func(taskName string, err error)) *SchedulerBuilder {
	b.errorHandler = fn
	return b
}

// Logger 设置日志记录器。
func (b *SchedulerBuilder) Logger(logger log.Logger) *SchedulerBuilder {
	b.logger = logger
	return b
}

// WithTask 添加任务。
func (b *SchedulerBuilder) WithTask(task Task) *SchedulerBuilder {
	b.tasks = append(b.tasks, task)
	return b
}

// WithCronTask 添加 Cron 表达式任务。
func (b *SchedulerBuilder) WithCronTask(name, cron string, fn func(ctx context.Context) error) *SchedulerBuilder {
	b.tasks = append(b.tasks, NewTask(name, cron, fn))
	return b
}

// WithFixedDelayTask 添加固定延迟任务。
func (b *SchedulerBuilder) WithFixedDelayTask(name string, delay time.Duration, fn func(ctx context.Context) error) *SchedulerBuilder {
	b.tasks = append(b.tasks, NewFixedDelayTask(name, delay, fn))
	return b
}

// WithFixedRateTask 添加固定频率任务。
func (b *SchedulerBuilder) WithFixedRateTask(name string, interval time.Duration, fn func(ctx context.Context) error) *SchedulerBuilder {
	b.tasks = append(b.tasks, NewFixedRateTask(name, interval, fn))
	return b
}

// Build 构建调度器。
func (b *SchedulerBuilder) Build() *DefaultScheduler {
	opts := []SchedulerOption{
		WithPoolSize(b.poolSize),
		WithErrorHandler(b.errorHandler),
	}

	if b.logger != nil {
		opts = append(opts, WithLogger(b.logger))
	}

	scheduler := NewScheduler(opts...)

	for _, task := range b.tasks {
		if err := scheduler.Register(task); err != nil {
			scheduler.logger.Error(context.Background(), "failed to register task",
				log.KeyValue{Key: "task", Value: task.Name()},
				log.KeyValue{Key: "error", Value: err},
			)
		}
	}

	return scheduler
}

// MustBuild 构建调度器，失败则 panic。
func (b *SchedulerBuilder) MustBuild() *DefaultScheduler {
	return b.Build()
}

// ScheduleHelper 调度器辅助工具，简化常见操作。
type ScheduleHelper struct {
	scheduler *DefaultScheduler
}

// NewScheduleHelper 创建调度器辅助工具。
func NewScheduleHelper(scheduler *DefaultScheduler) *ScheduleHelper {
	return &ScheduleHelper{scheduler: scheduler}
}

// RegisterCronTask 注册 Cron 表达式任务。
func (h *ScheduleHelper) RegisterCronTask(name, cron string, fn func(ctx context.Context) error) error {
	task := NewTask(name, cron, fn)
	return h.scheduler.Register(task)
}

// RegisterFixedDelayTask 注册固定延迟任务。
func (h *ScheduleHelper) RegisterFixedDelayTask(name string, delay time.Duration, fn func(ctx context.Context) error) error {
	task := NewFixedDelayTask(name, delay, fn)
	return h.scheduler.Register(task)
}

// RegisterFixedRateTask 注册固定频率任务。
func (h *ScheduleHelper) RegisterFixedRateTask(name string, interval time.Duration, fn func(ctx context.Context) error) error {
	task := NewFixedRateTask(name, interval, fn)
	return h.scheduler.Register(task)
}

// UnregisterTask 注销任务。
func (h *ScheduleHelper) UnregisterTask(name string) bool {
	return h.scheduler.Unregister(name)
}

// GetTaskCount 获取已注册任务数量。
func (h *ScheduleHelper) GetTaskCount() int {
	return len(h.scheduler.RegisteredTasks())
}

// HasTask 检查任务是否已注册。
func (h *ScheduleHelper) HasTask(name string) bool {
	for _, task := range h.scheduler.RegisteredTasks() {
		if task.Name() == name {
			return true
		}
	}
	return false
}

// StartAndBlock 启动调度器并阻塞，直到收到停止信号。
func (h *ScheduleHelper) StartAndBlock(ctx context.Context) error {
	if err := h.scheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler: %w", err)
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return h.scheduler.Shutdown(shutdownCtx)
}
