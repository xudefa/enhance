# tenant 包 — 多租户支持

> **所属层级**: Infrastructure Layer  
> **设计理念**: 租户隔离，SaaS 架构支持  
> **设计灵感**: Spring Multi-Tenant + Hibernate Multi-Tenancy

## 概述

`tenant` 包提供多租户架构支持，参考 Spring Boot 多租户设计。支持租户隔离、租户上下文、租户解析器等功能，适用于 SaaS 应用开发。

### 核心功能

| 功能 | 说明 |
|------|------|
| **租户解析器** | 支持请求头、子域名、JWT 等多种租户识别方式 |
| **租户上下文** | 线程安全的租户上下文管理 |
| **租户中间件** | HTTP 中间件自动解析和设置租户 |
| **租户隔离** | 支持数据层和业务层的租户隔离 |

---

## 核心接口

### Tenant 租户对象

```go
type Tenant struct {
    ID       string
    Name     string
    Domain   string
    Database string
    Enabled  bool
    Metadata map[string]string
}
```

### TenantResolver 租户解析器接口

```go
type TenantResolver interface {
    Resolve(req *http.Request) (string, error)
}
```

#### HeaderResolver 请求头解析器

```go
resolver := tenant.NewHeaderResolver("X-Tenant-ID")
```

从请求头 `X-Tenant-ID` 中提取租户 ID。

#### SubdomainResolver 子域名解析器

```go
resolver := tenant.NewSubdomainResolver("example.com")
```

从子域名中提取租户 ID，例如 `tenant1.example.com` 解析为 `tenant1`。

### TenantManager 租户管理器

```go
type TenantManager struct {
    // ...
}
```

#### 创建

```go
manager := tenant.NewTenantManager(resolver)
```

#### 租户管理

```go
// 设置当前租户
manager.SetCurrentTenant("tenant-123")

// 获取当前租户
currentTenant := manager.GetCurrentTenant()

// 清除当前租户
manager.ClearCurrentTenant()
```

---

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "github.com/xudefa/enhance/tenant"
)

func main() {
    // 创建请求头解析器
    resolver := tenant.NewHeaderResolver("X-Tenant-ID")

    // 创建租户管理器
    manager := tenant.NewTenantManager(resolver)

    // 设置当前租户
    manager.SetCurrentTenant("tenant-123")

    // 获取当前租户
    currentTenant := manager.GetCurrentTenant()
    fmt.Println("Current tenant:", currentTenant)
}
```

---

## API 参考

### 使用租户中间件

```go
resolver := tenant.NewHeaderResolver("X-Tenant-ID")
manager := tenant.NewTenantManager(resolver)

// 创建租户中间件
middleware := tenant.NewTenantMiddleware(manager)

// 包装处理器
handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    tenant := manager.GetCurrentTenant()
    fmt.Fprintf(w, "Hello from tenant: %s", tenant.ID)
}))

http.Handle("/api", handler)
http.ListenAndServe(":8080", nil)
```

### 使用子域名解析器

```go
resolver := tenant.NewSubdomainResolver("example.com")
manager := tenant.NewTenantManager(resolver)

middleware := tenant.NewTenantMiddleware(manager)

handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    tenant := manager.GetCurrentTenant()
    fmt.Fprintf(w, "Tenant: %s", tenant.ID)
}))

http.Handle("/", handler)
http.ListenAndServe(":8080", nil)
```

---

## 使用示例

### 场景 1: SaaS 应用多租户隔离

SaaS 应用中不同租户数据完全隔离：

```go
func (s *UserService) GetUsers(ctx context.Context) ([]User, error) {
    tenant := tenant.GetCurrentTenant(ctx)
    if tenant == nil {
        return nil, fmt.Errorf("no tenant context")
    }

    // 使用租户 ID 查询数据
    return s.repo.FindByTenant(tenant.ID)
}
```

**最佳实践**:
- 所有数据查询必须包含租户 ID
- 使用中间件自动设置租户上下文
- 记录租户操作日志

### 场景 2: 多租户数据库隔离

不同租户使用不同的数据库或 schema：

```go
func GetDBForTenant(tenantID string) (*sql.DB, error) {
    // 根据租户 ID 获取对应的数据库连接
    config := getTenantConfig(tenantID)
    return sql.Open(config.Driver, config.DSN)
}

