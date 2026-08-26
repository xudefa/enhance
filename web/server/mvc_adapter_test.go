package server

import (
	"net/http"
	"testing"
)

func TestNewHttpServerAdapter_Helper(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter(WithHost("127.0.0.1:0"))
	if adapter == nil {
		t.Fatal("NewHttpServerAdapter returned nil")
	}
	if adapter.server == nil {
		t.Error("server should be set")
	}
}

func TestHttpServerAdapter_SetHandler_Helper(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter(WithHost("127.0.0.1:0"))

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	adapter.SetHandler(handler)

	_ = called
}

func TestHttpServerAdapter_Use_Helper(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter(WithHost("127.0.0.1:0"))

	middlewareCalled := false
	adapter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	})

	_ = middlewareCalled
}

func TestHttpServerAdapter_GetServer_Helper(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter(WithHost("127.0.0.1:0"))

	srv := adapter.GetServer()
	if srv == nil {
		t.Error("GetServer should not return nil")
	}
	if srv != adapter.server {
		t.Error("GetServer should return the same server instance")
	}
}

func TestHttpServerAdapter_StopWithoutStart_Helper(t *testing.T) {
	t.Parallel()
	adapter := NewHttpServerAdapter(WithHost("127.0.0.1:0"))

	err := adapter.Stop(t.Context())
	if err != nil {
		t.Errorf("Stop on non-started server should not error, got %v", err)
	}
}
