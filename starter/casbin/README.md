# Casbin Starter

Casbin 权限控制自动配置模块，提供灵活的权限管理支持。

## 功能特性

- ✅ 自动配置 Casbin Enforcer
- ✅ 支持多种权限模型（ACL/RBAC/ABAC）
- ✅ 策略持久化
- ✅ 权限检查接口
- ✅ 角色管理支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/casbin"
)
```

### 2. 配置文件

在 `application.json` 中添加 Casbin 配置：

```json
{
  "casbin": {
    "enabled": true,
    "model_path": "config/casbin_model.conf",
    "policy_path": "config/casbin_policy.csv",
    "adapter": "file"
  }
}
```

### 3. 模型配置文件

创建 `config/casbin_model.conf`：

```ini
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```

### 4. 策略配置文件

创建 `config/casbin_policy.csv`：

```csv
p, alice, data1, read
p, bob, data2, write
p, admin, *, *
```

### 5. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/casbin"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("casbin-demo"),
    )
    defer app.Stop()
    
    // 获取 Enforcer
    enforcer := core.MustGetBean[*casbin.CasbinEnforcer](app.Container())
    
    // 检查权限
    allowed, err := enforcer.Enforce("alice", "data1", "read")
    if err != nil {
        // 处理错误
    }
    println("alice 可以读取 data1:", allowed)
    
    // 添加策略
    enforcer.AddPolicy("alice", "data2", "write")
    
    // 删除策略
    enforcer.RemovePolicy("alice", "data1", "read")
    
    // 获取用户角色
    roles, _ := enforcer.GetRolesForUser("alice")
    
    // 添加角色
    enforcer.AddRoleForUser("alice", "admin")
    
    // 检查角色权限
    allowed, _ = enforcer.Enforce("admin", "data1", "delete")
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `casbin.enabled` | bool | false | 是否启用 Casbin |
| `casbin.model_path` | string | config/casbin_model.conf | 模型配置文件路径 |
| `casbin.policy_path` | string | config/casbin_policy.csv | 策略文件路径 |
| `casbin.adapter` | string | file | 适配器类型 (file/gorm/xorm) |

## 权限模型

### ACL（访问控制列表）

最简单的权限模型，直接定义用户、资源和操作的权限。

```csv
p, alice, data1, read
p, bob, data2, write
```

### RBAC（基于角色的访问控制）

通过角色来管理权限。

```csv
p, admin, *, *
p, editor, data, write
p, viewer, data, read

g, alice, admin
g, bob, editor
```

### ABAC（基于属性的访问控制）

基于资源属性的权限控制。

```ini
[matchers]
m = r.sub == r.obj.Owner
```

## 高级用法

### 中间件集成

```go
import "github.com/xudefa/enhance/starter/casbin"

// HTTP 中间件
e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        user := getUser(c)
        resource := c.Request().URL.Path
        action := c.Request().Method
        
        allowed, _ := enforcer.Enforce(user, resource, action)
        if !allowed {
            return c.JSON(403, map[string]string{"error": "Forbidden"})
        }
        return next(c)
    }
})
```

### 动态策略

```go
// 运行时添加策略
enforcer.AddPolicy("alice", "data1", "read")

// 运行时删除策略
enforcer.RemovePolicy("alice", "data1", "read")

// 批量添加策略
enforcer.AddPolicies([][]string{
    {"alice", "data1", "read"},
    {"bob", "data2", "write"},
})

// 保存策略到文件
enforcer.SavePolicy()
```

## 启动顺序

- **优先级**: `OrderPriorityWebLayer` (0)
- **触发条件**: `casbin.enabled=true`

## 依赖

- `github.com/casbin/casbin/v2`