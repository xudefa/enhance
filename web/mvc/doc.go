// Package mvc 提供 MVC 控制器支持。
//
// 该模块提供控制器注册、路由组、内容协商等 Web MVC 功能。
// 参考 Spring MVC 的设计理念，提供完整的 MVC 架构支持。
//
// # 架构设计
//
//   - Context: HTTP 请求上下文接口
//   - Router: 路由器接口，负责路由注册和匹配
//   - Controller: 控制器接口，处理 HTTP 请求
//   - Server: HTTP 服务器接口
//   - HandlerFunc: HTTP 处理函数
//   - MiddlewareFunc: 中间件函数类型
//   - WebSocketServer: WebSocket 服务器接口
//   - MessageHandler: 消息处理器接口
//   - WebSocketMiddleware: WebSocket 中间件接口
//   - Connection: WebSocket 连接接口
//   - Room: WebSocket 房间接口
//
// # 核心功能
//
//   - 控制器注册: 支持自动扫描和注册控制器
//   - 路由组: 支持路由分组和前缀
//   - 内容协商: 支持 JSON、XML、HTML 等多种响应格式
//   - 视图渲染: 支持模板引擎渲染视图
//   - 拦截器: 支持请求拦截和预处理
//
// # 使用方式
//
// 定义控制器：
//
//	type UserController struct {
//	    mvc.RestController
//	}
//
//	// @GetMapping("/users/{id}")
//	func (c *UserController) GetUser(ctx *mvc.Context) {
//	    id := ctx.Param("id")
//	    user := c.userService.GetByID(id)
//	    ctx.JSON(user)
//	}
//
// 注册路由：
//
//	router := mvc.NewRouter()
//	router.Group("/api").
//	    AddController(&UserController{})
//
// # 内容协商
//
// 支持根据 Accept 头自动协商响应格式：
//
//   - application/json: JSON 响应
//   - application/xml: XML 响应
//   - text/html: HTML 响应
package mvc

import (
	"github.com/xudefa/enhance/web/core"
)

// 核心接口重新导出。
type (
	Context        = core.Context
	Router         = core.Router
	Server         = core.Server
	Controller     = core.Controller
	HandlerFunc    = core.HandlerFunc
	MiddlewareFunc = core.MiddlewareFunc
)
