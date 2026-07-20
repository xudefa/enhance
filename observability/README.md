# observability 包 — 可观测性

> **所属层级**: Infrastructure Layer  
> **设计理念**: 日志与指标，全面监控  
> **设计灵感**: Spring Boot Actuator + Micrometer

## 概述

`observability` 包提供应用可观测性支持，包含日志可观测性和指标可观测性两个子模块，帮助开发者监控应用运行状态。

### 子模块

| 子模块 | 说明 |
|--------|------|
| **metrics** | 指标可观测性：Counter 计数器、Gauge 仪表盘 |
| **logging** | 日志可观测性：结构化日志、日志级别控制 |

---

## 核心接口

### Counter 计数器

```go
type Counter interface {
    Inc(delta float64)
    Value() float64
}
```

#### 创建和使用

```go
counter := metrics.NewCounter("http_requests_total")
counter.Inc(1)

// 查询指标值
value := counter.Value()
```

### Gauge 仪表盘

```go
type Gauge interface {
    Set(value float64)
    Value() float64
}
```

#### 创建和使用

```go
gauge := metrics.NewGauge("goroutine_count")
gauge.Set(float64(runtime.NumGoroutine()))

// 查询当前值
value := gauge.Value()
```

### MeterRegistry 指标注册表

```go
type MeterRegistry interface {
    Counter(name string, tags ...string) Counter
    Gauge(name string, tags ...string) Gauge
    Collect() []Metric
}
```

#### 创建和使用

```go
registry := metrics.NewMeterRegistry()

// 创建计数器
counter := registry.Counter("http_requests_total", "method", "GET", "status", "200")
counter.Inc()

// 创建仪表盘
gauge := registry.Gauge("active_connections", "pool", "main")
gauge.Set(10)

// 收集所有指标
metrics := registry.Collect()
```

### Logger 日志接口

```go
type Logger interface {
    Debug(ctx context.Context, msg string, keys ...KeyValue)
    Info(ctx context.Context, msg string, keys ...KeyValue)
    Warn(ctx context.Context, msg string, keys ...KeyValue)
    Error(ctx context.Context, msg string, keys ...KeyValue)
    With(ctx context.Context, keys ...KeyValue) Logger
}
```

#### 创建和使用

```go
logger := logging.NewLogger()

// 结构化日志
logger.Info(ctx, "用户登录",
    logging.KeyValue{Key: "user_id", Value: 123},
    logging.KeyValue{Key: "ip", Value: "192.168.1.1"},
)

// 创建子日志器（携带上下文）
subLogger := logger.With(ctx, logging.KeyValue{Key: "request_id", Value: "req-123"})
subLogger.Info(ctx, "处理请求")
```

---

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "runtime"
    "github.com/xudefa/enhance/observability/metrics"
    "github.com/xudefa/enhance/observability/logging"
)

func main() {
    // 创建指标注册表
    registry := metrics.NewMeterRegistry()

    // 创建计数器
    requestCounter := registry.Counter("http_requests_total", "method", "GET")
    requestCounter.Inc()

    // 创建仪表盘
    goroutineGauge := registry.Gauge("goroutine_count")
    goroutineGauge.Set(float64(runtime.NumGoroutine()))

    // 创建日志器
    logger := logging.NewLogger()
    ctx := context.Background()

    // 记录日志
    logger.Info(ctx, "应用启动",
        logging.KeyValue{Key: "version", Value: "1.0.0"},
        logging.KeyValue{Key: "port", Value: 8080},
    )
}
```

---

## API 参考

### 带标签的指标

```go
registry := metrics.NewMeterRegistry()

// 创建带标签的计数器
counter := registry.Counter(
    "http_requests_total",
    "method", "GET",
    "status", "200",
    "path", "/api/users",
)
counter.Inc()

// 创建带标签的仪表盘
gauge := registry.Gauge(
    "active_connections",
    "pool", "main",
    "database", "postgres",
)
gauge.Set(10)
```

### 日志级别控制

```go
logger := logging.NewLogger()

// 设置日志级别
logger.SetLevel(logging.LevelDebug)

// 不同级别的日志
logger.Debug(ctx, "调试信息", logging.KeyValue{Key: "detail", Value: "verbose"})
logger.Info(ctx, "普通信息", logging.KeyValue{Key: "user_id", Value: 123})
logger.Warn(ctx, "警告信息", logging.KeyValue{Key: "retry_count", Value: 3})
logger.Error(ctx, "错误信息", logging.KeyValue{Key: "error", Value: "connection refused"})
```

### 收集指标数据

```go
registry := metrics.NewMeterRegistry()

// 创建多个指标
counter := registry.Counter("requests_total")
gauge := registry.Gauge("memory_usage")

