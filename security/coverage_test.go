package security

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
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
func TestLogoutFilter_AddLogoutHandler(t *testing.T) {
	t.Parallel()

	filter, err := NewLogoutFilter("/logout", nil)
	if err != nil {
		t.Fatalf("NewLogoutFilter error: %v", err)
	}

	handler := &mockLogoutHandler{}
	filter.AddLogoutHandler(handler)
	if len(filter.handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(filter.handlers))
	}
}

// TestLogoutFilter_DoFilter_InvalidTypes 测试登出过滤器的类型检查
func TestLogoutFilter_DoFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	f, err := NewLogoutFilter("/logout", nil)
	if err != nil {
		t.Fatalf("NewLogoutFilter error: %v", err)
	}

	t.Run("invalid context", func(t *testing.T) {
		err := f.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, filter.NewDefaultFilterChain())
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		err := f.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, filter.NewDefaultFilterChain())
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		err := f.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", filter.NewDefaultFilterChain())
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// TestLogoutFilter_DoFilter 测试登出过滤器
func TestLogoutFilter_DoFilter(t *testing.T) {
	t.Parallel()

	t.Run("no logout handlers", func(t *testing.T) {
		ctx := context.Background()
		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		f, err := NewLogoutFilter("/logout", nil)
		if err != nil {
			t.Fatalf("NewLogoutFilter error: %v", err)
		}

		err = f.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("non-logout URL passes through", func(t *testing.T) {
		ctx := context.Background()
		req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		f, err := NewLogoutFilter("/logout", nil)
		if err != nil {
			t.Fatalf("NewLogoutFilter error: %v", err)
		}

		err = f.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("expected chain to be called for non-logout URL")
		}
	})

	t.Run("GET method not allowed for logout", func(t *testing.T) {
		ctx := context.Background()
		req := &mockSecurityRequest{method: "GET", uri: "/logout"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}

		f, err := NewLogoutFilter("/logout", nil)
		if err != nil {
			t.Fatalf("NewLogoutFilter error: %v", err)
		}

		err = f.DoFilter(ctx, req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("expected chain to be called for GET logout")
		}
	})
}

// TestLogoutFilter_WithSuccessHandler 测试带成功处理器的登出过滤器
func TestLogoutFilter_WithSuccessHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	successHandler := &mockLogoutSuccessHandler{}
	f, err := NewLogoutFilter("/logout", nil)
	if err != nil {
		t.Fatalf("NewLogoutFilter error: %v", err)
	}
	f.SetSuccessHandler(successHandler)

	err = f.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !successHandler.called {
		t.Error("expected success handler to be called")
	}
}

// TestLogoutFilter_MustNew 测试 MustNewLogoutFilter
func TestLogoutFilter_MustNew(t *testing.T) {
	t.Parallel()

	t.Run("valid creation", func(t *testing.T) {
		f := MustNewLogoutFilter("/logout", nil)
		if f == nil {
			t.Fatal("expected non-nil filter")
		}
	})

	t.Run("panic on empty url", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for empty url")
			}
		}()
		MustNewLogoutFilter("", nil)
	})
}

// TestSimpleLogoutSuccessHandler 测试简单登出成功处理器
func TestSimpleLogoutSuccessHandler(t *testing.T) {
	t.Parallel()

	handler := NewSimpleLogoutSuccessHandler("/goodbye")
	resp := &mockSecurityResponse{headers: make(map[string]string)}
	handler.OnLogoutSuccess(context.Background(), &mockSecurityRequest{}, resp, nil)

	if resp.statusCode != 302 {
		t.Errorf("expected status 302, got %d", resp.statusCode)
	}
	if resp.headers["Location"] != "/goodbye" {
		t.Errorf("expected Location '/goodbye', got '%s'", resp.headers["Location"])
	}
}

// TestDefaultLogoutSuccessHandler 测试默认登出成功处理器
func TestDefaultLogoutSuccessHandler(t *testing.T) {
	t.Parallel()

	handler := NewDefaultLogoutSuccessHandler("/login?logout")
	resp := &mockSecurityResponse{headers: make(map[string]string)}
	handler.OnLogoutSuccess(context.Background(), &mockSecurityRequest{}, resp, nil)

	if resp.statusCode != 302 {
		t.Errorf("expected status 302, got %d", resp.statusCode)
	}
	if resp.headers["Location"] != "/login?logout" {
		t.Errorf("expected Location '/login?logout', got '%s'", resp.headers["Location"])
	}
}

