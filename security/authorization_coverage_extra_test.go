package security

import (
	"context"
	"testing"
)

// mockAuthForVoter 模拟认证对象
type mockAuthForVoter struct {
	principalName string
	authenticated bool
	authorities   []string
}

func (m *mockAuthForVoter) Principal() any          { return m.principalName }
func (m *mockAuthForVoter) Credentials() any        { return nil }
func (m *mockAuthForVoter) Authorities() []string   { return m.authorities }
func (m *mockAuthForVoter) Authenticated() bool     { return m.authenticated }
func (m *mockAuthForVoter) Name() string            { return m.principalName }
func (m *mockAuthForVoter) SetAuthenticated(bool)   {}
func (m *mockAuthForVoter) SetAuthorities([]string) {}

// TestRoleVoter_Vote_Coverage 测试 RoleVoter.Vote
func TestRoleVoter_Vote_Coverage(t *testing.T) {
	t.Parallel()

	voter := NewRoleVoter()

	t.Run("abstain when no attributes", func(t *testing.T) {
		t.Parallel()
		result := voter.Vote(context.Background(), nil, "/test", []string{})
		if result != ACCESS_ABSTAIN {
			t.Errorf("Expected ACCESS_ABSTAIN, got %d", result)
		}
	})

	t.Run("abstain when no matching attribute", func(t *testing.T) {
		t.Parallel()
		result := voter.Vote(context.Background(), nil, "/test", []string{"other"})
		if result != ACCESS_ABSTAIN {
			t.Errorf("Expected ACCESS_ABSTAIN, got %d", result)
		}
	})

	t.Run("grant when role matches", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authorities:   []string{"ROLE_ADMIN"},
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"ROLE_ADMIN"})
		if result != ACCESS_GRANTED {
			t.Errorf("Expected ACCESS_GRANTED, got %d", result)
		}
	})

	t.Run("deny when role doesn't match", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authorities:   []string{"ROLE_USER"},
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"ROLE_ADMIN"})
		if result != ACCESS_DENIED {
			t.Errorf("Expected ACCESS_DENIED, got %d", result)
		}
	})
}

// TestRoleVoter_Supports_Coverage 测试 RoleVoter.Supports
func TestRoleVoter_Supports_Coverage(t *testing.T) {
	t.Parallel()

	voter := NewRoleVoter()

	if !voter.Supports("ROLE_ADMIN") {
		t.Error("Expected Supports to return true for ROLE_ prefix")
	}

	if voter.Supports("OTHER_ADMIN") {
		t.Error("Expected Supports to return false for non-ROLE_ prefix")
	}
}

// TestRoleVoter_SetRolePrefix_Coverage 测试 SetRolePrefix
func TestRoleVoter_SetRolePrefix_Coverage(t *testing.T) {
	t.Parallel()

	voter := NewRoleVoter()
	voter.SetRolePrefix("CUSTOM_")

	if voter.rolePrefix != "CUSTOM_" {
		t.Errorf("Expected rolePrefix to be CUSTOM_, got %s", voter.rolePrefix)
	}
}

// TestAuthenticatedVoter_Vote_Coverage 测试 AuthenticatedVoter.Vote
func TestAuthenticatedVoter_Vote_Coverage(t *testing.T) {
	t.Parallel()

	voter := &AuthenticatedVoter{}

	t.Run("abstain when no attributes", func(t *testing.T) {
		t.Parallel()
		result := voter.Vote(context.Background(), nil, "/test", []string{})
		if result != ACCESS_ABSTAIN {
			t.Errorf("Expected ACCESS_ABSTAIN, got %d", result)
		}
	})

	t.Run("grant IS_AUTHENTICATED_FULLY when authenticated", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authenticated: true,
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"IS_AUTHENTICATED_FULLY"})
		if result != ACCESS_GRANTED {
			t.Errorf("Expected ACCESS_GRANTED, got %d", result)
		}
	})

	t.Run("deny IS_AUTHENTICATED_FULLY when not authenticated", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authenticated: false,
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"IS_AUTHENTICATED_FULLY"})
		if result != ACCESS_DENIED {
			t.Errorf("Expected ACCESS_DENIED, got %d", result)
		}
	})

	t.Run("grant IS_AUTHENTICATED_REMEMBERED when authenticated", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authenticated: true,
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"IS_AUTHENTICATED_REMEMBERED"})
		if result != ACCESS_GRANTED {
			t.Errorf("Expected ACCESS_GRANTED, got %d", result)
		}
	})

	t.Run("grant IS_AUTHENTICATED_ANONYMOUSLY always", func(t *testing.T) {
		t.Parallel()
		result := voter.Vote(context.Background(), nil, "/test", []string{"IS_AUTHENTICATED_ANONYMOUSLY"})
		if result != ACCESS_GRANTED {
			t.Errorf("Expected ACCESS_GRANTED, got %d", result)
		}
	})
}

