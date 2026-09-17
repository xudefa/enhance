package security

import (
	"context"
	"strings"
	"testing"
)

func testLogoutFilterLogoutSuccess(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})
	logoutCtx := ContextWithAuthentication(ctx, auth)

	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	logoutFilter := MustNewLogoutFilter("/logout", []LogoutHandler{NewSecurityContextLogoutHandler()})

	err := logoutFilter.DoFilter(logoutCtx, req, resp, chain)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if chain.called {
		t.Error("Filter chain should not be called after logout")
	}
}

func testLogoutFilterNonLogoutURLSkips(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	req := &mockSecurityRequest{method: "POST", uri: "/api/test"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	logoutFilter := MustNewLogoutFilter("/logout", []LogoutHandler{})

	err := logoutFilter.DoFilter(ctx, req, resp, chain)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !chain.called {
		t.Error("Filter chain should be called for non-logout URL")
	}
}

func testLogoutFilterCustomSuccessHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})
	logoutCtx := ContextWithAuthentication(ctx, auth)

	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}
	chain := &mockSecurityFilterChain{}

	customHandler := &mockLogoutSuccessHandler{targetCalled: "/custom"}
	logoutFilter := MustNewLogoutFilter("/logout", []LogoutHandler{})
	logoutFilter.SetSuccessHandler(customHandler)

	err := logoutFilter.DoFilter(logoutCtx, req, resp, chain)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.headers["Location"] != "/custom" {
		t.Errorf("Expected Location '/custom', got '%s'", resp.headers["Location"])
	}
}

func TestLogoutFilter(t *testing.T) {
	t.Parallel()
	t.Run("Logout success", testLogoutFilterLogoutSuccess)
	t.Run("Non-logout URL should skip", testLogoutFilterNonLogoutURLSkips)
	t.Run("Custom success handler", testLogoutFilterCustomSuccessHandler)
}

func TestLogoutSuccessHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("DefaultLogoutSuccessHandler", func(t *testing.T) {
		handler := NewDefaultLogoutSuccessHandler("/login?logout")
		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

		handler.OnLogoutSuccess(ctx, req, resp, auth)

		if resp.statusCode != 302 {
			t.Errorf("Expected status code 302, got %d", resp.statusCode)
		}
		if resp.headers["Location"] != "/login?logout" {
			t.Errorf("Expected Location '/login?logout', got '%s'", resp.headers["Location"])
		}
	})

	t.Run("SimpleLogoutSuccessHandler", func(t *testing.T) {
		handler := NewSimpleLogoutSuccessHandler("/home")
		req := &mockSecurityRequest{method: "POST", uri: "/logout"}
		resp := &mockSecurityResponse{}
		auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

		handler.OnLogoutSuccess(ctx, req, resp, auth)

		if resp.statusCode != 302 {
			t.Errorf("Expected status code 302, got %d", resp.statusCode)
		}
		if resp.headers["Location"] != "/home" {
			t.Errorf("Expected Location '/home', got '%s'", resp.headers["Location"])
		}
	})
}

func TestCookieClearingLogoutHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}
	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

	handler := NewCookieClearingLogoutHandler("session_id")

	handler.Logout(ctx, req, resp, auth)

	cookie := resp.headers["Set-Cookie"]
	if cookie == "" {
		t.Error("Expected Set-Cookie header")
	}
	if !strings.Contains(cookie, "session_id") {
		t.Errorf("Expected cookie 'session_id' to be cleared, got %s", cookie)
	}
	if !strings.Contains(cookie, "Max-Age=0") {
		t.Errorf("Expected cookie to have Max-Age=0, got %s", cookie)
	}
}

func TestSecurityContextLogoutHandler(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	handler := NewSecurityContextLogoutHandler()

	auth := NewAuthenticatedUsernamePasswordAuthenticationToken("admin", []string{"ROLE_ADMIN"})

	req := &mockSecurityRequest{method: "POST", uri: "/logout"}
	resp := &mockSecurityResponse{}

	handler.Logout(ctx, req, resp, auth)

	// SecurityContextLogoutHandler is a no-op; authentication clearing
	// is handled by LogoutFilter which reads from context.
}
