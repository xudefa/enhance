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

// mockRoleAuth 用于 RoleVoter 和 AuthenticatedVoter 测试的模拟认证
type mockRoleAuth struct {
	principalName string
	authenticated bool
	authorities   []string
}

func (m *mockRoleAuth) Principal() any          { return m.principalName }
func (m *mockRoleAuth) Credentials() any        { return nil }
func (m *mockRoleAuth) Authorities() []string   { return m.authorities }
func (m *mockRoleAuth) Authenticated() bool     { return m.authenticated }
func (m *mockRoleAuth) Name() string            { return m.principalName }
func (m *mockRoleAuth) SetAuthenticated(bool)   {}
func (m *mockRoleAuth) SetAuthorities([]string) {}

// TestRoleVoter_Vote 测试 RoleVoter.Vote 方法
func TestRoleVoter_Vote(t *testing.T) {
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
		auth := &mockRoleAuth{
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
		auth := &mockRoleAuth{
			principalName: "user",
			authorities:   []string{"ROLE_USER"},
		}
		voteResult := voter.Vote(context.Background(), auth, "/test", []string{"ROLE_ADMIN"})
		if voteResult != ACCESS_DENIED {
			t.Errorf("Expected ACCESS_DENIED, got %d", voteResult)
		}
	})
}

// TestAuthenticatedVoter_Vote 测试 AuthenticatedVoter.Vote 方法
func testAuthVoterVoteAbstainNoAttributes(t *testing.T) {
	t.Parallel()
	voter := &AuthenticatedVoter{}
	voteResult := voter.Vote(context.Background(), nil, "/test", []string{})
	if voteResult != ACCESS_ABSTAIN {
		t.Errorf("Expected ACCESS_ABSTAIN, got %d", voteResult)
	}
}

func testAuthVoterVoteGrantFullyAuthenticated(t *testing.T) {
	t.Parallel()
	voter := &AuthenticatedVoter{}
	auth := &mockRoleAuth{
		principalName: "user",
		authenticated: true,
	}
	voteResult := voter.Vote(context.Background(), auth, "/test", []string{"IS_AUTHENTICATED_FULLY"})
	if voteResult != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED, got %d", voteResult)
	}
}

func testAuthVoterVoteDenyFullyUnauthenticated(t *testing.T) {
	t.Parallel()
	voter := &AuthenticatedVoter{}
	auth := &mockRoleAuth{
		principalName: "user",
		authenticated: false,
	}
	voteResult := voter.Vote(context.Background(), auth, "/test", []string{"IS_AUTHENTICATED_FULLY"})
	if voteResult != ACCESS_DENIED {
		t.Errorf("Expected ACCESS_DENIED, got %d", voteResult)
	}
}

func testAuthVoterVoteGrantRememberedAuthenticated(t *testing.T) {
	t.Parallel()
	voter := &AuthenticatedVoter{}
	auth := &mockRoleAuth{
		principalName: "user",
		authenticated: true,
	}
	voteResult := voter.Vote(context.Background(), auth, "/test", []string{"IS_AUTHENTICATED_REMEMBERED"})
	if voteResult != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED, got %d", voteResult)
	}
}

func testAuthVoterVoteGrantAnonymous(t *testing.T) {
	t.Parallel()
	voter := &AuthenticatedVoter{}
	voteResult := voter.Vote(context.Background(), nil, "/test", []string{"IS_AUTHENTICATED_ANONYMOUSLY"})
	if voteResult != ACCESS_GRANTED {
		t.Errorf("Expected ACCESS_GRANTED, got %d", voteResult)
	}
}

func TestAuthenticatedVoter_Vote(t *testing.T) {
	t.Parallel()

	t.Run("abstain when no attributes", testAuthVoterVoteAbstainNoAttributes)
	t.Run("grant IS_AUTHENTICATED_FULLY when authenticated", testAuthVoterVoteGrantFullyAuthenticated)
	t.Run("deny IS_AUTHENTICATED_FULLY when not authenticated", testAuthVoterVoteDenyFullyUnauthenticated)
	t.Run("grant IS_AUTHENTICATED_REMEMBERED when authenticated", testAuthVoterVoteGrantRememberedAuthenticated)
	t.Run("grant IS_AUTHENTICATED_ANONYMOUSLY always", testAuthVoterVoteGrantAnonymous)
}
