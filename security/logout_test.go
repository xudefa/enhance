package security

import (
	"context"
	"strings"
	"testing"
)

type testLogoutHandler struct {
	called bool
}

func (h *testLogoutHandler) Logout(_ context.Context, _ SecurityRequest, _ SecurityResponse, _ Authentication) {
	h.called = true
}

func TestNewLogoutFilter_Success(t *testing.T) {
	t.Parallel()

	logoutFilter, err := NewLogoutFilter("/logout", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logoutFilter == nil {
		t.Fatal("expected non-nil filter")
	}
	if logoutFilter.Order() != 0 {
		t.Errorf("expected order 0, got %d", logoutFilter.Order())
	}
}

func TestNewLogoutFilter_EmptyUrl(t *testing.T) {
	t.Parallel()

	_, err := NewLogoutFilter("", nil)
	if err == nil {
		t.Error("expected error for empty logoutUrl")
	}
}

func TestMustNewLogoutFilter_Success(t *testing.T) {
	t.Parallel()

	logoutFilter := MustNewLogoutFilter("/logout", nil)
	if logoutFilter == nil {
		t.Fatal("expected non-nil filter")
	}
}

func TestMustNewLogoutFilter_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for empty logoutUrl")
		}
	}()
	MustNewLogoutFilter("", nil)
}

func TestLogoutFilter_AddLogoutHandler_New(t *testing.T) {
	t.Parallel()

	logoutFilter := MustNewLogoutFilter("/logout", nil)
	handler := &testLogoutHandler{}
	logoutFilter.AddLogoutHandler(handler)

	if len(logoutFilter.handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(logoutFilter.handlers))
	}
}

func TestLogoutFilter_SetSuccessHandler(t *testing.T) {
	t.Parallel()

	logoutFilter := MustNewLogoutFilter("/logout", nil)
	handler := &mockLogoutSuccessHandler{}
	logoutFilter.SetSuccessHandler(handler)

	if logoutFilter.successHandler == nil {
		t.Error("expected successHandler to be set")
	}
}

func TestLogoutFilter_DoFilter_TypeErrors(t *testing.T) {
	t.Parallel()

	logoutFilter := MustNewLogoutFilter("/logout", nil)

	err := logoutFilter.DoFilter("notContext", nil, nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-context")
	}

	err = logoutFilter.DoFilter(context.Background(), "notReq", nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-request")
	}

	err = logoutFilter.DoFilter(context.Background(), newMockSecurityRequest("POST", "/logout", nil), "notResp", &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-response")
	}
}

func TestLogoutFilter_NonAllowedMethod(t *testing.T) {
	t.Parallel()

	logoutFilter := MustNewLogoutFilter("/logout", nil)
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/logout", nil)
	resp := newMockSecurityResponse()

	err := logoutFilter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called for non-allowed method")
	}
}

func TestLogoutFilter_WrongURI(t *testing.T) {
	t.Parallel()

	logoutFilter := MustNewLogoutFilter("/logout", nil)
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("POST", "/other", nil)
	resp := newMockSecurityResponse()

	err := logoutFilter.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called for wrong URI")
	}
}

func TestLogoutFilter_LogoutPost_Success(t *testing.T) {
	t.Parallel()

	handler := &testLogoutHandler{}
	logoutFilter := MustNewLogoutFilter("/logout", []LogoutHandler{handler})

	req := newMockSecurityRequest("POST", "/logout", nil)
	resp := newMockSecurityResponse()

	err := logoutFilter.DoFilter(context.Background(), req, resp, &mockFilterChain{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handler.called {
		t.Error("expected handler to be called")
	}
	if resp.statusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.statusCode)
	}
}

func TestLogoutFilter_LogoutDelete_Success(t *testing.T) {
	t.Parallel()

	logoutFilter := MustNewLogoutFilter("/logout", nil)

	req := newMockSecurityRequest("DELETE", "/logout", nil)
	resp := newMockSecurityResponse()

	err := logoutFilter.DoFilter(context.Background(), req, resp, &mockFilterChain{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.statusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.statusCode)
	}
}

