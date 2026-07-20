// Package schedule 提供定时任务调度功能，用于 enhance 框架。
package schedule

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/xudefa/enhance/log"
)

// SchedulerOption 调度器配置选项函数。
type SchedulerOption func(*schedulerConfig)

// schedulerConfig 调度器内部配置。
type schedulerConfig struct {
	poolSize     int
	errorHandler func(taskName string, err error)
	logger       log.Logger
}

// WithPoolSize 设置任务执行池大小（最大并发数）。
func WithPoolSize(size int) SchedulerOption {
	return func(cfg *schedulerConfig) {
		if size > 0 {
			cfg.poolSize = size
		}
	}
}

// WithErrorHandler 设置执行错误处理函数。
func WithErrorHandler(fn func(taskName string, err error)) SchedulerOption {
	return func(cfg *schedulerConfig) {
		cfg.errorHandler = fn
	}
}

// WithLogger 设置日志记录器。
func WithLogger(logger log.Logger) SchedulerOption {
	return func(cfg *schedulerConfig) {
		cfg.logger = logger
	}
}

// scheduledTask 内部调度任务结构。
type scheduledTask struct {
	task    Task
	nextRun time.Time
	index   int // heap 中的索引
}

// taskHeap 基于最小堆的任务调度队列。
type taskHeap []*scheduledTask

func (h taskHeap) Len() int { return len(h) }

func (h taskHeap) Less(i, j int) bool {
	return h[i].nextRun.Before(h[j].nextRun)
}

func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *taskHeap) Push(x any) {
	n := len(*h)
	item := x.(*scheduledTask)
	item.index = n
	*h = append(*h, item)
}

func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[:n-1]
	return item
}

// DefaultScheduler 默认调度器实现。
//
// 基于最小堆实现任务调度，支持动态注册/注销任务、并发控制和优雅关闭。
type DefaultScheduler struct {
	mu           sync.RWMutex
	tasks        map[string]*scheduledTask
	heap         taskHeap
	poolSize     int
	semaphore    chan struct{}
	errorHandler func(taskName string, err error)
	logger       log.Logger
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc
	done         chan struct{}
}

// NewScheduler 创建调度器实例。
func NewScheduler(opts ...SchedulerOption) *DefaultScheduler {
	cfg := &schedulerConfig{
		poolSize: DefaultSchedulePoolSize,
		logger:   log.NewLoggerBuilder().Build(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())

	s := &DefaultScheduler{
		tasks:        make(map[string]*scheduledTask),
		semaphore:    make(chan struct{}, cfg.poolSize),
		errorHandler: cfg.errorHandler,
		logger:       cfg.logger,
		ctx:          ctx,
		cancel:       cancel,
		done:         make(chan struct{}),
	}

	heap.Init(&s.heap)

	return s
}

// Start 启动调度器，开始触发定时任务。
func (s *DefaultScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}

	s.running = true
	s.mu.Unlock()

	s.logger.Info(context.Background(), "scheduler started",
		log.KeyValue{Key: "pool_size", Value: s.poolSize})

	go s.run()

	return nil
}

// run 调度器主循环。
func (s *DefaultScheduler) run() {
	defer close(s.done)

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info(context.Background(), "scheduler stopping")
			return
		default:
			s.mu.Lock()
			if s.heap.Len() == 0 {
				s.mu.Unlock()
				time.Sleep(100 * time.Millisecond)
				continue
			}

			next := s.heap[0]
			now := time.Now()

			if next.nextRun.After(now) {
				waitTime := next.nextRun.Sub(now)
				s.mu.Unlock()

				select {
				case <-time.After(waitTime):
					// 等待完成，继续执行
				case <-s.ctx.Done():
					return
				}
			} else {
				heap.Pop(&s.heap)
				s.mu.Unlock()

				// 检查任务是否已被注销
				s.mu.RLock()
				_, exists := s.tasks[next.task.Name()]
				s.mu.RUnlock()

				if !exists {
					// 任务已注销，不再执行
					continue
				}

				s.executeTask(next)

				// 再次检查任务是否仍然存在
				s.mu.RLock()
				_, stillExists := s.tasks[next.task.Name()]
				s.mu.RUnlock()

				if stillExists {
					s.mu.Lock()
					next.nextRun = s.calculateNextRun(next.task, next.nextRun)
					heap.Push(&s.heap, next)
					s.mu.Unlock()
				}
			}
		}
	}
}

