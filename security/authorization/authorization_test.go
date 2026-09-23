package authorization

import (
	"context"
	"testing"
)

// mockAuthentication 模拟认证信息。
type mockAuthentication struct {
	principal     any
	credentials   any
	authorities   []string
	authenticated bool
}

func (m *mockAuthentication) Principal() any {
	return m.principal
}

func (m *mockAuthentication) Credentials() any {
	return m.credentials
}

func (m *mockAuthentication) Authorities() []string {
	return m.authorities
}

func (m *mockAuthentication) Authenticated() bool {
	return m.authenticated
}

// mockVoter 模拟投票者。
type mockVoter struct {
	voteResult int
	supported  bool
}

func (v *mockVoter) Vote(_ context.Context, _ Authentication, _ string, _ []string) int {
	return v.voteResult
}

func (v *mockVoter) Supports(_ string) bool {
	return v.supported
}

func TestAffirmativeBased_GrantIfAnyGrants(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	voter2 := &mockVoter{voteResult: AccessGranted, supported: true}
	manager := NewAffirmativeBased(voter1, voter2)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAffirmativeBased_DenyIfAllDeny(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessDenied, supported: true}
	voter2 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewAffirmativeBased(voter1, voter2)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/admin", []string{"hasRole('ADMIN')"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestAffirmativeBased_AllAbstainDefaultDeny(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewAffirmativeBased(voter1)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied when all abstain, got %v", err)
	}
}

func TestAffirmativeBased_AllAbstainAllowIfConfigured(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewAffirmativeBased(voter1)
	manager.(*affirmativeBased).SetAllowIfAllAbstainDecisions(true)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error when allowIfAllAbstainDecisions is true, got %v", err)
	}
}

func TestAffirmativeBased_Supports(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewAffirmativeBased(voter1)

	if !manager.Supports("hasRole('ADMIN')") {
		t.Error("expected Supports to return true")
	}
}

func TestAffirmativeBased_AddVoter(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewAffirmativeBased(voter1)

	newVoter := &mockVoter{voteResult: AccessGranted, supported: true}
	manager.(*affirmativeBased).AddVoter(newVoter)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error after adding grant voter, got %v", err)
	}
}

func TestUnanimousBased_AllGrant(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessGranted, supported: true}
	voter2 := &mockVoter{voteResult: AccessGranted, supported: true}
	manager := NewUnanimousBased(voter1, voter2)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestUnanimousBased_DenyIfAnyDenies(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessGranted, supported: true}
	voter2 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewUnanimousBased(voter1, voter2)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestUnanimousBased_AllAbstainDefaultDeny(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewUnanimousBased(voter1)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied when all abstain, got %v", err)
	}
}

func TestUnanimousBased_AllAbstainAllowIfConfigured(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewUnanimousBased(voter1)
	manager.(*unanimousBased).SetAllowIfAllAbstainDecisions(true)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error when allowIfAllAbstainDecisions is true, got %v", err)
	}
}

func TestConsensusBased_MajorityGrant(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessGranted, supported: true}
	voter2 := &mockVoter{voteResult: AccessGranted, supported: true}
	voter3 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewConsensusBased(voter1, voter2, voter3)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error for majority grant, got %v", err)
	}
}

func TestConsensusBased_MajorityDeny(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessGranted, supported: true}
	voter2 := &mockVoter{voteResult: AccessDenied, supported: true}
	voter3 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewConsensusBased(voter1, voter2, voter3)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied for majority deny, got %v", err)
	}
}

func TestConsensusBased_EqualDefaultDeny(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessGranted, supported: true}
	voter2 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewConsensusBased(voter1, voter2)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied when equal, got %v", err)
	}
}

func TestConsensusBased_EqualAllowIfConfigured(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessGranted, supported: true}
	voter2 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewConsensusBased(voter1, voter2)
	manager.(*consensusBased).SetAllowIfEqualGrantedDenied(true)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error when allowIfEqualGrantedDenied is true, got %v", err)
	}
}

func TestConsensusBased_AllAbstainAllowIfConfigured(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewConsensusBased(voter1)
	manager.(*consensusBased).SetAllowIfAllAbstainDecisions(true)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error when allowIfAllAbstainDecisions is true, got %v", err)
	}
}

func TestAffirmativeBased_WithWebExpressionVoter(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	manager := NewAffirmativeBased(voter)

	auth := &mockAuthentication{
		principal:     "admin",
		authorities:   []string{"ROLE_ADMIN"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/admin", []string{"hasRole('ADMIN')"})
	if err != nil {
		t.Errorf("expected no error for admin accessing admin route, got %v", err)
	}

	err = manager.Decide(context.Background(), auth, "/api/admin", []string{"hasRole('GUEST')"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied for admin without GUEST role, got %v", err)
	}
}

func TestAffirmativeBased_NilAuthentication(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	manager := NewAffirmativeBased(voter)

	err := manager.Decide(context.Background(), nil, "/api/users", []string{"authenticated"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied for nil authentication, got %v", err)
	}
}

func TestAffirmativeBased_PermitAllWithNilAuth(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	manager := NewAffirmativeBased(voter)

	err := manager.Decide(context.Background(), nil, "/api/public", []string{"permitAll"})
	if err != nil {
		t.Errorf("expected no error for permitAll, got %v", err)
	}
}

func TestAffirmativeBased_EmptyAttributes(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	manager := NewAffirmativeBased(voter)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := manager.Decide(context.Background(), auth, "/api/users", []string{})
	if err != nil {
		t.Errorf("expected no error for empty attributes (no constraints apply), got %v", err)
	}
}

func BenchmarkAffirmativeBased_Decide(b *testing.B) {
	voter := NewWebExpressionVoter()
	manager := NewAffirmativeBased(voter)

	auth := &mockAuthentication{
		principal:     "admin",
		authorities:   []string{"ROLE_ADMIN"},
		authenticated: true,
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.Decide(ctx, auth, "/api/admin", []string{"hasRole('ADMIN')"})
	}
}