// 收集所有指标
metrics := registry.Collect()
for _, m := range metrics {
    fmt.Printf("Metric: %s, Value: %f, Tags: %v\n", m.Name, m.Value, m.Tags)
}
```

---

## 使用示例

### 场景 1: HTTP 请求监控

```go
type MonitoringMiddleware struct {
    registry metrics.MeterRegistry
    logger   logging.Logger
}

func (m *MonitoringMiddleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // 记录请求
        counter := m.registry.Counter(
            "http_requests_total",
            "method", r.Method,
            "path", r.URL.Path,
        )
        counter.Inc()

        // 执行请求
        next.ServeHTTP(w, r)

        // 记录耗时
        duration := time.Since(start).Seconds()
        m.logger.Info(r.Context(), "请求完成",
            logging.KeyValue{Key: "method", Value: r.Method},
            logging.KeyValue{Key: "path", Value: r.URL.Path},
            logging.KeyValue{Key: "duration_ms", Value: duration * 1000},
        )
    })
}
```

### 场景 2: 数据库连接池监控

```go
type DBMonitor struct {
    db       *sql.DB
    registry metrics.MeterRegistry
}

func (m *DBMonitor) StartMonitoring() {
    ticker := time.NewTicker(10 * time.Second)
    go func() {
        for range ticker.C {
            stats := m.db.Stats()

            // 更新连接池指标
            m.registry.Gauge("db_connections_max").Set(float64(stats.MaxOpenConnections))
            m.registry.Gauge("db_connections_open").Set(float64(stats.OpenConnections))
            m.registry.Gauge("db_connections_in_use").Set(float64(stats.InUse))
            m.registry.Gauge("db_connections_idle").Set(float64(stats.Idle))
        }
    }()
}
```

### 场景 3: 业务指标监控

```go
type OrderMetrics struct {
    registry metrics.MeterRegistry
}

func (m *OrderMetrics) RecordOrderCreated(amount float64) {
    // 记录订单数量
    m.registry.Counter("orders_created_total").Inc()

    // 记录订单金额
    amountGauge := m.registry.Gauge("order_amount_total")
    amountGauge.Set(amountGauge.Value() + amount)
}

func (m *OrderMetrics) RecordOrderFailed(reason string) {
    // 记录失败订单
    m.registry.Counter(
        "orders_failed_total",
        "reason", reason,
    ).Inc()
}
```

---

## 最佳实践

### 1. 使用标签区分指标维度

```go
// ✅ 推荐：使用标签区分维度
counter := registry.Counter(
    "http_requests_total",
    "method", "GET",
    "status", "200",
    "path", "/api/users",
)

// ⚠️ 不推荐：不使用标签
counter := registry.Counter("http_requests_total")
```

### 2. 定期更新仪表盘指标

```go
// ✅ 推荐：定期更新仪表盘
func startMetricsCollection(registry metrics.MeterRegistry) {
    ticker := time.NewTicker(10 * time.Second)
    go func() {
        for range ticker.C {
            registry.Gauge("goroutine_count").Set(float64(runtime.NumGoroutine()))
            
            var mem runtime.MemStats
            runtime.ReadMemStats(&mem)
            registry.Gauge("memory_usage_bytes").Set(float64(mem.Alloc))
        }
    }()
}

// ⚠️ 不推荐：只设置一次
gauge := registry.Gauge("memory_usage")
gauge.Set(float64(mem.Alloc)) // 不会自动更新
```

### 3. 结构化日志记录

```go
// ✅ 推荐：使用结构化日志
logger.Info(ctx, "用户登录",
    logging.KeyValue{Key: "user_id", Value: 123},
    logging.KeyValue{Key: "ip", Value: "192.168.1.1"},
    logging.KeyValue{Key: "user_agent", Value: r.UserAgent()},
)

// ⚠️ 不推荐：使用字符串拼接
logger.Info(ctx, fmt.Sprintf("用户登录 user_id=123 ip=192.168.1.1"))
```

### 4. 与依赖注入集成

```go
// ✅ 推荐：将指标和日志组件注册为 Bean
container.Register(
    reflect.TypeOf(&metrics.MeterRegistry{}),
    core.Bean(createMeterRegistry()),
    core.Singleton(),
)

container.Register(
    reflect.TypeOf(&logging.Logger{}),
    core.Bean(createLogger()),
    core.Singleton(),
)

// 注入使用
type OrderService struct {
    Registry metrics.MeterRegistry `inject:"meterRegistry"`
    Logger   logging.Logger        `inject:"logger"`
}

func (s *OrderService) CreateOrder(order *Order) error {
    s.Registry.Counter("orders_created_total").Inc()
    s.Logger.Info(context.Background(), "订单创建",
        logging.KeyValue{Key: "order_id", Value: order.ID},
    )
    return nil
}
```

### 5. 设计原则

- **结构化日志**：支持字段化的日志记录，便于日志分析
- **线程安全**：所有组件都支持并发访问
- **可扩展**：支持自定义日志级别和指标类型
- **零外部依赖**：核心框架仅使用 Go 标准库