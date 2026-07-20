# resilience 包 — 弹性与容错

> **所属层级**: Infrastructure Layer  
> **设计理念**: 服务弹性，负载均衡，熔断容错  
> **设计灵感**: Spring Cloud + Netflix Hystrix + Resilience4j

## 概述

`resilience` 包提供服务弹性、负载均衡和容错能力，确保分布式系统的高可用性。

### 核心功能

| 功能 | 说明 |
|------|------|
| **服务注册与发现** | 抽象注册中心接口，支持服务实例注册、注销、发现和监听 |
| **负载均衡** | 10+ 种负载均衡策略（随机、轮询、加权、最少连接、一致性哈希等） |
| **熔断器** | 熔断器模式实现，支持状态转换（Closed/Open/Half-Open）和故障恢复 |

---

## 核心接口

### 服务注册与发现

#### Registry 接口

```go
type Registry interface {
    Register(ctx context.Context, info InstanceInfo) error
    Deregister(ctx context.Context, info InstanceInfo) error
    Discover(ctx context.Context, serviceName string) ([]InstanceInfo, error)
    Watch(ctx context.Context, serviceName string) (<-chan []InstanceInfo, error)
}
```

#### InstanceInfo 结构体

| 字段 | 类型 | 说明 |
|------|------|------|
| `ServiceName` | `string` | 服务名称 |
| `ID` | `string` | 实例唯一标识 |
| `Host` | `string` | 实例主机地址 |
| `Port` | `int` | 实例端口号 |
| `Weight` | `int` | 权重，用于加权选择策略 |
| `Healthy` | `bool` | 健康状态 |
| `Metadata` | `map[string]string` | 附加元数据（版本号、区域等） |

#### 使用示例

```go
var reg resilience.Registry // 具体实现由注册中心提供（如 etcd、consul、nacos）

// 注册实例
err := reg.Register(ctx, resilience.InstanceInfo{
    ServiceName: "user-service",
    ID:          "192.168.1.1:8080",
    Host:        "192.168.1.1",
    Port:        8080,
    Weight:      10,
    Healthy:     true,
    Metadata:    map[string]string{"version": "v1.0"},
})

// 发现服务
instances, err := reg.Discover(ctx, "user-service")

// 监听变化
ch, err := reg.Watch(ctx, "user-service")
for instances := range ch {
    // 处理实例变更
}
```

### 负载均衡

#### Selector 接口

```go
type Selector interface {
    Select(instances []*InstanceInfo) (*InstanceInfo, error)
}
```

#### 内置负载均衡策略

| 策略 | 说明 | 适用场景 |
|------|------|----------|
| `RandomSelector` | 随机选择 | 简单无状态场景 |
| `RoundRobinSelector` | 轮询选择 | 后端性能相近 |
| `WeightedRandomSelector` | 加权随机 | 后端性能不均 |
| `WeightedRoundRobinSelector` | 加权轮询 | 后端性能不均且需均匀分布 |
| `LeastConnSelector` | 最少连接 | 长连接场景 |
| `ResponseTimeSelector` | 响应时间优先 | 追求最低延迟 |
| `AdaptiveSelector` | 自适应权重 | 动态调整负载 |
| `HashSelector` | 一致性哈希 | 缓存亲和性 |
| `StickySelector` | 粘性会话 | 会话保持 |
| `HealthSelector` | 健康状态优先 | 高可用场景 |
| `IPHashSelector` | IP 哈希 | 客户端亲和性 |

#### 使用示例

```go
// 轮询选择
selector := resilience.NewRoundRobinSelector()
selected, err := selector.Select(instances)

// 加权随机
selector := resilience.NewWeightedRandomSelector()
selected, err := selector.Select(instances)

// 最少连接
selector := resilience.NewLeastConnSelector()
selected, err := selector.Select(instances)
```

### 熔断器

#### CircuitBreaker 接口

```go
type CircuitBreaker interface {
    Execute(fn func() (any, error)) (any, error)
    IsOpen() bool
    RecordSuccess()
    RecordFailure()
}
```

#### 熔断器状态

| 状态 | 常量 | 说明 |
|------|------|------|
| Closed | `StateClosed` | 关闭状态，正常处理请求 |
| Open | `StateOpen` | 打开状态，快速失败 |
| HalfOpen | `StateHalfOpen` | 半开状态，尝试恢复 |

#### 状态转换

```
Closed ──(失败率超过阈值)──> Open
Open ──(等待时间过后)──> HalfOpen
HalfOpen ──(请求成功)──> Closed
HalfOpen ──(请求失败)──> Open
```

#### 使用示例

```go
// 基本使用
breaker := resilience.NewCircuitBreaker(
    resilience.WithFailureThreshold(5),
    resilience.WithSuccessThreshold(3),
    resilience.WithTimeout(30 * time.Second),
)

result, err := breaker.Execute(func() (any, error) {
    return callExternalService()
})

// 带降级逻辑
breaker := resilience.NewCircuitBreaker(
    resilience.WithFailureThreshold(5),
    resilience.WithFallback(func() (any, error) {
        // 降级逻辑
        return cachedResult, nil
    }),
)
```

---

## 快速开始

### 服务注册与发现

```go
package main

import (
    "context"
    "fmt"
    "github.com/xudefa/enhance/resilience"
)

func main() {
    // 创建注册中心（以内存实现为例）
    reg := resilience.NewInMemoryRegistry()

    // 注册服务实例
    reg.Register(context.Background(), resilience.InstanceInfo{
        ServiceName: "user-service",
        ID:          "instance-1",
        Host:        "192.168.1.1",
        Port:        8080,
        Weight:      10,
        Healthy:     true,
    })

    // 发现服务
    instances, _ := reg.Discover(context.Background(), "user-service")
    for _, instance := range instances {
        fmt.Printf("Found instance: %s at %s:%d\n", instance.ID, instance.Host, instance.Port)
    }
}
```

