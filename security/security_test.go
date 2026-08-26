package security

import (
	"context"
	"testing"

	"github.com/xudefa/enhance/security/authorization"
)

// mockAccessDecisionVoter 模拟投票者
type mockAccessDecisionVoter struct {
	result int
}

func (m *mockAccessDecisionVoter) Vote(ctx context.Context, authentication authorization.Authentication, resource string, attributes []string) int {
	return m.result
}

func (m *mockAccessDecisionVoter) Supports(attribute string) bool {
	return true
}

func TestAnonymousAuthenticationProvider_Authenticate(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	// Test with nil token
	result, err := provider.Authenticate(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Principal() != "anonymousUser" {
		t.Errorf("expected 'anonymousUser', got %v", result.Principal())
	}

	// Test with authenticated token
	token := NewUsernamePasswordAuthenticationToken("user", "pass")
	token.SetAuthenticated(true)
	result, err = provider.Authenticate(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for authenticated token")
	}

	// Test with token that has credentials
	token2 := NewUsernamePasswordAuthenticationToken("user", "pass")
	result, err = provider.Authenticate(context.Background(), token2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for token with credentials")
	}
}

func TestAnonymousAuthenticationProvider_Supports(t *testing.T) {
	t.Parallel()

	provider := NewAnonymousAuthenticationProvider()

	// Test with UsernamePasswordAuthenticationToken
	token := NewUsernamePasswordAuthenticationToken("user", "pass")
	if !provider.Supports(token) {
		t.Error("expected to support UsernamePasswordAuthenticationToken")
	}

	// Test with nil
	if provider.Supports(nil) {
		t.Error("expected not to support nil")
	}
}

func TestHasAnyAuthority(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	// Create a mock authentication
	token := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER", "ROLE_ADMIN"})

	// Test with matching authority
	if !voter.hasAnyAuthority(token, []string{"ROLE_USER", "ROLE_OTHER"}) {
		t.Error("expected to have ROLE_USER authority")
	}

	// Test with no matching authority
	if voter.hasAnyAuthority(token, []string{"ROLE_OTHER", "ROLE_ANOTHER"}) {
		t.Error("expected not to have any of the authorities")
	}

	// Test with nil authentication
	if voter.hasAnyAuthority(nil, []string{"ROLE_USER"}) {
		t.Error("expected false for nil authentication")
	}
}

func TestRoleVoter_Vote_WithCustomPrefix(t *testing.T) {
	t.Parallel()

	voter := NewRoleVoter()
	voter.SetRolePrefix("CUSTOM_")

	token := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"CUSTOM_ADMIN"})

	// Test with matching role
	result := voter.Vote(context.Background(), token, "/admin", []string{"CUSTOM_ADMIN"})
	if result != ACCESS_GRANTED {
		t.Errorf("expected ACCESS_GRANTED, got %d", result)
	}

	// Test with non-matching role
	result = voter.Vote(context.Background(), token, "/admin", []string{"CUSTOM_USER"})
	if result != ACCESS_DENIED {
		t.Errorf("expected ACCESS_DENIED, got %d", result)
	}
}

func TestWebExpressionVoter_HasRole(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	token := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER", "ROLE_ADMIN"})

	// Test with matching role
	if !voter.hasRole(token, "USER") {
		t.Error("expected to have ROLE_USER")
	}

	// Test with non-matching role
	if voter.hasRole(token, "GUEST") {
		t.Error("expected not to have ROLE_GUEST")
	}

	// Test with nil authentication
	if voter.hasRole(nil, "USER") {
		t.Error("expected false for nil authentication")
	}
}

func TestWebExpressionVoter_HasAnyRole(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	token := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"ROLE_USER"})

	// Test with matching role
	if !voter.hasAnyRole(token, []string{"USER", "ADMIN"}) {
		t.Error("expected to have one of the roles")
	}

	// Test with no matching roles
	if voter.hasAnyRole(token, []string{"GUEST", "ANONYMOUS"}) {
		t.Error("expected not to have any of the roles")
	}

	// Test with nil authentication
	if voter.hasAnyRole(nil, []string{"USER"}) {
		t.Error("expected false for nil authentication")
	}
}

func TestWebExpressionVoter_HasAuthority(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	token := NewAuthenticatedUsernamePasswordAuthenticationToken("user", []string{"READ", "WRITE"})

	// Test with matching authority
	if !voter.hasAuthority(token, "READ") {
		t.Error("expected to have READ authority")
	}

	// Test with non-matching authority
	if voter.hasAuthority(token, "DELETE") {
		t.Error("expected not to have DELETE authority")
	}

	// Test with nil authentication
	if voter.hasAuthority(nil, "READ") {
		t.Error("expected false for nil authentication")
	}
}

func TestAffirmativeBased_AllAbstain(t *testing.T) {
	t.Parallel()

	voter := &mockAccessDecisionVoter{result: ACCESS_ABSTAIN}
	manager := NewAffirmativeBased(voter)
	manager.SetAllowIfAllAbstainDecisions(false)

	err := manager.Decide(context.Background(), nil, "/test", []string{"READ"})
	if err == nil {
		t.Error("expected error when all abstain and allowIfAllAbstainDecisions is false")
	}
}

func TestAffirmativeBased_AllAbstainAllow(t *testing.T) {
	t.Parallel()

	voter := &mockAccessDecisionVoter{result: ACCESS_ABSTAIN}
	manager := NewAffirmativeBased(voter)
	manager.SetAllowIfAllAbstainDecisions(true)

	err := manager.Decide(context.Background(), nil, "/test", []string{"READ"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConsensusBased_Supports(t *testing.T) {
	t.Parallel()

	manager := NewConsensusBased()
	if !manager.Supports("READ") {
		t.Error("expected to support READ")
	}
}

func TestConsensusBased_AddVoter(t *testing.T) {
	t.Parallel()

	manager := NewConsensusBased()
	voter := &mockAccessDecisionVoter{result: ACCESS_GRANTED}
	manager.AddVoter(voter)

	if len(manager.decisionVoters) != 1 {
		t.Errorf("expected 1 voter, got %d", len(manager.decisionVoters))
	}
}

func TestConsensusBased_SetAllowIfEqualGrantedDenied(t *testing.T) {
	t.Parallel()

	manager := NewConsensusBased()
	manager.SetAllowIfEqualGrantedDenied(true)

	if !manager.allowIfEqualGrantedDenied {
		t.Error("expected allowIfEqualGrantedDenied to be true")
	}
}

func TestConsensusBased_SetAllowIfAllAbstainDecisions(t *testing.T) {
	t.Parallel()

	manager := NewConsensusBased()
	manager.SetAllowIfAllAbstainDecisions(true)

	if !manager.allowIfAllAbstainDecisions {
		t.Error("expected allowIfAllAbstainDecisions to be true")
	}
}
