package mvc

import (
	"context"
	"testing"
)

func TestStats_Struct(t *testing.T) {
	t.Parallel()
	stats := Stats{
		TotalConnections:  10,
		ActiveConnections: 5,
		RoomsCount:        2,
		MessagesSent:      100,
		MessagesReceived:  50,
		BytesSent:         1024,
		BytesReceived:     512,
	}

	if stats.TotalConnections != 10 {
		t.Errorf("expected TotalConnections 10, got %d", stats.TotalConnections)
	}
	if stats.ActiveConnections != 5 {
		t.Errorf("expected ActiveConnections 5, got %d", stats.ActiveConnections)
	}
	if stats.RoomsCount != 2 {
		t.Errorf("expected RoomsCount 2, got %d", stats.RoomsCount)
	}
	if stats.MessagesSent != 100 {
		t.Errorf("expected MessagesSent 100, got %d", stats.MessagesSent)
	}
	if stats.MessagesReceived != 50 {
		t.Errorf("expected MessagesReceived 50, got %d", stats.MessagesReceived)
	}
	if stats.BytesSent != 1024 {
		t.Errorf("expected BytesSent 1024, got %d", stats.BytesSent)
	}
	if stats.BytesReceived != 512 {
		t.Errorf("expected BytesReceived 512, got %d", stats.BytesReceived)
	}
}

func TestMessageHandler_Interface(t *testing.T) {
	t.Parallel()
	var _ MessageHandler = (*testMessageHandler)(nil)
}

func TestWebSocketMiddleware_Interface(t *testing.T) {
	t.Parallel()
	var _ WebSocketMiddleware = (*testWebSocketMiddleware)(nil)
}

func TestConnection_Interface(t *testing.T) {
	t.Parallel()
	var _ Connection = (*testConnection)(nil)
}

func TestRoom_Interface(t *testing.T) {
	t.Parallel()
	var _ Room = (*testRoom)(nil)
}

func TestWebSocketServer_Interface(t *testing.T) {
	t.Parallel()
	var _ WebSocketServer = (*mockWebSocketServer)(nil)
}

// Test implementations

type testMessageHandler struct{}

func (h *testMessageHandler) Handle(conn Connection, message []byte) error {
	return nil
}

type testWebSocketMiddleware struct{}

func (m *testWebSocketMiddleware) Handle(conn Connection) error {
	return nil
}

type testConnection struct{}

func (c *testConnection) ID() string                         { return "test" }
func (c *testConnection) Send(message []byte) error          { return nil }
func (c *testConnection) Close() error                       { return nil }
func (c *testConnection) IsClosed() bool                     { return false }
func (c *testConnection) SetAttribute(key string, value any) {}
func (c *testConnection) GetAttribute(key string) (any, bool) {
	return nil, false
}
func (c *testConnection) Join(roomID string) error  { return nil }
func (c *testConnection) Leave(roomID string) error { return nil }
func (c *testConnection) Rooms() []string           { return []string{} }

type testRoom struct{}

func (r *testRoom) ID() string                     { return "test-room" }
func (r *testRoom) Broadcast(message []byte) error { return nil }
func (r *testRoom) Members() []Connection          { return []Connection{} }

type mockWebSocketServer struct{}

func (m *mockWebSocketServer) Start() error                   { return nil }
func (m *mockWebSocketServer) Stop(ctx context.Context) error { return nil }
func (m *mockWebSocketServer) HandleMessage(event string, handler MessageHandler) {
}
func (m *mockWebSocketServer) Use(middleware WebSocketMiddleware) {}
