# admin 包 — 应用管理监控

> **所属层级**: Infrastructure Layer  
> **设计理念**: Spring Boot Admin 风格，集中式监控

## 概述

`admin` 包位于 Infrastructure Layer，提供 Spring Boot Admin 风格的可视化监控功能，支持应用注册、状态监控、实例管理、健康检查等。适用于需要集中管理多个应用实例的场景。

### 核心功能

| 功能 | 说明 |
|------|------|
| **应用注册中心** | 支持应用实例的注册与注销 |
| **健康监控** | 实时查看应用和实例的健康状态 |
| **指标收集** | 支持自定义指标的上报与查询 |
| **HTTP API** | 提供 RESTful API 用于查询应用状态 |
| **状态聚合** | 自动聚合多实例状态计算应用整体状态 |

---

## Application 应用对象

```go
type Application struct {
    ID        string
    Name      string
    Version   string
    Instances []*ApplicationInstance
    Status    ApplicationStatus
    Metadata  map[string]string
}
```

### ApplicationStatus 应用状态

| 常量 | 值 | 说明 |
|------|----|------|
| `StatusUp` | `"UP"` | 运行中 |
| `StatusDown` | `"DOWN"` | 已停止 |
| `StatusUnknown` | `"UNKNOWN"` | 未知 |
| `StatusOutOfService` | `"OUT_OF_SERVICE"` | 服务不可用 |

---

## ApplicationInstance 应用实例

```go
type ApplicationInstance struct {
    ID            string
    ApplicationID string
    Name          string
    URL           string
    Status        ApplicationStatus
    Health        *HealthInfo
    Metrics       map[string]float64
    Metadata      map[string]string
    RegisteredAt  time.Time
    LastSeen      time.Time
}
```

### 创建实例

## admin.NewApplicationInstance

创建应用实例。

```go
instance := admin.NewApplicationInstance("my-app", "http://localhost:8080")
```

---

## HealthInfo 健康信息

```go
type HealthInfo struct {
    Status     ApplicationStatus
    Components map[string]*ComponentHealth
    Details    map[string]interface{}
}
```

### 创建健康信息

## admin.NewHealthInfo

创建健康信息对象。

```go
health := admin.NewHealthInfo(admin.StatusUp)
health.AddComponent("database", admin.StatusUp, map[string]interface{}{
    "latency": "5ms",
})
health.AddDetail("version", "1.0.0")
```

---

## ApplicationRegistry 应用注册中心

```go
type ApplicationRegistry struct {
    // ...
}
```

### 创建

## admin.NewApplicationRegistry

创建应用注册中心。

```go
registry := admin.NewApplicationRegistry()
```

### 注册与注销

## admin.ApplicationRegistry.Register

注册应用实例。

```go
instance := admin.NewApplicationInstance("my-app", "http://localhost:8080")
registry.Register(instance)
```

## admin.ApplicationRegistry.Deregister

注销应用实例。

```go
registry.Deregister(instanceID)
```

### 查询方法

## admin.ApplicationRegistry.GetApplication

获取应用信息。

```go
app, err := registry.GetApplication("my-app")
```

## admin.ApplicationRegistry.GetInstance

获取实例信息。

```go
instance, err := registry.GetInstance("my-app-123456")
```

## admin.ApplicationRegistry.ListApplications

列出所有应用。

```go
apps := registry.ListApplications()
```

## admin.ApplicationRegistry.ListInstances

列出所有实例。

```go
instances := registry.ListInstances()
```

### 更新状态

## admin.ApplicationRegistry.UpdateHealth

更新实例健康信息。

```go
registry.UpdateHealth(instanceID, healthInfo)
```

## admin.ApplicationRegistry.UpdateMetrics

更新实例指标信息。

```go
registry.UpdateMetrics(instanceID, map[string]float64{
    "cpu_usage": 45.2,
    "memory_mb": 512.0,
})
```

### 统计方法

| 方法 | 说明 |
|------|------|
| `GetApplicationCount()` | 获取应用数量 |
| `GetInstanceCount()` | 获取实例数量 |
| `GetUpInstanceCount()` | 获取运行中的实例数量 |
| `GetDownInstanceCount()` | 获取停止的实例数量 |

---

## AdminServer Admin 服务器

```go
type AdminServer struct {
    // ...
}
```

### 创建

## admin.NewAdminServer

创建 Admin 服务器。

```go
server := admin.NewAdminServer(registry)
http.Handle("/admin", server.Handler())
http.ListenAndServe(":9090", nil)
```

### API 端点

| 路径 | 方法 | 说明 |
|------|------|------|
| `/admin/applications` | GET | 列出所有应用 |
| `/admin/applications?id=xxx` | GET | 获取指定应用 |
| `/admin/instances` | GET | 列出所有实例 |
| `/admin/instances?id=xxx` | GET | 获取指定实例 |
| `/admin/instances/{id}/health` | GET | 获取实例健康 |
| `/admin/instances/{id}/metrics` | GET | 获取实例指标 |
| `/admin/health` | GET | 获取整体健康状态 |
| `/admin/register` | POST | 注册实例 |
| `/admin/deregister` | POST | 注销实例 |