// executeTask 执行单个任务。
func (s *DefaultScheduler) executeTask(st *scheduledTask) {
	select {
	case s.semaphore <- struct{}{}:
		go func() {
			defer func() { <-s.semaphore }()

			s.logger.Debug(s.ctx, "executing task",
				log.KeyValue{Key: "task", Value: st.task.Name()})

			err := st.task.Execute(s.ctx)
			if err != nil {
				s.logger.Error(s.ctx, "task execution failed",
					log.KeyValue{Key: "task", Value: st.task.Name()},
					log.KeyValue{Key: "error", Value: err},
				)

				if s.errorHandler != nil {
					s.errorHandler(st.task.Name(), err)
				}
			}
		}()
	default:
		s.logger.Warn(s.ctx, "task pool exhausted, skipping",
			log.KeyValue{Key: "task", Value: st.task.Name()},
		)
	}
}

// calculateNextRun 计算任务的下次执行时间。
func (s *DefaultScheduler) calculateNextRun(task Task, lastRun time.Time) time.Time {
	// 处理固定延迟任务
	if ft, ok := task.(*fixedDelayTask); ok {
		return lastRun.Add(ft.FixedDelay())
	}

	// 处理固定频率任务
	if ft, ok := task.(*fixedRateTask); ok {
		return lastRun.Add(ft.Interval())
	}

	// 处理 Cron 表达式任务
	cron, err := ParseCronExpression(task.Cron())
	if err != nil {
		s.logger.Error(s.ctx, "invalid cron expression",
			log.KeyValue{Key: "task", Value: task.Name()},
			log.KeyValue{Key: "cron", Value: task.Cron()},
			log.KeyValue{Key: "error", Value: err},
		)
		return time.Now().Add(time.Hour)
	}

	return cron.Next(lastRun)
}

// Shutdown 优雅关闭，等待正在执行的任务完成。
func (s *DefaultScheduler) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}

	s.running = false
	s.mu.Unlock()

	s.logger.Info(s.ctx, "scheduler shutting down")

	s.cancel()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.done:
		s.logger.Info(s.ctx, "scheduler stopped")
		return nil
	}
}

// Register 注册定时任务，任务名唯一，重复返回 error。
func (s *DefaultScheduler) Register(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.Name()]; exists {
		return fmt.Errorf("task %q already registered", task.Name())
	}

	nextRun := s.calculateNextRun(task, time.Now())
	st := &scheduledTask{
		task:    task,
		nextRun: nextRun,
	}

	s.tasks[task.Name()] = st
	heap.Push(&s.heap, st)

	s.logger.Info(s.ctx, "task registered",
		log.KeyValue{Key: "task", Value: task.Name()},
		log.KeyValue{Key: "cron", Value: task.Cron()},
		log.KeyValue{Key: "next_run", Value: nextRun},
	)

	return nil
}

// Unregister 注销定时任务。
func (s *DefaultScheduler) Unregister(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	st, exists := s.tasks[name]
	if !exists {
		return false
	}

	heap.Remove(&s.heap, st.index)
	delete(s.tasks, name)

	s.logger.Info(s.ctx, "task unregistered",
		log.KeyValue{Key: "task", Value: name})

	return true
}

// IsRunning 返回调度器是否正在运行。
func (s *DefaultScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// RegisteredTasks 返回所有已注册任务。
func (s *DefaultScheduler) RegisteredTasks() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]Task, 0, len(s.tasks))
	for _, st := range s.tasks {
		tasks = append(tasks, st.task)
	}

	return tasks
}
