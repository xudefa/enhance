# retry 包 — 重试机制

> **所属层级**: Infrastructure Layer  
> **设计理念**: 独立重试机制，统一重试策略  
> **设计灵感**: Spring Retry + Resilience4j Retry

## 概述

`retry` 包提供独立的重试机制支持，从 `event/deadletter.go` 和 `web/server/retry.go` 抽取而来，为事件处理、HTTP 请求等场景提供统一的重试策略和执行器。

### 核心功能

| 功能 | 说明 |
|------|------|
| **重试策略配置** | 最大次数、退避策略、延迟计算 |
| **退避策略** | None/Fixed/Linear/Exponential 四种策略 |
| **Jitter 防惊群** | 抖动比例 (0.0-1.0)，防止惊群效应 |
| **上下文取消** | 支持 context.Context 取消重试 |
| **回调通知** | 重试成功/失败回调通知 |
| **零依赖** | 仅使用 Go 标准库 |

---

## 核心类型

### RetryPolicy 重试策略配置

```go
type RetryPolicy struct {
    MaxAttempts  int             // 最大尝试次数（包含首次），0 表示不重试
    Strategy     BackoffStrategy // 退避策略
    InitialDelay time.Duration   // 初始延迟
    MaxDelay     time.Duration   // 最大延迟（指数退避上限）
    Multiplier   float64         // 退避乘数（指数退避用）
    Jitter       float64         // 抖动比例 (0.0-1.0)
}
```

### Executor 重试执行器

```go
executor := retry.NewExecutor(policy)
result, err := executor.Execute(ctx, func(ctx context.Context) (any, error) {
    return httpClient.Get(ctx, url)
})
```

---

## 便捷构造函数

| 函数 | 说明 |
|------|------|
| `NoRetry()` | 不重试策略（仅执行一次） |
| `FixedDelay(maxAttempts, delay)` | 固定延迟重试 |
| `LinearBackoff(maxAttempts, initialDelay, maxDelay)` | 线性退避 |
| `ExponentialBackoff(maxAttempts, initialDelay, maxDelay)` | 指数退避 |

---

## 退避策略

| 策略 | 常量 | 说明 |
|------|------|------|
| 无退避 | `BackoffNone` | 立即重试 |
| 固定间隔 | `BackoffFixed` | 每次延迟相同时间 |
| 线性 | `BackoffLinear` | 延迟线性递增 |
| 指数 | `BackoffExponential` | 延迟指数递增（推荐） |

---

## 使用示例

```go
// 指数退避：最多 3 次，初始 100ms，最大 10s
policy := retry.ExponentialBackoff(3, 100*time.Millisecond, 10*time.Second)
executor := retry.NewExecutor(policy)

result, err := executor.Execute(ctx, func(ctx context.Context) (any, error) {
    return httpClient.Get(ctx, url)
})
```

---

## 设计决策

- **ADR-004**: 零外部依赖，仅使用 Go 标准库
- **ADR-005**: 接口定义在 doc.go，实现隐藏在 retry.go
- 从 `event` 和 `web/server` 抽取，统一重试逻辑避免重复