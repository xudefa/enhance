# go-micro Starter

go-micro 微服务框架自动配置模块，提供微服务支持。

## 功能特性

- ✅ 自动配置 go-micro 服务
- ✅ 服务注册和发现
- ✅ 多注册中心支持
- ✅ 服务版本管理
- ✅ 优雅启动和关闭

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/micro"
)
```

### 2. 配置文件

在 `application.json` 中添加 go-micro 配置：

```json
{
  "micro": {
    "enabled": true,
    "service_name": "demo-service",
    "version": "1.0.0",
    "registry_addr": ""
  }
}
```

### 3. 使用示例

```go
package main

import (
    "context"
    
    "github.com/go-micro/go-micro/v5"
    
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/micro"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("micro-demo"),
    )
    defer app.Stop()
    
    // 获取 go-micro 服务
    service := core.MustGetBean[micro.Service](app.Container())
    
    // 注册处理器
    micro.RegisterHandler(service, &Handler{})
    
    // 启动服务
    app.Start()
    app.WaitForSignal()
}

// Handler 服务处理器
type Handler struct{}

func (h *Handler) Call(ctx context.Context, req *Request, rsp *Response) error {
    rsp.Message = "Hello, " + req.Name
    return nil
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `micro.enabled` | bool | false | 是否启用 go-micro |
| `micro.service_name` | string | enhance-service | 服务名称 |
| `micro.version` | string | latest | 服务版本 |
| `micro.registry_addr` | string | "" | 注册中心地址 |

## 高级用法

### 使用 Consul 注册中心

```go
import (
    "github.com/go-micro/go-micro/v5/registry"
    "github.com/go-micro/plugins/v5/registry/consul"
)

// 配置 Consul 注册中心
registry := consul.NewRegistry(
    registry.Addrs("localhost:8500"),
)

service := micro.NewService(
    micro.Name("demo-service"),
    micro.Registry(registry),
)
```

### 服务发现

```go
import "github.com/go-micro/go-micro/v5/registry"

// 发现服务
services, err := registry.GetService("other-service")
if err != nil {
    // 处理错误
}

for _, service := range services {
    println("服务:", service.Name, "版本:", service.Version)
}
```

### RPC 调用

```go
import "github.com/go-micro/go-micro/v5/client"

// 创建客户端
c := client.NewClient()

// 调用服务
req := c.NewRequest(
    "other-service",
    "Handler.Call",
    &Request{Name: "John"},
)

rsp := &Response{}
err := c.Call(context.Background(), req, rsp)
if err != nil {
    // 处理错误
}
```

## 启动顺序

- **优先级**: `OrderPriorityBusinessLayer` (1000)
- **触发条件**: `micro.enabled=true`

## 依赖

- `github.com/go-micro/go-micro/v5`