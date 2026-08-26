package stdlib

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xudefa/enhance/web/engine"
)

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

func TestFactory_CreateServer_WithOptions(t *testing.T) {
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
