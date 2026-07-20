# Nacos Starter

Nacos 配置中心自动配置模块，提供动态配置管理支持。

## 功能特性

- ✅ 自动配置 Nacos 客户端
- ✅ 动态配置获取
- ✅ 配置变更监听
- ✅ 配置发布支持
- ✅ 多命名空间支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/nacos"
)
```

### 2. 配置文件

在 `application.json` 中添加 Nacos 配置：

```json
{
  "nacos": {
    "enabled": true,
    "server_addr": "127.0.0.1",
    "port": 8848,
    "namespace_id": "public",
    "app_name": "demo-app",
    "username": "nacos",
    "password": "nacos",
    "timeout_ms": 10000,
    "log_dir": "/tmp/nacos/log",
    "cache_dir": "/tmp/nacos/cache",
    "log_level": "info"
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/nacos"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("nacos-demo"),
    )
    defer app.Stop()
    
    // 获取 Nacos 配置客户端
    nacosConfig := core.MustGetBean[*nacos.NacosAutoConfiguration](app.Container())
    
    // 获取配置
    content, err := nacosConfig.GetConfig("demo-config", "DEFAULT_GROUP")
    if err != nil {
        // 处理错误
    }
    println("配置内容:", content)
    
    // 监听配置变更
    err = nacosConfig.ListenConfig("demo-config", "DEFAULT_GROUP", func(newContent string) {
        println("配置已更新:", newContent)
    })
    if err != nil {
        // 处理错误
    }
    
    // 发布配置
    success, err := nacosConfig.PublishConfig("demo-config", "DEFAULT_GROUP", "new value")
    if err != nil {
        // 处理错误
    }
    
    app.WaitForSignal()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `nacos.enabled` | bool | false | 是否启用 Nacos |
| `nacos.server_addr` | string | 127.0.0.1 | Nacos 服务器地址 |
| `nacos.port` | uint64 | 8848 | Nacos 端口 |
| `nacos.namespace_id` | string | public | 命名空间 ID |
| `nacos.app_name` | string | "" | 应用名称 |
| `nacos.username` | string | "" | 用户名 |
| `nacos.password` | string | "" | 密码 |
| `nacos.timeout_ms` | int | 10000 | 超时时间（毫秒） |
| `nacos.log_dir` | string | "" | 日志目录 |
| `nacos.cache_dir` | string | "" | 缓存目录 |
| `nacos.log_level` | string | info | 日志级别 |

## 高级用法

### 多配置监听

```go
// 监听多个配置
configs := []struct{
    DataID string
    Group  string
}{
    {"db-config", "DEFAULT_GROUP"},
    {"redis-config", "DEFAULT_GROUP"},
    {"mq-config", "DEFAULT_GROUP"},
}

for _, cfg := range configs {
    nacosConfig.ListenConfig(cfg.DataID, cfg.Group, func(content string) {
        println("配置更新:", cfg.DataID, content)
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

content, _ := nacosConfig.GetConfig("db-config", "DEFAULT_GROUP")
var dbConfig DatabaseConfig
json.Unmarshal([]byte(content), &dbConfig)
```

## 启动顺序

- **优先级**: `OrderPriorityInfrastructure` (-3000)
- **触发条件**: `nacos.enabled=true`

## 依赖

- `github.com/nacos-group/nacos-sdk-go/v2`