// ============================================================
// filter.go 测试
// ============================================================

// TestFilterSecurityInterceptor_Setters 测试 FilterSecurityInterceptor 的 setter 方法
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
func TestStandardPasswordEncoder(t *testing.T) {
	t.Parallel()

	encoder := NewStandardPasswordEncoder("my-secret-key")

	password := "testPassword123"
	encoded := encoder.Encode(password)

	if encoded == password {
		t.Error("encoded password should differ from raw password")
	}

	if !encoder.Matches(password, encoded) {
		t.Error("should match correct password")
	}

	if encoder.Matches("wrongPassword", encoded) {
		t.Error("should not match wrong password")
	}

	// Same secret produces same hash
	encoder2 := NewStandardPasswordEncoder("my-secret-key")
	if encoder2.Encode(password) != encoded {
		t.Error("same secret should produce same hash")
	}

	// Different secret produces different hash
	encoder3 := NewStandardPasswordEncoder("different-secret")
	if encoder3.Encode(password) == encoded {
		t.Error("different secret should produce different hash")
	}
}

// TestMustNewDelegatingPasswordEncoder 测试 MustNewDelegatingPasswordEncoder
func TestMustNewDelegatingPasswordEncoder(t *testing.T) {
	t.Parallel()

	t.Run("valid creation", func(t *testing.T) {
		encoder := MustNewDelegatingPasswordEncoder("sha256", map[string]PasswordEncoder{
			"sha256": NewSha256PasswordEncoder(),
		})
		if encoder == nil {
			t.Fatal("expected non-nil encoder")
		}
	})

	t.Run("panic on invalid id", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic for invalid id")
			}
		}()
		MustNewDelegatingPasswordEncoder("unknown", map[string]PasswordEncoder{
			"sha256": NewSha256PasswordEncoder(),
		})
	})
}

