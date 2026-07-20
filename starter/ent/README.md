# ent Starter

ent ORM 框架自动配置模块，提供类型安全的 ORM 支持。

## 功能特性

- ✅ 自动配置 ent 数据库连接
- ✅ 支持多种数据库驱动（MySQL/PostgreSQL/SQLite）
- ✅ 类型安全的查询
- ✅ 自动迁移支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/ent"
)
```

### 2. 配置文件

在 `application.json` 中添加 ent 配置：

```json
{
  "ent": {
    "enabled": true,
    "driver": "mysql",
    "dsn": "root:root@tcp(localhost:3306)/enhance?parseTime=True",
    "database": "enhance"
  }
}
```

### 3. 使用示例

```go
package main

import (
    "context"
    
    "entgo.io/ent/dialect/sql"
    
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/xudefa/enhance/starter/ent"
    "your/ent/schema"
)

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("ent-demo"),
    )
    defer app.Stop()
    
    // 获取 SQL 驱动
    driver := core.MustGetBean[*sql.Driver](app.Container())
    
    // 创建 ent 客户端
    client := ent.NewClient(ent.Driver(driver))
    defer client.Close()
    
    ctx := context.Background()
    
    // 自动迁移
    client.Schema.Create(ctx)
    
    // 创建用户
    user, err := client.User.
        Create().
        SetName("John").
        SetAge(30).
        Save(ctx)
    if err != nil {
        // 处理错误
    }
    println("创建用户:", user.Name)
    
    // 查询用户
    users, err := client.User.
        Query().
        Where(user.AgeGT(18)).
        All(ctx)
    if err != nil {
        // 处理错误
    }
    
    // 更新用户
    err = client.User.
        UpdateOne(user).
        SetAge(31).
        Exec(ctx)
    if err != nil {
        // 处理错误
    }
    
    // 删除用户
    err = client.User.
        DeleteOne(user).
        Exec(ctx)
    if err != nil {
        // 处理错误
    }
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `ent.enabled` | bool | false | 是否启用 ent |
| `ent.driver` | string | mysql | 数据库驱动 (mysql/postgres/sqlite) |
| `ent.dsn` | string | "" | 数据库连接字符串 |
| `ent.database` | string | enhance | 数据库名称 |

## 支持的数据库驱动

| 驱动 | 说明 |
|------|------|
| mysql | MySQL 数据库 |
| postgres | PostgreSQL 数据库 |
| sqlite | SQLite 数据库 |

## 高级用法

### 事务支持

```go
tx, err := client.Tx(ctx)
if err != nil {
    // 处理错误
}

// 在事务中执行操作
user, err := tx.User.Create().SetName("John").Save(ctx)
if err != nil {
    tx.Rollback()
    return
}

// 提交事务
tx.Commit()
```

### 关联查询

```go
// 查询用户及其帖子
users, err := client.User.
    Query().
    WithPosts().
    All(ctx)

for _, user := range users {
    println("用户:", user.Name)
    for _, post := range user.Edges.Posts {
        println("  帖子:", post.Title)
    }
}
```

### 分页查询

```go
users, err := client.User.
    Query().
    Offset(0).
    Limit(10).
    Order(ent.Asc(user.FieldCreatedAt)).
    All(ctx)
```

## 启动顺序

- **优先级**: `OrderPriorityDataLayer` (-2000)
- **触发条件**: `ent.enabled=true`

## 依赖

- `entgo.io/ent`
- `entgo.io/ent/dialect/sql`