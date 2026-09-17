package security

import (
	"context"
	"net/http"
	"testing"
)

// ============================================================
// filter_chain.go 测试
// ============================================================

// TestSecurityFilterChainAdapter 测试安全过滤器链适配器
func testSecurityFilterChainAdapterDoFilterValid(t *testing.T) {
	t.Parallel()
	proxy := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
	adapter := &securityFilterChainAdapter{proxy: proxy}
	err := adapter.DoFilter(context.Background(), &mockSecurityRequest{}, &mockSecurityResponse{})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func testSecurityFilterChainAdapterDoFilterInvalidCtx(t *testing.T) {
	t.Parallel()
	proxy := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
	adapter := &securityFilterChainAdapter{proxy: proxy}
	err := adapter.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{})
	if err == nil {
		t.Error("expected error for invalid context")
	}
}

func testSecurityFilterChainAdapterDoFilterInvalidReq(t *testing.T) {
	t.Parallel()
	proxy := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
	adapter := &securityFilterChainAdapter{proxy: proxy}
	err := adapter.DoFilter(context.Background(), "invalid", &mockSecurityResponse{})
	if err == nil {
		t.Error("expected error for invalid request")
	}
}

func testSecurityFilterChainAdapterDoFilterInvalidResp(t *testing.T) {
	t.Parallel()
	proxy := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
	adapter := &securityFilterChainAdapter{proxy: proxy}
	err := adapter.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid")
	if err == nil {
		t.Error("expected error for invalid response")
	}
}

func testSecurityFilterChainAdapterMatches(t *testing.T) {
	t.Parallel()
	proxy := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
	adapter := &securityFilterChainAdapter{proxy: proxy}
	if !adapter.Matches(&mockSecurityRequest{}) {
		t.Error("expected Matches to return true for SecurityRequest")
	}
	if adapter.Matches("invalid") {
		t.Error("expected Matches to return false for non-SecurityRequest")
	}
}

func testSecurityFilterChainAdapterGetFilters(t *testing.T) {
	t.Parallel()
	proxy := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
	adapter := &securityFilterChainAdapter{proxy: proxy}
	filters := adapter.GetFilters()
	if len(filters) != 0 {
		t.Errorf("expected 0 filters, got %d", len(filters))
	}
}

func TestSecurityFilterChainAdapter(t *testing.T) {
	t.Parallel()
	t.Run("DoFilter with valid types", testSecurityFilterChainAdapterDoFilterValid)
	t.Run("DoFilter with invalid context", testSecurityFilterChainAdapterDoFilterInvalidCtx)
	t.Run("DoFilter with invalid request", testSecurityFilterChainAdapterDoFilterInvalidReq)
	t.Run("DoFilter with invalid response", testSecurityFilterChainAdapterDoFilterInvalidResp)
	t.Run("Matches returns true for SecurityRequest", testSecurityFilterChainAdapterMatches)
	t.Run("GetFilters returns copy of filters", testSecurityFilterChainAdapterGetFilters)
}

// TestFilterChainAdapter 测试过滤器链适配器

func testFilterChainAdapterDoFilterValid(t *testing.T) {
	t.Parallel()
	chainCalled := false
	chain := &chainFuncFilter{doFilter: func(ctx interface{}, req interface{}, resp interface{}) error {
		chainCalled = true
		return nil
	}}
	proxy := newFilterChainProxy(nil, chain)
	vfc := &virtualFilterChain{proxy: proxy, index: 0}
	adapter := &filterChainAdapter{vfc: vfc}
	err := adapter.DoFilter(context.Background(), &mockSecurityRequest{}, &mockSecurityResponse{})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if !chainCalled {
		t.Error("expected chain to be called")
	}
}

func testFilterChainAdapterDoFilterInvalidCtx(t *testing.T) {
	t.Parallel()
	adapter := &filterChainAdapter{vfc: &virtualFilterChain{}}
	err := adapter.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{})
	if err == nil {
		t.Error("expected error for invalid context")
	}
}

