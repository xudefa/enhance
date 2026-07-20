# metrics 包 — 指标收集

> **所属层级**: Infrastructure Layer  
> **设计理念**: 零外部依赖，轻量级指标收集  
> **设计灵感**: Micrometer

## 概述

`metrics` 包提供轻量级的指标收集接口和默认实现，**零外部依赖**，仅使用 Go 标准库。核心概念参考 Micrometer，提供 `Counter`（计数器）和 `Gauge`（仪表盘）两种指标类型，通过 `MeterRegistry` 统一管理。

### 核心功能

| 功能 | 说明 |
|------|------|
| **计数器** | Counter 只增不减，适用于请求总数、错误次数等 |
| **仪表盘** | Gauge 可增可减，适用于连接数、内存使用量等 |
| **指标注册表** | MeterRegistry 统一管理指标创建与收集 |
| **标签支持** | 支持为指标添加键值对标签 |
| **并发安全** | 所有实现均为并发安全 |
| **零依赖** | 仅使用 Go 标准库 |

---

## 核心接口

### Counter 接口

计数器，只增不减，适用于记录请求总数、错误次数等单调递增的数值：

```go
type Counter interface {
    Inc()          // 加 1
    Add(v float64) // 增加指定值
    Value() float64 // 获取当前值
}
```

### Gauge 接口

仪表盘，可增可减，适用于记录当前连接数、内存使用量、CPU 使用率等：

```go
type Gauge interface {
    Set(v float64)  // 设置当前值
    Add(v float64)  // 增加指定值（负数即减少）
    Value() float64 // 获取当前值
}
```

### MeterRegistry 接口

指标注册表，管理 Counter 和 Gauge 的创建与收集：

```go
type MeterRegistry interface {
    Counter(name string, tags ...string) Counter // 获取或创建计数器
    Gauge(name string, tags ...string) Gauge     // 获取或创建仪表盘
    Collect() []Metric                            // 收集所有指标快照
}
```

- `tags` 参数为偶数个的键值对序列（如 `"service", "auth", "version", "v1"`）
- `Collect()` 返回所有已注册指标的 `Metric` 快照

### Metric 结构体

指标快照结构体，用于采集和上报：

```go
type Metric struct {
    Name  string            // 指标名称
    Value float64           // 指标当前值
    Tags  map[string]string // 指标标签
}
```

---

## 快速开始

### 创建指标注册表

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/metrics"
)

func main() {
    registry := metrics.NewSimpleRegistry()

    // 创建计数器
    reqCounter := registry.Counter("requests_total", "service", "api")
    errCounter := registry.Counter("errors_total", "service", "api")

    // 创建仪表盘
    connGauge := registry.Gauge("active_connections", "pool", "main")

    // 模拟业务
    reqCounter.Inc()
    reqCounter.Inc()
    errCounter.Inc()
    connGauge.Set(42)
    connGauge.Add(-1)

    // 采集
    for _, m := range registry.Collect() {
        fmt.Printf("Metric: %s = %.0f\n", m.Name, m.Value)
    }
    // 输出:
    // Metric: requests_total = 2
    // Metric: errors_total = 1
    // Metric: active_connections = 41
}
```

### 独立使用 Counter / Gauge

`SimpleCounter` 和 `SimpleGauge` 可独立于 Registry 使用：

```go
counter := metrics.NewSimpleCounter()
counter.Inc()

gauge := metrics.NewSimpleGauge()
gauge.Set(100)
```

---

## API 参考

### SimpleRegistry — 默认实现

`SimpleRegistry` 是 `MeterRegistry` 的默认实现，使用 `map` 存储指标，`sync.Mutex` 保证并发安全：

```go
registry := metrics.NewSimpleRegistry()

counter := registry.Counter("http_requests_total", "method", "GET")
counter.Inc()

gauge := registry.Gauge("memory_usage", "type", "heap")
gauge.Set(1024.5)

