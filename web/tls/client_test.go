package tls

import (
	"net/http"
	"testing"
	"time"

	"github.com/xudefa/enhance/web/server"
)

func TestNewTLSClient_Basic(t *testing.T) {
	t.Parallel()
	c := NewClient("https://localhost:8443")
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewTLSClient_WithInsecureTLS(t *testing.T) {
	t.Parallel()
	c := NewClient("https://localhost:8443", WithInsecureTLS())
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewTLSClient_WithTimeout(t *testing.T) {
	t.Parallel()
	c := NewClient("https://localhost:8443",
		WithTimeout(10*time.Second),
	)
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewTLSClient_WithDefaultHeader(t *testing.T) {
	t.Parallel()
	c := NewClient("https://localhost:8443",
		WithDefaultHeader("X-Custom", "val"),
	)
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewTLSClient_WithTransport(t *testing.T) {
	t.Parallel()
	transport := &http.Transport{MaxIdleConns: 42}
	c := NewClient("https://localhost:8443",
		WithTransport(transport),
	)
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewTLSClient_WithTLSConfig(t *testing.T) {
	t.Parallel()
	c := NewClient("https://localhost:8443",
		WithTLSConfig(nil),
	)
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewTLSClient_WithAllOptions(t *testing.T) {
	t.Parallel()
	transport := &http.Transport{MaxIdleConns: 50}
	c := NewClient("https://example.com",
		WithInsecureTLS(),
		WithTimeout(20*time.Second),
		WithDefaultHeader("Authorization", "Bearer tok"),
		WithTransport(transport),
		WithTLSConfig(nil),
	)
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestTLSClientOptions_AreFunctional(t *testing.T) {
	t.Parallel()
	opts := []ClientOption{
		WithInsecureTLS(),
		WithTimeout(5 * time.Second),
		WithDefaultHeader("K", "V"),
	}

	for _, opt := range opts {
		b := &ClientBuilder{
			baseURL: "https://test.com",
			timeout: server.DefaultTimeout,
		}
		opt(b)
	}

	b := &ClientBuilder{
		baseURL: "https://test.com",
		timeout: server.DefaultTimeout,
	}
	for _, opt := range opts {
		opt(b)
	}
}
