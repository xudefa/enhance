// Package web 提供 Web 框架支持，用于 enhance 框架。
//
// 该模块提供控制器扫描和路由注册、中间件支持、参数绑定和响应序列化等 Web 开发核心功能。
// 参考 Spring MVC 的设计理念，采用面向接口编程，支持灵活更换网络库和 Web 框架。
//
// # 架构设计
//
// web 包采用分层架构:
//   - core: 核心接口定义(Context, Router, Server, Controller)
//   - engine: 引擎层(引擎注册表、工厂、适配器)
//   - engine/stdlib: 标准库实现
//   - mvc: MVC 框架层(WebStarter, 控制器注册)
//   - middleware: 中间件(RequestID, AccessLog, Error, CORS)
//   - binding: 参数绑定
//
// # 核心接口
//
//   - Context: HTTP 请求上下文接口
//   - Router: 路由器接口，负责路由注册和匹配
//   - Server: HTTP 服务器接口，定义服务器生命周期
//   - Controller: 控制器接口，处理 HTTP 请求
//   - HandlerFunc: HTTP 处理函数
//   - MiddlewareFunc: 中间件函数类型
//
// # 使用方式
//
// 使用 MVC 框架:
//
//	starter := mvc.NewWebStarter(
//	    mvc.WithConfig(mvc.DefaultConfig()),
//	    mvc.WithRouter(router),
//	    mvc.WithServer(server),
//	)
//	starter.Start()
//
// 直接使用引擎:
//
//	router := stdlib.NewRouter()
//	router.GET("/hello", func(ctx core.Context) {
//	    ctx.String(200, "Hello, World!")
//	})
//
//	server := stdlib.NewServer(
//	    engine.WithHost("0.0.0.0"),
//	    engine.WithPort(8080),
//	)
//	server.SetHandler(router)
//	server.Start()
//
// # 支持的引擎
//
// 默认引擎（已注册）：
//   - engine.StdLib: 标准库 net/http
//
// # 扩展指南
//
// 添加新的网络引擎：
//  1. 实现 engine.Factory 接口
//  2. 实现 core.Router 接口（或使用 engine.RouterAdapter）
//  3. 实现 core.Server 接口（或使用 engine.ServerAdapter）
//  4. 实现 core.Context 接口
//  5. 在 init() 中注册到 engine.GlobalRegistry
package web

import (
	"github.com/xudefa/enhance/web/core"
	"github.com/xudefa/enhance/web/engine"
	"github.com/xudefa/enhance/web/mvc"
)

// ==================== 核心接口 ====================

// Context HTTP 请求上下文接口。
type Context = core.Context

// Router 路由器接口。
type Router = core.Router

// Server HTTP 服务器接口。
type Server = core.Server

// Controller 控制器接口。
type Controller = core.Controller

// HandlerFunc HTTP 处理函数。
type HandlerFunc = core.HandlerFunc

// MiddlewareFunc 中间件函数类型。
type MiddlewareFunc = core.MiddlewareFunc

// ==================== 引擎相关 ====================

// EngineFactory 引擎工厂接口。
type EngineFactory = engine.Factory

// EngineType 引擎类型。
type EngineType = engine.Type

// EngineRegistry 引擎注册表。
type EngineRegistry = engine.Registry

// ServerOption 服务器配置选项。
type ServerOption = engine.ServerOption

// ServerConfig 服务器配置。
type ServerConfig = engine.ServerConfig

// ==================== MVC 相关 ====================

// WebStarter Web 启动器。
type WebStarter = mvc.WebStarter

// WebConfig Web 配置。
type WebConfig = mvc.WebConfig
