# Cobra Starter

Cobra CLI 框架自动配置模块，提供命令行应用支持。

## 功能特性

- ✅ 自动配置 Cobra CLI
- ✅ 子命令支持
- ✅ 标志位解析
- ✅ 自动生成帮助文档
- ✅ 版本管理

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/cobra"
)
```

### 2. 配置文件

在 `application.json` 中添加 Cobra 配置：

```json
{
  "cobra": {
    "enabled": true,
    "use": "mycli",
    "short": "My CLI Application",
    "long": "A powerful CLI application built with enhance",
    "version": "1.0.0"
  }
}
```

### 3. 使用示例

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/spf13/cobra"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("cobra-demo"),
    )
    defer app.Stop()
    
    // 获取根命令
    rootCmd := core.MustGetBean[*cobra.Command](app.Container())
    
    // 添加子命令
    rootCmd.AddCommand(&cobra.Command{
        Use:   "start",
        Short: "Start the application",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Println("Starting application...")
        },
    })
    
    // 添加带标志的命令
    rootCmd.AddCommand(&cobra.Command{
        Use:   "deploy",
        Short: "Deploy the application",
        Run: func(cmd *cobra.Command, args []string) {
            env, _ := cmd.Flags().GetString("env")
            fmt.Printf("Deploying to %s...\n", env)
        },
    })
    
    // 执行
    app.Start()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `cobra.enabled` | bool | false | 是否启用 Cobra |
| `cobra.use` | string | app | 命令名称 |
| `cobra.short` | string | A CLI application | 简短描述 |
| `cobra.long` | string | A CLI application built with enhance framework | 详细描述 |
| `cobra.version` | string | 1.0.0 | 版本号 |

## 高级用法

### 标志位

```go
cmd := &cobra.Command{
    Use:   "server",
    Short: "Start the server",
    Run: func(cmd *cobra.Command, args []string) {
        port, _ := cmd.Flags().GetInt("port")
        debug, _ := cmd.Flags().GetBool("debug")
        fmt.Printf("Starting server on port %d (debug=%v)\n", port, debug)
    },
}

// 添加标志
cmd.Flags().IntP("port", "p", 8080, "Server port")
cmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
```

### 持久标志

```go
// 持久标志对所有子命令可用
rootCmd.PersistentFlags().String("config", "", "Config file path")
```

### 参数验证

```go
cmd := &cobra.Command{
    Use:   "create [name]",
    Short: "Create a new resource",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        name := args[0]
        fmt.Printf("Creating %s...\n", name)
    },
}
```

## 启动顺序

- **优先级**: `OrderPriorityInfrastructure` (-4000)
- **触发条件**: `cobra.enabled=true`

## 依赖

- `github.com/spf13/cobra`