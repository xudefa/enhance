package authentication

import (
	"context"
	"testing"
)

func TestNewInMemoryUserDetails(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("admin", "pass123", []string{"ROLE_ADMIN"})

	if user.Username() != "admin" {
		t.Errorf("expected username 'admin', got %s", user.Username())
	}
	if user.Password() != "pass123" {
		t.Errorf("expected password 'pass123', got %s", user.Password())
	}
	if len(user.Authorities()) != 1 || user.Authorities()[0] != "ROLE_ADMIN" {
		t.Errorf("expected authorities [ROLE_ADMIN], got %v", user.Authorities())
	}
}

func TestInMemoryUserDetails_Defaults(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("user", "pass", nil)

	if !user.Enabled() {
		t.Error("expected Enabled to be true by default")
	}
	if !user.AccountNonLocked() {
		t.Error("expected AccountNonLocked to be true by default")
	}
	if !user.AccountNonExpired() {
		t.Error("expected AccountNonExpired to be true by default")
	}
	if !user.CredentialsNonExpired() {
		t.Error("expected CredentialsNonExpired to be true by default")
	}
}

func TestInMemoryUserDetails_Setters(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("user", "pass", nil)

	user.SetEnabled(false)
	if user.Enabled() {
		t.Error("expected Enabled to be false")
	}

	user.SetAccountNonLocked(false)
	if user.AccountNonLocked() {
		t.Error("expected AccountNonLocked to be false")
	}
}

func TestNewInMemoryUserDetailsService(t *testing.T) {
	t.Parallel()

	svc := NewInMemoryUserDetailsService()
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.UserCount() != 0 {
		t.Errorf("expected 0 users, got %d", svc.UserCount())
	}
}

func TestInMemoryUserDetailsService_CreateAndLoad(t *testing.T) {
	t.Parallel()

	svc := NewInMemoryUserDetailsService()
	svc.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	if svc.UserCount() != 1 {
		t.Errorf("expected 1 user, got %d", svc.UserCount())
	}

	ctx := context.Background()
	user, err := svc.LoadUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username() != "admin" {
		t.Errorf("expected username 'admin', got %s", user.Username())
	}
}

func TestInMemoryUserDetailsService_UserNotFound(t *testing.T) {
	t.Parallel()

	svc := NewInMemoryUserDetailsService()

	ctx := context.Background()
	_, err := svc.LoadUserByUsername(ctx, "nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestInMemoryUserDetailsService_DeleteUser(t *testing.T) {
	t.Parallel()

	svc := NewInMemoryUserDetailsService()
	svc.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})
	svc.DeleteUser("admin")

	if svc.UserCount() != 0 {
		t.Errorf("expected 0 users after delete, got %d", svc.UserCount())
	}
}

func TestInMemoryUserDetailsService_MultipleUsers(t *testing.T) {
	t.Parallel()

	svc := NewInMemoryUserDetailsService()
	svc.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})
	svc.CreateUser("user", "user123", []string{"ROLE_USER"})

	if svc.UserCount() != 2 {
		t.Errorf("expected 2 users, got %d", svc.UserCount())
	}

	ctx := context.Background()
	u1, err := svc.LoadUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u1.Username() != "admin" {
		t.Errorf("expected 'admin', got %s", u1.Username())
	}

	u2, err := svc.LoadUserByUsername(ctx, "user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u2.Username() != "user" {
		t.Errorf("expected 'user', got %s", u2.Username())
	}
}

func TestNewNoOpPasswordEncoder(t *testing.T) {
	t.Parallel()

	encoder := NewNoOpPasswordEncoder()
	if encoder == nil {
		t.Fatal("expected non-nil encoder")
	}
}

func TestNoOpPasswordEncoder_Encode(t *testing.T) {
	t.Parallel()

	encoder := NewNoOpPasswordEncoder()
	encoded := encoder.Encode("password")
	if encoded != "password" {
		t.Errorf("expected encoded password to equal raw, got %s", encoded)
	}
}

func TestNoOpPasswordEncoder_Matches(t *testing.T) {
	t.Parallel()

	encoder := NewNoOpPasswordEncoder()

	if !encoder.Matches("password", "password") {
		t.Error("expected passwords to match")
	}
	if encoder.Matches("wrong", "password") {
		t.Error("expected passwords to not match")
	}
}

func TestNoOpPasswordEncoder_EmptyPassword(t *testing.T) {
	t.Parallel()

	encoder := NewNoOpPasswordEncoder()

	if !encoder.Matches("", "") {
		t.Error("expected empty passwords to match")
	}
	if encoder.Matches("", "something") {
		t.Error("expected empty password to not match")
	}
}
