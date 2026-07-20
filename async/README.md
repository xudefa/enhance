# async 包 — 异步方法执行

> **所属层级**: Infrastructure Layer  
> **设计理念**: 简化异步编程，Future 模式  
> **设计灵感**: Spring @Async + CompletableFuture

## 概述

`async` 包提供异步方法执行功能，参考 Spring Boot 的 `@Async` 注解设计。基于 goroutine 池实现异步任务执行，支持 Future 模式返回值、优雅关闭等特性。

### 核心功能

| 功能 | 说明 |
|------|------|
| **线程池配置** | 支持核心线程数、最大线程数、队列容量配置 |
| **Future 返回值** | 支持阻塞获取结果、超时获取、状态检查 |
| **自定义线程名** | 支持自定义线程名前缀 |
| **优雅关闭** | 支持带超时的优雅关闭 |
| **零依赖** | 仅使用 Go 标准库 |

---

## 核心接口

### Future 异步结果

```go
type Future struct {
    // ...
}
```

#### 获取结果

| 方法 | 说明 |
|------|------|
| `Get()` | 阻塞获取结果 |
| `GetWithContext(ctx)` | 带上下文的阻塞获取 |
| `GetWithTimeout(timeout)` | 带超时的阻塞获取 |
| `IsDone()` | 检查任务是否完成 |

```go
// 阻塞获取
result, err := future.Get()

// 带超时获取
result, err := future.GetWithTimeout(5 * time.Second)

// 带上下文获取
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
result, err := future.GetWithContext(ctx)

// 状态检查
if future.IsDone() {
    result, err := future.Get()
}
```

### AsyncExecutor 异步执行器

```go
type AsyncExecutor struct {
    // ...
}
```

#### 创建

```go
executor := async.NewAsyncExecutor(
    4,   // 工作线程数
    100, // 任务队列容量
)
```

#### 生命周期管理

| 方法 | 说明 |
|------|------|
| `Start()` | 启动执行器 |
| `Shutdown()` | 优雅关闭执行器 |
| `ShutdownWithTimeout(timeout)` | 带超时的优雅关闭 |
| `IsRunning()` | 检查执行器是否运行 |

```go
// 启动
executor.Start()

// 优雅关闭
executor.Shutdown()

// 带超时关闭
err := executor.ShutdownWithTimeout(10 * time.Second)
```

#### 提交任务

| 方法 | 说明 |
|------|------|
| `Submit(fn)` | 提交有返回值的异步任务 |
| `SubmitVoid(fn)` | 提交无返回值的异步任务 |
| `GetQueueSize()` | 获取当前队列中的任务数 |

```go
// 提交有返回值的任务
future := executor.Submit(func() (any, error) {
    result := doSomething()
    return result, nil
})

// 提交无返回值的任务
executor.SubmitVoid(func() error {
    return doSomething()
})
```

---

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/async"
)

func main() {
    executor := async.NewAsyncExecutor(4, 100)
    defer executor.Shutdown()

    // 提交异步任务
    future := executor.Submit(func() (any, error) {
        // 模拟耗时操作
        time.Sleep(1 * time.Second)
        return "result", nil
    })

    // 阻塞获取结果
    result, err := future.Get()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Result:", result)
}
```

---

## API 参考

### 带超时获取

```go
executor := async.NewAsyncExecutor(4, 100)
defer executor.Shutdown()

future := executor.Submit(func() (any, error) {
    time.Sleep(5 * time.Second)
    return "slow result", nil
})

// 设置 3 秒超时
result, err := future.GetWithTimeout(3 * time.Second)
if err != nil {
    fmt.Println("Timeout:", err)
    return
}
```

### 使用 Context

```go
executor := async.NewAsyncExecutor(4, 100)
defer executor.Shutdown()

future := executor.Submit(func() (any, error) {
    return fetchDataFromAPI()
})

ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := future.GetWithContext(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        fmt.Println("Request timeout")
    }
    return
}
```

### 批量提交任务

```go
executor := async.NewAsyncExecutor(4, 100)
defer executor.Shutdown()

// 提交多个任务
var futures []*async.Future
for i := 0; i < 10; i++ {
    taskID := i
    future := executor.Submit(func() (any, error) {
        return processTask(taskID)
    })
    futures = append(futures, future)
}

// 收集所有结果
for i, future := range futures {
    result, err := future.Get()
    if err != nil {
        fmt.Printf("Task %d failed: %v\n", i, err)
        continue
    }
    fmt.Printf("Task %d result: %v\n", i, result)
}
```

### 无返回值任务

```go
executor := async.NewAsyncExecutor(4, 100)
defer executor.Shutdown()

// 提交后台任务
executor.SubmitVoid(func() error {
    return sendEmailNotification("user@example.com", "Welcome!")
})

