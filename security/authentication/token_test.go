package authentication

import (
	"testing"
)

func TestUsernamePasswordToken_Unauthenticated(t *testing.T) {
	t.Parallel()

	token := NewUsernamePasswordToken("user", "pass")

	if token.Principal() != "user" {
		t.Errorf("expected principal 'user', got %v", token.Principal())
	}
	if token.Credentials() != "pass" {
		t.Errorf("expected credentials 'pass', got %v", token.Credentials())
	}
	if token.Authenticated() {
		t.Error("expected not authenticated")
	}
	if len(token.Authorities()) != 0 {
		t.Errorf("expected nil authorities, got %v", token.Authorities())
	}
	if token.Name() != "user" {
		t.Errorf("expected name 'user', got %s", token.Name())
	}
}

func TestUsernamePasswordToken_Authenticated(t *testing.T) {
	t.Parallel()

	token := NewAuthenticatedUsernamePasswordToken("admin", nil, []string{"ROLE_ADMIN"})

	if !token.Authenticated() {
		t.Error("expected authenticated")
	}
	if token.Principal() != "admin" {
		t.Errorf("expected principal 'admin', got %v", token.Principal())
	}
	if token.Credentials() != nil {
		t.Errorf("expected nil credentials, got %v", token.Credentials())
	}
	if len(token.Authorities()) != 1 || token.Authorities()[0] != "ROLE_ADMIN" {
		t.Errorf("expected authorities [ROLE_ADMIN], got %v", token.Authorities())
	}
}

func TestUsernamePasswordToken_Setters(t *testing.T) {
	t.Parallel()

	token := NewUsernamePasswordToken("user", "pass")

	token.SetAuthorities([]string{"ROLE_USER", "ROLE_ADMIN"})
	if len(token.Authorities()) != 2 {
		t.Errorf("expected 2 authorities, got %d", len(token.Authorities()))
	}

	token.SetAuthenticated(true)
	if !token.Authenticated() {
		t.Error("expected authenticated to be true")
	}

	token.SetAuthenticated(false)
	if token.Authenticated() {
		t.Error("expected authenticated to be false")
	}
}

func TestUsernamePasswordToken_Name_WithUserDetails(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("admin", "pass", []string{"ROLE_ADMIN"})
	token := NewUsernamePasswordToken(user, "pass")

	if token.Name() != "admin" {
		t.Errorf("expected name 'admin', got %s", token.Name())
	}
}

func TestUsernamePasswordToken_Name_WithString(t *testing.T) {
	t.Parallel()

	token := NewUsernamePasswordToken("user", "pass")
	if token.Name() != "user" {
		t.Errorf("expected name 'user', got %s", token.Name())
	}
}

func TestUsernamePasswordToken_Name_NilPrincipal(t *testing.T) {
	t.Parallel()

	token := &UsernamePasswordToken{principal: 12345}
	if token.Name() != "" {
		t.Errorf("expected empty name, got %s", token.Name())
	}
}

func TestAnonymousToken_Basic(t *testing.T) {
	t.Parallel()

	token := NewAnonymousToken()

	if token.Principal() != "anonymousUser" {
		t.Errorf("expected principal 'anonymousUser', got %v", token.Principal())
	}
	if token.Credentials() != nil {
		t.Errorf("expected nil credentials, got %v", token.Credentials())
	}
	if token.Authenticated() {
		t.Error("expected not authenticated")
	}
}

func TestUsernamePasswordToken_NilAuthorities(t *testing.T) {
	t.Parallel()

	token := NewUsernamePasswordToken("user", "pass")
	if token.Authorities() != nil {
		t.Errorf("expected nil authorities for new token, got %v", token.Authorities())
	}
}
