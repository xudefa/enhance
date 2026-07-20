# exception 包 — 异常处理

> **所属层级**: Infrastructure Layer  
> **设计理念**: 全局异常处理，统一错误码  
> **设计灵感**: Spring @ControllerAdvice + Spring Error Handling

## 概述

`exception` 包提供统一的应用层异常处理机制，实现类似 Spring Boot `@ControllerAdvice` 的全局异常处理功能。

### 核心功能

| 功能 | 说明 |
|------|------|
| **全局异常处理** | 统一的异常捕获和响应机制 |
| **自定义异常映射** | 自定义异常到 HTTP 响应的映射 |
| **异常解析器链** | 优先级排序的异常解析器链 |
| **HTTP 中间件** | 自动捕获 HTTP 请求中的异常 |
| **日志和监控集成** | 与日志和监控系统无缝集成 |
| **安全模块适配** | AccessDeniedHandler 和 AuthenticationEntryPoint 适配器 |
| **统一错误码体系** | 业务错误码定义、注册和解析 |

---

## 核心接口

### ExceptionHandler 异常处理器

```go
type ExceptionHandler interface {
    Handle(ctx context.Context, err error, response ResponseWriter) *ErrorResponse
    RegisterResolver(resolver ExceptionResolver)
    RegisterHandlerFunc(handlerFunc ExceptionHandlerFunc)
}
```

### ExceptionResolver 异常解析器

```go
type ExceptionResolver interface {
    Supports(err error) bool
    Resolve(ctx context.Context, err error) *ErrorResponse
}
```

### ErrorCode 业务错误码

```go
type ErrorCode struct {
    Code    int    // HTTP 状态码
    Message string // 用户可见消息
    Detail  string // 开发者调试信息
}
```

### BusinessError 业务错误包装器

```go
type BusinessError struct {
    code    ErrorCode
    details map[string]any
}
```

---

## 快速开始

### 基本异常处理

```go
package main

import (
    "context"
    "github.com/xudefa/enhance/exception"
)

func main() {
    handler := exception.NewDefaultExceptionHandler()
    resp := handler.Handle(context.Background(), exception.ErrNotFound, response)
}
```

---

## API 参考

### HTTP 中间件

```go
handler := exception.NewDefaultExceptionHandler()
middleware := exception.ExceptionHandlingMiddleware(handler)
http.Handle("/", middleware(httpHandler))
```

### 自定义异常解析器

```go
type CustomExceptionResolver struct{}

func (r *CustomExceptionResolver) Supports(err error) bool {
    _, ok := err.(*CustomException)
    return ok
}

func (r *CustomExceptionResolver) Resolve(ctx context.Context, err error) *exception.ErrorResponse {
    return exception.NewErrorResponse().
        WithCode(400).
        WithMessage("Custom error").
        WithDetails(err.Error())
}

handler.RegisterResolver(&CustomExceptionResolver{})
```

### 业务错误码

```go
// 定义业务错误码
var ErrUserNotFound = exception.ErrorCode{
    Code:    404,
    Message: "用户不存在",
    Detail:  "user_not_found",
}

// 注册错误码
exception.RegisterErrorCode(ErrUserNotFound)

// 使用业务错误
err := exception.New(ErrUserNotFound).
    WithDetail("id", "123").
    WithDetail("type", "user")
```

### 与 ErrorCodeExceptionResolver 集成

```go
resolver := exception.NewErrorCodeExceptionResolver()
response := resolver.Resolve(ctx, err)
// response.Code == 404
// response.Message == "用户不存在"
```

---

## 预定义错误码

| 错误码 | HTTP 状态码 | 消息 | 详情 |
|--------|------------|------|------|
| `ErrCodeBadRequest` | 400 | 请求参数错误 | bad_request |
| `ErrCodeUnauthorized` | 401 | 未授权 | unauthorized |
| `ErrCodeForbidden` | 403 | 禁止访问 | forbidden |
| `ErrCodeNotFound` | 404 | 资源不存在 | not_found |
| `ErrCodeMethodNotAllowed` | 405 | 方法不允许 | method_not_allowed |
| `ErrCodeConflict` | 409 | 资源冲突 | conflict |
| `ErrCodeInternalServerError` | 500 | 服务器内部错误 | internal_server_error |
| `ErrCodeServiceUnavailable` | 503 | 服务不可用 | service_unavailable |

---

## 错误响应格式