executor.SubmitVoid(func() error {
    return updateCache("user:123", userData)
})
```

---

## 使用示例

### 场景 1: 异步邮件发送

用户注册后异步发送邮件通知，不阻塞主流程：

```go
func (s *UserService) RegisterUser(req *RegisterRequest) error {
    // 创建用户
    user := s.createUser(req)

    // 异步发送邮件
    s.executor.SubmitVoid(func() error {
        return s.emailService.SendWelcomeEmail(user.Email)
    })

    // 立即返回
    return nil
}
```

### 场景 2: 批量数据处理

批量处理数据时，使用异步提高处理效率：

```go
func (s *DataService) ProcessBatch(items []Item) ([]Result, error) {
    var futures []*async.Future

    // 提交所有处理任务
    for _, item := range items {
        item := item // 捕获循环变量
        future := s.executor.Submit(func() (any, error) {
            return s.processItem(item)
        })
        futures = append(futures, future)
    }

    // 收集结果
    results := make([]Result, 0, len(items))
    for _, future := range futures {
        result, err := future.Get()
        if err != nil {
            return nil, err
        }
        results = append(results, result.(Result))
    }

    return results, nil
}
```

### 场景 3: 后台定时任务

执行后台定时任务，如数据清理、缓存刷新等：

```go
func (s *CleanupService) StartCleanupJob() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for range ticker.C {
        s.executor.SubmitVoid(func() error {
            count, err := s.cleanupExpiredRecords()
            if err != nil {
                log.Printf("Cleanup failed: %v", err)
                return err
            }
            log.Printf("Cleaned up %d records", count)
            return nil
        })
    }
}
```

### 场景 4: 并发 API 调用

并发调用多个外部 API，减少总耗时：

```go
func (s *AggregationService) GetAggregatedData(userID string) (*AggregatedData, error) {
    // 并发调用多个 API
    userFuture := s.executor.Submit(func() (any, error) {
        return s.userAPI.GetUser(userID)
    })

    ordersFuture := s.executor.Submit(func() (any, error) {
        return s.orderAPI.GetOrders(userID)
    })

    statsFuture := s.executor.Submit(func() (any, error) {
        return s.statsAPI.GetUserStats(userID)
    })

    // 等待所有结果
    user, err := userFuture.Get()
    if err != nil {
        return nil, err
    }

    orders, err := ordersFuture.Get()
    if err != nil {
        return nil, err
    }

    stats, err := statsFuture.Get()
    if err != nil {
        return nil, err
    }

    return &AggregatedData{
        User:   user.(*User),
        Orders: orders.([]*Order),
        Stats:  stats.(*UserStats),
    }, nil
}
```

---

## 最佳实践

### 1. 非关键路径操作使用异步

```go
// ✅ 推荐：邮件发送使用异步
s.executor.SubmitVoid(func() error {
    return s.emailService.SendWelcomeEmail(user.Email)
})

// ⚠️ 不推荐：阻塞主流程
err := s.emailService.SendWelcomeEmail(user.Email)
```

### 2. 捕获循环变量避免闭包问题

```go
// ✅ 推荐：捕获循环变量
for _, item := range items {
    item := item // 捕获循环变量
    executor.Submit(func() (any, error) {
        return processItem(item)
    })
}

// ⚠️ 不推荐：直接使用循环变量
for _, item := range items {
    executor.Submit(func() (any, error) {
        return processItem(item) // 可能获取到错误的值
    })
}
```

### 3. 设置合理的超时时间

```go
// ✅ 推荐：设置超时
result, err := future.GetWithTimeout(5 * time.Second)

// ⚠️ 不推荐：无限期等待
result, err := future.Get()
```

### 4. 优雅关闭执行器

```go
// ✅ 推荐：使用 defer 确保关闭
executor := async.NewAsyncExecutor(4, 100)
defer executor.ShutdownWithTimeout(10 * time.Second)

// ⚠️ 不推荐：忘记关闭
executor := async.NewAsyncExecutor(4, 100)
```

### 5. 与依赖注入集成

```go
// ✅ 推荐：将 Executor 注册为 Bean
container.Register(
    reflect.TypeOf(&async.AsyncExecutor{}),
    core.Bean(async.NewAsyncExecutor(4, 100)),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    Executor *async.AsyncExecutor `inject:"asyncExecutor"`
}
```

---

## 设计要点

- `AsyncExecutor` 使用 `context.Context` 控制生命周期
- `Future` 使用 `chan struct{}` 实现阻塞等待
- 任务队列使用 `chan` 实现，支持缓冲
- 优雅关闭时等待所有任务完成
- 零外部依赖，仅使用 Go 标准库