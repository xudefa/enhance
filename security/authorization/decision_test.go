package authorization

import (
	"context"
	"testing"
)

func TestAffirmativeBased_decision_OneGrantPasses(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	manager := NewAffirmativeBased(v1, v2)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestAffirmativeBased_decision_AllDenyFails(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessDenied, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewAffirmativeBased(v1, v2)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestAffirmativeBased_decision_AllAbstainDefaultDeny(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewAffirmativeBased(v)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestAffirmativeBased_decision_AllAbstainAllowWhenConfigured(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewAffirmativeBased(v)
	manager.(*affirmativeBased).SetAllowIfAllAbstainDecisions(true)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestAffirmativeBased_decision_SupportsFunc(t *testing.T) {
	t.Parallel()

	v := &mockVoter{supported: true}
	manager := NewAffirmativeBased(v)

	if !manager.Supports("read") {
		t.Error("expected Supports to return true")
	}
}

func TestAffirmativeBased_decision_AddVoterFunc(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewAffirmativeBased(v1)

	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	manager.(*affirmativeBased).AddVoter(v2)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil after adding grant voter, got %v", err)
	}
}

func TestAffirmativeBased_decision_EmptyVoters(t *testing.T) {
	t.Parallel()

	manager := NewAffirmativeBased()
	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil for empty voters, got %v", err)
	}
}

func TestUnanimousBased_decision_AllGrantPasses(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	manager := NewUnanimousBased(v1, v2)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestUnanimousBased_decision_OneDenyFails(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewUnanimousBased(v1, v2)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestUnanimousBased_decision_AllAbstainDefaultDeny(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewUnanimousBased(v)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestUnanimousBased_decision_AllAbstainAllowWhenConfigured(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewUnanimousBased(v)
	manager.(*unanimousBased).SetAllowIfAllAbstainDecisions(true)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestUnanimousBased_decision_SupportsFunc(t *testing.T) {
	t.Parallel()

	v := &mockVoter{supported: true}
	manager := NewUnanimousBased(v)

	if !manager.Supports("read") {
		t.Error("expected Supports to return true")
	}
}

func TestUnanimousBased_decision_AddVoterFunc(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewUnanimousBased(v1)

	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	manager.(*unanimousBased).AddVoter(v2)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil after adding grant voter, got %v", err)
	}
}

func TestConsensusBased_decision_MajorityGrant(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	v3 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewConsensusBased(v1, v2, v3)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestConsensusBased_decision_MajorityDeny(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	v3 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewConsensusBased(v1, v2, v3)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestConsensusBased_decision_EqualDefaultDeny(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewConsensusBased(v1, v2)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestConsensusBased_decision_EqualAllowWhenConfigured(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessGranted, supported: true}
	v2 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewConsensusBased(v1, v2)
	manager.(*consensusBased).SetAllowIfEqualGrantedDenied(true)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestConsensusBased_decision_AllAbstainDefaultDeny(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewConsensusBased(v)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied, got %v", err)
	}
}

func TestConsensusBased_decision_AllAbstainAllowWhenConfigured(t *testing.T) {
	t.Parallel()

	v := &mockVoter{voteResult: AccessAbstain, supported: true}
	manager := NewConsensusBased(v)
	manager.(*consensusBased).SetAllowIfAllAbstainDecisions(true)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestConsensusBased_decision_SupportsFunc(t *testing.T) {
	t.Parallel()

	v := &mockVoter{supported: true}
	manager := NewConsensusBased(v)

	if !manager.Supports("read") {
		t.Error("expected Supports to return true")
	}
}

func TestConsensusBased_decision_AddVoterFunc(t *testing.T) {
	t.Parallel()

	v1 := &mockVoter{voteResult: AccessDenied, supported: true}
	manager := NewConsensusBased(v1)

	v2 := &mockVoter{voteResult: AccessGranted, supported: true}
	v3 := &mockVoter{voteResult: AccessGranted, supported: true}
	manager.(*consensusBased).AddVoter(v2)
	manager.(*consensusBased).AddVoter(v3)

	err := manager.Decide(context.Background(), nil, "/api", []string{"read"})
	if err != nil {
		t.Errorf("expected nil after adding grant voters, got %v", err)
	}
}

func TestWebExpressionVoter_decision_PermitAll(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	voteResult := voter.Vote(context.Background(), nil, "/api", []string{"permitAll"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_DenyAll(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	voteResult := voter.Vote(context.Background(), nil, "/api", []string{"denyAll"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_AuthenticatedCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	auth := &mockAuthentication{authenticated: true}
	voteResult := voter.Vote(context.Background(), auth, "/api", []string{"authenticated"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), nil, "/api", []string{"authenticated"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied for nil auth, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_HasRoleCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"ROLE_ADMIN", "ROLE_USER"}}

	voteResult := voter.Vote(context.Background(), auth, "/api", []string{"hasRole('ADMIN')"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), auth, "/api", []string{"hasRole('GUEST')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_HasAnyRoleCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"ROLE_USER"}}

	voteResult := voter.Vote(context.Background(), auth, "/api", []string{"hasAnyRole('ADMIN','USER')"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), auth, "/api", []string{"hasAnyRole('ADMIN','GUEST')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_HasAuthorityCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"read", "write"}}

	voteResult := voter.Vote(context.Background(), auth, "/api", []string{"hasAuthority('read')"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), auth, "/api", []string{"hasAuthority('delete')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_HasAnyAuthorityCheck(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"read"}}

	voteResult := voter.Vote(context.Background(), auth, "/api", []string{"hasAnyAuthority('read','write')"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted, got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), auth, "/api", []string{"hasAnyAuthority('write','delete')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_EmptyAttrs(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	voteResult := voter.Vote(context.Background(), nil, "/api", []string{})
	if voteResult != AccessAbstain {
		t.Errorf("expected AccessAbstain, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_UnsupportedExpr(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{authorities: []string{"ROLE_USER"}}
	voteResult := voter.Vote(context.Background(), auth, "/api", []string{"unknown"})
	if voteResult != AccessAbstain {
		t.Errorf("expected AccessAbstain, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_NilAuthWithRole(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	voteResult := voter.Vote(context.Background(), nil, "/api", []string{"hasRole('ADMIN')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied, got %d", voteResult)
	}
}

func TestWebExpressionVoter_decision_SupportsFunc(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	if !voter.Supports("anything") {
		t.Error("expected Supports to return true for anything")
	}
}
