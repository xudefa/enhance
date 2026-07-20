# MongoDB Starter

MongoDB 文档数据库自动配置模块，提供 NoSQL 数据库支持。

## 功能特性

- ✅ 自动配置 MongoDB 连接
- ✅ 连接池管理
- ✅ 支持认证
- ✅ 优雅关闭支持
- ✅ 便捷的 Database/Collection 获取方法

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/mongodb"
)
```

### 2. 配置文件

在 `application.json` 中添加 MongoDB 配置：

```json
{
  "mongodb": {
    "enabled": true,
    "host": "localhost",
    "port": 27017,
    "username": "",
    "password": "",
    "database": "enhance",
    "auth-source": "admin",
    "max-pool-size": 100,
    "min-pool-size": 10,
    "connect-timeout": 10,
    "server-selection-timeout": 5
  }
}
```

### 3. 使用示例

```go
package main

import (
    "context"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type User struct {
    ID    string `bson:"_id,omitempty"`
    Name  string `bson:"name"`
    Email string `bson:"email"`
    Age   int    `bson:"age"`
}

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("mongodb-demo"),
    )
    defer app.Stop()
    
    // 获取 MongoDB Client 实例
    client := core.MustGetBean[*mongo.Client](app.Container())
    
    // 获取数据库
    db := client.Database("enhance")
    
    // 获取集合
    collection := db.Collection("users")
    
    ctx := context.Background()
    
    // 插入文档
    user := User{
        Name:  "John",
        Email: "john@example.com",
        Age:   30,
    }
    result, err := collection.InsertOne(ctx, user)
    if err != nil {
        // 处理错误
    }
    
    // 查询文档
    var found User
    err = collection.FindOne(ctx, bson.M{"_id": result.InsertedID}).Decode(&found)
    if err != nil {
        // 处理错误
    }
    
    // 更新文档
    _, err = collection.UpdateOne(
        ctx,
        bson.M{"_id": result.InsertedID},
        bson.M{"$set": bson.M{"age": 31}},
    )
    
    // 删除文档
    _, err = collection.DeleteOne(ctx, bson.M{"_id": result.InsertedID})
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `mongodb.enabled` | bool | false | 是否启用 MongoDB |
| `mongodb.host` | string | localhost | 数据库主机地址 |
| `mongodb.port` | int | 27017 | 数据库端口 |
| `mongodb.username` | string | "" | 数据库用户名 |
| `mongodb.password` | string | "" | 数据库密码 |
| `mongodb.database` | string | enhance | 数据库名称 |
| `mongodb.auth-source` | string | admin | 认证数据库 |
| `mongodb.max-pool-size` | int | 100 | 最大连接池大小 |
| `mongodb.min-pool-size` | int | 10 | 最小连接池大小 |
| `mongodb.connect-timeout` | int | 10 | 连接超时（秒） |
| `mongodb.server-selection-timeout` | int | 5 | 服务器选择超时（秒） |

## 高级用法

### 事务支持

```go
client := core.MustGetBean[*mongo.Client](app.Container())

// 启动会话
session, err := client.StartSession()
if err != nil {
    // 处理错误
}
defer session.EndSession(ctx)

// 执行事务
result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
    collection := client.Database("enhance").Collection("users")
    
    // 执行多个操作
    collection.InsertOne(sessCtx, bson.M{"name": "Alice"})
    collection.InsertOne(sessCtx, bson.M{"name": "Bob"})
    
    return nil, nil
})
```

### 索引管理

```go
collection := client.Database("enhance").Collection("users")

// 创建索引
indexModel := mongo.IndexModel{
    Keys:    bson.D{{Key: "email", Value: 1}},
    Options: options.Index().SetUnique(true),
}
name, err := collection.Indexes().CreateOne(ctx, indexModel)
```

### 聚合查询

```go
pipeline := mongo.Pipeline{
    bson.D{{Key: "$match", Value: bson.M{"age": bson.M{"$gte": 18}}}},
    bson.D{{Key: "$group", Value: bson.D{
        {Key: "_id", Value: "$city"},
        {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
    }}},
    bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
}

cursor, err := collection.Aggregate(ctx, pipeline)
```

## 启动顺序

- **优先级**: `OrderPriorityDataLayer` (-2000)
- **触发条件**: `mongodb.enabled=true`

## 依赖

- `go.mongodb.org/mongo-driver`