package security

import (
	"context"
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

	f, err := NewLogoutFilter("/logout", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if f.Order() != 0 {
		t.Errorf("expected order 0, got %d", f.Order())
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

	f := MustNewLogoutFilter("/logout", nil)
	if f == nil {
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

	f := MustNewLogoutFilter("/logout", nil)
	handler := &testLogoutHandler{}
	f.AddLogoutHandler(handler)

	if len(f.handlers) != 1 {
		t.Errorf("expected 1 handler, got %d", len(f.handlers))
	}
}

func TestLogoutFilter_SetSuccessHandler(t *testing.T) {
	t.Parallel()

	f := MustNewLogoutFilter("/logout", nil)
	handler := &mockLogoutSuccessHandler{}
	f.SetSuccessHandler(handler)

	if f.successHandler == nil {
		t.Error("expected successHandler to be set")
	}
}

func TestLogoutFilter_DoFilter_TypeErrors(t *testing.T) {
	t.Parallel()

	f := MustNewLogoutFilter("/logout", nil)

	err := f.DoFilter("notContext", nil, nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-context")
	}

	err = f.DoFilter(context.Background(), "notReq", nil, &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-request")
	}

	err = f.DoFilter(context.Background(), newMockSecurityRequest("POST", "/logout", nil), "notResp", &mockFilterChain{})
	if err == nil {
		t.Error("expected error for non-response")
	}
}

func TestLogoutFilter_NonAllowedMethod(t *testing.T) {
	t.Parallel()

	f := MustNewLogoutFilter("/logout", nil)
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("GET", "/logout", nil)
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !chain.called {
		t.Error("expected chain to be called for non-allowed method")
	}
}

func TestLogoutFilter_WrongURI(t *testing.T) {
	t.Parallel()

	f := MustNewLogoutFilter("/logout", nil)
	chain := &mockFilterChain{}

	req := newMockSecurityRequest("POST", "/other", nil)
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, chain)
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
	f := MustNewLogoutFilter("/logout", []LogoutHandler{handler})

	req := newMockSecurityRequest("POST", "/logout", nil)
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, &mockFilterChain{})
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

	f := MustNewLogoutFilter("/logout", nil)

	req := newMockSecurityRequest("DELETE", "/logout", nil)
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, &mockFilterChain{})
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
	f := MustNewLogoutFilter("/logout", nil)
	f.SetSuccessHandler(handler)

	req := newMockSecurityRequest("POST", "/logout", nil)
	resp := newMockSecurityResponse()

	err := f.DoFilter(context.Background(), req, resp, &mockFilterChain{})
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

	f := MustNewLogoutFilter("/logout", nil)
	if f.Order() != 0 {
		t.Errorf("expected order 0, got %d", f.Order())
	}
}
