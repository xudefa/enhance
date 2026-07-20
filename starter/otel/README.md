# OpenTelemetry Starter

OpenTelemetry 链路追踪自动配置模块，提供分布式追踪支持。

## 功能特性

- ✅ 自动配置 OTLP 导出器
- ✅ 服务名称和版本配置
- ✅ 采样率控制
- ✅ 上下文传播
- ✅ 批量导出优化

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/otel"
)
```

### 2. 配置文件

在 `application.json` 中添加 OpenTelemetry 配置：

```json
{
  "otel": {
    "enabled": true,
    "endpoint": "localhost:4317",
    "service_name": "demo-service",
    "service_version": "1.0.0",
    "sampling_rate": 1.0
  }
}
```

### 3. 使用示例

```go
package main

import (
    "context"
    
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/trace"
    
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/otel"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("otel-demo"),
    )
    defer app.Stop()
    
    // 获取 TracerProvider
    otelConfig := core.MustGetBean[*otel.OtelAutoConfiguration](app.Container())
    tracerProvider := otelConfig.GetTracer("demo")
    tracer := tracerProvider.Tracer("demo")
    
    ctx := context.Background()
    
    // 创建 Span
    ctx, span := tracer.Start(ctx, "my-operation")
    defer span.End()
    
    // 设置属性
    span.SetAttributes(
        attribute.String("user.id", "123"),
        attribute.Int("order.count", 5),
    )
    
    // 记录事件
    span.AddEvent("processing started")
    
    // 记录错误
    span.RecordError(fmt.Errorf("something went wrong"))
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `otel.enabled` | bool | false | 是否启用 OpenTelemetry |
| `otel.endpoint` | string | localhost:4317 | OTLP 端点地址 |
| `otel.service_name` | string | enhance-app | 服务名称 |
| `otel.service_version` | string | 1.0.0 | 服务版本 |
| `otel.sampling_rate` | float64 | 1.0 | 采样率 (0.0-1.0) |

## 高级用法

### HTTP 中间件集成

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

// 包装 HTTP Handler
handler := otelhttp.NewHandler(myHandler, "my-handler")
```

### 数据库追踪

```go
import "go.opentelemetry.io/contrib/instrumentation/database/sql/otelsql"

// 包装数据库连接
db, err := otelsql.Open("mysql", dsn)
```

### 自定义导出器

```go
// 使用 Jaeger 导出器
import "go.opentelemetry.io/otel/exporters/jaeger"

exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
```

## 启动顺序

- **优先级**: `OrderPriorityMonitoringLayer` (2000)
- **触发条件**: `otel.enabled=true`

## 依赖

- `go.opentelemetry.io/otel`
- `go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`
- `go.opentelemetry.io/otel/sdk`