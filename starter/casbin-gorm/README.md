# Casbin-GORM Starter

Casbin + GORM 集成模块，提供基于数据库的权限策略持久化支持。

## 功能特性

- ✅ 自动配置 Casbin + GORM
- ✅ 策略数据库持久化
- ✅ 支持多种数据库（MySQL/PostgreSQL/SQLite）
- ✅ 自动创建策略表
- ✅ 策略缓存支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/casbin-gorm"
)
```

### 2. 配置文件

在 `application.json` 中添加配置：

```json
{
  "db": {
    "gorm": {
      "enabled": true,
      "host": "localhost",
      "port": 3306,
      "username": "root",
      "password": "root",
      "database": "enhance"
    }
  },
  "casbin": {
    "enabled": true,
    "model_path": "config/casbin_model.conf",
    "adapter": "gorm"
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/casbin-gorm"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("casbin-gorm-demo"),
    )
    defer app.Stop()
    
    // 获取 Enforcer
    enforcer := core.MustGetBean[*casbin_gorm.CasbinEnforcer](app.Container())
    
    // 添加策略（自动保存到数据库）
    enforcer.AddPolicy("alice", "data1", "read")
    enforcer.AddPolicy("bob", "data2", "write")
    
    // 检查权限
    allowed, _ := enforcer.Enforce("alice", "data1", "read")
    
    // 策略变更会自动同步到数据库
    enforcer.RemovePolicy("alice", "data1", "read")
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `db.gorm.enabled` | bool | false | 是否启用 GORM |
| `casbin.enabled` | bool | false | 是否启用 Casbin |
| `casbin.model_path` | string | config/casbin_model.conf | 模型配置文件路径 |
| `casbin.adapter` | string | gorm | 适配器类型 |

## 数据库表结构

模块会自动创建 `casbin_rule` 表：

```sql
CREATE TABLE casbin_rule (
    id    BIGINT AUTO_INCREMENT PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0    VARCHAR(100),
    v1    VARCHAR(100),
    v2    VARCHAR(100),
    v3    VARCHAR(100),
    v4    VARCHAR(100),
    v5    VARCHAR(100)
);
```

## 高级用法

### 多租户支持

```go
// 为不同租户使用不同的数据库
enforcer.AddPolicy("tenant1:alice", "data1", "read")
enforcer.AddPolicy("tenant2:bob", "data2", "write")
```

### 策略缓存

```go
// 加载策略到内存
enforcer.LoadPolicy()

// 保存策略到数据库
enforcer.SavePolicy()

// 清除缓存
enforcer.ClearPolicy()
```

## 启动顺序

- **优先级**: `OrderPriorityDataLayer` (-2000)
- **触发条件**: `casbin.enabled=true` 且 `db.gorm.enabled=true`

## 依赖

- `github.com/casbin/casbin/v2`
- `github.com/casbin/gorm-adapter/v3`
- `gorm.io/gorm`