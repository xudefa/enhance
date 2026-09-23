package authentication

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/log"
)

func TestAnonymousAuthenticationProvider(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	token := NewUsernamePasswordToken("user", nil)
	authenticated, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !authenticated.Authenticated() {
		t.Error("expected anonymous auth to be authenticated")
	}
	if authenticated.Principal() != "anonymousUser" {
		t.Errorf("expected principal 'anonymousUser', got '%v'", authenticated.Principal())
	}
	if len(authenticated.Authorities()) != 1 || authenticated.Authorities()[0] != "ROLE_ANONYMOUS" {
		t.Errorf("expected authorities [ROLE_ANONYMOUS], got %v", authenticated.Authorities())
	}
}

func TestAnonymousAuthenticationProviderSkipsLoginAttempt(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	token := NewUsernamePasswordToken("user", "pass")
	authenticated, err := provider.Authenticate(context.Background(), token)
	if authenticated != nil {
		t.Error("expected nil result for login attempt with credentials")
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAnonymousAuthenticationProviderAlreadyAuthenticated(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	token := NewAuthenticatedUsernamePasswordToken("admin", nil, []string{"ROLE_ADMIN"})
	authenticated, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authenticated != nil {
		t.Error("expected nil result for already authenticated token")
	}
}

func TestAnonymousAuthenticationProviderNilToken(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	authenticated, err := provider.Authenticate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !authenticated.Authenticated() {
		t.Error("expected anonymous auth to be authenticated")
	}
	if authenticated.Principal() != "anonymousUser" {
		t.Errorf("expected principal 'anonymousUser', got '%v'", authenticated.Principal())
	}
}

func TestAnonymousAuthenticationProviderSupports(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	usernameToken := NewUsernamePasswordToken("user", "pass")
	if !provider.Supports(usernameToken) {
		t.Error("expected anonymous provider to support UsernamePasswordToken")
	}

	anonToken := NewAnonymousToken()
	if provider.Supports(anonToken) {
		t.Error("expected anonymous provider to not support anonymousToken")
	}
}

func TestProviderManagerSuccess(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())
	manager := NewProviderManager(provider)

	token := NewUsernamePasswordToken("admin", "admin123")
	authenticated, err := manager.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !authenticated.Authenticated() {
		t.Error("expected authentication to be successful")
	}
	if user, ok := authenticated.Principal().(UserDetails); !ok || user.Username() != "admin" {
		t.Errorf("expected principal with username 'admin', got %v", authenticated.Principal())
	}
}

func TestProviderManagerMultipleProviders(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	daoProvider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())
	anonProvider := NewAnonymousAuthenticationProvider()

	// DAO provider must come first; anonymous would match everything
	manager := NewProviderManager(daoProvider, anonProvider)

	token := NewUsernamePasswordToken("admin", "admin123")
	authenticated, err := manager.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !authenticated.Authenticated() {
		t.Error("expected authentication to be successful")
	}
	if user, ok := authenticated.Principal().(UserDetails); !ok || user.Username() != "admin" {
		t.Errorf("expected principal with username 'admin', got %v", authenticated.Principal())
	}
}

func TestProviderManagerNoProviderMatches(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	manager := NewProviderManager(provider)

	token := NewUsernamePasswordToken("admin", "admin123")
	_, err := manager.Authenticate(context.Background(), token)
	if err != ErrBadCredentials {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestProviderManagerAllFail(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	daoProvider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	// Only the DAO provider; wrong password should fail
	manager := NewProviderManager(daoProvider)

	token := NewUsernamePasswordToken("admin", "wrongpassword")
	_, err := manager.Authenticate(context.Background(), token)
	if err != ErrBadCredentials {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestProviderManagerAddProvider(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	daoProvider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	manager := NewProviderManager().(*ProviderManager)
	manager.AddProvider(daoProvider)

	token := NewUsernamePasswordToken("admin", "admin123")
	authenticated, err := manager.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !authenticated.Authenticated() {
		t.Error("expected authentication to be successful")
	}
}

func BenchmarkProviderManagerAuthenticate(b *testing.B) {
	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())
	manager := NewProviderManager(provider)

	token := NewUsernamePasswordToken("admin", "admin123")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.Authenticate(ctx, token)
	}
}