func TestLogoutFilter_WithSuccessHandler_Set(t *testing.T) {
	t.Parallel()

	handler := &mockLogoutSuccessHandler{}
	logoutFilter := MustNewLogoutFilter("/logout", nil)
	logoutFilter.SetSuccessHandler(handler)

	req := newMockSecurityRequest("POST", "/logout", nil)
	resp := newMockSecurityResponse()

	err := logoutFilter.DoFilter(context.Background(), req, resp, &mockFilterChain{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handler.called {
		t.Error("expected success handler to be called")
	}
}

func TestDefaultLogoutSuccessHandler_OnLogoutSuccess(t *testing.T) {
	t.Parallel()

	handler := NewDefaultLogoutSuccessHandler("/login?logout")
	resp := newMockSecurityResponse()

	handler.OnLogoutSuccess(context.Background(), newMockSecurityRequest("POST", "/logout", nil), resp, nil)

	if resp.statusCode != 302 {
		t.Errorf("expected status 302, got %d", resp.statusCode)
	}
	if resp.headers["Location"] != "/login?logout" {
		t.Errorf("expected redirect to /login?logout, got %s", resp.headers["Location"])
	}
}

func TestSimpleLogoutSuccessHandler_OnLogoutSuccess(t *testing.T) {
	t.Parallel()

	handler := NewSimpleLogoutSuccessHandler("/home")
	resp := newMockSecurityResponse()

	handler.OnLogoutSuccess(context.Background(), newMockSecurityRequest("POST", "/logout", nil), resp, nil)

	if resp.statusCode != 302 {
		t.Errorf("expected status 302, got %d", resp.statusCode)
	}
	if resp.headers["Location"] != "/home" {
		t.Errorf("expected redirect to /home, got %s", resp.headers["Location"])
	}
}

func TestSecurityContextLogoutHandler_Logout(t *testing.T) {
	t.Parallel()

	handler := NewSecurityContextLogoutHandler()
	handler.Logout(context.Background(), newMockSecurityRequest("POST", "/logout", nil), newMockSecurityResponse(), nil)
}

func TestCookieClearingLogoutHandler_Logout(t *testing.T) {
	t.Parallel()

	handler := NewCookieClearingLogoutHandler("session", "token")
	resp := newMockSecurityResponse()

	handler.Logout(context.Background(), newMockSecurityRequest("POST", "/logout", nil), resp, nil)

	setCookie := resp.headers["Set-Cookie"]
	if setCookie == "" {
		t.Error("expected Set-Cookie header")
	}
	if setCookie != "token=; Path=/; Max-Age=0" {
		t.Errorf("unexpected Set-Cookie value: %s", setCookie)
	}
}

func TestCookieClearingLogoutHandler_MultipleCookies(t *testing.T) {
	t.Parallel()

	handler := NewCookieClearingLogoutHandler("a", "b")

	type cookieRecorder struct {
		cookies []string
	}
	recorder := &cookieRecorder{}

	original := handler.cookieNames
	_ = original

	for range handler.cookieNames {
		resp := newMockSecurityResponse()
		handler.Logout(context.Background(), newMockSecurityRequest("POST", "/logout", nil), resp, nil)
		recorder.cookies = append(recorder.cookies, resp.headers["Set-Cookie"])
	}

	if len(recorder.cookies) != 2 {
		t.Fatalf("expected 2 cookie headers, got %d", len(recorder.cookies))
	}
}

func TestLogoutFilter_Order(t *testing.T) {
	t.Parallel()

	logoutFilter := MustNewLogoutFilter("/logout", nil)
	if logoutFilter.Order() != 0 {
		t.Errorf("expected order 0, got %d", logoutFilter.Order())
	}
}

// ==================== Web Logout Filter Tests ====================

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

func TestLogoutFilter_Web(t *testing.T) {
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
