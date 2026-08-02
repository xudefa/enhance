// Package core 提供 Web 框架核心接口定义。
//
// 该模块定义了 Web 框架的核心接口，包括 HTTP 上下文、路由器、服务器和控制器。
// 所有 HTTP 框架实现都应遵循这些接口，以实现框架无关的代码。
//
// # 核心接口
//
//   - Context: HTTP 请求上下文接口，封装请求和响应操作
//   - Router: 路由器接口，负责路由注册和匹配
//   - Server: HTTP 服务器接口，定义服务器生命周期
//   - Controller: 控制器接口，处理 HTTP 请求
//   - HandlerFunc: HTTP 处理函数类型
//   - MiddlewareFunc: 中间件函数类型
//
// # 设计原则
//
// 接口设计遵循最小化原则，每个接口只包含必要的方法。
// 这使得实现者可以轻松地提供完整的实现，而不需要实现不必要的方法。
//
// # 使用方式
//
// 实现 Context 接口：
//
//	type MyContext struct {
//	    // 实现所有 Context 方法
//	}
//
// 使用 Router 接口：
//
//	router := myRouter.New()
//	router.GET("/users", handler)
//	router.POST("/users", createHandler)
package core

import (
	"context"
	"net/http"
)

// ==================== 核心接口 ====================

// Context HTTP 请求上下文接口。
//
// 封装 HTTP 请求和响应操作，提供统一的处理接口。
// 所有 HTTP 框架实现都应实现此接口。
type Context interface {
	// RequestMethod 返回请求方法(GET、POST 等)。
	RequestMethod() string

	// RequestURI 返回请求 URI。
	RequestURI() string

	// PathParam 获取路径参数，如 /users/:id 中的 id。
	PathParam(name string) string

	// Query 获取查询参数，如 ?name=value 中的 name。
	Query(name string) string

	// QueryDefault 获取查询参数(带默认值)，参数不存在时返回默认值。
	QueryDefault(name, defaultVal string) string

	// Header 获取指定请求头的值。
	Header(key string) string

	// BindJSON 解析 JSON 请求体到目标结构。
	BindJSON(target any) error

	// SetStatusCode 设置响应状态码。
	SetStatusCode(code int)

	// SetHeader 设置响应头。
	SetHeader(key, value string)

	// JSON 返回 JSON 响应。
	JSON(code int, data any) error

	// String 返回字符串响应。
	String(code int, format string, args ...any)

	// AbortWithStatus 中止请求并返回指定状态码。
	AbortWithStatus(code int)

	// AbortWithStatusJSON 中止请求并返回 JSON 响应。
	AbortWithStatusJSON(code int, body any)

	// Next 调用下一个中间件或处理器。
	Next()

	// IsAborted 判断请求是否已被中止。
	IsAborted() bool

	// Context 获取请求上下文。
	Context() context.Context

	// Request 获取底层 HTTP 请求。
	Request() *http.Request

	// SetContext 设置请求上下文。
	SetContext(ctx context.Context)
}

// Router 路由器接口。
//
// 提供路由注册和路由组功能。
type Router interface {
	// GET 注册 GET 路由。
	GET(path string, handler HandlerFunc)

	// POST 注册 POST 路由。
	POST(path string, handler HandlerFunc)

	// PUT 注册 PUT 路由。
	PUT(path string, handler HandlerFunc)

	// DELETE 注册 DELETE 路由。
	DELETE(path string, handler HandlerFunc)

	// PATCH 注册 PATCH 路由。
	PATCH(path string, handler HandlerFunc)

	// Group 创建路由组。
	Group(prefix string) Router

	// Use 注册中间件。
	Use(middleware MiddlewareFunc)
}

// Server HTTP 服务器接口。
//
// 定义 HTTP 服务器的生命周期管理。
type Server interface {
	// Start 启动服务器。
	Start() error

	// Stop 停止服务器。
	Stop(ctx context.Context) error

	// SetHandler 设置 HTTP 处理器。
	SetHandler(handler any)

	// Use 注册全局中间件。
	Use(middleware any)
}

// Controller 控制器接口。
//
// 控制器负责处理 HTTP 请求。
// 实现此接口的结构体可以自动注册到路由器。
type Controller interface {
	// Routes 注册路由到路由器。
	//
	// 控制器在此方法中注册所有路由到给定的路由器。
	Routes(router Router)
}

// ==================== 函数类型 ====================

// HandlerFunc HTTP 处理函数。
type HandlerFunc func(ctx Context)

// MiddlewareFunc 中间件函数。
type MiddlewareFunc func(ctx Context)