func testFilterChainAdapterDoFilterInvalidReq(t *testing.T) {
	t.Parallel()
	adapter := &filterChainAdapter{vfc: &virtualFilterChain{}}
	err := adapter.DoFilter(context.Background(), "invalid", &mockSecurityResponse{})
	if err == nil {
		t.Error("expected error for invalid request")
	}
}

func testFilterChainAdapterDoFilterInvalidResp(t *testing.T) {
	t.Parallel()
	adapter := &filterChainAdapter{vfc: &virtualFilterChain{}}
	err := adapter.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid")
	if err == nil {
		t.Error("expected error for invalid response")
	}
}

func testFilterChainAdapterAddGetFilters(t *testing.T) {
	t.Parallel()
	adapter := &filterChainAdapter{vfc: &virtualFilterChain{}}
	adapter.AddFilter(nil)
	if filters := adapter.GetFilters(); filters != nil {
		t.Errorf("expected nil, got %v", filters)
	}
}

func TestFilterChainAdapter(t *testing.T) {
	t.Parallel()
	t.Run("DoFilter with valid types", testFilterChainAdapterDoFilterValid)
	t.Run("DoFilter with invalid context", testFilterChainAdapterDoFilterInvalidCtx)
	t.Run("DoFilter with invalid request", testFilterChainAdapterDoFilterInvalidReq)
	t.Run("DoFilter with invalid response", testFilterChainAdapterDoFilterInvalidResp)
	t.Run("AddFilter and GetFilters", testFilterChainAdapterAddGetFilters)
}

// TestSecurityFilterChainHandler_SetNextHandler 测试设置下一个处理器

func TestSecurityFilterChainHandler_SetNextHandler(t *testing.T) {
	t.Parallel()

	handler := NewSecurityFilterChainHandler(&DefaultSecurityFilterChain{}, nil)
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	handler.SetNextHandler(next)
	handler.ServeHTTP(newMockResponseWriter(), &http.Request{URL: mustParseURL("/test")})

	if !nextCalled {
		t.Error("expected next handler to be called")
	}
}

// TestSecurityFilterChainHandler_AuthenticationPropagation 测试认证信息传播

func TestSecurityFilterChainHandler_AuthenticationPropagation(t *testing.T) {
	t.Parallel()

	chain := &chainFuncFilter{doFilter: func(ctx interface{}, req interface{}, resp interface{}) error {
		if r, ok := req.(SecurityRequest); ok {
			r.SetAttribute("security.currentAuthentication", &mockAuthentication{principal: "testuser", authenticated: true})
		}
		return nil
	}}

	handler := NewSecurityFilterChainHandler(chain, nil)
	rec := newMockResponseWriter()
	handler.ServeHTTP(rec, &http.Request{URL: mustParseURL("/test")})
}

// TestBasicAuthenticationFilter_Order 测试 BasicAuthenticationFilter 的 Order 方法

func TestBasicAuthenticationFilter_Order(t *testing.T) {
	t.Parallel()

	authFilter := NewBasicAuthenticationFilter(nil)
	if authFilter.Order() != 0 {
		t.Errorf("expected order 0, got %d", authFilter.Order())
	}
}

// ============================================================
// http_security.go 测试
// ============================================================

// TestDefaultSecurityFilterChain 测试默认安全过滤器链

func TestDefaultSecurityFilterChain(t *testing.T) {
	t.Parallel()

	chain := &DefaultSecurityFilterChain{}

	err := chain.DoFilter(context.Background(), &mockSecurityRequest{}, &mockSecurityResponse{})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	if !chain.Matches(&mockSecurityRequest{}) {
		t.Error("expected Matches to return true")
	}

	if filters := chain.GetFilters(); filters != nil {
		t.Errorf("expected nil, got %v", filters)
	}
}
