// Package mvc 提供 MVC 控制器支持。
package mvc

import "context"

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
