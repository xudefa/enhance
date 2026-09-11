package stdlib

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xudefa/enhance/web/engine"
)

func TestNewServer_Default(t *testing.T) {
	t.Parallel()
	s := NewServer()
	if s == nil {
		t.Fatal("expected server to be created")
	}
	if s.host == "" {
		t.Error("expected host to be set")
	}
}

func TestNewServer_WithOptions(t *testing.T) {
	t.Parallel()
	s := NewServer(
		engine.WithHost("localhost"),
		engine.WithPort(9090),
		engine.WithReadTimeout(60),
		engine.WithWriteTimeout(120),
		engine.WithIdleTimeout(180),
	)

	if s.host != "localhost:9090" {
		t.Errorf("expected host 'localhost:9090', got %s", s.host)
	}
	if s.readTimeout != 60*time.Second {
		t.Errorf("expected read timeout 60s, got %v", s.readTimeout)
	}
	if s.writeTimeout != 120*time.Second {
		t.Errorf("expected write timeout 120s, got %v", s.writeTimeout)
	}
	if s.idleTimeout != 180*time.Second {
		t.Errorf("expected idle timeout 180s, got %v", s.idleTimeout)
	}
}

func TestServer_SetHandler(t *testing.T) {
	t.Parallel()
	s := NewServer()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	s.SetHandler(handler)
	if s.handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestServer_Use(t *testing.T) {
	t.Parallel()
	s := NewServer()
	middleware := func(next http.Handler) http.Handler {
		return next
	}
	s.Use(middleware)
	s.mu.RLock()
	count := len(s.middlewares)
	s.mu.RUnlock()

	if count != 1 {
		t.Errorf("expected 1 middleware, got %d", count)
	}
}

func TestServer_Use_Multiple(t *testing.T) {
	t.Parallel()
	s := NewServer()
	middleware := func(next http.Handler) http.Handler {
		return next
	}
	s.Use(middleware)
	s.Use(middleware)
	s.Use(middleware)

	s.mu.RLock()
	count := len(s.middlewares)
	s.mu.RUnlock()

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
	f := &Factory{}
	if f.Type() != engine.StdLib {
		t.Errorf("expected StdLib type, got %v", f.Type())
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
	s := NewServer()

	middlewareCalled := false
	s.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	})

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	wrapped := s.wrapHTTPHandler(handler)
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
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func TestServer_Start_ListenAndServe(t *testing.T) {
	t.Parallel()

	port := getFreePort(t)

	s := NewServer(
		engine.WithHost("127.0.0.1"),
		engine.WithPort(port),
	)

	var called int32
	mux := http.NewServeMux()
	mux.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&called, 1)
		w.WriteHeader(http.StatusOK)
	})
	s.SetHandler(mux)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
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
	if err := s.Stop(ctx); err != nil {
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

func TestServer_Stop_ShutdownWaitsForInFlight(t *testing.T) {
	t.Parallel()

	port := getFreePort(t)

	s := NewServer(
		engine.WithHost("127.0.0.1"),
		engine.WithPort(port),
	)

	handlerStarted := make(chan struct{})
	handlerDone := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		close(handlerStarted)
		<-handlerDone
		w.WriteHeader(http.StatusOK)
	})
	s.SetHandler(mux)

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	addr := "http://127.0.0.1:" + fmt.Sprintf("%d", port)
	for i := 0; i < 50; i++ {
		time.Sleep(10 * time.Millisecond)
		resp, err := http.Get(addr + "/health")
		if err == nil {
			resp.Body.Close()
			break
		}
	}

	// Start the slow request in a goroutine
	slowDone := make(chan struct{})
	go func() {
		defer close(slowDone)
		http.Get(addr + "/slow")
	}()

	<-handlerStarted

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopDone := make(chan error, 1)
	go func() {
		stopDone <- s.Stop(ctx)
	}()

	close(handlerDone)

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

func TestNewServer_WithTLSOptions(t *testing.T) {
	t.Parallel()

	s := NewServer(
		engine.WithHost("localhost"),
		engine.WithPort(8443),
	)
	s.certFile = "/path/to/cert.pem"
	s.keyFile = "/path/to/key.pem"

	if s.certFile != "/path/to/cert.pem" {
		t.Errorf("certFile = %s, want /path/to/cert.pem", s.certFile)
	}
	if s.keyFile != "/path/to/key.pem" {
		t.Errorf("keyFile = %s, want /path/to/key.pem", s.keyFile)
	}
}

func TestServer_ConcurrentUseAndWrap(t *testing.T) {
	t.Parallel()
	s := NewServer()

	done := make(chan bool)

	go func() {
		for i := 0; i < 10; i++ {
			s.Use(func(next http.Handler) http.Handler {
				return next
			})
		}
		done <- true
	}()

	go func() {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		s.wrapHTTPHandler(handler)
		done <- true
	}()

	<-done
	<-done
}
