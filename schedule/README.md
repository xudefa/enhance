# schedule 包 — 定时任务调度

> **所属层级**: Infrastructure Layer  
> **设计理念**: 零外部依赖，Spring 风格调度  
> **设计灵感**: Spring @Scheduled + Quartz

## 概述

`schedule` 包提供零外部依赖的 Spring 风格定时任务调度框架，为应用提供 Cron 表达式解析、任务调度和生命周期管理能力。

### 核心功能

| 功能 | 说明 |
|------|------|
| **Cron 表达式解析** | 6 字段 Cron 表达式，位图编码，`Next()` 算法 |
| **任务调度** | 基于最小堆的调度器，支持动态注册/注销 |
| **固定延迟任务** | 从上次任务完成后延迟固定时间再执行 |
| **固定频率任务** | 从任务开始时计算下一次执行时间 |
| **自动配置** | `ScheduleAutoConfiguration` 在 `schedule.enabled=true` 时自动启用 |
| **并发控制** | 可配置并发池大小，限制同时执行的任务数 |
| **优雅关闭** | 支持超时控制的优雅关闭 |
| **Builder 模式** | 链式配置，简化调度器创建 |
| **Helper 工具** | 简化常见调度器操作 |

## 快速开始

### 创建调度器

```go
package main

import (
    "context"
    "fmt"
    "github.com/xudefa/enhance/schedule"
    "time"
)

func main() {
    // 创建调度器
    scheduler := schedule.NewScheduler(
        schedule.WithPoolSize(10),
        schedule.WithErrorHandler(func(taskName string, err error) {
            fmt.Printf("task %s failed: %v\n", taskName, err)
        }),
    )

    // 创建任务
    task := schedule.NewTask("my-task", "0 */5 * * * *", func(ctx context.Context) error {
        fmt.Println("task executed")
        return nil
    })

    // 注册任务
    if err := scheduler.Register(task); err != nil {
        panic(err)
    }

    // 启动调度器
    ctx := context.Background()
    if err := scheduler.Start(ctx); err != nil {
        panic(err)
    }

    // 运行一段时间后关闭
    // ...

    // 优雅关闭
    scheduler.Shutdown(ctx)
}
```

### 使用 Builder 模式

```go
// 使用 Builder 模式创建调度器
scheduler := schedule.NewSchedulerBuilder().
    PoolSize(10).
    WithCronTask("cron-task", "0 */5 * * * *", func(ctx context.Context) error {
        fmt.Println("cron task executed")
        return nil
    }).
    WithFixedDelayTask("delay-task", 5*time.Second, func(ctx context.Context) error {
        fmt.Println("fixed delay task executed")
        return nil
    }).
    WithFixedRateTask("rate-task", 10*time.Second, func(ctx context.Context) error {
        fmt.Println("fixed rate task executed")
        return nil
    }).
    Build()

// 启动调度器
ctx := context.Background()
scheduler.Start(ctx)
```

### 使用 Helper 工具

```go
// 创建调度器和 Helper
scheduler := schedule.NewScheduler()
helper := schedule.NewScheduleHelper(scheduler)

// 注册任务
helper.RegisterCronTask("my-cron", "0 */5 * * * *", func(ctx context.Context) error {
    return nil
})

helper.RegisterFixedDelayTask("my-delay", 5*time.Second, func(ctx context.Context) error {
    return nil
})

// 检查任务
if helper.HasTask("my-cron") {
    fmt.Println("task exists")
}

fmt.Printf("task count: %d\n", helper.GetTaskCount())

// 注销任务
helper.UnregisterTask("my-cron")
```

### Cron 表达式格式

支持 6 字段 Spring 风格 Cron 表达式：**秒 分 时 日 月 周**

| 字段 | 范围 | 特殊字符 |
|------|------|----------|
| 秒 | 0-59 | `*` `,` `-` `/` |
| 分 | 0-59 | `*` `,` `-` `/` |
| 时 | 0-23 | `*` `,` `-` `/` |
| 日 | 1-31 | `*` `,` `-` `/` |
| 月 | 1-12 或 JAN-DEC | `*` `,` `-` `/` |
| 周 | 0-6 或 SUN-SAT | `*` `,` `-` `/` |

### 常用示例

| 表达式 | 说明 |
|--------|------|
| `0 */5 * * * *` | 每 5 分钟执行 |
| `0 0 */1 * * *` | 每小时执行 |
| `0 0 0 * * *` | 每天零点执行 |
| `0 0 0 * * MON-FRI` | 工作日零点执行 |
| `0 0 0 1 * *` | 每月 1 号零点执行 |
| `0 30 9 * * MON-FRI` | 工作日 9:30 执行 |

## API 参考

### Task 接口

```go
type Task interface {
    Name() string                        // 任务名称
    Cron() string                        // Cron 表达式
    Execute(ctx context.Context) error   // 执行任务
}
```

### 任务创建函数

