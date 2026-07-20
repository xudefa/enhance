// Package fiber 提供 Fiber Web 框架自动配置。
//
// Fiber 是基于 fasthttp 的高性能 Web 框架，性能优于 Gin。
//
// 功能特性：
//   - 自动配置 Fiber Web 服务器
//   - 支持 Prefork 模式
//   - 高性能路由和中间件
//   - 兼容 Express.js API
//
// 配置示例：
//
//	{
//	  "fiber": {
//	    "enabled": true,
//	    "host": "0.0.0.0",
//	    "port": 3000
//	  }
//	}
//
// 使用示例：
//
//	app := core.MustGetBean[*fiber.App](app.Container())
//	app.Get("/ping", func(c *fiber.Ctx) error {
//	    return c.SendString("pong")
//	})
package fiber
