# XORM Starter

XORM ORM 框架自动配置模块，提供轻量级数据库访问支持。

## 功能特性

- ✅ 自动配置 XORM 数据库连接
- ✅ 连接池管理
- ✅ 自动迁移支持
- ✅ 多数据库驱动支持
- ✅ 缓存支持

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/xorm"
)
```

### 2. 配置文件

在 `application.json` 中添加 XORM 配置：

```json
{
  "db": {
    "xorm": {
      "enabled": true,
      "host": "localhost",
      "port": 3306,
      "username": "root",
      "password": "root",
      "database": "enhance",
      "charset": "utf8mb4",
      "max-open-conns": 100,
      "max-idle-conns": 10,
      "conn-max-lifetime": 3600
    }
  }
}
```

### 3. 使用示例

```go
package main

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "xorm.io/xorm"
)

type User struct {
    ID   int64  `xorm:"pk autoincr"`
    Name string `xorm:"varchar(255)"`
    Age  int
}

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("xorm-demo"),
    )
    defer app.Stop()
    
    // 获取 XORM Engine 实例
    engine := core.MustGetBean[*xorm.Engine](app.Container())
    
    // 同步表结构
    engine.Sync2(&User{})
    
    // 创建记录
    engine.Insert(&User{Name: "John", Age: 30})
    
    // 查询记录
    var user User
    engine.ID(1).Get(&user)
    
    // 更新记录
    engine.ID(1).Update(&User{Age: 31})
    
    // 删除记录
    engine.ID(1).Delete(&User{})
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `db.xorm.enabled` | bool | false | 是否启用 XORM |
| `db.xorm.host` | string | localhost | 数据库主机地址 |
| `db.xorm.port` | int | 3306 | 数据库端口 |
| `db.xorm.username` | string | root | 数据库用户名 |
| `db.xorm.password` | string | root | 数据库密码 |
| `db.xorm.database` | string | enhance | 数据库名称 |
| `db.xorm.charset` | string | utf8mb4 | 字符集 |
| `db.xorm.max-open-conns` | int | 100 | 最大打开连接数 |
| `db.xorm.max-idle-conns` | int | 10 | 最大空闲连接数 |
| `db.xorm.conn-max-lifetime` | int | 3600 | 连接最大生命周期（秒） |

## 高级用法

### 事务支持

```go
engine := core.MustGetBean[*xorm.Engine](app.Container())

// 使用事务
session := engine.NewSession()
defer session.Close()

err := session.Begin()
if err != nil {
    // 处理错误
}

// 执行数据库操作
session.Insert(&User{Name: "John"})
session.Insert(&User{Name: "Jane"})

// 提交事务
err = session.Commit()
if err != nil {
    // 处理错误
}
```

### 缓存支持

```go
// 启用缓存
cacher := xorm.NewLRUCacher2(xorm.NewMemoryStore(), 1000)
engine.SetDefaultCacher(cacher)
```

### 条件查询

```go
var users []User
engine.Where("age > ?", 18).
    OrderBy("id DESC").
    Limit(10).
    Find(&users)
```

## 启动顺序

- **优先级**: `OrderPriorityDataLayer` (-2000)
- **触发条件**: `db.xorm.enabled=true`

## 依赖

- `xorm.io/xorm`
- `github.com/go-sql-driver/mysql`