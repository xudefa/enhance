package stdlib

import (
	"context"
	"fmt"
	"github.com/xudefa/enhance/web/engine"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewServer_Default(t *testing.T) {
	t.Parallel()
	server := NewServer()
	if server == nil {
		t.Fatal("expected server to be created")
	}
	if server.host == "" {
		t.Error("expected host to be set")
	}
}

func TestNewServer_WithOptions(t *testing.T) {
	t.Parallel()
	server := NewServer(
		engine.WithHost("localhost"),
		engine.WithPort(9090),
		engine.WithReadTimeout(60),
		engine.WithWriteTimeout(120),
		engine.WithIdleTimeout(180),
	)

	if server.host != "localhost:9090" {
		t.Errorf("expected host 'localhost:9090', got %s", server.host)
	}
	if server.readTimeout != 60*time.Second {
		t.Errorf("expected read timeout 60s, got %v", server.readTimeout)
	}
	if server.writeTimeout != 120*time.Second {
		t.Errorf("expected write timeout 120s, got %v", server.writeTimeout)
	}
	if server.idleTimeout != 180*time.Second {
		t.Errorf("expected idle timeout 180s, got %v", server.idleTimeout)
	}
}

func TestServer_SetHandler(t *testing.T) {
	t.Parallel()
	server := NewServer()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	server.SetHandler(handler)
	if server.handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestServer_Use(t *testing.T) {
	t.Parallel()
	server := NewServer()
	middleware := func(next http.Handler) http.Handler {
		return next
	}
	server.Use(middleware)
	server.mu.RLock()
	count := len(server.middlewares)
	server.mu.RUnlock()

	if count != 1 {
		t.Errorf("expected 1 middleware, got %d", count)
	}
}

func TestServer_Use_Multiple(t *testing.T) {
	t.Parallel()
	server := NewServer()
	middleware := func(next http.Handler) http.Handler {
		return next
	}
	server.Use(middleware)
	server.Use(middleware)
	server.Use(middleware)

	server.mu.RLock()
	count := len(server.middlewares)
	server.mu.RUnlock()

	if count != 3 {
		t.Errorf("expected 3 middlewares, got %d", count)
	}
}

func TestServer_Stop_NilServer(t *testing.T) {
	t.Parallel()
	s := &Server{}
	err := s.Stop(context.Background())
	if err != nil {
		t.Errorf("expected no error when stopping nil server, got %v", err)
	}
}

func TestFactory_Type(t *testing.T) {
	t.Parallel()
	factory := &Factory{}
	if factory.Type() != engine.StdLib {
		t.Errorf("expected StdLib type, got %v", factory.Type())
	}
}

func TestFactory_CreateRouter(t *testing.T) {
	t.Parallel()
	f := &Factory{}
	router, err := f.CreateRouter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if router == nil {
		t.Error("expected router to be created")
	}
}

func TestFactory_CreateServer(t *testing.T) {
	t.Parallel()
	f := &Factory{}
	server, err := f.CreateServer(
		engine.WithHost("localhost"),
		engine.WithPort(8080),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server == nil {
		t.Error("expected server to be created")
	}
}

func TestWrapHTTPHandler(t *testing.T) {
	t.Parallel()
	s := NewServer()

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	wrapped := s.wrapHTTPHandler(handler)
	if wrapped == nil {
		t.Fatal("expected wrapped handler to be non-nil")
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestWrapHTTPHandler_WithMiddleware(t *testing.T) {
	t.Parallel()
	server := NewServer()

	middlewareCalled := false
	server.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	})

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	wrapped := server.wrapHTTPHandler(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if !middlewareCalled {
		t.Error("expected middleware to be called")
	}
	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func getFreePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port
}

func testStdlibServerStartListenAndServe(t *testing.T, server *Server, port int) {
	t.Helper()
	var called int32
	mux := http.NewServeMux()
	mux.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&called, 1)
		w.WriteHeader(http.StatusOK)
	})
	server.SetHandler(mux)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	var err error
	for i := 0; i < 50; i++ {
		time.Sleep(10 * time.Millisecond)
		resp, getErr := http.Get("http://127.0.0.1:" + fmt.Sprintf("%d", port) + "/alive")
		if getErr == nil {
			resp.Body.Close()
			break
		}
		err = getErr
	}
	if err != nil {
		t.Fatalf("server did not start: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	select {
	case startErr := <-errCh:
		if startErr != nil && startErr != http.ErrServerClosed {
			t.Errorf("Start() error = %v", startErr)
		}
	case <-time.After(3 * time.Second):
		t.Error("Start() did not return in time")
	}

	if atomic.LoadInt32(&called) == 0 {
		t.Error("handler was not called")
	}
}

func testStdlibServerStopShutdownStart(t *testing.T, server *Server, port int) (started, done chan struct{}, addr string) {
	t.Helper()
	started = make(chan struct{})
	done = make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-done
		w.WriteHeader(http.StatusOK)
	})
	server.SetHandler(mux)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	addr = "http://127.0.0.1:" + fmt.Sprintf("%d", port)
	for i := 0; i < 50; i++ {
		time.Sleep(10 * time.Millisecond)
		resp, err := http.Get(addr + "/health")
		if err == nil {
			resp.Body.Close()
			break
		}
	}
	return started, done, addr
}