// TestConsensusBased_Decide_Coverage 测试 ConsensusBased.Decide
func TestConsensusBased_Decide_Coverage(t *testing.T) {
	t.Parallel()

	t.Run("grant when grant > deny", func(t *testing.T) {
		t.Parallel()
		mgr := NewConsensusBased(
			&mockAccessDecisionVoter{result: ACCESS_GRANTED},
			&mockAccessDecisionVoter{result: ACCESS_DENIED},
			&mockAccessDecisionVoter{result: ACCESS_GRANTED},
		)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("deny when deny > grant", func(t *testing.T) {
		t.Parallel()
		mgr := NewConsensusBased(
			&mockAccessDecisionVoter{result: ACCESS_DENIED},
			&mockAccessDecisionVoter{result: ACCESS_DENIED},
			&mockAccessDecisionVoter{result: ACCESS_GRANTED},
		)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != ErrAccessDenied {
			t.Errorf("Expected ErrAccessDenied, got %v", err)
		}
	})

	t.Run("deny when grant == deny", func(t *testing.T) {
		t.Parallel()
		mgr := NewConsensusBased(
			&mockAccessDecisionVoter{result: ACCESS_GRANTED},
			&mockAccessDecisionVoter{result: ACCESS_DENIED},
		)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != ErrAccessDenied {
			t.Errorf("Expected ErrAccessDenied, got %v", err)
		}
	})

	t.Run("allow when all abstain and allowIfAllAbstainDecisions=true", func(t *testing.T) {
		t.Parallel()
		mgr := NewConsensusBased(
			&mockAccessDecisionVoter{result: ACCESS_ABSTAIN},
		)
		mgr.SetAllowIfEqualGrantedDenied(true)
		mgr.SetAllowIfAllAbstainDecisions(true)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// TestAffirmativeBased_Decide_Coverage 测试 AffirmativeBased.Decide
func TestAffirmativeBased_Decide_Coverage(t *testing.T) {
	t.Parallel()

	t.Run("grant when any voter grants", func(t *testing.T) {
		t.Parallel()
		mgr := NewAffirmativeBased(
			&mockAccessDecisionVoter{result: ACCESS_ABSTAIN},
			&mockAccessDecisionVoter{result: ACCESS_GRANTED},
			&mockAccessDecisionVoter{result: ACCESS_DENIED},
		)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("deny when all deny", func(t *testing.T) {
		t.Parallel()
		mgr := NewAffirmativeBased(
			&mockAccessDecisionVoter{result: ACCESS_DENIED},
			&mockAccessDecisionVoter{result: ACCESS_DENIED},
		)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != ErrAccessDenied {
			t.Errorf("Expected ErrAccessDenied, got %v", err)
		}
	})

	t.Run("allow when all abstain and allowIfAllAbstainDecisions=true", func(t *testing.T) {
		t.Parallel()
		mgr := NewAffirmativeBased(
			&mockAccessDecisionVoter{result: ACCESS_ABSTAIN},
		)
		mgr.SetAllowIfAllAbstainDecisions(true)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("allow when all abstain and abstain==0", func(t *testing.T) {
		t.Parallel()
		mgr := NewAffirmativeBased()
		err := mgr.Decide(context.Background(), nil, "/test", []string{})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// TestUnanimousBased_Decide_Coverage 测试 UnanimousBased.Decide
func TestUnanimousBased_Decide_Coverage(t *testing.T) {
	t.Parallel()

	t.Run("grant when all grant or abstain", func(t *testing.T) {
		t.Parallel()
		mgr := NewUnanimousBased(
			&mockAccessDecisionVoter{result: ACCESS_GRANTED},
			&mockAccessDecisionVoter{result: ACCESS_ABSTAIN},
		)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("deny when any deny", func(t *testing.T) {
		t.Parallel()
		mgr := NewUnanimousBased(
			&mockAccessDecisionVoter{result: ACCESS_GRANTED},
			&mockAccessDecisionVoter{result: ACCESS_DENIED},
		)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != ErrAccessDenied {
			t.Errorf("Expected ErrAccessDenied, got %v", err)
		}
	})

	t.Run("allow when all abstain and allowIfAllAbstainDecisions=true", func(t *testing.T) {
		t.Parallel()
		mgr := NewUnanimousBased(
			&mockAccessDecisionVoter{result: ACCESS_ABSTAIN},
		)
		mgr.SetAllowIfAllAbstainDecisions(true)
		err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// TestAffirmativeBased_AddVoter_Coverage 测试 AddVoter
func TestAffirmativeBased_AddVoter_Coverage(t *testing.T) {
	t.Parallel()

	mgr := NewAffirmativeBased()
	mgr.AddVoter(&mockAccessDecisionVoter{result: ACCESS_GRANTED})

	if len(mgr.decisionVoters) != 1 {
		t.Errorf("Expected 1 voter, got %d", len(mgr.decisionVoters))
	}
}

// TestUnanimousBased_AddVoter_Coverage 测试 AddVoter
func TestUnanimousBased_AddVoter_Coverage(t *testing.T) {
	t.Parallel()

	mgr := NewUnanimousBased()
	mgr.AddVoter(&mockAccessDecisionVoter{result: ACCESS_GRANTED})

	if len(mgr.decisionVoters) != 1 {
		t.Errorf("Expected 1 voter, got %d", len(mgr.decisionVoters))
	}
}

// TestWebExpressionVoter_Vote_Coverage 测试 WebExpressionVoter.Vote
func TestWebExpressionVoter_Vote_Coverage(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	t.Run("abstain when no attributes", func(t *testing.T) {
		t.Parallel()
		result := voter.Vote(context.Background(), nil, "/test", []string{})
		if result != ACCESS_ABSTAIN {
			t.Errorf("Expected ACCESS_ABSTAIN, got %d", result)
		}
	})

	t.Run("permitAll grants", func(t *testing.T) {
		t.Parallel()
		result := voter.Vote(context.Background(), nil, "/test", []string{"permitAll"})
		if result != ACCESS_GRANTED {
			t.Errorf("Expected ACCESS_GRANTED, got %d", result)
		}
	})

	t.Run("denyAll denies", func(t *testing.T) {
		t.Parallel()
		result := voter.Vote(context.Background(), nil, "/test", []string{"denyAll"})
		if result != ACCESS_DENIED {
			t.Errorf("Expected ACCESS_DENIED, got %d", result)
		}
	})

	t.Run("authenticated grants when authenticated", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authenticated: true,
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"authenticated"})
		if result != ACCESS_GRANTED {
			t.Errorf("Expected ACCESS_GRANTED, got %d", result)
		}
	})

	t.Run("authenticated denies when not authenticated", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authenticated: false,
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"authenticated"})
		if result != ACCESS_DENIED {
			t.Errorf("Expected ACCESS_DENIED, got %d", result)
		}
	})

	t.Run("hasRole grants when role matches", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authorities:   []string{"ROLE_ADMIN"},
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"hasRole('ROLE_ADMIN')"})
		if result != ACCESS_GRANTED {
			t.Errorf("Expected ACCESS_GRANTED, got %d", result)
		}
	})

	t.Run("hasRole denies when role doesn't match", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authorities:   []string{"ROLE_USER"},
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"hasRole('ROLE_ADMIN')"})
		if result != ACCESS_DENIED {
			t.Errorf("Expected ACCESS_DENIED, got %d", result)
		}
	})

	t.Run("hasAuthority grants when authority matches", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authorities:   []string{"WRITE"},
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"hasAuthority('WRITE')"})
		if result != ACCESS_GRANTED {
			t.Errorf("Expected ACCESS_GRANTED, got %d", result)
		}
	})

	t.Run("hasAuthority denies when authority doesn't match", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authorities:   []string{"READ"},
		}
		result := voter.Vote(context.Background(), auth, "/test", []string{"hasAuthority('WRITE')"})
		if result != ACCESS_DENIED {
			t.Errorf("Expected ACCESS_DENIED, got %d", result)
		}
	})

	t.Run("unknown expression abstains", func(t *testing.T) {
		t.Parallel()
		result := voter.Vote(context.Background(), nil, "/test", []string{"unknownExpression"})
		if result != ACCESS_ABSTAIN {
			t.Errorf("Expected ACCESS_ABSTAIN, got %d", result)
		}
	})
}
