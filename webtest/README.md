# webtest 包 — Web 测试

> **所属层级**: Infrastructure Layer  
> **设计理念**: 链式 API，简化测试  
> **设计灵感**: Spring WebTestClient

## 概述

`webtest` 包提供 Web 测试客户端功能，参考 Spring WebTestClient 设计。支持链式构建 HTTP 请求和断言响应，简化 Web 层集成测试的编写。

### 核心功能

| 功能 | 说明 |
|------|------|
| **链式 API** | 流畅的链式调用，提高测试代码可读性 |
| **请求构建** | 支持 GET/POST/PUT/DELETE/PATCH 请求 |
| **响应断言** | 丰富的状态码、响应头、响应体断言方法 |
| **JSON 支持** | 自动序列化/反序列化 JSON 数据 |
| **调试支持** | 提供响应打印功能，方便调试 |

---

## 核心接口

### WebTestClient Web 测试客户端

```go
type WebTestClient struct {
    handler http.Handler
    baseURL string
    headers map[string]string
}
```

#### 创建

```go
client := webtest.NewWebTestClient(handler)
```

#### 便捷函数

```go
client := webtest.CreateWebTestClient(handler)
```

#### 配置方法

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `BaseURL(url string)` | 设置基础 URL | `*WebTestClient` |
| `Header(name, value string)` | 设置默认请求头 | `*WebTestClient` |

#### HTTP 方法

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `Get(path string)` | 发起 GET 请求 | `*RequestSpec` |
| `Post(path string)` | 发起 POST 请求 | `*RequestSpec` |
| `Put(path string)` | 发起 PUT 请求 | `*RequestSpec` |
| `Delete(path string)` | 发起 DELETE 请求 | `*RequestSpec` |
| `Patch(path string)` | 发起 PATCH 请求 | `*RequestSpec` |

### RequestSpec 请求规范

```go
type RequestSpec struct {
    // ...
}
```

#### 请求配置

| 方法 | 说明 | 返回值 |
|------|------|--------|
| `Header(name, value string)` | 设置请求头 | `*RequestSpec` |
| `ContentType(contentType string)` | 设置 Content-Type | `*RequestSpec` |
| `Body(body io.Reader)` | 设置请求体 | `*RequestSpec` |
| `JSON(data any)` | 设置 JSON 请求体 | `*RequestSpec` |
| `Exchange()` | 发起请求并返回响应规范 | `*ResponseSpec` |

### ResponseSpec 响应规范

```go
type ResponseSpec struct {
    recorder *httptest.ResponseRecorder
}
```

#### 状态码断言

| 方法 | 说明 |
|------|------|
| `Status(expected int)` | 断言状态码 |
| `StatusIsOk()` | 断言 200 |
| `StatusIsCreated()` | 断言 201 |
| `StatusIsNoContent()` | 断言 204 |
| `StatusIsBadRequest()` | 断言 400 |
| `StatusIsUnauthorized()` | 断言 401 |
| `StatusIsForbidden()` | 断言 403 |
| `StatusIsNotFound()` | 断言 404 |

#### 响应头断言

| 方法 | 说明 |
|------|------|
| `Header(name, expected string)` | 断言响应头 |

#### 响应体断言

| 方法 | 说明 |
|------|------|
| `Body()` | 获取响应体字符串 |
| `JSONBody(target any)` | 解析 JSON 响应体 |
| `BodyContains(substring string)` | 断言响应体包含子串 |
| `BodyEquals(expected string)` | 断言响应体等于字符串 |

#### 调试方法

| 方法 | 说明 |
|------|------|
| `Print()` | 打印响应信息 |
| `Recorder()` | 获取底层 ResponseRecorder |

---

## 快速开始

### 基本使用

```go
package main

import (
    "net/http"
    "testing"
    "github.com/xudefa/enhance/webtest"
)

func TestHelloEndpoint(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Hello, World!"))
    })

    client := webtest.NewWebTestClient(handler)
    client.Get("/api/hello").
        Exchange().
        StatusIsOk().
        BodyContains("Hello")
}
```

---

## API 参考

### POST JSON 请求

```go
func TestCreateUser(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var user User
        json.NewDecoder(r.Body).Decode(&user)
        
        user.ID = "123"
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(user)
    })

    client := webtest.NewWebTestClient(handler)

    var result User
    client.Post("/api/users").
        JSON(map[string]string{
            "name": "John",
            "email": "john@example.com",
        }).
        Exchange().
        StatusIsCreated().
        JSONBody(&result)

    if result.ID != "123" {
        t.Errorf("expected ID 123, got %s", result.ID)
    }
}
```

