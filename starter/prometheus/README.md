# Prometheus Starter

Prometheus 监控指标自动配置模块，提供应用性能监控支持。

## 功能特性

- ✅ 自动配置 Prometheus 指标收集
- ✅ 支持 OpenMetrics 格式
- ✅ 自定义指标创建（Counter/Gauge/Histogram）
- ✅ 独立 HTTP 服务器暴露指标
- ✅ 可配置指标路径

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/prometheus"
)
```

### 2. 配置文件

在 `application.json` 中添加 Prometheus 配置：

```json
{
  "prometheus": {
    "enabled": true,
    "host": "0.0.0.0",
    "port": 9090,
    "metrics_path": "/metrics",
    "enable_open_metrics": false
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/prometheus"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("prometheus-demo"),
    )
    defer app.Stop()
    
    // 获取 Prometheus 配置器
    prom := core.MustGetBean[*prometheus.PrometheusAutoConfiguration](app.Container())
    
    // 创建 Counter 指标
    requestCount := prom.NewCounter(
        "http_requests_total",
        "Total number of HTTP requests",
    )
    requestCount.Inc()
    
    // 创建 Gauge 指标
    activeUsers := prom.NewGauge(
        "active_users",
        "Number of active users",
    )
    activeUsers.Set(100)
    
    // 创建 Histogram 指标
    requestDuration := prom.NewHistogram(
        "http_request_duration_seconds",
        "HTTP request duration in seconds",
        []float64{0.1, 0.5, 1.0, 2.0, 5.0},
    )
    requestDuration.Observe(0.5)
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `prometheus.enabled` | bool | false | 是否启用 Prometheus |
| `prometheus.host` | string | 0.0.0.0 | 服务器监听地址 |
| `prometheus.port` | int | 9090 | 服务器端口 |
| `prometheus.metrics_path` | string | /metrics | 指标暴露路径 |
| `prometheus.enable_open_metrics` | bool | false | 是否启用 OpenMetrics 格式 |

## 高级用法

### HTTP 中间件集成

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// 自动创建指标
var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
)

// 在中间件中使用
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rw := &responseWriter{w, http.StatusOK}
        next.ServeHTTP(rw, r)
        httpRequestsTotal.WithLabelValues(
            r.Method,
            r.URL.Path,
            strconv.Itoa(rw.statusCode),
        ).Inc()
    })
}
```

### 自定义 Registry

```go
prom := core.MustGetBean[*prometheus.PrometheusAutoConfiguration](app.Container())
registry := prom.GetRegistry()

// 注册自定义收集器
registry.MustRegister(myCustomCollector)
```

## 启动顺序

- **优先级**: `OrderPriorityMonitoringLayer` (-1000)
- **触发条件**: `prometheus.enabled=true`

## 依赖

- `github.com/prometheus/client_golang`