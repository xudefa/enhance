# Viper Starter

Viper 配置管理增强自动配置模块，提供灵活的配置管理支持。

## 功能特性

- ✅ 自动配置 Viper 实例
- ✅ 支持 YAML/JSON/TOML 等格式
- ✅ 配置文件热更新
- ✅ 环境变量覆盖
- ✅ 便捷的配置获取方法

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/viper"
)
```

### 2. 配置文件

在 `application.json` 中添加 Viper 配置：

```json
{
  "viper": {
    "enabled": true,
    "config-name": "application",
    "config-type": "yaml",
    "config-path": ".",
    "watch-changes": false
  }
}
```

### 3. 创建配置文件

创建 `application.yaml`：

```yaml
app:
  name: my-application
  version: 1.0.0
  port: 8080

database:
  host: localhost
  port: 3306
  username: root
  password: root
  name: mydb
```

### 4. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/spf13/viper"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("viper-demo"),
    )
    defer app.Stop()
    
    // 获取 Viper 实例
    v := core.MustGetBean[*viper.Viper](app.Container())
    
    // 获取配置值
    appName := v.GetString("app.name")
    appPort := v.GetInt("app.port")
    
    // 获取嵌套配置
    dbHost := v.GetString("database.host")
    dbPort := v.GetInt("database.port")
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `viper.enabled` | bool | false | 是否启用 Viper |
| `viper.config-name` | string | application | 配置文件名（不含扩展名） |
| `viper.config-type` | string | yaml | 配置文件类型（yaml/json/toml） |
| `viper.config-path` | string | . | 配置文件路径 |
| `viper.watch-changes` | bool | false | 是否监听配置文件变化 |

## 高级用法

### 环境变量覆盖

```go
v := core.MustGetBean[*viper.Viper](app.Container())

// 绑定环境变量
v.AutomaticEnv()
v.SetEnvPrefix("MYAPP")

// 现在可以通过环境变量覆盖配置
// export MYAPP_APP_NAME=my-custom-app
```

### 配置文件热更新

```json
{
  "viper": {
    "enabled": true,
    "watch-changes": true
  }
}
```

当配置文件发生变化时，Viper 会自动重新加载配置。

### 默认值设置

```go
v := core.MustGetBean[*viper.Viper](app.Container())

// 设置默认值
v.SetDefault("app.port", 8080)
v.SetDefault("database.host", "localhost")
```

## 启动顺序

- **优先级**: `OrderPriorityInfrastructure` (-4000)
- **触发条件**: `viper.enabled=true`

## 依赖

- `github.com/spf13/viper`