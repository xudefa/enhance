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
	"context"

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

// Stats WebSocket 服务器统计信息。
type Stats struct {
	TotalConnections  int
	ActiveConnections int
	RoomsCount        int
	MessagesSent      int64
	MessagesReceived  int64
	BytesSent         int64
	BytesReceived     int64
}

// WebSocketServer WebSocket 服务器接口。
type WebSocketServer interface {
	Start() error
	Stop(ctx context.Context) error
	HandleMessage(event string, handler MessageHandler)
	Use(middleware WebSocketMiddleware)
}

// MessageHandler 消息处理器接口。
type MessageHandler interface {
	Handle(conn Connection, message []byte) error
}

// WebSocketMiddleware WebSocket 中间件接口。
type WebSocketMiddleware interface {
	Handle(conn Connection) error
}

// Connection WebSocket 连接接口。
type Connection interface {
	ID() string
	Send(message []byte) error
	Close() error
	IsClosed() bool
	SetAttribute(key string, value any)
	GetAttribute(key string) (any, bool)
	Join(roomID string) error
	Leave(roomID string) error
	Rooms() []string
}

// Room WebSocket 房间接口。
type Room interface {
	ID() string
	Broadcast(message []byte) error
	Members() []Connection
}
