package stdlib

import (
	"testing"

	"github.com/xudefa/enhance/web/engine"
)

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