func (s *OrderService) CreateOrder(order *Order) error {
    tenant := tenant.GetCurrentTenant()
    
    db, err := GetDBForTenant(tenant.ID)
    if err != nil {
        return err
    }

    return db.Exec("INSERT INTO orders ...", order)
}
```

**最佳实践**:
- 使用连接池管理租户数据库连接
- 缓存租户配置避免重复查询
- 定期清理空闲连接

### 场景 3: 租户配额管理

限制每个租户的资源使用量：

```go
func (s *QuotaService) CheckQuota(tenantID string, resource string) error {
    quota := s.getTenantQuota(tenantID, resource)
    usage := s.getTenantUsage(tenantID, resource)

    if usage >= quota {
        return fmt.Errorf("tenant %s exceeded %s quota", tenantID, resource)
    }

    return nil
}
```

**最佳实践**:
- 在关键操作前检查配额
- 提供配额使用情况查询接口
- 支持配额动态调整

---

## 最佳实践

### 1. 使用中间件自动解析租户

```go
// ✅ 推荐：使用中间件自动解析
resolver := tenant.NewHeaderResolver("X-Tenant-ID")
manager := tenant.NewTenantManager(resolver)
middleware := tenant.NewTenantMiddleware(manager)

// ⚠️ 不推荐：手动解析每个请求
func handler(w http.ResponseWriter, r *http.Request) {
    tenantID := r.Header.Get("X-Tenant-ID")
    if tenantID == "" {
        http.Error(w, "Missing tenant", 400)
        return
    }
    manager.SetCurrentTenant(tenantID)
}
```

### 2. 使用上下文传递租户信息

```go
// ✅ 推荐：使用 context 传递租户
ctx := context.WithValue(r.Context(), "tenant", tenant)
next.ServeHTTP(w, r.WithContext(ctx))

// 获取租户
func GetCurrentTenant(ctx context.Context) *Tenant {
    if t, ok := ctx.Value("tenant").(*Tenant); ok {
        return t
    }
    return nil
}

// ⚠️ 不推荐：使用全局变量
var currentTenant *Tenant
```

### 3. 数据查询包含租户隔离

```go
// ✅ 推荐：所有查询包含租户条件
func (r *UserRepository) FindByTenant(tenantID string) ([]User, error) {
    return r.db.Where("tenant_id = ?", tenantID).Find(&users)
}

// ⚠️ 不推荐：查询不包含租户条件
func (r *UserRepository) FindAll() ([]User, error) {
    return r.db.Find(&users)
}
```

### 4. 缓存租户配置

```go
// ✅ 推荐：缓存租户配置
type TenantCache struct {
    cache map[string]*TenantConfig
    mu    sync.RWMutex
}

func (c *TenantCache) Get(tenantID string) (*TenantConfig, error) {
    c.mu.RLock()
    if config, ok := c.cache[tenantID]; ok {
        c.mu.RUnlock()
        return config, nil
    }
    c.mu.RUnlock()
    
    // 从数据库加载
    config, err := loadFromDB(tenantID)
    if err != nil {
        return nil, err
    }
    
    c.mu.Lock()
    c.cache[tenantID] = config
    c.mu.Unlock()
    
    return config, nil
}

// ⚠️ 不推荐：每次查询都从数据库加载
func GetTenantConfig(tenantID string) (*TenantConfig, error) {
    return loadFromDB(tenantID) // 每次都查询数据库
}
```

### 5. 与依赖注入集成

```go
// ✅ 推荐：将 TenantManager 注册为 Bean
container.Register(
    reflect.TypeOf(&tenant.TenantManager{}),
    core.Bean(createTenantManager()),
    core.Singleton(),
)

// 注入使用
type UserService struct {
    TenantManager *tenant.TenantManager `inject:"tenantManager"`
}

func (s *UserService) GetUsers(ctx context.Context) ([]User, error) {
    tenant := s.TenantManager.GetCurrentTenant()
    return s.repo.FindByTenant(tenant.ID)
}
```

### 6. 设计要点

- 支持多种租户解析方式
- 使用上下文传递租户信息
- 中间件自动解析和设置租户
- 线程安全的租户上下文管理
- 零外部依赖，仅使用 Go 标准库