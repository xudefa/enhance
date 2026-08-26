package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/security/authentication"
	"github.com/xudefa/enhance/security/filter"
)

func TestAuthContextFilter_DoFilter(t *testing.T) {
	t.Parallel()

	f := NewAuthContextFilter()

	if f.Order() != AuthContextFilterOrder {
		t.Errorf("expected order %d, got %d", AuthContextFilterOrder, f.Order())
	}

	req := &mockSecurityRequest{method: "GET", uri: "/test"}
	resp := &mockSecurityResponse{}
	chain := filter.NewDefaultFilterChain()

	ctx := context.Background()
	err := f.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("DoFilter error: %v", err)
	}
}

func TestAnonymousAuthenticationFilter_DoFilter(t *testing.T) {
	t.Parallel()

	f := NewAnonymousAuthenticationFilter()

	if f.Order() != AnonymousAuthenticationFilterOrder {
		t.Errorf("expected order %d, got %d", AnonymousAuthenticationFilterOrder, f.Order())
	}

	req := &mockSecurityRequest{method: "GET", uri: "/test"}
	resp := &mockSecurityResponse{}
	chain := filter.NewDefaultFilterChain()

	ctx := context.Background()
	err := f.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("DoFilter error: %v", err)
	}
}

func TestNewAnonymousAuthenticationToken(t *testing.T) {
	t.Parallel()

	token := NewAnonymousAuthenticationToken("testKey", "anonymousUser", []string{"ROLE_ANONYMOUS"})

	if token.Principal() != "anonymousUser" {
		t.Errorf("expected principal 'anonymousUser', got %v", token.Principal())
	}
	if token.Credentials() != nil {
		t.Errorf("expected credentials to be nil, got %v", token.Credentials())
	}
	if len(token.Authorities()) != 1 || token.Authorities()[0] != "ROLE_ANONYMOUS" {
		t.Errorf("expected authorities [ROLE_ANONYMOUS], got %v", token.Authorities())
	}
	if token.Authenticated() {
		t.Error("expected authenticated to be false")
	}
	if token.Name() != "anonymousUser" {
		t.Errorf("expected name 'anonymousUser', got %s", token.Name())
	}
}

func TestAnonymousAuthenticationToken_NameWithNonStringPrincipal(t *testing.T) {
	t.Parallel()

	token := NewAnonymousAuthenticationToken("testKey", 123, []string{})

	if token.Name() != "" {
		t.Errorf("expected empty name for non-string principal, got %s", token.Name())
	}
}

func TestExceptionTranslationFilter_DoFilter(t *testing.T) {
	t.Parallel()

	accessDeniedHandler := &mockAccessDeniedHandler{}
	authEntryPoint := &mockAuthenticationEntryPoint{}
	f := NewExceptionTranslationFilter(accessDeniedHandler, authEntryPoint)

	req := &mockSecurityRequest{method: "GET", uri: "/test"}
	resp := &mockSecurityResponse{}
	chain := filter.NewDefaultFilterChain()

	ctx := context.Background()
	err := f.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("DoFilter error: %v", err)
	}
}

func TestExceptionTranslationFilter_DoFilter_InvalidContext(t *testing.T) {
	t.Parallel()

	f := NewExceptionTranslationFilter(&mockAccessDeniedHandler{}, &mockAuthenticationEntryPoint{})

	req := &mockSecurityRequest{}
	resp := &mockSecurityResponse{}
	chain := filter.NewDefaultFilterChain()

	err := f.DoFilter("invalid", req, resp, chain)
	if err == nil {
		t.Fatal("expected error for invalid context")
	}
}

func TestExceptionTranslationFilter_DoFilter_InvalidRequest(t *testing.T) {
	t.Parallel()

	f := NewExceptionTranslationFilter(&mockAccessDeniedHandler{}, &mockAuthenticationEntryPoint{})

	resp := &mockSecurityResponse{}
	chain := filter.NewDefaultFilterChain()

	err := f.DoFilter(context.Background(), "invalid", resp, chain)
	if err == nil {
		t.Fatal("expected error for invalid request")
	}
}

func TestExceptionTranslationFilter_DoFilter_InvalidResponse(t *testing.T) {
	t.Parallel()

	f := NewExceptionTranslationFilter(&mockAccessDeniedHandler{}, &mockAuthenticationEntryPoint{})

	req := &mockSecurityRequest{}
	chain := filter.NewDefaultFilterChain()

	err := f.DoFilter(context.Background(), req, "invalid", chain)
	if err == nil {
		t.Fatal("expected error for invalid response")
	}
}

func TestFilterSecurityInterceptor_DoFilter(t *testing.T) {
	t.Parallel()

	authManager := &mockAuthenticationManager{}
	decisionManager := NewAffirmativeBased()
	metadataSource := NewExpressionBasedFilterInvocationSecurityMetadataSource()

	f := NewFilterSecurityInterceptor(metadataSource, decisionManager, authManager)

	if f.Order() != FilterSecurityInterceptorOrder {
		t.Errorf("expected order %d, got %d", FilterSecurityInterceptorOrder, f.Order())
	}

	req := &mockSecurityRequest{method: "GET", uri: "/test"}
	resp := &mockSecurityResponse{}
	chain := filter.NewDefaultFilterChain()

	ctx := context.Background()
	err := f.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("DoFilter error: %v", err)
	}
}

func TestFilterSecurityInterceptor_DoFilter_InvalidContext(t *testing.T) {
	t.Parallel()

	f := NewFilterSecurityInterceptor(
		NewExpressionBasedFilterInvocationSecurityMetadataSource(),
		NewAffirmativeBased(),
		&mockAuthenticationManager{},
	)

	req := &mockSecurityRequest{}
	resp := &mockSecurityResponse{}
	chain := filter.NewDefaultFilterChain()

	err := f.DoFilter("invalid", req, resp, chain)
	if err == nil {
		t.Fatal("expected error for invalid context")
	}
}

func TestFilterSecurityInterceptor_DoFilter_InvalidRequest(t *testing.T) {
	t.Parallel()

	f := NewFilterSecurityInterceptor(
		NewExpressionBasedFilterInvocationSecurityMetadataSource(),
		NewAffirmativeBased(),
		&mockAuthenticationManager{},
	)

	resp := &mockSecurityResponse{}
	chain := filter.NewDefaultFilterChain()

	err := f.DoFilter(context.Background(), "invalid", resp, chain)
	if err == nil {
		t.Fatal("expected error for invalid request")
	}
}

func TestFilterSecurityInterceptor_DoFilter_InvalidResponse(t *testing.T) {
	t.Parallel()

	f := NewFilterSecurityInterceptor(
		NewExpressionBasedFilterInvocationSecurityMetadataSource(),
		NewAffirmativeBased(),
		&mockAuthenticationManager{},
	)

	req := &mockSecurityRequest{}
	chain := filter.NewDefaultFilterChain()

	err := f.DoFilter(context.Background(), req, "invalid", chain)
	if err == nil {
		t.Fatal("expected error for invalid response")
	}
}

type mockAccessDeniedHandler struct{}

func (m *mockAccessDeniedHandler) Handle(ctx context.Context, request SecurityRequest, response SecurityResponse, accessDenied error) error {
	return nil
}

type mockAuthenticationEntryPoint struct{}

func (m *mockAuthenticationEntryPoint) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, authError error) error {
	return nil
}

type mockAuthenticationManager struct{}

func (m *mockAuthenticationManager) Authenticate(ctx context.Context, token authentication.AuthenticationToken) (authentication.Authentication, error) {
	return nil, nil
}
