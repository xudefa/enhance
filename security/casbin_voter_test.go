package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/security/authorization"
)

func TestNewCasbinVoter(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{}
	voter, err := NewCasbinVoter(enforcer)
	if err != nil {
		t.Fatalf("NewCasbinVoter error: %v", err)
	}
	if voter == nil {
		t.Fatal("expected non-nil voter")
	}
}

func TestNewCasbinVoter_NilEnforcer(t *testing.T) {
	t.Parallel()

	_, err := NewCasbinVoter(nil)
	if err == nil {
		t.Fatal("expected error for nil enforcer")
	}
}

func TestMustNewCasbinVoter(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{}
	voter := MustNewCasbinVoter(enforcer)
	if voter == nil {
		t.Fatal("expected non-nil voter")
	}
}

func TestMustNewCasbinVoter_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil enforcer")
		}
	}()

	MustNewCasbinVoter(nil)
}

func TestCasbinVoter_Vote_Granted(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{allowed: true}
	voter := MustNewCasbinVoter(enforcer)

	auth := &mockAuthentication{authenticated: true, principal: "admin"}
	ctx := context.Background()

	result := voter.Vote(ctx, auth, "GET:/api/users", []string{"read"})
	if result != ACCESS_GRANTED {
		t.Errorf("expected ACCESS_GRANTED, got %d", result)
	}
}

func TestCasbinVoter_Vote_Denied(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{allowed: false}
	voter := MustNewCasbinVoter(enforcer)

	auth := &mockAuthentication{authenticated: true, principal: "user"}
	ctx := context.Background()

	result := voter.Vote(ctx, auth, "DELETE:/api/users/1", []string{"delete"})
	if result != ACCESS_DENIED {
		t.Errorf("expected ACCESS_DENIED, got %d", result)
	}
}

func TestCasbinVoter_Vote_Abstain(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{}
	voter := MustNewCasbinVoter(enforcer)

	// 未认证的用户应该弃权
	auth := &mockAuthentication{authenticated: false}
	ctx := context.Background()

	result := voter.Vote(ctx, auth, "GET:/api/users", []string{"read"})
	if result != ACCESS_ABSTAIN {
		t.Errorf("expected ACCESS_ABSTAIN, got %d", result)
	}
}

func TestCasbinVoter_Vote_NilAuthentication(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{}
	voter := MustNewCasbinVoter(enforcer)

	ctx := context.Background()
	result := voter.Vote(ctx, nil, "GET:/api/users", []string{"read"})
	if result != ACCESS_ABSTAIN {
		t.Errorf("expected ACCESS_ABSTAIN for nil authentication, got %d", result)
	}
}

func TestCasbinVoter_Vote_EnforcerError(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{err: true}
	voter := MustNewCasbinVoter(enforcer)

	auth := &mockAuthentication{authenticated: true, principal: "admin"}
	ctx := context.Background()

	result := voter.Vote(ctx, auth, "GET:/api/users", []string{"read"})
	if result != ACCESS_DENIED {
		t.Errorf("expected ACCESS_DENIED on enforcer error, got %d", result)
	}
}

func TestCasbinVoter_Supports(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{}
	voter := MustNewCasbinVoter(enforcer)

	if !voter.Supports("any-attribute") {
		t.Error("expected CasbinVoter to support any attribute")
	}
}

type mockCasbinEnforcer struct {
	allowed bool
	err     bool
}

func (m *mockCasbinEnforcer) Enforce(ctx context.Context, subject, object, action string) (bool, error) {
	if m.err {
		return false, context.DeadlineExceeded
	}
	return m.allowed, nil
}

func (m *mockCasbinEnforcer) AddPolicy(ctx context.Context, sub, obj, act string) error {
	return nil
}

func (m *mockCasbinEnforcer) RemovePolicy(ctx context.Context, sub, obj, act string) error {
	return nil
}

func (m *mockCasbinEnforcer) GetPolicy(ctx context.Context) ([][]string, error) {
	return nil, nil
}

func (m *mockCasbinEnforcer) LoadPolicy(ctx context.Context) error {
	return nil
}

func (m *mockCasbinEnforcer) SavePolicy(ctx context.Context) error {
	return nil
}

// 确保mockAuthentication实现authorization.Authentication接口
var _ authorization.Authentication = (*mockAuthentication)(nil)

type mockAuthentication struct {
	authenticated bool
	principal     string
}

func (m *mockAuthentication) Principal() any          { return m.principal }
func (m *mockAuthentication) Credentials() any        { return nil }
func (m *mockAuthentication) Authorities() []string   { return []string{"ROLE_USER"} }
func (m *mockAuthentication) Authenticated() bool     { return m.authenticated }
func (m *mockAuthentication) Name() string            { return m.principal }
func (m *mockAuthentication) SetAuthenticated(bool)   {}
func (m *mockAuthentication) SetAuthorities([]string) {}
