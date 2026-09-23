package authentication

import (
	"context"
	"strings"
	"testing"

	"github.com/xudefa/enhance/log"
)

func TestDaoAuthenticationProviderSuccess(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	token := NewUsernamePasswordToken("admin", "admin123")
	authenticated, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !authenticated.Authenticated() {
		t.Error("expected authentication to be successful")
	}
	if authenticated.Principal() == nil {
		t.Error("expected non-nil principal")
	}
	if user, ok := authenticated.Principal().(UserDetails); !ok || user.Username() != "admin" {
		t.Errorf("expected principal with username 'admin', got %v", authenticated.Principal())
	}
	if len(authenticated.Authorities()) != 1 || authenticated.Authorities()[0] != "ROLE_ADMIN" {
		t.Errorf("expected authorities [ROLE_ADMIN], got %v", authenticated.Authorities())
	}
}

func TestDaoAuthenticationProviderWrongPassword(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	token := NewUsernamePasswordToken("admin", "wrongpassword")
	_, err := provider.Authenticate(context.Background(), token)
	if err != ErrBadCredentials {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestDaoAuthenticationProviderUserNotFound(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	token := NewUsernamePasswordToken("nonexistent", "password")
	_, err := provider.Authenticate(context.Background(), token)
	if err != ErrBadCredentials {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestDaoAuthenticationProviderDisabledUser(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	user, _ := userDetailsService.LoadUserByUsername(context.Background(), "admin")
	user.(*InMemoryUserDetails).SetEnabled(false)

	token := NewUsernamePasswordToken("admin", "admin123")
	_, err := provider.Authenticate(context.Background(), token)
	if err == nil || !strings.Contains(err.Error(), "user is disabled") {
		t.Errorf("expected 'user is disabled' error, got %v", err)
	}
}

func TestDaoAuthenticationProviderLockedUser(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	user, _ := userDetailsService.LoadUserByUsername(context.Background(), "admin")
	user.(*InMemoryUserDetails).SetAccountNonLocked(false)

	token := NewUsernamePasswordToken("admin", "admin123")
	_, err := provider.Authenticate(context.Background(), token)
	if err == nil || !strings.Contains(err.Error(), "user account is locked") {
		t.Errorf("expected 'user account is locked' error, got %v", err)
	}
}

func TestDaoAuthenticationProviderSupports(t *testing.T) {
	t.Parallel()

	provider := NewDaoAuthenticationProvider(nil, nil, log.Build())

	usernameToken := NewUsernamePasswordToken("user", "pass")
	if !provider.Supports(usernameToken) {
		t.Error("expected provider to support UsernamePasswordToken")
	}

	anonToken := NewAnonymousToken()
	if provider.Supports(anonToken) {
		t.Error("expected provider to not support anonymousToken")
	}
}

func TestDaoAuthenticationProviderWithUserDetails(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	user := NewInMemoryUserDetails("admin", "admin123", []string{"ROLE_ADMIN"})
	token := NewUsernamePasswordToken(user, "admin123")
	authenticated, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !authenticated.Authenticated() {
		t.Error("expected authentication to be successful")
	}
	if principal, ok := authenticated.Principal().(UserDetails); !ok || principal.Username() != "admin" {
		t.Errorf("expected principal with username 'admin', got %v", authenticated.Principal())
	}
}

func TestDaoAuthenticationProviderUnsupportedPrincipal(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	token := NewUsernamePasswordToken(12345, "pass")
	_, err := provider.Authenticate(context.Background(), token)
	if err == nil {
		t.Error("expected error for unsupported principal type")
	}
}
