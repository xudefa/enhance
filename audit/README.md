# audit 包 — 审计日志

> **所属层级**: Infrastructure Layer  
> **设计理念**: 操作审计，异步处理  
> **设计灵感**: Spring Audit + Hibernate Envers

## 概述

`audit` 包提供操作审计日志记录功能，参考 Spring Boot 的 Spring Audit 设计。支持记录用户操作、数据变更、安全事件等，适用于需要审计追踪的企业级应用。

### 核心功能

| 功能 | 说明 |
|------|------|
| **多种事件类型** | 支持 CREATE/UPDATE/DELETE/LOGIN/SECURITY 等内置事件类型 |
| **异步处理** | 支持异步事件写入，提高性能，缓冲区满时自动降级为同步 |
| **多写入器** | 支持控制台、文件等多种 EventWriter 实现 |
| **拦截器** | 提供 AuditInterceptor 用于自动审计，可与 AOP 框架配合使用 |
| **日志助手** | 提供 AuditLogger 简化常见审计场景的日志记录 |

---

## 核心接口

### Event 审计事件

```go
type Event struct {
    ID           string
    Timestamp    time.Time
    Actor        string
    Action       EventType
    Resource     string
    Target       string
    Details      map[string]any
    Severity     EventSeverity
    Source       string
    Result       string
    ErrorMessage string
    Duration     time.Duration
    Tags         []string
}
```

#### 事件类型 (EventType)

| 常量 | 值 | 说明 |
|------|----|------|
| `EventCreate` | `"CREATE"` | 创建事件 |
| `EventUpdate` | `"UPDATE"` | 更新事件 |
| `EventDelete` | `"DELETE"` | 删除事件 |
| `EventRead` | `"READ"` | 读取事件 |
| `EventLogin` | `"LOGIN"` | 登录事件 |
| `EventLogout` | `"LOGOUT"` | 登出事件 |
| `EventAccess` | `"ACCESS"` | 访问事件 |
| `EventPermission` | `"PERMISSION"` | 权限事件 |
| `EventSecurity` | `"SECURITY"` | 安全事件 |
| `EventCustom` | `"CUSTOM"` | 自定义事件 |

#### 严重程度 (EventSeverity)

| 常量 | 值 | 说明 |
|------|----|------|
| `SeverityInfo` | `"INFO"` | 普通信息 |
| `SeverityWarning` | `"WARNING"` | 警告 |
| `SeverityError` | `"ERROR"` | 错误 |
| `SeverityCritical` | `"CRITICAL"` | 严重错误 |

### EventWriter 事件写入器

```go
type EventWriter interface {
    Write(event Event) error
    Close() error
}
```

#### ConsoleWriter 控制台写入器

将审计事件以 JSON 格式输出到标准输出：

```go
writer := audit.NewConsoleWriter()
writer.Write(event)
```

#### FileWriter 文件写入器

将审计事件追加到指定文件：

```go
writer, err := audit.NewFileWriter("/var/log/audit.log")
if err != nil {
    // 处理错误
}
defer writer.Close()

writer.Write(event)
```

### Auditor 审计日志器

```go
type Auditor struct {
    // ...
}
```

#### 创建

```go
auditor := audit.NewAuditor(
    audit.WithWriter(fileWriter),
    audit.WithAsync(),
    audit.WithBufferSize(1000),
)
defer auditor.Close()
```

#### 选项函数

| 函数 | 说明 | 默认值 |
|------|------|--------|
| `WithWriter(writer)` | 设置事件写入器 | `ConsoleWriter` |
| `WithBufferSize(size)` | 设置缓冲区大小(异步模式) | `1000` |
| `WithAsync()` | 启用异步写入模式 | 关闭 |

#### 记录事件

```go
auditor.Log(audit.Event{
    Actor:    "user123",
    Action:   audit.EventCreate,
    Resource: "user",
    Target:   "user:456",
    Details:  map[string]any{"name": "John"},
})
```

#### 便捷方法

