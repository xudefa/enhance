package security

import (
	"context"
	"net/http"
	"testing"
)

// TestFilterSecurityInterceptor_Setters 测试添加登出处理器

func TestFilterSecurityInterceptor_Setters(t *testing.T) {
	t.Parallel()

	interceptor := &FilterSecurityInterceptor{}

	source := NewExpressionBasedFilterInvocationSecurityMetadataSource()
	interceptor.SetSecurityMetadataSource(source)
	if interceptor.securityMetadataSource != source {
		t.Error("expected securityMetadataSource to be set")
	}

	manager := NewAffirmativeBased()
	interceptor.SetAccessDecisionManager(manager)
	if interceptor.accessDecisionManager != manager {
		t.Error("expected accessDecisionManager to be set")
	}

	authManager := &mockAuthenticationManager{}
	interceptor.SetAuthenticationManager(authManager)
	if interceptor.authenticationManager != authManager {
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
