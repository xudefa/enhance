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
func TestSecurityFilterChainAdapter(t *testing.T) {
	t.Parallel()

	t.Run("DoFilter with valid types", func(t *testing.T) {
		p := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
		adapter := &securityFilterChainAdapter{proxy: p}
		err := adapter.DoFilter(context.Background(), &mockSecurityRequest{}, &mockSecurityResponse{})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("DoFilter with invalid context", func(t *testing.T) {
		p := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
		adapter := &securityFilterChainAdapter{proxy: p}
		err := adapter.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("DoFilter with invalid request", func(t *testing.T) {
		p := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
		adapter := &securityFilterChainAdapter{proxy: p}
		err := adapter.DoFilter(context.Background(), "invalid", &mockSecurityResponse{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("DoFilter with invalid response", func(t *testing.T) {
		p := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
		adapter := &securityFilterChainAdapter{proxy: p}
		err := adapter.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid")
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})

	t.Run("Matches returns true for SecurityRequest", func(t *testing.T) {
		p := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
		adapter := &securityFilterChainAdapter{proxy: p}
		if !adapter.Matches(&mockSecurityRequest{}) {
			t.Error("expected Matches to return true for SecurityRequest")
		}
		if adapter.Matches("invalid") {
			t.Error("expected Matches to return false for non-SecurityRequest")
		}
	})

	t.Run("GetFilters returns copy of filters", func(t *testing.T) {
		p := newFilterChainProxy(nil, &DefaultSecurityFilterChain{})
		adapter := &securityFilterChainAdapter{proxy: p}
		filters := adapter.GetFilters()
		if len(filters) != 0 {
			t.Errorf("expected 0 filters, got %d", len(filters))
		}
	})
}

// TestFilterChainAdapter 测试过滤器链适配器

func TestFilterChainAdapter(t *testing.T) {
	t.Parallel()

	t.Run("DoFilter with valid types", func(t *testing.T) {
		chainCalled := false
		chain := &chainFuncFilter{doFilter: func(ctx interface{}, req interface{}, resp interface{}) error {
			chainCalled = true
			return nil
		}}
		p := newFilterChainProxy(nil, chain)
		vfc := &virtualFilterChain{proxy: p, index: 0}
		adapter := &filterChainAdapter{vfc: vfc}
		err := adapter.DoFilter(context.Background(), &mockSecurityRequest{}, &mockSecurityResponse{})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
		if !chainCalled {
			t.Error("expected chain to be called")
		}
	})

	t.Run("DoFilter with invalid context", func(t *testing.T) {
		adapter := &filterChainAdapter{vfc: &virtualFilterChain{}}
		err := adapter.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("DoFilter with invalid request", func(t *testing.T) {
		adapter := &filterChainAdapter{vfc: &virtualFilterChain{}}
		err := adapter.DoFilter(context.Background(), "invalid", &mockSecurityResponse{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("DoFilter with invalid response", func(t *testing.T) {
		adapter := &filterChainAdapter{vfc: &virtualFilterChain{}}
		err := adapter.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid")
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})

	t.Run("AddFilter and GetFilters", func(t *testing.T) {
		adapter := &filterChainAdapter{vfc: &virtualFilterChain{}}
		adapter.AddFilter(nil)
		if filters := adapter.GetFilters(); filters != nil {
			t.Errorf("expected nil, got %v", filters)
		}
	})
}

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
	val, ok := adapter.GetAttribute("testKey")
	if !ok || val != "testValue" {
		t.Errorf("expected attribute 'testValue', got %v (exists=%v)", val, ok)
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

	f := NewBasicAuthenticationFilter(nil)
	if f.Order() != 0 {
		t.Errorf("expected order 0, got %d", f.Order())
	}
}

// ============================================================
// http_security.go 测试
// ============================================================

// TestHttpSecurity_ExceptionHandling 测试异常处理配置

func TestHttpSecurity_ExceptionHandling(t *testing.T) {
	t.Parallel()

	httpSec := NewHttpSecurity().(*httpSecurity)
	handler := NewHttp403ForbiddenAccessDeniedHandler()
	entryPoint := NewHttp401UnauthorizedEntryPoint()

	result := httpSec.ExceptionHandling(handler, entryPoint)
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if httpSec.exceptionTranslationFilter == nil {
		t.Error("expected exceptionTranslationFilter to be set")
	}
}

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

// TestExpressionInterceptUrlRegistry 测试 URL 拦截注册的各种方法

func TestExpressionInterceptUrlRegistry(t *testing.T) {
	t.Parallel()

	httpSec := &httpSecurity{
		authorizeRules: make([]authorizeRule, 0),
	}

	t.Run("PermitAll", func(t *testing.T) {
		httpSec2 := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
		r := &expressionInterceptUrlRegistry{httpSecurity: httpSec2, patterns: []string{"/public/**"}}
		r.PermitAll()
		if len(httpSec2.authorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(httpSec2.authorizeRules))
		}
		if httpSec2.authorizeRules[0].attrs[0] != "permitAll" {
			t.Errorf("expected 'permitAll', got '%s'", httpSec2.authorizeRules[0].attrs[0])
		}
	})

	t.Run("Authenticated", func(t *testing.T) {
		httpSec2 := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
		r := &expressionInterceptUrlRegistry{httpSecurity: httpSec2, patterns: []string{"/api/**"}}
		r.Authenticated()
		if len(httpSec2.authorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(httpSec2.authorizeRules))
		}
	})

	t.Run("HasRole", func(t *testing.T) {
		httpSec2 := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
		r := &expressionInterceptUrlRegistry{httpSecurity: httpSec2, patterns: []string{"/admin/**"}}
		r.HasRole("ADMIN")
		if len(httpSec2.authorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(httpSec2.authorizeRules))
		}
	})

	t.Run("HasAnyRole", func(t *testing.T) {
		httpSec2 := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
		r := &expressionInterceptUrlRegistry{httpSecurity: httpSec2, patterns: []string{"/api/**"}}
		r.HasAnyRole("ADMIN", "USER")
		if len(httpSec2.authorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(httpSec2.authorizeRules))
		}
	})

	t.Run("HasAuthority", func(t *testing.T) {
		httpSec2 := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
		r := &expressionInterceptUrlRegistry{httpSecurity: httpSec2, patterns: []string{"/api/**"}}
		r.HasAuthority("READ")
		if len(httpSec2.authorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(httpSec2.authorizeRules))
		}
	})

	t.Run("HasAnyAuthority", func(t *testing.T) {
		httpSec2 := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
		r := &expressionInterceptUrlRegistry{httpSecurity: httpSec2, patterns: []string{"/api/**"}}
		r.HasAnyAuthority("READ", "WRITE")
		if len(httpSec2.authorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(httpSec2.authorizeRules))
		}
	})

	t.Run("DenyAll", func(t *testing.T) {
		httpSec2 := &httpSecurity{authorizeRules: make([]authorizeRule, 0)}
		r := &expressionInterceptUrlRegistry{httpSecurity: httpSec2, patterns: []string{"/secret/**"}}
		r.DenyAll()
		if len(httpSec2.authorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(httpSec2.authorizeRules))
		}
	})

	t.Run("addRule with empty patterns", func(t *testing.T) {
		r := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{}}
		result := r.addRule([]string{"permitAll"})
		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("addRule with empty attrs", func(t *testing.T) {
		r := &expressionInterceptUrlRegistry{httpSecurity: httpSec, patterns: []string{"/test"}}
		result := r.addRule([]string{})
		if result == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("AnyRequest", func(t *testing.T) {
		authorizer := &httpSecurityAuthorizer{httpSecurity: httpSec}
		reg := authorizer.AnyRequest()
		if reg == nil {
			t.Error("expected non-nil registry")
		}
	})
}

// TestWebSecurity 测试 WebSecurity 入口

func TestWebSecurity(t *testing.T) {
	t.Parallel()

	ws := NewWebSecurity()
	if ws == nil {
		t.Fatal("expected non-nil WebSecurity")
	}

	hs := ws.HttpSecurity()
	if hs == nil {
		t.Fatal("expected non-nil httpSecurity")
	}
	if hs.filters == nil {
		t.Error("expected filters to be initialized")
	}

	_, err := ws.Build()
	if err == nil {
		t.Error("expected error from WebSecurity.Build")
	}
}

// ============================================================
// logout.go 测试
// ============================================================

// TestLogoutFilter_AddLogoutHandler 测试添加登出处理器

func TestFilterSecurityInterceptor_Setters(t *testing.T) {
	t.Parallel()

	f := &FilterSecurityInterceptor{}

	source := NewExpressionBasedFilterInvocationSecurityMetadataSource()
	f.SetSecurityMetadataSource(source)
	if f.securityMetadataSource != source {
		t.Error("expected securityMetadataSource to be set")
	}

	manager := NewAffirmativeBased()
	f.SetAccessDecisionManager(manager)
	if f.accessDecisionManager != manager {
		t.Error("expected accessDecisionManager to be set")
	}

	authManager := &mockAuthenticationManager{}
	f.SetAuthenticationManager(authManager)
	if f.authenticationManager != authManager {
		t.Error("expected authenticationManager to be set")
	}
}

// TestFilterSecurityInterceptor_DoFilter_FilterApplied 测试同一请求只执行一次过滤

func TestFilterSecurityInterceptor_DoFilter_FilterApplied(t *testing.T) {
	t.Parallel()

	interceptor := NewFilterSecurityInterceptor(
		NewExpressionBasedFilterInvocationSecurityMetadataSource(),
		NewAffirmativeBased(),
		&mockAuthenticationManager{},
	)

	req := &mockSecurityRequest{method: "GET", uri: "/test"}
	req.SetAttribute(filterAppliedKey, true)
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	err := interceptor.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called")
	}
}

// TestMatchWildcard 测试通配符匹配

func TestMatchWildcard(t *testing.T) {
	t.Parallel()

	source := NewExpressionBasedFilterInvocationSecurityMetadataSource()
	source.AddMapping("/api/*/detail", []string{"authenticated"})

	req := NewHttpRequestAdapter(&http.Request{
		Method: "GET",
		URL:    mustParseURL("/api/user/detail"),
	})

	attrs, err := source.GetAttributes(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attrs) != 1 || attrs[0] != "authenticated" {
		t.Errorf("expected [authenticated], got %v", attrs)
	}

	// Test non-matching wildcard
	req2 := NewHttpRequestAdapter(&http.Request{
		Method: "GET",
		URL:    mustParseURL("/api/user/profile"),
	})
	attrs, err = source.GetAttributes(context.Background(), req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(attrs) != 0 {
		t.Errorf("expected 0 attributes, got %v", attrs)
	}
}

// TestHttp403ForbiddenEntryPoint 测试 403 入口点

func TestHttp403ForbiddenEntryPoint(t *testing.T) {
	t.Parallel()

	entryPoint := NewHttp403ForbiddenEntryPoint()
	resp := &mockSecurityResponse{}
	err := entryPoint.Commence(context.Background(), &mockSecurityRequest{}, resp, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 403 {
		t.Errorf("expected status 403, got %d", resp.statusCode)
	}
}

// TestHttp401UnauthorizedEntryPoint 测试 401 入口点

func TestHttp401UnauthorizedEntryPoint(t *testing.T) {
	t.Parallel()

	entryPoint := NewHttp401UnauthorizedEntryPoint()
	resp := &mockSecurityResponse{}
	err := entryPoint.Commence(context.Background(), &mockSecurityRequest{}, resp, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
}

// TestHttp403ForbiddenAccessDeniedHandler 测试 403 访问拒绝处理器

func TestHttp403ForbiddenAccessDeniedHandler(t *testing.T) {
	t.Parallel()

	handler := NewHttp403ForbiddenAccessDeniedHandler()
	resp := &mockSecurityResponse{}
	err := handler.Handle(context.Background(), &mockSecurityRequest{}, resp, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 403 {
		t.Errorf("expected status 403, got %d", resp.statusCode)
	}
}

// ============================================================
// password_encoder.go 测试
// ============================================================

// TestStandardPasswordEncoder 测试标准密码编码器
