// Package openapi 提供 OpenAPI 文档生成功能，用于 enhance 框架。
//
// 该模块自动从控制器注解生成 OpenAPI 3.0 规范文档，支持 Swagger UI 集成。
// 参考 SpringDoc OpenAPI 的设计理念。
//
// # 架构设计
//
//   - OpenAPI: OpenAPI 文档结构，定义 API 规范
//   - Operation: 操作定义，描述单个 API 端点
//   - Schema: 数据模型定义，描述请求和响应结构
//   - SwaggerUI: Swagger UI 集成，提供可视化文档界面
//
// # 核心功能
//
//   - 文档生成: 自动从代码注解生成 OpenAPI 3.0 文档
//   - Swagger UI: 集成 Swagger UI，提供交互式 API 文档
//   - 注解支持: 支持 @Operation、@Parameter、@Response 等注解
//   - 数据模型: 自动生成请求和响应的数据模型定义
//
// # 使用方式
//
// 在控制器中使用注解：
//
//	// @Operation(summary: "Get user by ID")
//	// @Parameter(name: "id", in: "path", required: true)
//	// @Response(code: 200, description: "Success")
//	func GetUser(c *gin.Context) {
//	    // 处理逻辑
//	}
//
// 启用 Swagger UI：
//
//	import _ "github.com/xudefa/enhance/openapi"
//
// 访问文档：
//
//	浏览器访问: http://localhost:8080/swagger/index.html
//
// # 配置属性
//
//   - openapi.enabled: 是否启用 OpenAPI 文档（默认 true）
//   - openapi.title: API 文档标题
//   - openapi.version: API 版本
//   - openapi.description: API 描述
//
// # 配置示例
//
// 环境变量：
//
//	export OPENAPI_ENABLED=true
//	export OPENAPI_TITLE="My API"
//	export OPENAPI_VERSION="1.0.0"
package openapi
