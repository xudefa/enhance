package security

import (
	"context"
	"testing"
)

func TestNewRole(t *testing.T) {
	t.Parallel()

	role := NewRole("ROLE_ADMIN")
	if role == nil {
		t.Fatal("expected non-nil role")
	}
	if role.Authority() != "ROLE_ADMIN" {
		t.Errorf("expected authority 'ROLE_ADMIN', got '%s'", role.Authority())
	}
}

func TestRole_Authority(t *testing.T) {
	t.Parallel()

	role := NewRole("ROLE_USER")
	if role.Authority() != "ROLE_USER" {
		t.Errorf("expected authority 'ROLE_USER', got '%s'", role.Authority())
	}
}

func TestNewAuthority(t *testing.T) {
	t.Parallel()

	auth := NewAuthority("ROLE_EDITOR")
	if auth == nil {
		t.Fatal("expected non-nil authority")
	}
	if auth.Authority() != "ROLE_EDITOR" {
		t.Errorf("expected authority 'ROLE_EDITOR', got '%s'", auth.Authority())
	}
}

func TestInMemoryUserDetails_AllFields(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails(
		"testuser",
		"password123",
		[]string{"ROLE_USER", "ROLE_ADMIN"},
	)

	if user.Username() != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username())
	}
	if user.Password() != "password123" {
		t.Errorf("expected password 'password123', got '%s'", user.Password())
	}

	authorities := user.Authorities()
	if len(authorities) != 2 {
		t.Errorf("expected 2 authorities, got %d", len(authorities))
	}

	if !user.Enabled() {
		t.Error("expected user to be enabled")
	}
	if !user.AccountNonExpired() {
		t.Error("expected account to be non-expired")
	}
	if !user.AccountNonLocked() {
		t.Error("expected account to be non-locked")
	}
	if !user.CredentialsNonExpired() {
		t.Error("expected credentials to be non-expired")
	}
}

func TestInMemoryUserDetailsService_CreateAndLoad(t *testing.T) {
	t.Parallel()

	service := NewInMemoryUserDetailsService()

	service.CreateUser("user1", "pass1", []string{"ROLE_USER"})

	loaded, err := service.LoadUserByUsername(context.Background(), "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.Username() != "user1" {
		t.Errorf("expected username 'user1', got '%s'", loaded.Username())
	}
}

func TestInMemoryUserDetailsService_DeleteUser(t *testing.T) {
	t.Parallel()

	service := NewInMemoryUserDetailsService()

	service.CreateUser("user1", "pass1", []string{"ROLE_USER"})

	if service.UserCount() != 1 {
		t.Errorf("expected 1 user, got %d", service.UserCount())
	}

	service.DeleteUser("user1")
	if service.UserCount() != 0 {
		t.Errorf("expected 0 users after delete, got %d", service.UserCount())
	}
}

func TestInMemoryUserDetailsService_LoadNonExistentUser(t *testing.T) {
	t.Parallel()

	service := NewInMemoryUserDetailsService()

	_, err := service.LoadUserByUsername(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}
