package authentication

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/log"
)

type testUserDetailsService struct {
	user UserDetails
	err  error
}

func (s *testUserDetailsService) LoadUserByUsername(_ context.Context, _ string) (UserDetails, error) {
	return s.user, s.err
}

type testPasswordEncoder struct {
	matches bool
}

func (e *testPasswordEncoder) Encode(rawPassword string) string { return rawPassword }
func (e *testPasswordEncoder) Matches(_, _ string) bool         { return e.matches }

func TestNewDaoAuthenticationProvider_NilLogger(t *testing.T) {
	t.Parallel()

	provider := NewDaoAuthenticationProvider(nil, nil, nil)
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestDaoAuthenticationProvider_Supports_UsernamePasswordToken(t *testing.T) {
	t.Parallel()

	provider := NewDaoAuthenticationProvider(nil, nil, log.Build())
	token := NewUsernamePasswordToken("user", "pass")
	if !provider.Supports(token) {
		t.Error("expected Supports to return true for UsernamePasswordToken")
	}
}

func TestDaoAuthenticationProvider_Supports_OtherToken(t *testing.T) {
	t.Parallel()

	provider := NewDaoAuthenticationProvider(nil, nil, log.Build())
	token := NewAnonymousToken()
	if provider.Supports(token) {
		t.Error("expected Supports to return false for anonymousToken")
	}
}

func TestDaoAuthenticationProvider_EmptyUsername(t *testing.T) {
	t.Parallel()

	svc := &testUserDetailsService{}
	encoder := &testPasswordEncoder{matches: true}
	provider := NewDaoAuthenticationProvider(svc, encoder, log.Build())

	token := NewUsernamePasswordToken("", "pass")
	_, err := provider.Authenticate(context.Background(), token)
	if err != ErrBadCredentials {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestDaoAuthenticationProvider_UserNotFound(t *testing.T) {
	t.Parallel()

	svc := &testUserDetailsService{err: ErrUserNotFound}
	encoder := &testPasswordEncoder{matches: true}
	provider := NewDaoAuthenticationProvider(svc, encoder, log.Build())

	token := NewUsernamePasswordToken("nobody", "pass")
	_, err := provider.Authenticate(context.Background(), token)
	if err != ErrBadCredentials {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestDaoAuthenticationProvider_WrongPassword(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("admin", "correct", []string{"ROLE_ADMIN"})
	svc := &testUserDetailsService{user: user}
	encoder := &testPasswordEncoder{matches: false}
	provider := NewDaoAuthenticationProvider(svc, encoder, log.Build())

	token := NewUsernamePasswordToken("admin", "wrong")
	_, err := provider.Authenticate(context.Background(), token)
	if err != ErrBadCredentials {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestDaoAuthenticationProvider_Success(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("admin", "pass", []string{"ROLE_ADMIN"})
	svc := &testUserDetailsService{user: user}
	encoder := &testPasswordEncoder{matches: true}
	provider := NewDaoAuthenticationProvider(svc, encoder, log.Build())

	token := NewUsernamePasswordToken("admin", "pass")
	auth, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !auth.Authenticated() {
		t.Error("expected authenticated")
	}
	if auth.Principal() == nil {
		t.Error("expected non-nil principal")
	}
}

func TestDaoAuthenticationProvider_UserDetailsPrincipal(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("admin", "pass", []string{"ROLE_ADMIN"})
	svc := &testUserDetailsService{user: user}
	encoder := &testPasswordEncoder{matches: true}
	provider := NewDaoAuthenticationProvider(svc, encoder, log.Build())

	token := NewUsernamePasswordToken(user, "pass")
	auth, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !auth.Authenticated() {
		t.Error("expected authenticated")
	}
}

func TestDaoAuthenticationProvider_UnsupportedPrincipal(t *testing.T) {
	t.Parallel()

	svc := &testUserDetailsService{}
	encoder := &testPasswordEncoder{matches: true}
	provider := NewDaoAuthenticationProvider(svc, encoder, log.Build())

	token := NewUsernamePasswordToken(12345, "pass")
	_, err := provider.Authenticate(context.Background(), token)
	if err == nil {
		t.Error("expected error for unsupported principal")
	}
}

func TestDaoAuthenticationProvider_DisabledUser(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("admin", "pass", []string{})
	user.SetEnabled(false)
	svc := &testUserDetailsService{user: user}
	encoder := &testPasswordEncoder{matches: true}
	provider := NewDaoAuthenticationProvider(svc, encoder, log.Build())

	token := NewUsernamePasswordToken("admin", "pass")
	_, err := provider.Authenticate(context.Background(), token)
	if err == nil || err.Error() != "user is disabled" {
		t.Errorf("expected 'user is disabled', got %v", err)
	}
}

func TestDaoAuthenticationProvider_LockedUser(t *testing.T) {
	t.Parallel()

	user := NewInMemoryUserDetails("admin", "pass", []string{})
	user.SetAccountNonLocked(false)
	svc := &testUserDetailsService{user: user}
	encoder := &testPasswordEncoder{matches: true}
	provider := NewDaoAuthenticationProvider(svc, encoder, log.Build())

	token := NewUsernamePasswordToken("admin", "pass")
	_, err := provider.Authenticate(context.Background(), token)
	if err == nil || err.Error() != "user account is locked" {
		t.Errorf("expected 'user account is locked', got %v", err)
	}
}

func TestAnonymousAuthenticationProvider_NilToken(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()
	result, err := provider.Authenticate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Principal() != "anonymousUser" {
		t.Errorf("expected principal 'anonymousUser', got %v", result.Principal())
	}
}

func TestAnonymousAuthenticationProvider_UnauthenticatedToken(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()
	token := NewUsernamePasswordToken("user", nil)
	result, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestAnonymousAuthenticationProvider_AuthenticatedToken(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()
	token := NewAuthenticatedUsernamePasswordToken("admin", nil, []string{"ROLE_ADMIN"})
	result, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for already authenticated token")
	}
}

func TestAnonymousAuthenticationProvider_TokenWithCredentials(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()
	token := NewUsernamePasswordToken("user", "pass")
	result, err := provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for token with credentials")
	}
}

func TestAnonymousAuthenticationProvider_Supports(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()
	token := NewUsernamePasswordToken("user", "pass")
	if !provider.Supports(token) {
		t.Error("expected Supports to return true for UsernamePasswordToken")
	}

	anonToken := NewAnonymousToken()
	if provider.Supports(anonToken) {
		t.Error("expected Supports to return false for anonymousToken")
	}
}
