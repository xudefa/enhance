// Package mongodb 提供 MongoDB 数据库自动配置。
//
// MongoDB 是最流行的文档型 NoSQL 数据库。
//
// 功能特性：
//   - 自动配置 MongoDB 连接
//   - 连接池管理
//   - 支持认证
//   - 优雅关闭支持
//
// 配置示例：
//
//	{
//	  "mongodb": {
//	    "enabled": true,
//	    "host": "localhost",
//	    "port": 27017,
//	    "database": "enhance"
//	  }
//	}
//
// 使用示例：
//
//	client := core.MustGetBean[*mongo.Client](app.Container())
//	db := client.Database("enhance")
//	collection := db.Collection("users")
//	collection.InsertOne(ctx, bson.M{"name": "John"})
package mongodb