### 带请求头的请求

```go
func TestAuthenticatedRequest(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token != "Bearer valid-token" {
            w.WriteHeader(http.StatusUnauthorized)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("Authenticated"))
    })

    client := webtest.NewWebTestClient(handler)
    client.Get("/api/protected").
        Header("Authorization", "Bearer valid-token").
        Exchange().
        StatusIsOk().
        BodyContains("Authenticated")
}
```

### 测试 404 响应

```go
func TestNotFound(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
        w.Write([]byte("Not Found"))
    })

    client := webtest.NewWebTestClient(handler)
    client.Get("/api/nonexistent").
        Exchange().
        StatusIsNotFound()
}
```

### 调试响应

```go
func TestDebugResponse(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"message": "Hello"}`))
    })

    client := webtest.NewWebTestClient(handler)
    client.Get("/api/hello").
        Exchange().
        Print() // 打印响应信息用于调试
}
```

输出:
```
Status: 200
Headers: map[Content-Type:[application/json]]
Body: {"message": "Hello"}
```

---

## 使用示例

### 场景 1: REST API 集成测试

测试 REST API 的 CRUD 操作：

```go
func TestUserCRUD(t *testing.T) {
    handler := setupTestHandler()
    client := webtest.NewWebTestClient(handler)

    // Create
    var created User
    client.Post("/api/users").
        JSON(CreateUserRequest{Name: "John", Email: "john@example.com"}).
        Exchange().
        StatusIsCreated().
        JSONBody(&created)

    // Read
    var retrieved User
    client.Get("/api/users/" + created.ID).
        Exchange().
        StatusIsOk().
        JSONBody(&retrieved)

    if retrieved.Name != "John" {
        t.Errorf("expected name John, got %s", retrieved.Name)
    }

    // Update
    client.Put("/api/users/" + created.ID).
        JSON(UpdateUserRequest{Name: "Jane"}).
        Exchange().
        StatusIsOk()

    // Delete
    client.Delete("/api/users/" + created.ID).
        Exchange().
        StatusIsNoContent()

    // Verify deleted
    client.Get("/api/users/" + created.ID).
        Exchange().
        StatusIsNotFound()
}
```

**最佳实践**:
- 测试完整的 CRUD 流程
- 验证每个操作的响应状态
- 验证数据一致性

### 场景 2: 认证和授权测试

测试 API 的认证和授权逻辑：

```go
func TestAuthentication(t *testing.T) {
    handler := setupTestHandler()
    client := webtest.NewWebTestClient(handler)

    // 无认证头访问受保护资源
    client.Get("/api/protected").
        Exchange().
        StatusIsUnauthorized()

    // 使用有效 token 访问
    client.Get("/api/protected").
        Header("Authorization", "Bearer valid-token").
        Exchange().
        StatusIsOk()

    // 使用无效 token 访问
    client.Get("/api/protected").
        Header("Authorization", "Bearer invalid-token").
        Exchange().
        StatusIsUnauthorized()
}

func TestAuthorization(t *testing.T) {
    handler := setupTestHandler()
    client := webtest.NewWebTestClient(handler)

    // 普通用户访问管理员接口
    client.Get("/api/admin/users").
        Header("Authorization", "Bearer user-token").
        Exchange().
        StatusIsForbidden()

    // 管理员访问管理员接口
    client.Get("/api/admin/users").
        Header("Authorization", "Bearer admin-token").
        Exchange().
        StatusIsOk()
}
```

**最佳实践**:
- 测试未认证、认证失败、认证成功场景
- 测试权限不足和权限足够场景
- 使用不同的 token 测试不同角色

### 场景 3: 错误处理测试

测试 API 的错误处理逻辑：

```go
func TestValidationErrors(t *testing.T) {
    handler := setupTestHandler()
    client := webtest.NewWebTestClient(handler)

    // 缺少必填字段
    client.Post("/api/users").
        JSON(map[string]string{"email": "test@example.com"}).
        Exchange().
        StatusIsBadRequest().
        BodyContains("name is required")

    // 邮箱格式错误
    client.Post("/api/users").
        JSON(map[string]string{
            "name": "John",
            "email": "invalid-email",
        }).
        Exchange().
        StatusIsBadRequest().
        BodyContains("invalid email format")
}

