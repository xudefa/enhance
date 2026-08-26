package authentication

import (
	"context"
	"testing"
)

type mockAuthProvider struct {
	supportsToken bool
	authResult    Authentication
	authErr       error
}

func (p *mockAuthProvider) Supports(_ AuthenticationToken) bool {
	return p.supportsToken
}

func (p *mockAuthProvider) Authenticate(_ context.Context, _ AuthenticationToken) (Authentication, error) {
	return p.authResult, p.authErr
}

func TestNewProviderManager_NoProviders(t *testing.T) {
	t.Parallel()

	m := NewProviderManager()
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestProviderManager_Authenticate_NoProviders(t *testing.T) {
	t.Parallel()

	m := NewProviderManager()
	token := NewUsernamePasswordToken("user", "pass")
	_, err := m.Authenticate(context.Background(), token)
	if err != ErrAuthenticationFailed {
		t.Errorf("expected ErrAuthenticationFailed, got %v", err)
	}
}

func TestProviderManager_Authenticate_Success(t *testing.T) {
	t.Parallel()

	result := NewAuthenticatedUsernamePasswordToken("user", nil, []string{"ROLE_USER"})
	provider := &mockAuthProvider{
		supportsToken: true,
		authResult:    result,
	}
	m := NewProviderManager(provider)

	token := NewUsernamePasswordToken("user", "pass")
	auth, err := m.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth == nil {
		t.Fatal("expected non-nil auth")
	}
	if auth.Principal() != "user" {
		t.Errorf("expected principal 'user', got %v", auth.Principal())
	}
}

func TestProviderManager_Authenticate_ProviderError(t *testing.T) {
	t.Parallel()

	provider := &mockAuthProvider{
		supportsToken: true,
		authErr:       ErrBadCredentials,
	}
	m := NewProviderManager(provider)

	token := NewUsernamePasswordToken("user", "pass")
	_, err := m.Authenticate(context.Background(), token)
	if err != ErrBadCredentials {
		t.Errorf("expected ErrBadCredentials, got %v", err)
	}
}

func TestProviderManager_Authenticate_ProviderSkips(t *testing.T) {
	t.Parallel()

	provider := &mockAuthProvider{
		supportsToken: false,
	}
	m := NewProviderManager(provider)

	token := NewUsernamePasswordToken("user", "pass")
	_, err := m.Authenticate(context.Background(), token)
	if err != ErrAuthenticationFailed {
		t.Errorf("expected ErrAuthenticationFailed when no provider matches, got %v", err)
	}
}

func TestProviderManager_Authenticate_ProviderReturnsNil(t *testing.T) {
	t.Parallel()

	provider := &mockAuthProvider{
		supportsToken: true,
		authResult:    nil,
	}
	m := NewProviderManager(provider)

	token := NewUsernamePasswordToken("user", "pass")
	_, err := m.Authenticate(context.Background(), token)
	if err != ErrAuthenticationFailed {
		t.Errorf("expected ErrAuthenticationFailed when provider returns nil, got %v", err)
	}
}

func TestProviderManager_AddProvider(t *testing.T) {
	t.Parallel()

	m := NewProviderManager().(*ProviderManager)
	provider := &mockAuthProvider{supportsToken: true}
	m.AddProvider(provider)

	if len(m.providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(m.providers))
	}
}

func TestProviderManager_MultipleProviders_FirstMatches(t *testing.T) {
	t.Parallel()

	result1 := NewAuthenticatedUsernamePasswordToken("user1", nil, []string{})
	p1 := &mockAuthProvider{supportsToken: true, authResult: result1}
	p2 := &mockAuthProvider{supportsToken: true, authResult: NewAuthenticatedUsernamePasswordToken("user2", nil, []string{})}

	m := NewProviderManager(p1, p2)
	token := NewUsernamePasswordToken("user1", "pass")
	auth, err := m.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth.Principal() != "user1" {
		t.Errorf("expected principal 'user1', got %v", auth.Principal())
	}
}

func TestProviderManager_MultipleProviders_FirstSkipsSecondSucceeds(t *testing.T) {
	t.Parallel()

	p1 := &mockAuthProvider{supportsToken: false}
	result2 := NewAuthenticatedUsernamePasswordToken("user2", nil, []string{})
	p2 := &mockAuthProvider{supportsToken: true, authResult: result2}

	m := NewProviderManager(p1, p2)
	token := NewUsernamePasswordToken("user2", "pass")
	auth, err := m.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if auth.Principal() != "user2" {
		t.Errorf("expected principal 'user2', got %v", auth.Principal())
	}
}
