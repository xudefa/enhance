# JWT Starter

JWT 认证自动配置模块，提供 Token 生成和验证支持。

## 功能特性

- ✅ 自动配置 JWT 认证
- ✅ Token 生成和验证
- ✅ 自定义 Claims 支持
- ✅ Token 刷新机制
- ✅ HTTP 中间件集成

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/jwt"
)
```

### 2. 配置文件

在 `application.json` 中添加 JWT 配置：

```json
{
  "jwt": {
    "enabled": true,
    "secret": "your-secret-key",
    "expire": 3600,
    "refresh_expire": 86400,
    "issuer": "enhance-app",
    "header": "Authorization",
    "prefix": "Bearer "
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/jwt"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("jwt-demo"),
    )
    defer app.Stop()
    
    // 获取 JWT Provider
    provider := core.MustGetBean[*jwt.JWTTokenProvider](app.Container())
    
    // 生成 Token
    claims := jwt.NewClaims("user-123", "admin")
    token, err := provider.Generate(claims)
    if err != nil {
        // 处理错误
    }
    
    // 验证 Token
    claims, err = provider.Validate(token)
    if err != nil {
        // 处理错误
    }
    
    // 刷新 Token
    newToken, err := provider.Refresh(token)
    if err != nil {
        // 处理错误
    }
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `jwt.enabled` | bool | false | 是否启用 JWT |
| `jwt.secret` | string | "" | 签名密钥 |
| `jwt.expire` | int | 3600 | Token 过期时间（秒） |
| `jwt.refresh_expire` | int | 86400 | 刷新 Token 过期时间（秒） |
| `jwt.issuer` | string | enhance-app | 签发者 |
| `jwt.header` | string | Authorization | HTTP Header 名称 |
| `jwt.prefix` | string | Bearer | Token 前缀 |

## 高级用法

### HTTP 中间件

```go
import (
    "github.com/xudefa/enhance/starter/jwt"
    "github.com/labstack/echo/v4"
)

// 在 Echo 中使用 JWT 中间件
e := echo.New()
e.Use(jwt.JWTMiddleware())

// 受保护的路由
e.GET("/protected", func(c echo.Context) error {
    claims := jwt.GetClaims(c)
    return c.JSON(200, claims)
})

// 公开路由
e.GET("/public", func(c echo.Context) error {
    return c.String(200, "Public")
})
```

### 自定义 Claims

```go
type CustomClaims struct {
    jwt.Claims
    Department string `json:"department"`
    Role       string `json:"role"`
}

// 生成带自定义 Claims 的 Token
claims := &CustomClaims{
    Claims: jwt.NewClaims("user-123", "admin"),
    Department: "Engineering",
    Role: "Developer",
}
token, _ := provider.Generate(claims)
```

### Token 黑名单

```go
// 将 Token 加入黑名单
provider.Blacklist(token)

// 检查 Token 是否在黑名单中
if provider.IsBlacklisted(token) {
    // Token 已失效
}
```

## 启动顺序

- **优先级**: `OrderPriorityWebLayer` (0)
- **触发条件**: `jwt.enabled=true`

## 依赖

- `github.com/golang-jwt/jwt/v5`