// TestDelegatingPasswordEncoder_EncodeWithMissingEncoder 测试 Encode 时编码器丢失的情况
func TestDelegatingPasswordEncoder_EncodeWithMissingEncoder(t *testing.T) {
	t.Parallel()

	// 创建一个内部编码器映射为空的委托编码器（绕过构造检查）
	encoder := &DelegatingPasswordEncoder{
		idForEncode:      "missing",
		passwordEncoders: map[string]PasswordEncoder{},
	}

	result := encoder.Encode("password")
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

// ============================================================
// rate_limit.go 测试
// ============================================================

// TestSlidingWindowRateLimiter_Close 测试关闭限流器
func TestSlidingWindowRateLimiter_Close(t *testing.T) {
	t.Parallel()

	limiter := NewSlidingWindowRateLimiter(100*time.Millisecond, 10)
	limiter.Close()
	// Closing twice should not panic
	limiter.Close()
}

// TestEnhancedRateLimitFilter_ClientKey 测试客户端键计算
func TestEnhancedRateLimitFilter_ClientKey(t *testing.T) {
	t.Parallel()

	strategy := &mockRateLimitStrategy{}
	filter := NewEnhancedRateLimitFilter(strategy)

	t.Run("without trusted proxies", func(t *testing.T) {
		req := &mockSecurityRequest{method: "GET", uri: "/test"}
		req.SetHeader("X-Forwarded-For", "10.0.0.1")
		key := filter.clientKey(req)
		if key == "" {
			t.Error("expected non-empty key")
		}
	})

	t.Run("with trusted proxies", func(t *testing.T) {
		filterWithProxy := NewEnhancedRateLimitFilter(strategy, WithTrustedProxies("192.168.0.0/16"))
		req := &mockSecurityRequest{method: "GET", uri: "/test", remoteAddr: "192.168.1.1:8080"}
		req.SetHeader("X-Forwarded-For", "10.0.0.1")
		key := filterWithProxy.clientKey(req)
		if key == "" {
			t.Error("expected non-empty key")
		}
	})
}

// TestEnhancedRateLimitFilter_DoFilter 测试增强限流过滤器的各种场景
func TestEnhancedRateLimitFilter_DoFilter(t *testing.T) {
	t.Parallel()

	t.Run("exclude path", func(t *testing.T) {
		filterWithExclude := NewEnhancedRateLimitFilter(&mockRateLimitStrategy{}, WithExcludePaths("/health"))
		req := &mockSecurityRequest{method: "GET", uri: "/health"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := filterWithExclude.DoFilter(context.Background(), req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("expected chain to be called for excluded path")
		}
	})

	t.Run("rate limit triggered", func(t *testing.T) {
		denyStrategy := &mockRateLimitStrategy{allow: false}
		filterDeny := NewEnhancedRateLimitFilter(denyStrategy)
		req := &mockSecurityRequest{method: "GET", uri: "/api"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := filterDeny.DoFilter(context.Background(), req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.statusCode != 429 {
			t.Errorf("expected status 429, got %d", resp.statusCode)
		}
	})

	t.Run("custom onRateLimit callback", func(t *testing.T) {
		customCalled := false
		denyStrategy := &mockRateLimitStrategy{allow: false}
		filterCustom := NewEnhancedRateLimitFilter(denyStrategy, WithOnRateLimit(func(ctx context.Context, request SecurityRequest, response SecurityResponse) {
			customCalled = true
			response.SetStatusCode(429)
		}))
		req := &mockSecurityRequest{method: "GET", uri: "/api"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := filterCustom.DoFilter(context.Background(), req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !customCalled {
			t.Error("expected custom onRateLimit to be called")
		}
	})

	t.Run("allowed request passes through", func(t *testing.T) {
		allowStrategy := &mockRateLimitStrategy{allow: true}
		filterAllow := NewEnhancedRateLimitFilter(allowStrategy)
		req := &mockSecurityRequest{method: "GET", uri: "/api"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := filterAllow.DoFilter(context.Background(), req, resp, chain)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !chain.called {
			t.Error("expected chain to be called for allowed request")
		}
	})
}

// TestEnhancedRateLimitFilter_InvalidTypes 测试 EnhancedRateLimitFilter 的类型检查
func TestEnhancedRateLimitFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	f := NewEnhancedRateLimitFilter(&mockRateLimitStrategy{})

	t.Run("invalid context", func(t *testing.T) {
		err := f.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		err := f.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		err := f.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// ============================================================
// cors_filter.go 测试
// ============================================================

// TestCorsFilter_IsOriginAllowed 测试 CORS 来源验证
func TestCorsFilter_IsOriginAllowed(t *testing.T) {
	t.Parallel()

	t.Run("empty allowed origins returns false", func(t *testing.T) {
		filter := NewCorsFilter(CorsConfig{})
		if filter.isOriginAllowed("http://example.com") {
			t.Error("expected false for empty allowed origins")
		}
	})

	t.Run("exact match", func(t *testing.T) {
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"http://example.com"},
		})
		if !filter.isOriginAllowed("http://example.com") {
			t.Error("expected true for exact match")
		}
		if filter.isOriginAllowed("http://other.com") {
			t.Error("expected false for non-matching origin")
		}
	})

	t.Run("wildcard suffix match", func(t *testing.T) {
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"http://*.example.com"},
		})
		if !filter.isOriginAllowed("http://sub.example.com") {
			t.Error("expected true for wildcard match")
		}
		if !filter.isOriginAllowed("http://sub.example.com/") {
			t.Error("expected true for wildcard match with trailing slash")
		}
		if filter.isOriginAllowed("http://evil.com") {
			t.Error("expected false for non-matching origin")
		}
	})

	t.Run("star wildcard", func(t *testing.T) {
		filter := NewCorsFilter(CorsConfig{
			AllowedOrigins: []string{"*"},
		})
		if !filter.isOriginAllowed("http://any.com") {
			t.Error("expected true for * wildcard")
		}
	})
}

// TestCorsFilter_InvalidTypes 测试 CorsFilter 的类型检查
func TestCorsFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	f := NewCorsFilter(CorsConfig{})
	t.Run("invalid context", func(t *testing.T) {
		err := f.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})
	t.Run("invalid request", func(t *testing.T) {
		err := f.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})
	t.Run("invalid response", func(t *testing.T) {
		err := f.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// TestCorsFilter_DoFilter_OPTIONS 测试 CORS 预检请求
func TestCorsFilter_DoFilter_OPTIONS(t *testing.T) {
	t.Parallel()

	filter := NewCorsFilter(CorsConfig{
		AllowedOrigins:   []string{"http://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"X-Custom"},
		AllowCredentials: true,
		MaxAge:           7200,
	})

	req := &mockSecurityRequest{method: "OPTIONS", uri: "/api/test"}
	req.SetHeader("Origin", "http://example.com")
	resp := &mockSecurityResponse{headers: make(map[string]string)}
	chain := &mockSecurityFilterChain{}

	err := filter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.statusCode != 204 {
		t.Errorf("expected status 204, got %d", resp.statusCode)
	}
	if resp.headers["Access-Control-Allow-Origin"] != "http://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin header")
	}
	if resp.headers["Access-Control-Allow-Credentials"] != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials header")
	}
}

// ============================================================
// csrf.go 测试
// ============================================================

// TestCsrfFilter_NewCsrfFilter_Error 测试 NewCsrfFilter 的错误路径
func TestCsrfFilter_NewCsrfFilter_Error(t *testing.T) {
	t.Parallel()

	_, err := NewCsrfFilter(nil)
	if err == nil {
		t.Error("expected error for nil token repository")
	}
}

// TestCsrfFilter_MustNewCsrfFilter_Panic 测试 MustNewCsrfFilter 的 panic
func TestCsrfFilter_MustNewCsrfFilter_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil token repository")
		}
	}()
	MustNewCsrfFilter(nil)
}

