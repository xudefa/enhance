# Apollo Starter

Apollo 配置中心自动配置模块，提供携程配置中心支持。

## 功能特性

- ✅ 自动配置 Apollo 客户端
- ✅ 动态配置获取
- ✅ 配置变更监听
- ✅ 多集群支持
- ✅ 本地缓存支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/apollo"
)
```

### 2. 配置文件

在 `application.json` 中添加 Apollo 配置：

```json
{
  "apollo": {
    "enabled": true,
    "app_id": "demo-app",
    "cluster": "default",
    "meta_addr": "http://localhost:8080",
    "namespace": "application",
    "is_backup_config": true,
    "secret": ""
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/apollo"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("apollo-demo"),
    )
    defer app.Stop()
    
    // 获取 Apollo 配置客户端
    apolloConfig := core.MustGetBean[*apollo.ApolloAutoConfiguration](app.Container())
    
    // 获取配置
    value, err := apolloConfig.GetConfig("demo.key", "application")
    if err != nil {
        // 处理错误
    }
    println("配置值:", value)
    
    // 监听配置变更
    apolloConfig.WatchConfig("application", func(event *config.ChangeEvent) {
        for key, value := range event.Changes {
            println("配置变更:", key, value.NewValue)
        }
    })
    
    app.WaitForSignal()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `apollo.enabled` | bool | false | 是否启用 Apollo |
| `apollo.app_id` | string | "" | 应用 ID |
| `apollo.cluster` | string | default | 集群名称 |
| `apollo.meta_addr` | string | "" | Apollo Meta 服务地址 |
| `apollo.namespace` | string | application | 命名空间 |
| `apollo.is_backup_config` | bool | true | 是否启用本地缓存 |
| `apollo.secret` | string | "" | 访问密钥 |

## 高级用法

### 多命名空间

```go
// 监听多个命名空间
namespaces := []string{"application", "database", "redis"}

for _, ns := range namespaces {
    apolloConfig.WatchConfig(ns, func(event *config.ChangeEvent) {
        println("命名空间变更:", ns)
        for key, value := range event.Changes {
            println("  ", key, value.NewValue)
        }
    })
}
```

### 配置解析为结构体

```go
import "encoding/json"

type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
}

content, _ := apolloConfig.GetConfig("db.host", "application")
var dbConfig DatabaseConfig
// 解析配置
```

## 启动顺序

- **优先级**: `OrderPriorityInfrastructure` (-3000)
- **触发条件**: `apollo.enabled=true`

## 依赖

- `github.com/apolloconfig/agollo/v4`