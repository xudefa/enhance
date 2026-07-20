// Package webtest 提供 Web 测试工具，用于 enhance 框架。
//
// 该模块提供 HTTP 客户端测试工具，用于测试 Web 端点和控制器。
// 支持模拟 HTTP 请求、验证响应、测试中间件等功能。
//
// # 架构设计
//
//   - TestServer: 测试服务器接口，用于启动临时 HTTP 服务
//   - TestClient: 测试客户端接口，用于发送 HTTP 请求
//   - RequestBuilder: 请求构建器接口，支持链式配置
//   - ResponseVerifier: 响应验证器接口，提供便捷的断言方法
//
// # 核心功能
//
//   - 模拟请求: 支持 GET、POST、PUT、DELETE 等 HTTP 方法
//   - 响应验证: 支持状态码、响应头、响应体验证
//   - JSON 验证: 支持 JSON 响应体验证
//   - 中间件测试: 支持测试 HTTP 中间件
//
// # 使用方式
//
// 创建测试服务器：
//
//	server := webtest.NewTestServer(router)
//	defer server.Close()
//
// 发送请求：
//
//	resp := server.GET("/api/users").
//	    WithHeader("Authorization", "Bearer token").
//	    Execute()
//
// 验证响应：
//
//	resp.AssertStatus(200)
//	resp.AssertJSONPath("$.users[0].name", "John")
//	resp.AssertBodyContains("success")
//
// # 请求构建
//
// 支持链式构建请求：
//
//	server.POST("/api/users").
//	    WithJSON(map[string]any{"name": "John"}).
//	    WithHeader("Content-Type", "application/json").
//	    WithQuery("page", "1").
//	    Execute()
//
// # 响应断言
//
// 提供丰富的断言方法：
//
//   - AssertStatus: 断言 HTTP 状态码
//   - AssertHeader: 断言响应头
//   - AssertBody: 断言响应体
//   - AssertJSONPath: 断言 JSON 路径值
//   - AssertBodyContains: 断言响应体包含指定字符串
package webtest

// TestServer 测试服务器接口。
//
// 用于启动临时 HTTP 服务，测试 Web 端点和控制器。
type TestServer interface {
	// GET 发起 GET 请求并返回请求构建器。
	GET(path string) RequestBuilder

	// POST 发起 POST 请求并返回请求构建器。
	POST(path string) RequestBuilder

	// PUT 发起 PUT 请求并返回请求构建器。
	PUT(path string) RequestBuilder

	// DELETE 发起 DELETE 请求并返回请求构建器。
	DELETE(path string) RequestBuilder

	// Patch 发起 PATCH 请求并返回请求构建器。
	Patch(path string) RequestBuilder

	// Close 关闭测试服务器。
	Close() error

	// URL 获取服务器 URL。
	URL() string
}

// RequestBuilder 请求构建器接口。
//
// 支持链式构建 HTTP 请求。
type RequestBuilder interface {
	// WithHeader 设置请求头。
	WithHeader(name, value string) RequestBuilder

	// WithJSON 设置 JSON 请求体。
	WithJSON(data any) RequestBuilder

	// WithBody 设置请求体。
	WithBody(body []byte) RequestBuilder

	// WithQuery 设置查询参数。
	WithQuery(key, value string) RequestBuilder

	// Execute 发起请求并返回响应验证器。
	Execute() ResponseVerifier
}

// ResponseVerifier 响应验证器接口。
//
// 提供便捷的断言方法验证 HTTP 响应。
type ResponseVerifier interface {
	// AssertStatus 断言 HTTP 状态码。
	AssertStatus(expected int) ResponseVerifier

	// AssertHeader 断言响应头。
	AssertHeader(name, expected string) ResponseVerifier

	// AssertBody 断言响应体。
	AssertBody(expected string) ResponseVerifier

	// AssertBodyContains 断言响应体包含指定字符串。
	AssertBodyContains(substring string) ResponseVerifier

	// Body 获取响应体字符串。
	Body() string

	// StatusCode 获取 HTTP 状态码。
	StatusCode() int
}
