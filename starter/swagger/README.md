# Swagger Starter

Swagger API 文档自动配置模块，提供交互式 API 文档支持。

## 功能特性

- ✅ 自动配置 Swagger UI
- ✅ 支持 OpenAPI 规范
- ✅ 交互式 API 文档
- ✅ API 测试支持
- ✅ 自动生成文档

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/swagger"
)
```

### 2. 配置文件

在 `application.json` 中添加 Swagger 配置：

```json
{
  "swagger": {
    "enabled": true,
    "host": "localhost",
    "port": 8080,
    "url": "/swagger/*"
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/swagger"
    "github.com/gin-gonic/gin"
)

// @title Enhance API
// @version 1.0
// @description This is a sample server.
// @host localhost:8080
// @BasePath /api/v1
func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("swagger-demo"),
    )
    defer app.Stop()
    
    engine := gin.Default()
    
    // 获取 Swagger 配置器
    swagger := core.MustGetBean[*swagger.SwaggerAutoConfiguration](app.Container())
    
    // 注册 Swagger UI 路由
    engine.GET("/swagger/*", gin.WrapH(swagger.GetHandler()))
    
    app.Start()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `swagger.enabled` | bool | false | 是否启用 Swagger |
| `swagger.host` | string | localhost | 服务器地址 |
| `swagger.port` | int | 8080 | 服务器端口 |
| `swagger.url` | string | /swagger/* | Swagger UI 路径 |
| `swagger.title` | string | Enhance API | API 标题 |

## 高级用法

### API 注释

```go
// GetUser 获取用户信息
// @Summary 获取用户信息
// @Description 根据用户 ID 获取用户详细信息
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} User
// @Failure 404 {object} Error
// @Router /users/{id} [get]
func GetUser(c *gin.Context) {
    // 实现
}
```

### 生成文档

```bash
# 安装 swag 工具
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
swag init -g main.go
```

## 启动顺序

- **优先级**: `OrderPriorityWebLayer` (0)
- **触发条件**: `swagger.enabled=true`

## 依赖

- `github.com/swaggo/http-swagger`
- `github.com/swaggo/swag`