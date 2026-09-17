package authentication

import (
	"context"
	"testing"
)

func TestUsernamePasswordToken(t *testing.T) {
	t.Parallel()

	token := NewUsernamePasswordToken("admin", "password123")

	if token.Principal() != "admin" {
		t.Errorf("expected principal 'admin', got '%v'", token.Principal())
	}
	if token.Credentials() != "password123" {
		t.Errorf("expected credentials 'password123', got '%v'", token.Credentials())
	}
	if token.Authenticated() {
		t.Error("expected token to not be authenticated")
	}
	if token.Name() != "admin" {
		t.Errorf("expected name 'admin', got '%s'", token.Name())
	}
}

func TestAuthenticatedUsernamePasswordToken(t *testing.T) {
	t.Parallel()

	token := NewAuthenticatedUsernamePasswordToken("admin", nil, []string{"ROLE_ADMIN"})

	if !token.Authenticated() {
		t.Error("expected token to be authenticated")
	}
	if token.Principal() != "admin" {
		t.Errorf("expected principal 'admin', got '%v'", token.Principal())
	}
	if token.Credentials() != nil {
		t.Errorf("expected nil credentials, got '%v'", token.Credentials())
	}
	if len(token.Authorities()) != 1 || token.Authorities()[0] != "ROLE_ADMIN" {
		t.Errorf("expected authorities [ROLE_ADMIN], got %v", token.Authorities())
	}
}

func TestUsernamePasswordTokenSetters(t *testing.T) {
	t.Parallel()

	token := NewUsernamePasswordToken("user", "pass")

	token.SetAuthorities([]string{"ROLE_USER"})
	if len(token.Authorities()) != 1 || token.Authorities()[0] != "ROLE_USER" {
		t.Errorf("expected authorities [ROLE_USER], got %v", token.Authorities())
	}

	token.SetAuthenticated(true)
	if !token.Authenticated() {
		t.Error("expected token to be authenticated")
	}
}

func TestAnonymousToken(t *testing.T) {
	t.Parallel()

	token := NewAnonymousToken()

	if token.Principal() != "anonymousUser" {
		t.Errorf("expected principal 'anonymousUser', got '%v'", token.Principal())
	}
	if token.Credentials() != nil {
		t.Errorf("expected nil credentials, got '%v'", token.Credentials())
	}
	if token.Authenticated() {
		t.Error("expected anonymous token to not be authenticated")
	}
}

func TestInMemoryUserDetailsService(t *testing.T) {
	t.Parallel()

	service := NewInMemoryUserDetailsService()

	service.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})
	service.CreateUser("user", "user123", []string{"ROLE_USER"})

	if service.UserCount() != 2 {
		t.Errorf("expected 2 users, got %d", service.UserCount())
	}

	ctx := context.Background()
	user, err := service.LoadUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("failed to load user: %v", err)
	}

	if user.Username() != "admin" {
		t.Errorf("expected username 'admin', got '%s'", user.Username())
	}
	if user.Password() != "admin123" {
		t.Errorf("expected password 'admin123', got '%s'", user.Password())
	}

	_, err = service.LoadUserByUsername(ctx, "nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}

	service.DeleteUser("admin")
	if service.UserCount() != 1 {
		t.Errorf("expected 1 user after delete, got %d", service.UserCount())
	}
}

func TestNoOpPasswordEncoder(t *testing.T) {
	t.Parallel()

	encoder := NewNoOpPasswordEncoder()

	if encoder.Encode("password") != "password" {
		t.Error("expected encoded password to be same as raw")
	}
	if !encoder.Matches("password", "password") {
		t.Error("expected password to match")
	}
	if encoder.Matches("wrong", "password") {
		t.Error("expected wrong password to not match")
	}
}

func TestUsernamePasswordTokenWithUserDetails(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("admin", "pass", []string{"ROLE_ADMIN"})
	token := NewUsernamePasswordToken(user, "pass")

	if token.Name() != "admin" {
		t.Errorf("expected name 'admin', got '%s'", token.Name())
	}
}

func TestInMemoryUserDetailsAccountStatus(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("admin", "pass", []string{"ROLE_ADMIN"})

	if !user.AccountNonExpired() {
		t.Error("expected AccountNonExpired to be true by default")
	}
	if !user.CredentialsNonExpired() {
		t.Error("expected CredentialsNonExpired to be true by default")
	}
	if !user.Enabled() {
		t.Error("expected Enabled to be true by default")
	}
	if !user.AccountNonLocked() {
		t.Error("expected AccountNonLocked to be true by default")
	}
}
