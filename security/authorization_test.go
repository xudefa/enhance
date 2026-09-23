package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/security/authorization"
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

// ==================== Casbin Voter Tests ====================

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

	voteResult := voter.Vote(ctx, auth, "GET:/api/users", []string{"read"})
	if voteResult != ACCESS_GRANTED {
		t.Errorf("expected ACCESS_GRANTED, got %d", voteResult)
	}
}

func TestCasbinVoter_Vote_Denied(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{allowed: false}
	voter := MustNewCasbinVoter(enforcer)

	auth := &mockAuthentication{authenticated: true, principal: "user"}
	ctx := context.Background()

	voteResult := voter.Vote(ctx, auth, "DELETE:/api/users/1", []string{"delete"})
	if voteResult != ACCESS_DENIED {
		t.Errorf("expected ACCESS_DENIED, got %d", voteResult)
	}
}

func TestCasbinVoter_Vote_Abstain(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{}
	voter := MustNewCasbinVoter(enforcer)

	auth := &mockAuthentication{authenticated: false}
	ctx := context.Background()

	voteResult := voter.Vote(ctx, auth, "GET:/api/users", []string{"read"})
	if voteResult != ACCESS_ABSTAIN {
		t.Errorf("expected ACCESS_ABSTAIN, got %d", voteResult)
	}
}

func TestCasbinVoter_Vote_NilAuthentication(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{}
	voter := MustNewCasbinVoter(enforcer)

	ctx := context.Background()
	voteResult := voter.Vote(ctx, nil, "GET:/api/users", []string{"read"})
	if voteResult != ACCESS_ABSTAIN {
		t.Errorf("expected ACCESS_ABSTAIN for nil authentication, got %d", voteResult)
	}
}

func TestCasbinVoter_Vote_EnforcerError(t *testing.T) {
	t.Parallel()

	enforcer := &mockCasbinEnforcer{err: true}
	voter := MustNewCasbinVoter(enforcer)

	auth := &mockAuthentication{authenticated: true, principal: "admin"}
	ctx := context.Background()

	voteResult := voter.Vote(ctx, auth, "GET:/api/users", []string{"read"})
	if voteResult != ACCESS_DENIED {
		t.Errorf("expected ACCESS_DENIED on enforcer error, got %d", voteResult)
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

// ==================== Authorization Coverage Extra Tests ====================

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

func TestRoleVoter_SetRolePrefix_Coverage(t *testing.T) {
	t.Parallel()

	voter := NewRoleVoter()
	voter.SetRolePrefix("CUSTOM_")

	if voter.rolePrefix != "CUSTOM_" {
		t.Errorf("Expected rolePrefix to be CUSTOM_, got %s", voter.rolePrefix)
	}
}

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

func TestAuthenticatedVoter_Vote_Coverage(t *testing.T) {
	t.Parallel()

	voter := &AuthenticatedVoter{}

	tests := []struct {
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

	for _, tt := range tests {
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

func TestAffirmativeBased_AddVoter_Coverage(t *testing.T) {
	t.Parallel()

	mgr := NewAffirmativeBased()
	mgr.AddVoter(&mockAccessDecisionVoter{result: ACCESS_GRANTED})

	if len(mgr.decisionVoters) != 1 {
		t.Errorf("Expected 1 voter, got %d", len(mgr.decisionVoters))
	}
}

func TestUnanimousBased_AddVoter_Coverage(t *testing.T) {
	t.Parallel()

	mgr := NewUnanimousBased()
	mgr.AddVoter(&mockAccessDecisionVoter{result: ACCESS_GRANTED})

	if len(mgr.decisionVoters) != 1 {
		t.Errorf("Expected 1 voter, got %d", len(mgr.decisionVoters))
	}
}

func TestWebExpressionVoter_Vote_Coverage(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	tests := []struct {
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

	for _, tt := range tests {
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