---

## API 参考

### 负载均衡

```go
// 创建服务实例列表
instances := []*resilience.InstanceInfo{
    {ID: "instance-1", Host: "192.168.1.1", Port: 8080, Weight: 10, Healthy: true},
    {ID: "instance-2", Host: "192.168.1.2", Port: 8080, Weight: 20, Healthy: true},
    {ID: "instance-3", Host: "192.168.1.3", Port: 8080, Weight: 15, Healthy: true},
}

// 使用加权随机选择器
selector := resilience.NewWeightedRandomSelector()
selected, err := selector.Select(instances)
if err != nil {
    // 处理错误
}
fmt.Printf("Selected instance: %s\n", selected.ID)
```

### 熔断器

```go
// 创建熔断器
breaker := resilience.NewCircuitBreaker(
    resilience.WithFailureThreshold(5),
    resilience.WithSuccessThreshold(3),
    resilience.WithTimeout(30 * time.Second),
    resilience.WithFallback(func() (any, error) {
        return "fallback response", nil
    }),
)

// 执行受保护的调用
result, err := breaker.Execute(func() (any, error) {
    return http.Get("https://api.example.com/data")
})

if err != nil {
    // 处理错误（可能是熔断器打开）
}
```

---

## 使用示例

### 场景 1: 微服务调用链

```go
type OrderService struct {
    registry   resilience.Registry
    selector   resilience.Selector
    breaker    resilience.CircuitBreaker
}

func (s *OrderService) CreateOrder(ctx context.Context, order *Order) error {
    // 发现库存服务
    instances, err := s.registry.Discover(ctx, "inventory-service")
    if err != nil {
        return err
    }

    // 选择实例
    instance, err := s.selector.Select(instances)
    if err != nil {
        return err
    }

    // 使用熔断器调用库存服务
    _, err = s.breaker.Execute(func() (any, error) {
        return http.Post(fmt.Sprintf("http://%s:%d/reserve", instance.Host, instance.Port), "application/json", order)
    })

    return err
}
```

### 场景 2: 动态服务发现与负载均衡

```go
func (s *GatewayService) StartLoadBalancer() {
    // 监听服务变化
    ch, _ := s.registry.Watch(context.Background(), "user-service")
    
    go func() {
        for instances := range ch {
            // 更新本地实例列表
            s.updateInstances(instances)
        }
    }()
}

func (s *GatewayService) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // 选择实例
    instance, err := s.selector.Select(s.instances)
    if err != nil {
        http.Error(w, "No available instances", 503)
        return
    }

    // 转发请求
    proxy := httputil.NewSingleHostReverseProxy(&url.URL{
        Scheme: "http",
        Host:   fmt.Sprintf("%s:%d", instance.Host, instance.Port),
    })
    proxy.ServeHTTP(w, r)
}
```

---

## 最佳实践

### 1. 选择合适的负载均衡策略

```go
// ✅ 推荐：根据场景选择策略
// 后端性能相近：使用轮询
selector := resilience.NewRoundRobinSelector()

// 后端性能不均：使用加权随机
selector := resilience.NewWeightedRandomSelector()

// 长连接场景：使用最少连接
selector := resilience.NewLeastConnSelector()

// 缓存亲和性：使用一致性哈希
selector := resilience.NewHashSelector()

// ⚠️ 不推荐：不考虑场景随意选择
selector := resilience.NewRandomSelector() // 可能导致负载不均
```

### 2. 熔断器配置建议

```go
// ✅ 推荐：合理配置熔断器
breaker := resilience.NewCircuitBreaker(
    resilience.WithFailureThreshold(5),      // 失败 5 次后打开
    resilience.WithSuccessThreshold(3),      // 成功 3 次后关闭
    resilience.WithTimeout(30*time.Second),  // 打开 30 秒后尝试半开
    resilience.WithFallback(fallbackFunc),   // 提供降级逻辑
)

// ⚠️ 不推荐：不配置降级逻辑
breaker := resilience.NewCircuitBreaker(
    resilience.WithFailureThreshold(5),
    // 没有 Fallback，熔断后直接返回错误
)
```

### 3. 服务健康检查

```go
// ✅ 推荐：定期检查服务健康状态
func (s *ServiceMonitor) StartHealthCheck() {
    ticker := time.NewTicker(10 * time.Second)
    go func() {
        for range ticker.C {
            for _, instance := range s.instances {
                if !s.checkHealth(instance) {
                    instance.Healthy = false
                    s.registry.Deregister(context.Background(), instance)
                }
            }
        }
    }()
}

// ⚠️ 不推荐：不检查健康状态
// 可能导致请求发送到不健康的实例
```

### 4. 与依赖注入集成

```go
// ✅ 推荐：将弹性组件注册为 Bean
container.Register(
    reflect.TypeOf(&resilience.CircuitBreaker{}),
    core.Bean(createCircuitBreaker()),
    core.Singleton(),
)

container.Register(
    reflect.TypeOf(&resilience.Selector{}),
    core.Bean(createSelector()),
    core.Singleton(),
)

// 注入使用
type OrderService struct {
    Breaker  resilience.CircuitBreaker `inject:"circuitBreaker"`
    Selector resilience.Selector       `inject:"selector"`
}
```

### 5. 设计要点

- 参考 Spring Cloud 设计理念
- 接口抽象，易于扩展
- 零外部依赖（仅使用 Go 标准库）
- 并发安全设计