// TestCsrfFilter_AddExcludePath_Error 测试 AddExcludePath 的错误路径
func TestCsrfFilter_AddExcludePath_Error(t *testing.T) {
	t.Parallel()

	repo := NewCookieCsrfTokenRepository()
	filter, err := NewCsrfFilter(repo)
	if err != nil {
		t.Fatalf("NewCsrfFilter error: %v", err)
	}

	t.Run("empty path", func(t *testing.T) {
		err := filter.AddExcludePath("")
		if err == nil {
			t.Error("expected error for empty path")
		}
	})

	t.Run("path without leading slash", func(t *testing.T) {
		err := filter.AddExcludePath("api/public")
		if err == nil {
			t.Error("expected error for path without leading slash")
		}
	})
}

// TestCsrfFilter_DoFilter_InvalidTypes 测试 CsrfFilter 的类型检查
func TestCsrfFilter_DoFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	repo := NewCookieCsrfTokenRepository()
	f, err := NewCsrfFilter(repo)
	if err != nil {
		t.Fatalf("NewCsrfFilter error: %v", err)
	}

	t.Run("invalid context", func(t *testing.T) {
		err := f.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		err := f.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		err := f.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// TestCsrfFilter_DoFilter_CSRFProtection 测试 CSRF 保护
func TestCsrfFilter_DoFilter_CSRFProtection(t *testing.T) {
	t.Parallel()

	repo := NewCookieCsrfTokenRepository()
	f, err := NewCsrfFilter(repo)
	if err != nil {
		t.Fatalf("NewCsrfFilter error: %v", err)
	}

	t.Run("POST without CSRF token returns error", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/data"}
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := f.DoFilter(context.Background(), req, resp, chain)
		if err == nil {
			t.Error("expected error for missing CSRF token")
		}
	})

	t.Run("POST with invalid CSRF token returns error", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/data"}
		req.SetHeader("X-CSRF-Token", "invalid-token")
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := f.DoFilter(context.Background(), req, resp, chain)
		if err == nil {
			t.Error("expected error for invalid CSRF token")
		}
	})

	t.Run("X-XSRF-Token header", func(t *testing.T) {
		req := &mockSecurityRequest{method: "POST", uri: "/api/data"}
		req.SetHeader("X-XSRF-Token", "invalid-token")
		resp := &mockSecurityResponse{}
		chain := &mockSecurityFilterChain{}
		err := f.DoFilter(context.Background(), req, resp, chain)
		if err == nil {
			t.Error("expected error for invalid X-XSRF-Token")
		}
	})
}

