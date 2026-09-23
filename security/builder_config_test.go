package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/security/authorization"
	"github.com/xudefa/enhance/security/filter"
)

func TestSecurityBuilder_BasicConfig(t *testing.T) {
	t.Parallel()
	authManager := &testAuthManager{}
	userDetailsService := &testUserDetailsService{}
	passwordEncoder := NewNoOpPasswordEncoder()

	config := NewSecurityBuilder().
		AuthenticationManager(authManager).
		UserDetailsService(userDetailsService).
		PasswordEncoder(passwordEncoder).
		EnableAnonymous().
		EnableHttpBasic().
		Build()

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_WithFilters(t *testing.T) {
	t.Parallel()
	filter1 := &testSecurityFilter{}
	filter2 := &testSecurityFilter{}

	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		AddFilter(filter1).
		AddFilterAfter(filter2, filter1).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_FormLogin(t *testing.T) {
	t.Parallel()
	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		EnableFormLogin("/login", "/home").
		EnableCsrf().
		EnableLogout("/logout").
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	chain, err := http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
	if chain == nil {
		t.Error("expected non-nil filter chain")
	}
}

func TestSecurityBuilder_LogoutWithHandler(t *testing.T) {
	t.Parallel()
	handler := &testLogoutSuccessHandler{}
	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		EnableLogout("/logout", handler).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	chain, err := http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
	if chain == nil {
		t.Error("expected non-nil filter chain")
	}
}

func TestSecurityBuilder_AddFilterBefore(t *testing.T) {
	t.Parallel()
	filter1 := &testSecurityFilter{}
	filter2 := &testSecurityFilter{}

	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		AddFilter(filter1).
		AddFilterBefore(filter2, filter1).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_AccessDecisionManager(t *testing.T) {
	t.Parallel()
	accessDecisionMgr := &testAccessDecisionManager{}
	config := NewSecurityBuilder().
		AuthenticationManager(&testAuthManager{}).
		AccessDecisionManager(accessDecisionMgr).
		Build()

	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err != nil {
		t.Errorf("unexpected build error: %v", err)
	}
}

func TestSecurityBuilder_EmptyBuild(t *testing.T) {
	t.Parallel()
	config := NewSecurityBuilder().Build()
	http := NewHttpSecurity()
	err := config.Configure(http)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = http.Build()
	if err == nil {
		t.Error("expected build to fail without auth manager")
	}
}

// 测试模拟实现

type testAuth struct {
	principal   string
	credentials string
}

func (a *testAuth) Principal() any        { return a.principal }
func (a *testAuth) Credentials() any      { return a.credentials }
func (a *testAuth) Authorities() []string { return []string{"ROLE_USER"} }
func (a *testAuth) Authenticated() bool   { return true }

type testAuthManager struct{}

func (m *testAuthManager) Authenticate(ctx context.Context, auth AuthenticationToken) (Authentication, error) {
	return &testAuth{principal: "testuser", credentials: "testpass"}, nil
}

type testUserDetailsService struct{}

func (s *testUserDetailsService) LoadUserByUsername(ctx context.Context, username string) (UserDetails, error) {
	return nil, nil
}

type testAccessDecisionManager struct{}

func (m *testAccessDecisionManager) Decide(ctx context.Context, auth authorization.Authentication, resource string, attrs []string) error {
	return nil
}

func (m *testAccessDecisionManager) Supports(attribute string) bool {
	return true
}

type testSecurityFilter struct{}

func (f *testSecurityFilter) DoFilter(ctx interface{}, req interface{}, resp interface{}, chain filter.FilterChain) error {
	return chain.DoFilter(ctx, req, resp)
}

func (f *testSecurityFilter) Order() int { return 0 }

type testLogoutSuccessHandler struct{}

func (h *testLogoutSuccessHandler) OnLogoutSuccess(ctx context.Context, req SecurityRequest, resp SecurityResponse, auth Authentication) {
	resp.SetStatusCode(302)
	resp.SetHeader("Location", "/login")
}