// 收集所有指标
allMetrics := registry.Collect()
for _, m := range allMetrics {
    fmt.Printf("%s = %.2f (tags: %v)\n", m.Name, m.Value, m.Tags)
}
```

### 设计要点

- `SimpleCounter` 使用 `sync.RWMutex`：读操作（Value）使用 RLock，写操作（Inc/Add）使用 Lock
- `SimpleGauge` 使用 `sync.Mutex`（无区分读写锁的必要）
- `SimpleRegistry` 按名称索引，同名 Counter 或 Gauge 在注册表中共享同一实例
- `tags` 参数通过 `parseTags()` 解析为 `map[string]string`，偶数索引为 key，奇数索引为 value
- 所有实现均为并发安全，适用于多 goroutine 场景

---

## 使用示例

### HTTP 请求指标

```go
registry := metrics.NewSimpleRegistry()

// 创建指标
reqCounter := registry.Counter("http_requests_total", "method", "GET", "status", "200")
errCounter := registry.Counter("http_errors_total", "method", "POST")
durationGauge := registry.Gauge("http_request_duration", "path", "/api/users")

// 记录请求
reqCounter.Inc()

// 记录错误
errCounter.Inc()

// 记录耗时
durationGauge.Set(150.5) // 毫秒
```

### 数据库连接池指标

```go
registry := metrics.NewSimpleRegistry()

// 连接数
activeConnGauge := registry.Gauge("db_active_connections", "pool", "main")
idleConnGauge := registry.Gauge("db_idle_connections", "pool", "main")

// 查询统计
queryCounter := registry.Counter("db_queries_total", "type", "select")
errorCounter := registry.Counter("db_errors_total", "type", "timeout")

// 更新指标
activeConnGauge.Set(10)
idleConnGauge.Set(5)
queryCounter.Add(100)
errorCounter.Inc()
```

### 与 Actuator 集成

```go
// 将 metrics 注册表注入到 Actuator
actuator := actuator.New(ctx)
actuator.SetMetricsRegistry(registry)

// 通过 /actuator/metrics 端点暴露指标
```

---

## 最佳实践

### 1. 使用标签区分指标维度

```go
// ✅ 推荐：使用标签区分不同维度
reqCounter := registry.Counter("http_requests_total", 
    "method", "GET",
    "status", "200",
    "path", "/api/users",
)

// ⚠️ 不推荐：为每个组合创建独立指标
reqCounterGet200 := registry.Counter("http_requests_get_200_users")
reqCounterPost200 := registry.Counter("http_requests_post_200_users")
```

### 2. 合理选择 Counter 和 Gauge

```go
// ✅ 推荐：请求总数使用 Counter（只增不减）
reqCounter := registry.Counter("requests_total")
reqCounter.Inc()

// ✅ 推荐：连接数使用 Gauge（可增可减）
connGauge := registry.Gauge("active_connections")
connGauge.Set(10)
connGauge.Add(-1)

// ⚠️ 不推荐：使用 Gauge 记录请求总数
reqGauge := registry.Gauge("requests_total")
reqGauge.Set(100) // 容易误操作减少
```

### 3. 定期采集指标

```go
// ✅ 推荐：定期采集并上报指标
ticker := time.NewTicker(10 * time.Second)
go func() {
    for range ticker.C {
        metrics := registry.Collect()
        reportToMonitoring(metrics)
    }
}()

// ⚠️ 不推荐：只在程序结束时采集
defer func() {
    reportToMonitoring(registry.Collect())
}()
```

### 4. 与依赖注入集成

```go
// ✅ 推荐：将 Registry 注册为 Bean
container.Register(
    reflect.TypeOf(&metrics.SimpleRegistry{}),
    core.Bean(metrics.NewSimpleRegistry()),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    Metrics metrics.MeterRegistry `inject:"metrics"`
}

func (s *UserService) GetUser(id int) (*User, error) {
    counter := s.Metrics.Counter("user_get_total")
    counter.Inc()
    return s.db.GetUser(id)
}
```

### 5. 命名规范

```go
// ✅ 推荐：使用下划线分隔的命名
registry.Counter("http_requests_total")
registry.Gauge("active_connections")

// ✅ 推荐：添加单位后缀
registry.Gauge("request_duration_ms")
registry.Gauge("memory_usage_bytes")

// ⚠️ 不推荐：使用驼峰命名
registry.Counter("httpRequestsTotal")
```