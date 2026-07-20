// Package gin 提供 Gin Web 框架自动配置。
//
// Gin 是 Go 语言中最流行的 Web 框架，以高性能和简洁 API 著称。
//
// 功能特性：
//   - 自动配置 Gin Web 服务器
//   - 支持 debug/release/test 模式
//   - 内置 Recover 和 Logger 中间件
//   - 优雅关闭支持
//   - 自定义中间件支持
//
// 配置示例：
//
//	{
//	  "gin": {
//	    "enabled": true,
//	    "host": "0.0.0.0",
//	    "port": 8080,
//	    "mode": "debug"
//	  }
//	}
//
// 使用示例：
//
//	engine := core.MustGetBean[*gin.Engine](app.Container())
//	engine.GET("/ping", func(c *gin.Context) {
//	    c.JSON(200, gin.H{"message": "pong"})
//	})
package gin