func TestNotFoundError(t *testing.T) {
    handler := setupTestHandler()
    client := webtest.NewWebTestClient(handler)

    client.Get("/api/users/nonexistent").
        Exchange().
        StatusIsNotFound().
        BodyContains("user not found")
}
```

**最佳实践**:
- 测试各种验证错误场景
- 验证错误响应格式和消息
- 测试资源不存在场景

### 场景 4: 分页和过滤测试

测试 API 的分页和过滤功能：

```go
func TestPagination(t *testing.T) {
    handler := setupTestHandler()
    client := webtest.NewWebTestClient(handler)

    // 第一页
    var page1 UserListResponse
    client.Get("/api/users?page=1&size=10").
        Exchange().
        StatusIsOk().
        JSONBody(&page1)

    if len(page1.Users) > 10 {
        t.Errorf("expected max 10 users, got %d", len(page1.Users))
    }

    // 第二页
    var page2 UserListResponse
    client.Get("/api/users?page=2&size=10").
        Exchange().
        StatusIsOk().
        JSONBody(&page2)

    // 验证分页信息
    if page2.Page != 2 {
        t.Errorf("expected page 2, got %d", page2.Page)
    }
}

func TestFiltering(t *testing.T) {
    handler := setupTestHandler()
    client := webtest.NewWebTestClient(handler)

    // 按状态过滤
    var activeUsers UserListResponse
    client.Get("/api/users?status=active").
        Exchange().
        StatusIsOk().
        JSONBody(&activeUsers)

    for _, user := range activeUsers.Users {
        if user.Status != "active" {
            t.Errorf("expected active user, got %s", user.Status)
        }
    }
}
```

**最佳实践**:
- 测试分页参数验证
- 测试过滤条件组合
- 验证分页响应格式

---

## 最佳实践

### 1. 使用链式调用提高可读性

```go
// ✅ 推荐：使用链式调用
client.Post("/api/users").
    JSON(CreateUserRequest{Name: "John"}).
    Exchange().
    StatusIsCreated().
    JSONBody(&result)

// ⚠️ 不推荐：分步调用
req := client.Post("/api/users")
req = req.JSON(CreateUserRequest{Name: "John"})
resp := req.Exchange()
resp.StatusIsCreated()
resp.JSONBody(&result)
```

### 2. 测试完整的错误场景

```go
// ✅ 推荐：测试各种错误场景
client.Post("/api/users").
    JSON(map[string]string{}).
    Exchange().
    StatusIsBadRequest()

client.Post("/api/users").
    JSON(map[string]string{"email": "invalid"}).
    Exchange().
    StatusIsBadRequest()

client.Get("/api/users/nonexistent").
    Exchange().
    StatusIsNotFound()

// ⚠️ 不推荐：只测试成功场景
client.Post("/api/users").
    JSON(CreateUserRequest{Name: "John"}).
    Exchange().
    StatusIsCreated()
```

### 3. 使用 JSONBody 自动反序列化

```go
// ✅ 推荐：使用 JSONBody 自动反序列化
var result User
client.Get("/api/users/123").
    Exchange().
    StatusIsOk().
    JSONBody(&result)

// ⚠️ 不推荐：手动解析 JSON
resp := client.Get("/api/users/123").Exchange()
body := resp.Body()
json.Unmarshal([]byte(body), &result)
```

### 4. 使用 Print 调试

```go
// ✅ 推荐：调试时使用 Print
client.Get("/api/users/123").
    Exchange().
    Print(). // 打印响应信息
    StatusIsOk()

// ⚠️ 不推荐：手动打印响应
resp := client.Get("/api/users/123").Exchange()
fmt.Printf("Status: %d\n", resp.Recorder().Code)
fmt.Printf("Body: %s\n", resp.Recorder().Body.String())
```

### 5. 与测试框架集成

```go
// ✅ 推荐：在测试套件中复用客户端
func TestSuite(t *testing.T) {
    handler := setupTestHandler()
    client := webtest.NewWebTestClient(handler)

    t.Run("CreateUser", func(t *testing.T) {
        testCreateUser(t, client)
    })

    t.Run("GetUser", func(t *testing.T) {
        testGetUser(t, client)
    })

    t.Run("UpdateUser", func(t *testing.T) {
        testUpdateUser(t, client)
    })
}

// ⚠️ 不推荐：每个测试都创建新的 handler
func TestCreateUser(t *testing.T) {
    handler := setupTestHandler() // 重复创建
    client := webtest.NewWebTestClient(handler)
    // ...
}
```

### 6. 设计要点

- 使用 `httptest.ResponseRecorder` 记录响应
- 链式 API 提高测试代码可读性
- 丰富的断言方法简化测试编写
- JSON 自动序列化/反序列化
- 零外部依赖，仅使用 Go 标准库