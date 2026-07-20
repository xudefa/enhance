# GORM Starter

GORM ORM 框架自动配置模块，提供数据库访问支持。

## 功能特性

- ✅ 自动配置 GORM 数据库连接
- ✅ 连接池管理
- ✅ 自动迁移支持
- ✅ 多数据库驱动支持（MySQL/PostgreSQL/SQLite）
- ✅ 日志集成
- ✅ 生命周期管理

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/gorm"
)
```

### 2. 配置文件

在 `application.json` 中添加 GORM 配置：

```json
{
  "db": {
    "gorm": {
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
    gormlib "gorm.io/gorm"
)

type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:255"`
    Age  int
}

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("gorm-demo"),
    )
    defer app.Stop()
    
    // 获取 GORM DB 实例
    db := core.MustGetBean[*gormlib.DB](app.Container())
    
    // 自动迁移
    db.AutoMigrate(&User{})
    
    // 创建记录
    db.Create(&User{Name: "John", Age: 30})
    
    // 查询记录
    var user User
    db.First(&user, 1)
    
    // 更新记录
    db.Model(&user).Update("Age", 31)
    
    // 删除记录
    db.Delete(&user)
}
```

## 配置说明

### 基础配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `db.gorm.enabled` | bool | false | 是否启用 GORM |
| `db.gorm.driver` | string | mysql | 数据库驱动（mysql/postgres/sqlite） |
| `db.gorm.host` | string | localhost | 数据库主机地址 |
| `db.gorm.port` | int | 3306 | 数据库端口 |
| `db.gorm.username` | string | root | 数据库用户名 |
| `db.gorm.password` | string | root | 数据库密码 |
| `db.gorm.database` | string | enhance | 数据库名称 |
| `db.gorm.charset` | string | utf8mb4 | 字符集 |
| `db.gorm.max-open-conns` | int | 100 | 最大打开连接数 |
| `db.gorm.max-idle-conns` | int | 10 | 最大空闲连接数 |
| `db.gorm.conn-max-lifetime` | int | 3600 | 连接最大生命周期（秒） |

### 多数据库配置示例

#### MySQL 配置

```json
{
  "db": {
    "gorm": {
      "enabled": true,
      "driver": "mysql",
      "host": "localhost",
      "port": 3306,
      "username": "root",
      "password": "root",
      "database": "enhance",
      "charset": "utf8mb4"
    }
  }
}
```

#### PostgreSQL 配置

```json
{
  "db": {
    "gorm": {
      "enabled": true,
      "driver": "postgres",
      "host": "localhost",
      "port": 5432,
      "username": "postgres",
      "password": "postgres",
      "database": "enhance"
    }
  }
}
```

#### SQLite 配置

```json
{
  "db": {
    "gorm": {
      "enabled": true,
      "driver": "sqlite",
      "database": "enhance.db"
    }
  }
}
```

## 高级用法

### 事务支持

```go
db := core.MustGetBean[*gormlib.DB](app.Container())

// 使用事务
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Error; err != nil {
    // 处理错误
}

// 执行数据库操作
tx.Create(&User{Name: "John"})
tx.Create(&User{Name: "Jane"})

// 提交事务
tx.Commit()
```

### 钩子函数

```go
type User struct {
    ID        uint      `gorm:"primaryKey"`
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// BeforeCreate 钩子
func (u *User) BeforeCreate(tx *gormlib.DB) error {
    u.CreatedAt = time.Now()
    u.UpdatedAt = time.Now()
    return nil
}
```

### 预加载关联

```go
type Post struct {
    ID      uint
    Title   string
    User    User
    UserID  uint
}

// 预加载用户
var posts []Post
db.Preload("User").Find(&posts)
```

## 启动顺序

- **优先级**: `OrderPriorityDataLayer` (-2000)
- **触发条件**: `db.gorm.enabled=true`

## 依赖

- `gorm.io/gorm`
- `gorm.io/driver/mysql`
- `gorm.io/driver/postgres`
- `gorm.io/driver/sqlite`