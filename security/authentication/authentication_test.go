package authentication

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/log"
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

func TestDaoAuthenticationProviderSuccess(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	token := NewUsernamePasswordToken("admin", "admin123")
	result, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Authenticated() {
		t.Error("expected authentication to be successful")
	}
	if result.Principal() == nil {
		t.Error("expected non-nil principal")
	}
	if user, ok := result.Principal().(UserDetails); !ok || user.Username() != "admin" {
		t.Errorf("expected principal with username 'admin', got %v", result.Principal())
	}
	if len(result.Authorities()) != 1 || result.Authorities()[0] != "ROLE_ADMIN" {
		t.Errorf("expected authorities [ROLE_ADMIN], got %v", result.Authorities())
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
	if err == nil || err.Error() != "user is disabled" {
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
	if err == nil || err.Error() != "user account is locked" {
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

func TestAnonymousAuthenticationProvider(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	token := NewUsernamePasswordToken("user", "pass")
	result, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Authenticated() {
		t.Error("expected anonymous auth to be authenticated")
	}
	if result.Principal() != "anonymousUser" {
		t.Errorf("expected principal 'anonymousUser', got '%v'", result.Principal())
	}
	if len(result.Authorities()) != 1 || result.Authorities()[0] != "ROLE_ANONYMOUS" {
		t.Errorf("expected authorities [ROLE_ANONYMOUS], got %v", result.Authorities())
	}
}

func TestAnonymousAuthenticationProviderAlreadyAuthenticated(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	token := NewAuthenticatedUsernamePasswordToken("admin", nil, []string{"ROLE_ADMIN"})
	result, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for already authenticated token")
	}
}

func TestAnonymousAuthenticationProviderNilToken(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	result, err := provider.Authenticate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Authenticated() {
		t.Error("expected anonymous auth to be authenticated")
	}
	if result.Principal() != "anonymousUser" {
		t.Errorf("expected principal 'anonymousUser', got '%v'", result.Principal())
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
	result, err := manager.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Authenticated() {
		t.Error("expected authentication to be successful")
	}
	if user, ok := result.Principal().(UserDetails); !ok || user.Username() != "admin" {
		t.Errorf("expected principal with username 'admin', got %v", result.Principal())
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
	result, err := manager.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Authenticated() {
		t.Error("expected authentication to be successful")
	}
	if user, ok := result.Principal().(UserDetails); !ok || user.Username() != "admin" {
		t.Errorf("expected principal with username 'admin', got %v", result.Principal())
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
	result, err := manager.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Authenticated() {
		t.Error("expected authentication to be successful")
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

func TestDaoAuthenticationProviderWithUserDetails(t *testing.T) {
	t.Parallel()

	userDetailsService := NewInMemoryUserDetailsService()
	userDetailsService.CreateUser("admin", "admin123", []string{"ROLE_ADMIN"})

	encoder := NewNoOpPasswordEncoder()
	provider := NewDaoAuthenticationProvider(userDetailsService, encoder, log.Build())

	user := NewInMemoryUserDetails("admin", "admin123", []string{"ROLE_ADMIN"})
	token := NewUsernamePasswordToken(user, "admin123")
	result, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Authenticated() {
		t.Error("expected authentication to be successful")
	}
	if principal, ok := result.Principal().(UserDetails); !ok || principal.Username() != "admin" {
		t.Errorf("expected principal with username 'admin', got %v", result.Principal())
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
