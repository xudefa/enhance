// Package swagger 提供 Swagger API 文档自动配置。
//
// Swagger 是最流行的 API 文档生成工具。
//
// 功能特性：
//   - 自动配置 Swagger UI
//   - 支持 OpenAPI 规范
//   - 交互式 API 文档
//   - API 测试支持
//
// 配置示例：
//
//	{
//	  "swagger": {
//	    "enabled": true,
//	    "host": "localhost",
//	    "port": 8080,
//	    "url": "/swagger/*"
//	  }
//	}
//
// 使用示例：
//
//	swagger := core.MustGetBean[*swagger.SwaggerAutoConfiguration](app.Container())
//	engine.GET("/swagger/*", gin.WrapH(swagger.GetHandler()))
package swagger
