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
		voteResult := voter.Vote(context.Background(), nil, "/test", []string{})
		if voteResult != ACCESS_ABSTAIN {
			t.Errorf("Expected ACCESS_ABSTAIN, got %d", voteResult)
		}
	})

	t.Run("abstain when no matching attribute", func(t *testing.T) {
		t.Parallel()
		voteResult := voter.Vote(context.Background(), nil, "/test", []string{"other"})
		if voteResult != ACCESS_ABSTAIN {
			t.Errorf("Expected ACCESS_ABSTAIN, got %d", voteResult)
		}
	})

	t.Run("grant when role matches", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authorities:   []string{"ROLE_ADMIN"},
		}
		voteResult := voter.Vote(context.Background(), auth, "/test", []string{"ROLE_ADMIN"})
		if voteResult != ACCESS_GRANTED {
			t.Errorf("Expected ACCESS_GRANTED, got %d", voteResult)
		}
	})

	t.Run("deny when role doesn't match", func(t *testing.T) {
		t.Parallel()
		auth := &mockAuthForVoter{
			principalName: "user",
			authorities:   []string{"ROLE_USER"},
		}
		voteResult := voter.Vote(context.Background(), auth, "/test", []string{"ROLE_ADMIN"})
		if voteResult != ACCESS_DENIED {
			t.Errorf("Expected ACCESS_DENIED, got %d", voteResult)
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
func testAuthenticatedVoterVoteCases() []struct {
	name       string
	auth       *mockAuthForVoter
	attributes []string
	want       int
} {
	return []struct {
		name       string
		auth       *mockAuthForVoter
		attributes []string
		want       int
	}{
		{"abstain when no attributes", nil, []string{}, ACCESS_ABSTAIN},
		{"grant IS_AUTHENTICATED_FULLY when authenticated", &mockAuthForVoter{principalName: "user", authenticated: true}, []string{"IS_AUTHENTICATED_FULLY"}, ACCESS_GRANTED},
		{"deny IS_AUTHENTICATED_FULLY when not authenticated", &mockAuthForVoter{principalName: "user", authenticated: false}, []string{"IS_AUTHENTICATED_FULLY"}, ACCESS_DENIED},
		{"grant IS_AUTHENTICATED_REMEMBERED when authenticated", &mockAuthForVoter{principalName: "user", authenticated: true}, []string{"IS_AUTHENTICATED_REMEMBERED"}, ACCESS_GRANTED},
		{"grant IS_AUTHENTICATED_ANONYMOUSLY always", nil, []string{"IS_AUTHENTICATED_ANONYMOUSLY"}, ACCESS_GRANTED},
	}
}

func TestAuthenticatedVoter_Vote_Coverage(t *testing.T) {
	t.Parallel()

	voter := &AuthenticatedVoter{}

	for _, tt := range testAuthenticatedVoterVoteCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			voteResult := voter.Vote(context.Background(), tt.auth, "/test", tt.attributes)
			if voteResult != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, voteResult)
			}
		})
	}
}

// TestConsensusBased_Decide_Coverage 测试 ConsensusBased.Decide
func testConsensusGrantWhenGrantGreater(t *testing.T) {
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
}

func testConsensusDenyWhenDenyGreater(t *testing.T) {
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
}

func testConsensusDenyWhenEqual(t *testing.T) {
	t.Parallel()
	mgr := NewConsensusBased(
		&mockAccessDecisionVoter{result: ACCESS_GRANTED},
		&mockAccessDecisionVoter{result: ACCESS_DENIED},
	)
	err := mgr.Decide(context.Background(), nil, "/test", []string{"test"})
	if err != ErrAccessDenied {
		t.Errorf("Expected ErrAccessDenied, got %v", err)
	}
}

func testConsensusAllowAllAbstain(t *testing.T) {
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
}

func TestConsensusBased_Decide_Coverage(t *testing.T) {
	t.Parallel()

	t.Run("grant when grant > deny", testConsensusGrantWhenGrantGreater)
	t.Run("deny when deny > grant", testConsensusDenyWhenDenyGreater)
	t.Run("deny when grant == deny", testConsensusDenyWhenEqual)
	t.Run("allow when all abstain and allowIfAllAbstainDecisions=true", testConsensusAllowAllAbstain)
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
func testWebExpressionVoterVoteCases() []struct {
	name       string
	auth       *mockAuthForVoter
	attributes []string
	want       int
} {
	return []struct {
		name       string
		auth       *mockAuthForVoter
		attributes []string
		want       int
	}{
		{"abstain when no attributes", nil, []string{}, ACCESS_ABSTAIN},
		{"permitAll grants", nil, []string{"permitAll"}, ACCESS_GRANTED},
		{"denyAll denies", nil, []string{"denyAll"}, ACCESS_DENIED},
		{"authenticated grants when authenticated", &mockAuthForVoter{principalName: "user", authenticated: true}, []string{"authenticated"}, ACCESS_GRANTED},
		{"authenticated denies when not authenticated", &mockAuthForVoter{principalName: "user", authenticated: false}, []string{"authenticated"}, ACCESS_DENIED},
		{"hasRole grants when role matches", &mockAuthForVoter{principalName: "user", authorities: []string{"ROLE_ADMIN"}}, []string{"hasRole('ROLE_ADMIN')"}, ACCESS_GRANTED},
		{"hasRole denies when role doesn't match", &mockAuthForVoter{principalName: "user", authorities: []string{"ROLE_USER"}}, []string{"hasRole('ROLE_ADMIN')"}, ACCESS_DENIED},
		{"hasAuthority grants when authority matches", &mockAuthForVoter{principalName: "user", authorities: []string{"WRITE"}}, []string{"hasAuthority('WRITE')"}, ACCESS_GRANTED},
		{"hasAuthority denies when authority doesn't match", &mockAuthForVoter{principalName: "user", authorities: []string{"READ"}}, []string{"hasAuthority('WRITE')"}, ACCESS_DENIED},
		{"unknown expression abstains", nil, []string{"unknownExpression"}, ACCESS_ABSTAIN},
	}
}

func TestWebExpressionVoter_Vote_Coverage(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	for _, tt := range testWebExpressionVoterVoteCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			voteResult := voter.Vote(context.Background(), tt.auth, "/test", tt.attributes)
			if voteResult != tt.want {
				t.Errorf("Expected %d, got %d", tt.want, voteResult)
			}
		})
	}
}
