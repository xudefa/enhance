package security

import (
	"context"
	"testing"
)

func TestCookieCsrfTokenRepository_ClearToken(t *testing.T) {
	t.Parallel()

	repo := NewCookieCsrfTokenRepository()
	req := &mockSecurityRequest{method: "GET", uri: "/test"}
	resp := &mockSecurityResponse{}

	repo.ClearToken(context.Background(), req, resp)

	cookie := resp.Header("Set-Cookie")
	if cookie == "" {
		t.Error("expected Set-Cookie header to be set")
	}
}

func TestGenerateSecureToken(t *testing.T) {
	t.Parallel()

	token, err := generateSecureToken(32)
	if err != nil {
		t.Fatalf("generateSecureToken error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestCsrfTokenManager_GenerateToken(t *testing.T) {
	t.Parallel()

	manager := NewCsrfTokenManager()
	token, err := manager.GenerateToken("user1")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestCsrfTokenManager_ValidateToken(t *testing.T) {
	t.Parallel()

	manager := NewCsrfTokenManager()
	token, err := manager.GenerateToken("user1")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	if !manager.ValidateToken("user1", token) {
		t.Error("expected token to be valid")
	}
}

func TestCsrfTokenManager_ValidateToken_Invalid(t *testing.T) {
	t.Parallel()

	manager := NewCsrfTokenManager()
	_, err := manager.GenerateToken("user1")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	if manager.ValidateToken("user1", "invalid-token") {
		t.Error("expected token to be invalid")
	}
}

func TestCsrfTokenManager_ValidateToken_NonExistent(t *testing.T) {
	t.Parallel()

	manager := NewCsrfTokenManager()

	if manager.ValidateToken("nonexistent", "some-token") {
		t.Error("expected token to be invalid for non-existent principal")
	}
}

func TestCsrfAuthenticationStrategy_OnAuthentication(t *testing.T) {
	t.Parallel()

	strategy := NewCsrfAuthenticationStrategy()
	req := &mockSecurityRequest{method: "POST", uri: "/login"}
	resp := &mockSecurityResponse{}
	auth := &mockAuthentication{authenticated: true, principal: "user1"}

	strategy.OnAuthentication(context.Background(), auth, req, resp)
}