---

## 使用示例

### 基本使用

```go
package main

import (
    "net/http"
    "github.com/xudefa/enhance/admin"
)

func main() {
    // 创建注册中心
    registry := admin.NewApplicationRegistry()

    // 创建 Admin 服务器
    server := admin.NewAdminServer(registry)

    // 启动 HTTP 服务
    http.Handle("/admin", server.Handler())
    http.ListenAndServe(":9090", nil)
}
```

### 注册应用实例

```go
registry := admin.NewApplicationRegistry()

// 创建并注册实例
instance1 := admin.NewApplicationInstance("user-service", "http://localhost:8081")
instance2 := admin.NewApplicationInstance("user-service", "http://localhost:8082")

registry.Register(instance1)
registry.Register(instance2)

// 查看应用数量
fmt.Println("应用数量:", registry.GetApplicationCount()) // 1
fmt.Println("实例数量:", registry.GetInstanceCount())   // 2
```

### 更新健康状态

```go
// 创建健康信息
health := admin.NewHealthInfo(admin.StatusUp)
health.AddComponent("database", admin.StatusUp, map[string]interface{}{
    "connections": 10,
    "latency":     "5ms",
})
health.AddComponent("cache", admin.StatusUp, map[string]interface{}{
    "hit_rate": 0.95,
})

// 更新实例健康
registry.UpdateHealth(instance1.ID, health)
```

### 更新指标

```go
metrics := map[string]float64{
    "cpu_usage":    45.2,
    "memory_mb":    512.0,
    "goroutines":   128.0,
    "requests_per_sec": 1500.0,
}

registry.UpdateMetrics(instance1.ID, metrics)
```

### 客户端注册

```bash
# 注册实例
curl -X POST http://localhost:9090/admin/register \
  -H "Content-Type: application/json" \
  -d '{
    "id": "user-service-1",
    "application_id": "user-service",
    "name": "user-service",
    "url": "http://localhost:8081"
  }'

# 查询应用列表
curl http://localhost:9090/admin/applications

# 查询整体健康状态
curl http://localhost:9090/admin/health
```

---

## 使用场景

### 场景 1: 微服务监控

**描述**: 集中监控多个微服务实例的健康状态和性能指标。

```go
registry := admin.NewApplicationRegistry()

// 注册多个服务实例
services := []string{"user-service", "order-service", "payment-service"}
for _, service := range services {
    for i := 1; i <= 3; i++ {
        instance := admin.NewApplicationInstance(
            service,
            fmt.Sprintf("http://localhost:%d", 8080+i),
        )
        registry.Register(instance)
    }
}

// 定期更新健康状态
go func() {
    for {
        for _, instance := range registry.ListInstances() {
            health := checkServiceHealth(instance.URL)
            registry.UpdateHealth(instance.ID, health)
        }
        time.Sleep(10 * time.Second)
    }
}()
```

**最佳实践**:
- 定期更新实例健康状态
- 设置合理的检查间隔
- 使用 LastSeen 时间检测失联实例

### 场景 2: 应用注册中心

**描述**: 作为应用注册中心,管理服务实例的注册与发现。

```go
registry := admin.NewApplicationRegistry()

// 应用启动时注册
instance := admin.NewApplicationInstance("my-app", "http://localhost:8080")
registry.Register(instance)

// 应用关闭时注销
defer registry.Deregister(instance.ID)

// 服务发现
func getServiceURL(serviceName string) (string, error) {
    apps := registry.ListApplications()
    for _, app := range apps {
        if app.Name == serviceName && len(app.Instances) > 0 {
            return app.Instances[0].URL, nil
        }
    }
    return "", fmt.Errorf("service %s not found", serviceName)
}
```

**最佳实践**:
- 应用启动时自动注册
- 应用关闭时自动注销
- 定期检查实例存活状态

### 场景 3: 健康检查聚合

**描述**: 聚合多个组件的健康状态,提供整体健康视图。

```go
func buildHealthInfo() *admin.HealthInfo {
    health := admin.NewHealthInfo(admin.StatusUp)

    // 数据库健康
    if err := checkDatabase(); err != nil {
        health.AddComponent("database", admin.StatusDown, map[string]interface{}{
            "error": err.Error(),
        })
    } else {
        health.AddComponent("database", admin.StatusUp, nil)
    }

    // Redis 健康
    if err := checkRedis(); err != nil {
        health.AddComponent("redis", admin.StatusDown, map[string]interface{}{
            "error": err.Error(),
        })
    } else {
        health.AddComponent("redis", admin.StatusUp, nil)
    }

    return health
}
```

**最佳实践**:
- 包含所有关键组件的健康状态
- 提供详细的错误信息
- 使用组件级别的状态聚合

---

## 设计要点

- `ApplicationRegistry` 使用 `sync.RWMutex` 保证并发安全
- 应用状态根据实例状态自动聚合计算
- `AdminServer` 提供标准 RESTful API
- 支持动态注册和注销实例
- 零外部依赖,仅使用 Go 标准库