```go
// Cron 表达式任务
func NewTask(name, cron string, fn func(ctx context.Context) error) Task

// 固定延迟任务（从任务完成后开始计时）
func NewFixedDelayTask(name string, delay time.Duration, fn func(ctx context.Context) error) Task

// 固定频率任务（从任务开始时开始计时）
func NewFixedRateTask(name string, interval time.Duration, fn func(ctx context.Context) error) Task
```

### Scheduler 接口

```go
type Scheduler interface {
    Start(ctx context.Context) error           // 启动调度器
    Shutdown(ctx context.Context) error        // 优雅关闭
    Register(task Task) error                  // 注册任务
    Unregister(name string) bool               // 注销任务
    IsRunning() bool                           // 是否运行中
    RegisteredTasks() []Task                   // 已注册任务列表
}
```

### 配置选项

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithPoolSize(size int)` | 任务执行池大小 | 10 |
| `WithErrorHandler(fn func(taskName string, err error))` | 错误处理函数 | nil |
| `WithLogger(logger log.Logger)` | 日志记录器 | log.NewLoggerBuilder().Build() |

### Builder 模式

```go
type SchedulerBuilder struct{}

func NewSchedulerBuilder() *SchedulerBuilder
func (b *SchedulerBuilder) PoolSize(size int) *SchedulerBuilder
func (b *SchedulerBuilder) ErrorHandler(fn func(taskName string, err error)) *SchedulerBuilder
func (b *SchedulerBuilder) Logger(logger log.Logger) *SchedulerBuilder
func (b *SchedulerBuilder) WithTask(task Task) *SchedulerBuilder
func (b *SchedulerBuilder) WithCronTask(name, cron string, fn func(ctx context.Context) error) *SchedulerBuilder
func (b *SchedulerBuilder) WithFixedDelayTask(name string, delay time.Duration, fn func(ctx context.Context) error) *SchedulerBuilder
func (b *SchedulerBuilder) WithFixedRateTask(name string, interval time.Duration, fn func(ctx context.Context) error) *SchedulerBuilder
func (b *SchedulerBuilder) Build() *DefaultScheduler
```

### Helper 工具

```go
type ScheduleHelper struct{}

func NewScheduleHelper(scheduler *DefaultScheduler) *ScheduleHelper
func (h *ScheduleHelper) RegisterCronTask(name, cron string, fn func(ctx context.Context) error) error
func (h *ScheduleHelper) RegisterFixedDelayTask(name string, delay time.Duration, fn func(ctx context.Context) error) error
func (h *ScheduleHelper) RegisterFixedRateTask(name string, interval time.Duration, fn func(ctx context.Context) error) error
func (h *ScheduleHelper) UnregisterTask(name string) bool
func (h *ScheduleHelper) GetTaskCount() int
func (h *ScheduleHelper) HasTask(name string) bool
func (h *ScheduleHelper) StartAndBlock(ctx context.Context) error
```

## 架构设计

```
┌──────────────────────────────────────────────────────┐
│              DefaultScheduler                        │
│                                                      │
│  ┌──────────────┐  ┌──────────────────────────────┐  │
│  │  taskHeap    │  │      taskPool (semaphore)    │  │
│  │  (min-heap)  │  │      (concurrency control)   │  │
│  │              │  │                              │  │
│  │  Register()  │  │  executeTask()               │  │
│  │  Unregister()│  │  calculateNextRun()          │  │
│  └──────────────┘  └──────────────────────────────┘  │
│                                                      │
│  ┌──────────────┐  ┌──────────────────────────────┐  │
│  │ CronExpr     │  │      Task                    │  │
│  │ (bitmap)     │  │  (function wrapper)          │  │
│  │              │  │                              │  │
│  │  Next()      │  │  Execute()                   │  │
│  └──────────────┘  └──────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

## 最佳实践

### 1. 合理设置并发池大小

```go
// ✅ 推荐：根据任务特性设置池大小
scheduler := schedule.NewScheduler(
    schedule.WithPoolSize(20), // CPU 密集型：CPU 核心数
)

// ✅ 推荐：IO 密集型可以设置更大
scheduler := schedule.NewScheduler(
    schedule.WithPoolSize(50), // IO 密集型：CPU 核心数 * 2
)
```

### 2. 使用错误处理函数

```go
// ✅ 推荐：注册错误处理函数
scheduler := schedule.NewScheduler(
    schedule.WithErrorHandler(func(taskName string, err error) {
        log.Printf("task %s failed: %v", taskName, err)
        // 可以发送告警、记录指标等
    }),
)
```

### 3. 优雅关闭

```go
// ✅ 推荐：使用带超时的关闭
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err := scheduler.Shutdown(ctx); err != nil {
    log.Printf("shutdown timeout: %v", err)
}
```

### 4. 任务幂等性

```go
// ✅ 推荐：确保任务可重复执行
task := schedule.NewTask("cleanup-task", "0 0 0 * * *", func(ctx context.Context) error {
    // 确保多次执行结果一致
    return cleanupOldFiles(ctx)
})
```