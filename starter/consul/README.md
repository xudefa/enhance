# Consul Starter

Consul 服务发现自动配置模块，提供服务注册与发现支持。

## 功能特性

- ✅ 自动配置 Consul 客户端
- ✅ 服务注册与发现
- ✅ 健康检查
- ✅ KV 存储支持
- ✅ 多数据中心支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/consul"
)
```

### 2. 配置文件

在 `application.json` 中添加 Consul 配置：

```json
{
  "consul": {
    "enabled": true,
    "host": "localhost",
    "port": 8500,
    "token": ""
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    consulapi "github.com/hashicorp/consul/api"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("consul-demo"),
    )
    defer app.Stop()
    
    // 获取 Consul 配置器
    consul := core.MustGetBean[*consul.ConsulAutoConfiguration](app.Container())
    
    // 注册服务
    consul.RegisterService(&consulapi.AgentServiceRegistration{
        ID:      "web-1",
        Name:    "web",
        Address: "192.168.1.100",
        Port:    8080,
        Check: &consulapi.AgentServiceCheck{
            HTTP:     "http://192.168.1.100:8080/health",
            Interval: "10s",
            Timeout:  "5s",
        },
    })
    
    // 发现服务
    services, err := consul.GetHealthyServices("web")
    if err != nil {
        // 处理错误
    }
    
    for _, service := range services {
        println(service.Service.Address, service.Service.Port)
    }
    
    // 注销服务
    consul.DeregisterService("web-1")
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `consul.enabled` | bool | false | 是否启用 Consul |
| `consul.host` | string | localhost | Consul 服务器地址 |
| `consul.port` | int | 8500 | Consul 服务器端口 |
| `consul.token` | string | "" | 访问令牌 |

## 高级用法

### KV 存储

```go
consul := core.MustGetBean[*consul.ConsulAutoConfiguration](app.Container())
client := consul.GetClient()

// 写入 KV
kv := &consulapi.KVPair{
    Key:   "config/app/name",
    Value: []byte("my-app"),
}
client.KV().Put(kv, nil)

// 读取 KV
pair, _, err := client.KV().Get("config/app/name", nil)
if err != nil {
    // 处理错误
}
println(string(pair.Value))
```

### 健康检查

```go
// 注册带健康检查的服务
consul.RegisterService(&consulapi.AgentServiceRegistration{
    ID:      "api-1",
    Name:    "api",
    Address: "192.168.1.100",
    Port:    3000,
    Checks: consulapi.AgentServiceChecks{
        {
            HTTP:     "http://192.168.1.100:3000/health",
            Interval: "10s",
            Timeout:  "5s",
        },
        {
            TCP:      "192.168.1.100:3000",
            Interval: "10s",
            Timeout:  "5s",
        },
    },
})
```

## 启动顺序

- **优先级**: `OrderPriorityServiceDiscovery` (-1500)
- **触发条件**: `consul.enabled=true`

## 依赖

- `github.com/hashicorp/consul/api`