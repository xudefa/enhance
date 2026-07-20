package engine

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/xudefa/enhance/web/core"
)

// mockFactory 模拟引擎工厂。
type mockFactory struct {
	engineType Type
}

func (m *mockFactory) Type() Type {
	return m.engineType
}

func (m *mockFactory) CreateRouter() (core.Router, error) {
	return nil, nil
}

func (m *mockFactory) CreateServer(opts ...ServerOption) (core.Server, error) {
	return nil, nil
}

func TestNewRegistry(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("expected registry to be created")
	}

	if registry.GetDefault() != StdLib {
		t.Errorf("expected default engine to be StdLib, got %s", registry.GetDefault())
	}
}

func TestRegister(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()
	factory := &mockFactory{engineType: "mock"}

	registry.Register(factory)

	if !registry.HasEngine("mock") {
		t.Error("expected mock engine to be registered")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()
	factory := &mockFactory{engineType: "mock"}

	registry.Register(factory)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for duplicate registration")
		}
	}()

	registry.Register(factory)
}

func TestGet(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()
	factory := &mockFactory{engineType: "mock"}

	registry.Register(factory)

	retrieved, err := registry.Get("mock")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved.Type() != "mock" {
		t.Errorf("expected mock engine, got %s", retrieved.Type())
	}
}

func TestGetNotRegistered(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()

	_, err := registry.Get("unknown")
	if err == nil {
		t.Fatal("expected error for unregistered engine")
	}
}

func TestSetDefault(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()
	factory := &mockFactory{engineType: "mock"}

	registry.Register(factory)
	registry.SetDefault("mock")

	if registry.GetDefault() != "mock" {
		t.Errorf("expected default to be mock, got %s", registry.GetDefault())
	}
}

func TestCreateRouter(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()
	factory := &mockFactory{engineType: "mock"}

	registry.Register(factory)
	registry.SetDefault("mock")

	router, err := registry.CreateRouter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if router != nil {
		t.Error("expected nil router from mock factory")
	}
}

func TestCreateServer(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()
	factory := &mockFactory{engineType: "mock"}

	registry.Register(factory)
	registry.SetDefault("mock")

	server, err := registry.CreateServer()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if server != nil {
		t.Error("expected nil server from mock factory")
	}
}

func TestListEngines(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()

	registry.Register(&mockFactory{engineType: "engine1"})
	registry.Register(&mockFactory{engineType: "engine2"})
	registry.Register(&mockFactory{engineType: "engine3"})

	engines := registry.ListEngines()

	if len(engines) != 3 {
		t.Fatalf("expected 3 engines, got %d", len(engines))
	}
}

func TestHasEngine(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()
	factory := &mockFactory{engineType: "mock"}

	registry.Register(factory)

	if !registry.HasEngine("mock") {
		t.Error("expected mock engine to exist")
	}

	if registry.HasEngine("unknown") {
		t.Error("expected unknown engine to not exist")
	}
}

func TestGlobalRegistry(t *testing.T) {
	t.Parallel()
	if GlobalRegistry == nil {
		t.Fatal("expected GlobalRegistry to be initialized")
	}
}

func TestServerConfigOptions(t *testing.T) {
	t.Parallel()
	config := &ServerConfig{}

	WithHost("0.0.0.0")(config)
	if config.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", config.Host)
	}

	WithPort(9090)(config)
	if config.Port != 9090 {
		t.Errorf("expected port 9090, got %d", config.Port)
	}

	WithReadTimeout(60)(config)
	if config.ReadTimeout != 60 {
		t.Errorf("expected read timeout 60, got %d", config.ReadTimeout)
	}

	WithWriteTimeout(60)(config)
	if config.WriteTimeout != 60 {
		t.Errorf("expected write timeout 60, got %d", config.WriteTimeout)
	}

	WithIdleTimeout(180)(config)
	if config.IdleTimeout != 180 {
		t.Errorf("expected idle timeout 180, got %d", config.IdleTimeout)
	}

	WithTLS("/cert.pem", "/key.pem")(config)
	if config.TLSCertFile != "/cert.pem" {
		t.Errorf("expected cert file /cert.pem, got %s", config.TLSCertFile)
	}

	if config.TLSKeyFile != "/key.pem" {
		t.Errorf("expected key file /key.pem, got %s", config.TLSKeyFile)
	}
}

func TestDefaultServerConfig(t *testing.T) {
	t.Parallel()
	config := DefaultServerConfig()

	if config.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", config.Host)
	}

	if config.Port != 8080 {
		t.Errorf("expected port 8080, got %d", config.Port)
	}

	if config.ReadTimeout != 30 {
		t.Errorf("expected read timeout 30, got %d", config.ReadTimeout)
	}

	if config.WriteTimeout != 30 {
		t.Errorf("expected write timeout 30, got %d", config.WriteTimeout)
	}

	if config.IdleTimeout != 120 {
		t.Errorf("expected idle timeout 120, got %d", config.IdleTimeout)
	}
}

func TestRegistryConcurrent(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()

	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			factory := &mockFactory{engineType: Type("engine" + string(rune(i)))}
			registry.Register(factory)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			registry.ListEngines()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			registry.HasEngine("test")
		}
		done <- true
	}()

	for i := 0; i < 3; i++ {
		<-done
	}
}

func BenchmarkRegistry_Register(b *testing.B) {
	registry := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		factory := &mockFactory{engineType: Type(fmt.Sprintf("engine%d", i))}
		registry.Register(factory)
	}
}

func BenchmarkRegistry_Get(b *testing.B) {
	registry := NewRegistry()
	registry.Register(&mockFactory{engineType: "test"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.Get("test")
	}
}

func BenchmarkRegistry_HasEngine(b *testing.B) {
	registry := NewRegistry()
	registry.Register(&mockFactory{engineType: "test"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.HasEngine("test")
	}
}

func BenchmarkRegistry_Concurrent(b *testing.B) {
	registry := NewRegistry()

	var counter int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := atomic.AddInt64(&counter, 1)
			factory := &mockFactory{engineType: Type(fmt.Sprintf("engine%d", id))}
			registry.Register(factory)
		}
	})
}