| 方法 | 说明 |
|------|------|
| `LogAction(actor, action, resource, target, details)` | 记录操作事件 |
| `LogSecurity(actor, action, source, details)` | 记录安全事件 |
| `LogError(actor, action, resource, err)` | 记录错误事件 |
| `LogWithDuration(event, duration)` | 记录带耗时的审计事件 |

```go
// 记录操作事件
auditor.LogAction("user123", audit.EventCreate, "user", "user:456", map[string]any{
    "name": "John",
})

// 记录安全事件
auditor.LogSecurity("user123", audit.EventLogin, "192.168.1.1", map[string]any{
    "reason": "invalid password",
})

// 记录错误事件
auditor.LogError("user123", audit.EventCreate, "user", err)
```

### AuditInterceptor 审计拦截器

用于拦截方法调用并自动记录审计日志，通常与 AOP 框架配合使用。

#### 创建

```go
interceptor := audit.NewAuditInterceptor(auditor)
interceptor.SetActorFunc(func() string {
    return getCurrentUser()
})
interceptor.SetSourceFunc(func() string {
    return getRequestIP()
})
```

#### 使用

```go
// 方法执行前
interceptor.Before("CreateUser", []any{"John", "john@example.com"})

// 方法执行后
interceptor.After("CreateUser", nil, 100*time.Millisecond)
```

### AuditLogger 审计日志助手

提供便捷的审计日志记录方法，封装常用的审计场景。

#### 创建

```go
logger := audit.NewAuditLogger(auditor, "user123", "web-app")
```

#### 方法

| 方法 | 说明 |
|------|------|
| `Create(resource, target, details)` | 记录创建事件 |
| `Update(resource, target, details)` | 记录更新事件 |
| `Delete(resource, target)` | 记录删除事件 |
| `Login(source, details)` | 记录登录事件 |
| `LoginFailure(source, reason)` | 记录登录失败事件 |
| `PermissionDenied(resource, target)` | 记录权限拒绝事件 |

```go
// 记录用户创建
logger.Create("user", "user:456", map[string]any{
    "name":  "John",
    "email": "john@example.com",
})

// 记录登录失败
logger.LoginFailure("192.168.1.1", "invalid password")

// 记录权限拒绝
logger.PermissionDenied("user", "user:789")
```

---

## 快速开始

### 基本使用

```go
package main

import (
    "github.com/xudefa/enhance/audit"
)

func main() {
    auditor := audit.NewAuditor()
    defer auditor.Close()

    auditor.Log(audit.Event{
        Actor:    "user123",
        Action:   audit.EventCreate,
        Resource: "user",
        Target:   "user:456",
        Details:  map[string]any{"name": "John"},
    })
}
```

### 异步模式

```go
auditor := audit.NewAuditor(
    audit.WithAsync(),
    audit.WithBufferSize(1000),
)
defer auditor.Close()

// 异步记录事件
for i := 0; i < 100; i++ {
    auditor.Log(audit.Event{
        Actor:  "user123",
        Action: audit.EventCreate,
    })
}
```

---

## API 参考

### 使用审计日志助手

```go
auditor := audit.NewAuditor()
defer auditor.Close()

logger := audit.NewAuditLogger(auditor, "admin", "admin-panel")

// 记录用户创建
logger.Create("user", "user:456", map[string]any{
    "name":  "John",
    "email": "john@example.com",
})

// 记录登录失败
logger.LoginFailure("192.168.1.1", "invalid password")

// 记录权限拒绝
logger.PermissionDenied("user", "user:789")
```

### 与 AOP 集成

```go
type UserService struct {
    auditor *audit.Auditor
}

func (s *UserService) CreateUser(name string) error {
    interceptor := audit.NewAuditInterceptor(s.auditor)
    
    interceptor.Before("CreateUser", []any{name})
    defer func() {
        interceptor.After("CreateUser", nil, 0)
    }()
    
    // 创建用户逻辑...
    return nil
}
```

---

## 使用示例

### 场景 1: 用户操作审计

记录用户的关键操作，用于安全审计和追溯：

```go
logger := audit.NewAuditLogger(auditor, currentUser, "web-app")

// 记录资源创建
logger.Create("user", userID, map[string]any{
    "name":  userName,
    "email": userEmail,
})

// 记录资源更新
logger.Update("user", userID, map[string]any{
    "role": "admin",
})

// 记录资源删除
logger.Delete("user", userID)
```

