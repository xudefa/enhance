package server

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"
)

func TestNewTLSClientHelper(t *testing.T) {
	t.Parallel()
	c := NewTLSClient("https://localhost:8443")
	if c == nil {
		t.Fatal("NewTLSClient returned nil")
	}
	if c.baseURL != "https://localhost:8443" {
		t.Errorf("baseURL = %s, want https://localhost:8443", c.baseURL)
	}
}

func TestWithTLSConfigHelper(t *testing.T) {
	t.Parallel()
	cfg := &tls.Config{InsecureSkipVerify: true}
	opt := WithTLSConfig(cfg)

	builder := &TLSClientBuilder{}
	opt(builder)

	if builder.tlsConfig != cfg {
		t.Error("tlsConfig should be set")
	}
}

func TestWithInsecureTLSHelper(t *testing.T) {
	t.Parallel()
	opt := WithInsecureTLS()

	builder := &TLSClientBuilder{}
	opt(builder)

	if builder.tlsConfig == nil {
		t.Fatal("tlsConfig should be set")
	}
	if !builder.tlsConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
	if builder.tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Error("MinVersion should be TLS 1.2")
	}
}

func TestWithTLSRequestTimeoutHelper(t *testing.T) {
	t.Parallel()
	opt := WithTLSRequestTimeout(10 * time.Second)

	builder := &TLSClientBuilder{}
	opt(builder)

	if builder.timeout != 10*time.Second {
		t.Errorf("timeout = %v, want 10s", builder.timeout)
	}
}

func TestWithTLSDefaultHeaderHelper(t *testing.T) {
	t.Parallel()
	opt := WithTLSDefaultHeader("X-Custom", "value")

	builder := &TLSClientBuilder{}
	opt(builder)

	if builder.headers == nil {
		t.Fatal("headers should be initialized")
	}
	if builder.headers["X-Custom"] != "value" {
		t.Errorf("header X-Custom = %s, want value", builder.headers["X-Custom"])
	}
}

func TestWithTLSTransportHelper(t *testing.T) {
	t.Parallel()
	transport := &http.Transport{MaxIdleConns: 42}
	opt := WithTLSTransport(transport)

	builder := &TLSClientBuilder{}
	opt(builder)

	if builder.transport != transport {
		t.Error("transport should be set")
	}
	if builder.transport.MaxIdleConns != 42 {
		t.Errorf("MaxIdleConns = %d, want 42", builder.transport.MaxIdleConns)
	}
}

func TestNewTLSClient_AllOptionsHelper(t *testing.T) {
	t.Parallel()
	transport := &http.Transport{MaxIdleConns: 99}

	c := NewTLSClient("https://example.com",
		WithInsecureTLS(),
		WithTLSRequestTimeout(15*time.Second),
		WithTLSDefaultHeader("Authorization", "Bearer token"),
		WithTLSTransport(transport),
	)

	if c == nil {
		t.Fatal("NewTLSClient returned nil")
	}
	if c.baseURL != "https://example.com" {
		t.Errorf("baseURL = %s", c.baseURL)
	}
	if c.httpClient.Timeout != 15*time.Second {
		t.Errorf("Timeout = %v, want 15s", c.httpClient.Timeout)
	}
}