// TestCookieCsrfTokenRepository_LoadCookieValue 测试 Cookie 加载
func TestCookieCsrfTokenRepository_LoadCookieValue(t *testing.T) {
	t.Parallel()

	t.Run("cookie not found", func(t *testing.T) {
		repo := NewCookieCsrfTokenRepository()
		req := &mockSecurityRequest{method: "GET", uri: "/test"}
		_, exists := repo.loadCookieValue(req)
		if exists {
			t.Error("expected false for missing cookie")
		}
	})

	t.Run("cookie found", func(t *testing.T) {
		repo := NewCookieCsrfTokenRepository()
		req := &mockSecurityRequest{method: "GET", uri: "/test"}
		req.SetHeader("Cookie", "_csrf_token=abc123; other=val")
		val, exists := repo.loadCookieValue(req)
		if !exists {
			t.Error("expected true for existing cookie")
		}
		if val != "abc123" {
			t.Errorf("expected 'abc123', got '%s'", val)
		}
	})

	t.Run("ValidateToken with cookie", func(t *testing.T) {
		repo := NewCookieCsrfTokenRepository()
		req := &mockSecurityRequest{method: "POST", uri: "/test"}
		req.SetHeader("Cookie", "_csrf_token=valid-token-value")
		if !repo.ValidateToken(context.Background(), req, "valid-token-value") {
			t.Error("expected valid token to be validated")
		}
	})

	t.Run("ValidateToken with non-string attribute", func(t *testing.T) {
		repo := NewCookieCsrfTokenRepository()
		req := &mockSecurityRequest{method: "POST", uri: "/test"}
		req.SetAttribute("csrf.token", 12345) // non-string attribute
		if repo.ValidateToken(context.Background(), req, "token") {
			t.Error("expected false for non-string attribute")
		}
	})
}

// TestSameSiteString 测试 SameSite 字符串转换
func TestSameSiteString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode http.SameSite
		want string
	}{
		{http.SameSiteLaxMode, "Lax"},
		{http.SameSiteStrictMode, "Strict"},
		{http.SameSiteNoneMode, "None"},
		{http.SameSiteDefaultMode, "Lax"},
	}

	for _, tt := range tests {
		got := sameSiteString(tt.mode)
		if got != tt.want {
			t.Errorf("sameSiteString(%d) = %s, want %s", tt.mode, got, tt.want)
		}
	}
}

// ============================================================
// config.go 测试
// ============================================================

// TestConfigRegistry_AddRule 测试 configRegistry 的 addRule 方法
func TestConfigRegistry_AddRule(t *testing.T) {
	t.Parallel()

	t.Run("empty patterns returns nil", func(t *testing.T) {
		r := &configRegistry{cfg: &SecurityConfig{}, patterns: []string{}}
		result := r.addRule([]string{"permitAll"})
		if result != nil {
			t.Error("expected nil for empty patterns")
		}
	})

	t.Run("empty attrs returns nil", func(t *testing.T) {
		r := &configRegistry{cfg: &SecurityConfig{}, patterns: []string{"/test"}}
		result := r.addRule([]string{})
		if result != nil {
			t.Error("expected nil for empty attrs")
		}
	})

	t.Run("valid rule added", func(t *testing.T) {
		cfg := &SecurityConfig{}
		r := &configRegistry{cfg: cfg, patterns: []string{"/api/**"}}
		result := r.addRule([]string{"authenticated"})
		if result != nil {
			t.Error("expected nil for valid rule")
		}
		if len(cfg.AuthorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(cfg.AuthorizeRules))
		}
	})
}

// TestConfigAuthorizer 测试 configAuthorizer
func TestConfigAuthorizer(t *testing.T) {
	t.Parallel()

	cfg := &SecurityConfig{}
	authorizer := &configAuthorizer{cfg: cfg}

	registry := authorizer.AntMatchers("/api/**")
	if registry == nil {
		t.Error("expected non-nil registry")
	}

	anyRegistry := authorizer.AnyRequest()
	if anyRegistry == nil {
		t.Error("expected non-nil registry")
	}
}