### 场景 2: 安全事件监控

监控和记录安全相关事件，如登录失败、权限拒绝等：

```go
// 登录失败
auditor.LogSecurity(username, audit.EventLogin, clientIP, map[string]any{
    "reason": "invalid password",
    "attempt": attemptCount,
})

// 权限拒绝
logger.PermissionDenied(resource, target)

// 安全告警
auditor.Log(audit.Event{
    Actor:    "system",
    Action:   audit.EventSecurity,
    Severity: audit.SeverityCritical,
    Details:  map[string]any{"alert": "brute force detected"},
})
```

### 场景 3: API 访问审计

记录 API 访问日志，用于性能分析和故障排查：

```go
func auditMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        interceptor := audit.NewAuditInterceptor(auditor)
        interceptor.SetActorFunc(func() string {
            return getUserFromRequest(r)
        })
        interceptor.SetSourceFunc(func() string {
            return r.RemoteAddr
        })
        
        interceptor.Before(r.URL.Path, nil)
        
        next.ServeHTTP(w, r)
        
        interceptor.After(r.URL.Path, nil, time.Since(start))
    })
}
```

### 场景 4: 数据变更追踪

记录数据的变更历史，用于审计和回滚：

```go
func (s *OrderService) UpdateOrder(orderID string, updates map[string]any) error {
    oldOrder := s.getOrder(orderID)
    
    if err := s.applyUpdates(orderID, updates); err != nil {
        auditor.LogError(currentUser, audit.EventUpdate, "order", err)
        return err
    }
    
    auditor.Log(audit.Event{
        Actor:    currentUser,
        Action:   audit.EventUpdate,
        Resource: "order",
        Target:   orderID,
        Details: map[string]any{
            "old": oldOrder,
            "new": updates,
        },
    })
    
    return nil
}
```

---

## 最佳实践

### 1. 使用异步模式提升性能

```go
// ✅ 推荐：高并发场景使用异步
auditor := audit.NewAuditor(
    audit.WithAsync(),
    audit.WithBufferSize(1000),
)
defer auditor.Close()

// ⚠️ 不推荐：同步模式影响性能
auditor := audit.NewAuditor()
```

### 2. 记录足够的上下文信息

```go
// ✅ 推荐：包含详细信息
auditor.Log(audit.Event{
    Actor:    "user123",
    Action:   audit.EventCreate,
    Resource: "user",
    Target:   "user:456",
    Details: map[string]any{
        "name":  "John",
        "email": "john@example.com",
        "role":  "admin",
    },
    Source: "192.168.1.1",
})

// ⚠️ 不推荐：信息不足
auditor.Log(audit.Event{
    Actor:  "user123",
    Action: audit.EventCreate,
})
```

### 3. 使用合适的 Severity 级别

```go
// ✅ 推荐：根据事件重要性设置 Severity
auditor.Log(audit.Event{
    Action:   audit.EventSecurity,
    Severity: audit.SeverityCritical,
    Details:  map[string]any{"alert": "brute force detected"},
})

// ⚠️ 不推荐：所有事件使用相同级别
auditor.Log(audit.Event{
    Action:   audit.EventCreate,
    Severity: audit.SeverityCritical, // 过度使用严重级别
})
```

### 4. 与 AOP 集成自动审计

```go
// ✅ 推荐：使用拦截器自动审计
interceptor := audit.NewAuditInterceptor(auditor)
interceptor.SetActorFunc(func() string {
    return getCurrentUser()
})

// ⚠️ 不推荐：手动记录每个操作
auditor.Log(audit.Event{...})
```

### 5. 与依赖注入集成

```go
// ✅ 推荐：将 Auditor 注册为 Bean
container.Register(
    reflect.TypeOf(&audit.Auditor{}),
    core.Bean(audit.NewAuditor(
        audit.WithAsync(),
        audit.WithBufferSize(1000),
    )),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    Auditor *audit.Auditor `inject:"auditor"`
}
```