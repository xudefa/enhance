package security

import (
	"context"
	"testing"
)

func TestAffirmativeBased_Supports(t *testing.T) {
	t.Parallel()

	manager := NewAffirmativeBased()
	if !manager.Supports("any-attribute") {
		t.Error("expected AffirmativeBased to support any attribute")
	}
}

func TestAffirmativeBased_AddVoter(t *testing.T) {
	t.Parallel()

	manager := NewAffirmativeBased()
	voter := &mockAccessDecisionVoter{}
	manager.AddVoter(voter)

	if len(manager.decisionVoters) != 1 {
		t.Errorf("expected 1 voter, got %d", len(manager.decisionVoters))
	}
}

func TestAffirmativeBased_SetAllowIfAllAbstainDecisions(t *testing.T) {
	t.Parallel()

	manager := NewAffirmativeBased()
	manager.SetAllowIfAllAbstainDecisions(true)

	if !manager.allowIfAllAbstainDecisions {
		t.Error("expected allowIfAllAbstainDecisions to be true")
	}
}

func TestUnanimousBased_Supports(t *testing.T) {
	t.Parallel()

	manager := NewUnanimousBased()
	if !manager.Supports("any-attribute") {
		t.Error("expected UnanimousBased to support any attribute")
	}
}

func TestUnanimousBased_AddVoter(t *testing.T) {
	t.Parallel()

	manager := NewUnanimousBased()
	voter := &mockAccessDecisionVoter{}
	manager.AddVoter(voter)

	if len(manager.decisionVoters) != 1 {
		t.Errorf("expected 1 voter, got %d", len(manager.decisionVoters))
	}
}

func TestUnanimousBased_SetAllowIfAllAbstainDecisions(t *testing.T) {
	t.Parallel()

	manager := NewUnanimousBased()
	manager.SetAllowIfAllAbstainDecisions(true)

	if !manager.allowIfAllAbstainDecisions {
		t.Error("expected allowIfAllAbstainDecisions to be true")
	}
}

func TestWebExpressionVoter_Supports(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	if !voter.Supports("hasRole('ROLE_ADMIN')") {
		t.Error("expected WebExpressionVoter to support expression attributes")
	}
	if !voter.Supports("unknown-attribute") {
		t.Error("expected WebExpressionVoter to support any attribute")
	}
}

func TestRoleVoter_Supports(t *testing.T) {
	t.Parallel()

	voter := NewRoleVoter()
	if !voter.Supports("ROLE_ADMIN") {
		t.Error("expected RoleVoter to support ROLE_ prefixed attributes")
	}
	if voter.Supports("ADMIN") {
		t.Error("expected RoleVoter to not support non-ROLE_ prefixed attributes")
	}
}

func TestRoleVoter_SetRolePrefix(t *testing.T) {
	t.Parallel()

	voter := NewRoleVoter()
	voter.SetRolePrefix("CUSTOM_")

	if voter.rolePrefix != "CUSTOM_" {
		t.Errorf("expected rolePrefix 'CUSTOM_', got '%s'", voter.rolePrefix)
	}
}

func TestAuthenticatedVoter_Supports(t *testing.T) {
	t.Parallel()

	voter := NewAuthenticatedVoter()
	if !voter.Supports("IS_AUTHENTICATED_FULLY") {
		t.Error("expected AuthenticatedVoter to support IS_AUTHENTICATED_FULLY")
	}
	if !voter.Supports("IS_AUTHENTICATED_REMEMBERED") {
		t.Error("expected AuthenticatedVoter to support IS_AUTHENTICATED_REMEMBERED")
	}
	if !voter.Supports("IS_AUTHENTICATED_ANONYMOUSLY") {
		t.Error("expected AuthenticatedVoter to support IS_AUTHENTICATED_ANONYMOUSLY")
	}
	if !voter.Supports("UNKNOWN_ATTRIBUTE") {
		t.Error("expected AuthenticatedVoter to support any attribute")
	}
}

func TestAffirmativeBased_Decide_AllAbstain(t *testing.T) {
	t.Parallel()

	voter := &mockAccessDecisionVoter{result: ACCESS_ABSTAIN}
	manager := NewAffirmativeBased(voter)
	manager.SetAllowIfAllAbstainDecisions(false)

	auth := &mockAuthentication{authenticated: true, principal: "user"}
	ctx := context.Background()

	err := manager.Decide(ctx, auth, "GET:/api/users", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied when all abstain and allowIfAllAbstainDecisions=false, got %v", err)
	}
}

func TestAffirmativeBased_Decide_AllAbstain_Allow(t *testing.T) {
	t.Parallel()

	voter := &mockAccessDecisionVoter{result: ACCESS_ABSTAIN}
	manager := NewAffirmativeBased(voter)
	manager.SetAllowIfAllAbstainDecisions(true)

	auth := &mockAuthentication{authenticated: true, principal: "user"}
	ctx := context.Background()

	err := manager.Decide(ctx, auth, "GET:/api/users", []string{"read"})
	if err != nil {
		t.Errorf("expected no error when allowIfAllAbstainDecisions=true, got %v", err)
	}
}

func TestUnanimousBased_Decide_OneDeny(t *testing.T) {
	t.Parallel()

	voter1 := &mockAccessDecisionVoter{result: ACCESS_GRANTED}
	voter2 := &mockAccessDecisionVoter{result: ACCESS_DENIED}
	manager := NewUnanimousBased(voter1, voter2)

	auth := &mockAuthentication{authenticated: true, principal: "user"}
	ctx := context.Background()

	err := manager.Decide(ctx, auth, "GET:/api/users", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied when one voter denies, got %v", err)
	}
}

func TestUnanimousBased_Decide_AllAbstain(t *testing.T) {
	t.Parallel()

	voter := &mockAccessDecisionVoter{result: ACCESS_ABSTAIN}
	manager := NewUnanimousBased(voter)
	manager.SetAllowIfAllAbstainDecisions(false)

	auth := &mockAuthentication{authenticated: true, principal: "user"}
	ctx := context.Background()

	err := manager.Decide(ctx, auth, "GET:/api/users", []string{"read"})
	if err != ErrAccessDenied {
		t.Errorf("expected ErrAccessDenied when all abstain, got %v", err)
	}
}

func TestUnanimousBased_Decide_AllAbstain_Allow(t *testing.T) {
	t.Parallel()

	voter := &mockAccessDecisionVoter{result: ACCESS_ABSTAIN}
	manager := NewUnanimousBased(voter)
	manager.SetAllowIfAllAbstainDecisions(true)

	auth := &mockAuthentication{authenticated: true, principal: "user"}
	ctx := context.Background()

	err := manager.Decide(ctx, auth, "GET:/api/users", []string{"read"})
	if err != nil {
		t.Errorf("expected no error when allowIfAllAbstainDecisions=true, got %v", err)
	}
}