// TestConfigRegistry_AllMethods 测试 configRegistry 的所有授权方法
func TestConfigRegistry_AllMethods(t *testing.T) {
	t.Parallel()

	t.Run("HasRole", func(t *testing.T) {
		cfg := &SecurityConfig{}
		r := &configRegistry{cfg: cfg, patterns: []string{"/admin/**"}}
		r.HasRole("ADMIN")
		if len(cfg.AuthorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(cfg.AuthorizeRules))
		}
	})

	t.Run("HasAnyRole", func(t *testing.T) {
		cfg := &SecurityConfig{}
		r := &configRegistry{cfg: cfg, patterns: []string{"/api/**"}}
		r.HasAnyRole("ADMIN", "USER")
		if len(cfg.AuthorizeRules) != 1 {
			t.Errorf("expected 1 rule, got %d", len(cfg.AuthorizeRules))
		}
	})
}

// ============================================================
// rate_limit_filter.go 测试 - 补充覆盖
// ============================================================

// TestRateLimitFilter_DoFilter_InvalidTypes 测试 TokenBucket 的类型检查
func TestRateLimitFilter_DoFilter_InvalidTypes(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled: true,
		Rate:    10,
		Burst:   20,
		Log:     newTestLogger(),
	})

	t.Run("invalid context", func(t *testing.T) {
		err := f.DoFilter("invalid", &mockSecurityRequest{}, &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid context")
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		err := f.DoFilter(context.Background(), "invalid", &mockSecurityResponse{}, &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid request")
		}
	})

	t.Run("invalid response", func(t *testing.T) {
		err := f.DoFilter(context.Background(), &mockSecurityRequest{}, "invalid", &mockSecurityFilterChain{})
		if err == nil {
			t.Error("expected error for invalid response")
		}
	})
}

// TestRateLimitFilter_DoFilter_RateLimited 测试令牌桶限流
func TestRateLimitFilter_DoFilter_RateLimited(t *testing.T) {
	t.Parallel()

	f := NewRateLimitFilter(RateLimitConfig{
		Enabled: true,
		Rate:    1,
		Burst:   1,
		Log:     newTestLogger(),
	})

	req := &mockSecurityRequest{method: "GET", uri: "/api", remoteAddr: "10.0.0.1:8080"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	// First request should be allowed
	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called for first request")
	}

	// Second request should be rate limited
	chain.called = false
	err = f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain.called {
		t.Error("expected chain not to be called for rate limited request")
	}
}

// ============================================================
// 辅助 Mock 类型
// ============================================================

// filterFuncFilter 用于测试的自定义过滤器
type filterFuncFilter struct {
	doFilter func(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error
}

func (f *filterFuncFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	if f.doFilter != nil {
		return f.doFilter(ctx, request, response, chain)
	}
	return chain.DoFilter(ctx, request, response)
}

func (f *filterFuncFilter) Order() int { return 0 }

// chainFuncFilter 用于测试的 SecurityFilterChain 实现
type chainFuncFilter struct {
	doFilter func(ctx interface{}, request interface{}, response interface{}) error
}

func (c *chainFuncFilter) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	if c.doFilter != nil {
		return c.doFilter(ctx, request, response)
	}
	return nil
}

func (c *chainFuncFilter) Matches(request interface{}) bool { return true }
func (c *chainFuncFilter) GetFilters() []filter.Filter      { return nil }

// mockLogoutHandler 用于测试的登出处理器 Mock
type mockLogoutHandler struct {
	called bool
}

func (m *mockLogoutHandler) Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication) {
	m.called = true
}

// mockRateLimitStrategy 用于测试的限流策略 Mock
type mockRateLimitStrategy struct {
	allow bool
}

func (m *mockRateLimitStrategy) Allow(key string) bool {
	return m.allow
}

// mockResponseWriter 用于测试的 http.ResponseWriter Mock
type mockResponseWriter struct {
	header http.Header
	code   int
	body   []byte
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{header: make(http.Header)}
}

func (m *mockResponseWriter) Header() http.Header {
	return m.header
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(code int) {
	m.code = code
}

// newTestLogger 创建静默测试日志器（输出丢弃，避免污染测试输出）
func newTestLogger() log.Logger {
	return log.NewSlogLogger(log.WithLevel(log.ErrorLevel), log.WithOutput(io.Discard))
}