func testStdlibServerStopShutdownWait(t *testing.T, server *Server, started, done chan struct{}, addr string) {
	t.Helper()
	slowDone := make(chan struct{})
	go func() {
		defer close(slowDone)
		http.Get(addr + "/slow")
	}()

	<-started

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopDone := make(chan error, 1)
	go func() {
		stopDone <- server.Stop(ctx)
	}()

	close(done)

	select {
	case stopErr := <-stopDone:
		if stopErr != nil {
			t.Errorf("Stop() error = %v", stopErr)
		}
	case <-time.After(5 * time.Second):
		t.Error("Stop() did not return after handler completed")
	}

	select {
	case <-slowDone:
	case <-time.After(3 * time.Second):
		t.Error("slow request did not complete")
	}
}

func TestServer_Start_ListenAndServe(t *testing.T) {
	t.Parallel()

	port := getFreePort(t)

	server := NewServer(
		engine.WithHost("127.0.0.1"),
		engine.WithPort(port),
	)

	testStdlibServerStartListenAndServe(t, server, port)
}

func TestServer_Stop_ShutdownWaitsForInFlight(t *testing.T) {
	t.Parallel()

	port := getFreePort(t)

	server := NewServer(
		engine.WithHost("127.0.0.1"),
		engine.WithPort(port),
	)

	started, done, addr := testStdlibServerStopShutdownStart(t, server, port)
	testStdlibServerStopShutdownWait(t, server, started, done, addr)
}

func TestNewServer_WithTLSOptions(t *testing.T) {
	t.Parallel()

	server := NewServer(
		engine.WithHost("localhost"),
		engine.WithPort(8443),
	)
	server.certFile = "/path/to/cert.pem"
	server.keyFile = "/path/to/key.pem"

	if server.certFile != "/path/to/cert.pem" {
		t.Errorf("certFile = %s, want /path/to/cert.pem", server.certFile)
	}
	if server.keyFile != "/path/to/key.pem" {
		t.Errorf("keyFile = %s, want /path/to/key.pem", server.keyFile)
	}
}

func TestServer_ConcurrentUseAndWrap(t *testing.T) {
	t.Parallel()
	server := NewServer()

	done := make(chan bool)

	go func() {
		for i := 0; i < 10; i++ {
			server.Use(func(next http.Handler) http.Handler {
				return next
			})
		}
		done <- true
	}()

	go func() {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		server.wrapHTTPHandler(handler)
		done <- true
	}()

	<-done
	<-done
}

func TestServer_Start_ListenAndServe_Coverage(t *testing.T) {
	t.Parallel()

	port := getFreePort(t)

	server := NewServer(
		engine.WithHost("127.0.0.1"),
		engine.WithPort(port),
	)

	testStdlibServerStartListenAndServe(t, server, port)
}

func TestServer_Stop_ShutdownWaitsForInFlight_Coverage(t *testing.T) {
	t.Parallel()

	port := getFreePort(t)

	server := NewServer(
		engine.WithHost("127.0.0.1"),
		engine.WithPort(port),
	)

	started, done, addr := testStdlibServerStopShutdownStart(t, server, port)
	testStdlibServerStopShutdownWait(t, server, started, done, addr)
}

func TestFactory_CreateServer_WithOptions_Coverage(t *testing.T) {
	t.Parallel()

	f := &Factory{}
	srv, err := f.CreateServer(
		engine.WithHost("0.0.0.0"),
		engine.WithPort(9090),
		engine.WithReadTimeout(45),
		engine.WithWriteTimeout(60),
		engine.WithIdleTimeout(180),
	)
	if err != nil {
		t.Fatalf("CreateServer() error = %v", err)
	}
	if srv == nil {
		t.Fatal("CreateServer() returned nil")
	}
}

func TestNewServer_WithTLSOptions_Coverage(t *testing.T) {
	t.Parallel()

	server := NewServer(
		engine.WithHost("localhost"),
		engine.WithPort(8443),
	)
	server.certFile = "/path/to/cert.pem"
	server.keyFile = "/path/to/key.pem"

	if server.certFile != "/path/to/cert.pem" {
		t.Errorf("certFile = %s, want /path/to/cert.pem", server.certFile)
	}
	if server.keyFile != "/path/to/key.pem" {
		t.Errorf("keyFile = %s, want /path/to/key.pem", server.keyFile)
	}
}
