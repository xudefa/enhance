package authorization

import (
	"context"
	"testing"
)

func TestAffirmativeBased_decision_OneGrantPasses(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	m := NewAffirmativeBased(v1, v2)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestAffirmativeBased_decision_AllDenyFails(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessDenied, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	m := NewAffirmativeBased(v1, v2)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestAffirmativeBased_decision_AllAbstainDefaultDeny(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	m := NewAffirmativeBased(v)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestAffirmativeBased_decision_AllAbstainAllowWhenConfigured(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	m := NewAffirmativeBased(v)
	m.(*affirmativeBased).SetAllowIfAllAbstainDecisions(true)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestAffirmativeBased_decision_SupportsFunc(t *testing.T) {
	t.Parallel()

	v := &mockVoter{supported: true}
	m := NewAffirmativeBased(v)

	if !m.Supports("read") {
		t.Error("expected Supports to return true")
	}
}

func TestAffirmativeBased_decision_AddVoterFunc(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessDenied, supported: true}
	m := NewAffirmativeBased(v1)

	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	m.(*affirmativeBased).AddVoter(v2)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil after adding grant voter, got %v", err)
	}
}

func TestAffirmativeBased_decision_EmptyVoters(t *testing.T) {
	t.Parallel()

	m := NewAffirmativeBased()
	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil for empty voters, got %v", err)
	}
}

func TestUnanimousBased_decision_AllGrantPasses(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	m := NewUnanimousBased(v1, v2)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestUnanimousBased_decision_OneDenyFails(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	m := NewUnanimousBased(v1, v2)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestUnanimousBased_decision_AllAbstainDefaultDeny(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	m := NewUnanimousBased(v)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestUnanimousBased_decision_AllAbstainAllowWhenConfigured(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	m := NewUnanimousBased(v)
	m.(*unanimousBased).SetAllowIfAllAbstainDecisions(true)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestUnanimousBased_decision_SupportsFunc(t *testing.T) {
	t.Parallel()

	v := &mockVoter{supported: true}
	m := NewUnanimousBased(v)

	if !m.Supports("read") {
		t.Error("expected Supports to return true")
	}
}

func TestUnanimousBased_decision_AddVoterFunc(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	m := NewUnanimousBased(v1)

	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	m.(*unanimousBased).AddVoter(v2)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil after adding grant voter, got %v", err)
	}
}

func TestConsensusBased_decision_MajorityGrant(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	v3 := &mockVoter{voteResult: AccessDenied, supported: true}
	m := NewConsensusBased(v1, v2, v3)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestConsensusBased_decision_MajorityDeny(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	v3 := &mockVoter{voteResult: AccessDenied, supported: true}
	m := NewConsensusBased(v1, v2, v3)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestConsensusBased_decision_EqualDefaultDeny(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	m := NewConsensusBased(v1, v2)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestConsensusBased_decision_EqualAllowWhenConfigured(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	m := NewConsensusBased(v1, v2)
	m.(*consensusBased).SetAllowIfEqualGrantedDenied(true)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestConsensusBased_decision_AllAbstainDefaultDeny(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	m := NewConsensusBased(v)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestConsensusBased_decision_AllAbstainAllowWhenConfigured(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	m := NewConsensusBased(v)
	m.(*consensusBased).SetAllowIfAllAbstainDecisions(true)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestConsensusBased_decision_SupportsFunc(t *testing.T) {
	t.Parallel()

	v := &mockVoter{supported: true}
	m := NewConsensusBased(v)

	if !m.Supports("read") {
		t.Error("expected Supports to return true")
	}
}

func TestConsensusBased_decision_AddVoterFunc(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessDenied, supported: true}
	m := NewConsensusBased(v1)

	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	v3 := &mockVoter{voteResult: AccessGranted, supported: true}
	m.(*consensusBased).AddVoter(v2)
	m.(*consensusBased).AddVoter(v3)

	err := m.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil after adding grant voters, got %v", err)
	}
}

func TestWebExpressionVoter_decision_PermitAll(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	result := voter.Vote(context.Background(), nil, "/api", []string{"permitAll"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", result)
	}
}

func TestWebExpressionVoter_decision_DenyAll(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	result := voter.Vote(context.Background(), nil, "/api", []string{"denyAll"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", result)
	}
}

func TestWebExpressionVoter_decision_AuthenticatedCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	auth := &mockAuthentication{authenticated: true}
	result := voter.Vote(context.Background(), auth, "/api", []string{"authenticated"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", result)
	}

	result = voter.Vote(context.Background(), nil, "/api", []string{"authenticated"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied for nil auth, got %d", result)
	}
}

func TestWebExpressionVoter_decision_HasRoleCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"ROLE_ADMIN", "ROLE_USER"}}

	result := voter.Vote(context.Background(), auth, "/api", []string{"hasRole('ADMIN')"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", result)
	}

	result = voter.Vote(context.Background(), auth, "/api", []string{"hasRole('GUEST')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", result)
	}
}

func TestWebExpressionVoter_decision_HasAnyRoleCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"ROLE_USER"}}

	result := voter.Vote(context.Background(), auth, "/api", []string{"hasAnyRole('ADMIN','USER')"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", result)
	}

	result = voter.Vote(context.Background(), auth, "/api", []string{"hasAnyRole('ADMIN','GUEST')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", result)
	}
}

func TestWebExpressionVoter_decision_HasAuthorityCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"read", "write"}}

	result := voter.Vote(context.Background(), auth, "/api", []string{"hasAuthority('read')"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", result)
	}

	result = voter.Vote(context.Background(), auth, "/api", []string{"hasAuthority('delete')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", result)
	}
}

func TestWebExpressionVoter_decision_HasAnyAuthorityCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"read"}}

	result := voter.Vote(context.Background(), auth, "/api", []string{"hasAnyAuthority('read','write')"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", result)
	}

	result = voter.Vote(context.Background(), auth, "/api", []string{"hasAnyAuthority('write','delete')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", result)
	}
}

func TestWebExpressionVoter_decision_EmptyAttrs(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	result := voter.Vote(context.Background(), nil, "/api", []string{})
	if result != AccessAbstain {
		t.Errorf("expected AccessAbstain, got %d", result)
	}
}

func TestWebExpressionVoter_decision_UnsupportedExpr(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"ROLE_USER"}}
	result := voter.Vote(context.Background(), auth, "/api", []string{"unknown"})
	if result != AccessAbstain {
		t.Errorf("expected AccessAbstain, got %d", result)
	}
}

func TestWebExpressionVoter_decision_NilAuthWithRole(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	result := voter.Vote(context.Background(), nil, "/api", []string{"hasRole('ADMIN')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", result)
	}
}

func TestWebExpressionVoter_decision_SupportsFunc(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	if !voter.Supports("anything") {
		t.Error("expected Supports to return true for anything")
	}
}