```json
{
    "code": 404,
    "message": "用户不存在",
    "requestId": "req-123",
    "traceId": "trace-456",
    "details": "user_not_found",
    "timestamp": 1640995200
}
```

---

## 使用示例

### Web 全局异常处理

```go
func main() {
    handler := exception.NewDefaultExceptionHandler()
    
    // 注册自定义解析器
    handler.RegisterResolver(&validationExceptionResolver{})
    handler.RegisterResolver(&businessExceptionResolver{})
    
    // 创建中间件
    middleware := exception.ExceptionHandlingMiddleware(handler)
    
    // 应用到路由
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", usersHandler)
    
    http.ListenAndServe(":8080", middleware(mux))
}
```

### 业务异常处理

```go
type UserService struct{}

func (s *UserService) GetUser(id string) (*User, error) {
    user, err := s.findUser(id)
    if err != nil {
        return nil, exception.New(exception.ErrUserNotFound).
            WithDetail("id", id).
            WithDetail("type", "user")
    }
    return user, nil
}
```

### 验证异常解析器

```go
type validationExceptionResolver struct{}

func (r *validationExceptionResolver) Supports(err error) bool {
    _, ok := err.(*validation.ValidationException)
    return ok
}

func (r *validationExceptionResolver) Resolve(ctx context.Context, err error) *exception.ErrorResponse {
    ve := err.(*validation.ValidationException)
    return exception.NewErrorResponse().
        WithCode(400).
        WithMessage("Validation failed").
        WithDetails(ve.Errors)
}
```

---

## 最佳实践

### 1. 使用统一错误码体系

```go
// ✅ 推荐：定义业务错误码
var (
    ErrUserNotFound = exception.ErrorCode{
        Code:    404,
        Message: "用户不存在",
        Detail:  "user_not_found",
    }
    ErrUserAlreadyExists = exception.ErrorCode{
        Code:    409,
        Message: "用户已存在",
        Detail:  "user_already_exists",
    }
)

// ⚠️ 不推荐：硬编码错误响应
func handler(w http.ResponseWriter, r *http.Request) {
    http.Error(w, "User not found", 404)
}
```

### 2. 使用中间件统一捕获异常

```go
// ✅ 推荐：使用中间件统一处理
handler := exception.NewDefaultExceptionHandler()
middleware := exception.ExceptionHandlingMiddleware(handler)
http.ListenAndServe(":8080", middleware(mux))

// ⚠️ 不推荐：每个处理器重复捕获逻辑
func handler(w http.ResponseWriter, r *http.Request) {
    defer func() {
        if err := recover(); err != nil {
            // 手动处理异常
        }
    }()
}
```

### 3. 注册自定义异常解析器

```go
// ✅ 推荐：按优先级注册解析器
handler.RegisterResolver(&validationExceptionResolver{})
handler.RegisterResolver(&businessExceptionResolver{})
handler.RegisterResolver(&defaultExceptionResolver{})

// ⚠️ 不推荐：不注册解析器，使用默认处理
handler := exception.NewDefaultExceptionHandler()
```

### 4. 使用 BusinessError 包装业务异常

```go
// ✅ 推荐：使用 BusinessError 包装
err := exception.New(exception.ErrUserNotFound).
    WithDetail("id", id).
    WithDetail("type", "user")

// ⚠️ 不推荐：直接返回普通错误
err := fmt.Errorf("user not found: %s", id)
```

### 5. 与日志和监控集成

```go
// ✅ 推荐：记录异常日志
func (h *ExceptionHandler) Handle(ctx context.Context, err error, response ResponseWriter) *ErrorResponse {
    resp := h.resolveError(ctx, err)
    
    // 记录日志
    if resp.Code >= 500 {
        log.Error(ctx, "Internal server error", 
            log.KeyValue{Key: "error", Value: err.Error()},
            log.KeyValue{Key: "code", Value: resp.Code},
        )
    }
    
    // 上报监控指标
    metrics.IncrementCounter("http_errors", "code", strconv.Itoa(resp.Code))
    
    return resp
}
```

### 6. 设计原则

- **参考 Spring @ControllerAdvice**：借鉴 Spring 的全局异常处理设计理念
- **统一错误响应**：所有异常都转换为统一的错误响应格式
- **可扩展**：支持自定义异常解析器和处理函数
- **优先级链**：支持多个解析器按优先级处理异常
- **零外部依赖**：核心框架仅使用 Go 标准库