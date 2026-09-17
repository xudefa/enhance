package security

import (
	"net/http"
	"testing"
)

// ============================================================
// http_adapter.go 测试
// ============================================================

// TestHttpRequestAdapter_AdditionalMethods 测试 HttpRequestAdapter 的其他方法

func TestHttpRequestAdapter_AdditionalMethods(t *testing.T) {
	t.Parallel()

	req := &http.Request{
		Method:     "POST",
		URL:        mustParseURL("/api/test?q=1"),
		RemoteAddr: "192.168.1.1:8080",
		Header:     http.Header{"X-Custom": {"value1"}},
	}

	adapter := NewHttpRequestAdapter(req)

	if adapter.GetHeader("X-Custom") != "value1" {
		t.Errorf("expected 'value1', got '%s'", adapter.GetHeader("X-Custom"))
	}
	if adapter.GetHeader("X-Missing") != "" {
		t.Error("expected empty string for missing header")
	}

	if adapter.RemoteAddress() != "192.168.1.1:8080" {
		t.Errorf("expected '192.168.1.1:8080', got '%s'", adapter.RemoteAddress())
	}

	adapter.SetAttribute("testKey", "testValue")
	attributeValue, ok := adapter.GetAttribute("testKey")
	if !ok || attributeValue != "testValue" {
		t.Errorf("expected attribute 'testValue', got %v (exists=%v)", attributeValue, ok)
	}

	// Test GetAttribute for non-existent key
	_, ok = adapter.GetAttribute("nonExistent")
	if ok {
		t.Error("expected false for non-existent attribute")
	}
}

// TestHttpResponseAdapter_StatusCodeGetter 测试 HttpResponseAdapter 的状态码获取

func TestHttpResponseAdapter_StatusCodeGetter(t *testing.T) {
	t.Parallel()

	rec := newMockResponseWriter()
	adapter := NewHttpResponseAdapter(rec)

	if adapter.StatusCode() != http.StatusOK {
		t.Errorf("expected default status %d, got %d", http.StatusOK, adapter.StatusCode())
	}

	adapter.SetStatusCode(http.StatusNotFound)
	if adapter.StatusCode() != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, adapter.StatusCode())
	}
}
