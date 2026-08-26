package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security/filter"
)

type mockAuthManager struct {
	authenticateResult Authentication
	authenticateErr    error
}

func (m *mockAuthManager) Authenticate(_ context.Context, _ AuthenticationToken) (Authentication, error) {
	return m.authenticateResult, m.authenticateErr
}

type mockFilterChain struct {
	called bool
}

func (c *mockFilterChain) DoFilter(_ interface{}, _ interface{}, _ interface{}) error {
	c.called = true
	return nil
}
func (c *mockFilterChain) AddFilter(_ filter.Filter) {}
func (c *mockFilterChain) GetFilters() []filter.Filter { return nil }

func TestNewUsernamePasswordAuthenticationFilterWithDefaults(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/", "/login?error", mgr, logger)

	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if f.Order() != 0 {
		t.Errorf("expected order 0, got %d", f.Order())
	}
}

func TestUsernamePasswordAuthenticationFilter_DoFilter_TypeErrors(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/", "/login?error", mgr, logger)

	err := f.DoFilter("notContext", nil, nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-context")
	}

	err = f.DoFilter(context.Background(), "notReq", nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-request")
	}

	err = f.DoFilter(context.Background(), &mockSecurityRequest{method: "GET", uri: "/"}, "notResp", &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-response")
	}
}

func TestUsernamePasswordAuthenticationFilter_NonPostPassesThrough(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/", "/login?error", mgr, logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "GET", uri: "/login"}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called for non-POST")
	}
}

func TestUsernamePasswordAuthenticationFilter_WrongURI_PassesThrough(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/", "/login?error", mgr, logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "POST", uri: "/other"}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called for wrong URI")
	}
}

func TestUsernamePasswordAuthenticationFilter_MissingCredentials(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/", "/login?error", mgr, logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "POST", uri: "/login", headers: map[string]string{"username": "user"}}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
	if chain.called {
		t.Error("chain should not be called when credentials are missing")
	}
}

func TestUsernamePasswordAuthenticationFilter_AuthSuccess(t *testing.T) {
	t.Parallel()

	authResult := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})
	mgr := &mockAuthManager{authenticateResult: authResult}
	logger := log.Build()
	f := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/dashboard", "/login?error", mgr, logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "POST", uri: "/login", headers: map[string]string{
		"username": "user",
		"password": "pass",
	}}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 302 {
		t.Errorf("expected status 302, got %d", resp.statusCode)
	}
	if resp.headers["Location"] != "/dashboard" {
		t.Errorf("expected redirect to /dashboard, got %s", resp.headers["Location"])
	}
}

func TestUsernamePasswordAuthenticationFilter_AuthFailure(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{authenticateErr: ErrBadCredentials}
	logger := log.Build()
	f := NewUsernamePasswordAuthenticationFilterWithDefaults("/login", "/dashboard", "/login?error", mgr, logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "POST", uri: "/login", headers: map[string]string{
		"username": "user",
		"password": "wrong",
	}}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
	if resp.headers["Location"] != "/login?error" {
		t.Errorf("expected redirect to /login?error, got %s", resp.headers["Location"])
	}
}

func TestBasicAuthenticationEntryPointWithRealm_Commence(t *testing.T) {
	t.Parallel()

	logger := log.Build()
	ep := NewBasicAuthenticationEntryPointWithRealm("TestRealm", logger)
	resp := &mockSecurityResponse{}

	err := ep.Commence(context.Background(), &mockSecurityRequest{method: "GET", uri: "/"}, resp, nil)
	if err != ErrBadCredentials {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
	expectedHeader := `Basic realm="TestRealm"`
	if resp.headers["WWW-Authenticate"] != expectedHeader {
		t.Errorf("expected WWW-Authenticate %q, got %q", expectedHeader, resp.headers["WWW-Authenticate"])
	}
}

func TestBasicAuthenticationEntryPointWithRealm_CommenceWithErr(t *testing.T) {
	t.Parallel()

	logger := log.Build()
	ep := NewBasicAuthenticationEntryPointWithRealm("TestRealm", logger)
	resp := &mockSecurityResponse{}

	originalErr := ErrAccessDenied
	returnedErr := ep.Commence(context.Background(), &mockSecurityRequest{method: "GET", uri: "/"}, resp, originalErr)
	if returnedErr != originalErr {
		t.Errorf("expected original error to be returned, got %v", returnedErr)
	}
}

func TestBasicAuthenticationFilterWithRealm_DoFilter_TypeErrors(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewBasicAuthenticationFilterWithRealm(mgr, "TestRealm", logger)

	err := f.DoFilter("notContext", nil, nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-context")
	}

	err = f.DoFilter(context.Background(), "notReq", nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-request")
	}

	err = f.DoFilter(context.Background(), &mockSecurityRequest{method: "GET", uri: "/"}, "notResp", &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-response")
	}
}

func TestBasicAuthenticationFilterWithRealm_NoAuthHeader(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewBasicAuthenticationFilterWithRealm(mgr, "TestRealm", logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "GET", uri: "/"}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err == nil {
		t.Error("expected error for missing Authorization header")
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
}

func TestBasicAuthenticationFilterWithRealm_InvalidPrefix(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewBasicAuthenticationFilterWithRealm(mgr, "TestRealm", logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "GET", uri: "/", headers: map[string]string{"Authorization": "Bearer token"}}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err == nil {
		t.Error("expected error for invalid prefix")
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
}

func TestBasicAuthenticationFilterWithRealm_InvalidBase64(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewBasicAuthenticationFilterWithRealm(mgr, "TestRealm", logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "GET", uri: "/", headers: map[string]string{"Authorization": "Basic !!!invalid!!!"}}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err == nil {
		t.Error("expected error for invalid base64")
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
}

func TestBasicAuthenticationFilterWithRealm_MissingSeparator(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewBasicAuthenticationFilterWithRealm(mgr, "TestRealm", logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "GET", uri: "/", headers: map[string]string{"Authorization": "Basic bm9jb2xvbg=="}}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err == nil {
		t.Error("expected error for missing separator")
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
}

func TestBasicAuthenticationFilterWithRealm_Success(t *testing.T) {
	t.Parallel()

	authResult := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})
	mgr := &mockAuthManager{authenticateResult: authResult}
	logger := log.Build()
	f := NewBasicAuthenticationFilterWithRealm(mgr, "TestRealm", logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "GET", uri: "/", headers: map[string]string{
		"Authorization": "Basic dXNlcjpwYXNz",
	}}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called after successful auth")
	}
}

func TestBasicAuthenticationFilterWithRealm_AuthFailure(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{authenticateErr: ErrBadCredentials}
	logger := log.Build()
	f := NewBasicAuthenticationFilterWithRealm(mgr, "TestRealm", logger)
	chain := &mockFilterChain{}

	req := &mockSecurityRequest{method: "GET", uri: "/", headers: map[string]string{
		"Authorization": "Basic dXNlcjpwYXNz",
	}}
	resp := &mockSecurityResponse{}

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err == nil {
		t.Error("expected error for auth failure")
	}
	if resp.statusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.statusCode)
	}
	if chain.called {
		t.Error("chain should not be called after auth failure")
	}
}

func TestBasicAuthenticationFilterWithRealm_Order(t *testing.T) {
	t.Parallel()

	mgr := &mockAuthManager{}
	logger := log.Build()
	f := NewBasicAuthenticationFilterWithRealm(mgr, "TestRealm", logger)

	if f.Order() != 0 {
		t.Errorf("expected order 0, got %d", f.Order())
	}
}
