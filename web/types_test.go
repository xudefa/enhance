package web

import (
	"testing"
)

func TestEngineTypeConstants(t *testing.T) {
	t.Parallel()
	if EngineStdLib == "" {
		t.Error("expected EngineStdLib to be non-empty")
	}
}

func TestGlobalEngineRegistry(t *testing.T) {
	t.Parallel()
	if GlobalEngineRegistry == nil {
		t.Fatal("expected GlobalEngineRegistry to be initialized")
	}
}

func TestNewEngineRegistry(t *testing.T) {
	t.Parallel()
	registry := NewEngineRegistry()
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestDefaultServerConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultServerConfig()
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
}

func TestWithHost(t *testing.T) {
	t.Parallel()
	cfg := DefaultServerConfig()
	opt := WithHost("localhost")
	opt(cfg)
	if cfg.Host != "localhost" {
		t.Errorf("expected host 'localhost', got %s", cfg.Host)
	}
}

func TestWithPort(t *testing.T) {
	t.Parallel()
	cfg := DefaultServerConfig()
	opt := WithPort(9090)
	opt(cfg)
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
}

func TestWithReadTimeout(t *testing.T) {
	t.Parallel()
	cfg := DefaultServerConfig()
	opt := WithReadTimeout(30)
	opt(cfg)
	if cfg.ReadTimeout != 30 {
		t.Errorf("expected read timeout 30, got %d", cfg.ReadTimeout)
	}
}

func TestWithWriteTimeout(t *testing.T) {
	t.Parallel()
	cfg := DefaultServerConfig()
	opt := WithWriteTimeout(60)
	opt(cfg)
	if cfg.WriteTimeout != 60 {
		t.Errorf("expected write timeout 60, got %d", cfg.WriteTimeout)
	}
}

func TestWithIdleTimeout(t *testing.T) {
	t.Parallel()
	cfg := DefaultServerConfig()
	opt := WithIdleTimeout(120)
	opt(cfg)
	if cfg.IdleTimeout != 120 {
		t.Errorf("expected idle timeout 120, got %d", cfg.IdleTimeout)
	}
}
