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
	m := NewAffirmativeBased(voter1)
	m.(*affirmativeBased).SetAllowIfAllAbstainDecisions(true)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := m.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
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
	m := NewAffirmativeBased(voter1)

	newVoter := &mockVoter{voteResult: AccessGranted, supported: true}
	m.(*affirmativeBased).AddVoter(newVoter)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := m.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
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
	m := NewUnanimousBased(voter1)
	m.(*unanimousBased).SetAllowIfAllAbstainDecisions(true)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := m.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
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
	m := NewConsensusBased(voter1, voter2)
	m.(*consensusBased).SetAllowIfEqualGrantedDenied(true)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := m.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error when allowIfEqualGrantedDenied is true, got %v", err)
	}
}

func TestConsensusBased_AllAbstainAllowIfConfigured(t *testing.T) {
	t.Parallel()

	voter1 := &mockVoter{voteResult: AccessAbstain, supported: true}
	m := NewConsensusBased(voter1)
	m.(*consensusBased).SetAllowIfAllAbstainDecisions(true)

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	err := m.Decide(context.Background(), auth, "/api/users", []string{"hasRole('USER')"})
	if err != nil {
		t.Errorf("expected no error when allowIfAllAbstainDecisions is true, got %v", err)
	}
}

func TestWebExpressionVoter_PermitAll(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	result := voter.Vote(context.Background(), auth, "/api/public", []string{"permitAll"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted for permitAll, got %d", result)
	}
}

func TestWebExpressionVoter_DenyAll(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	result := voter.Vote(context.Background(), auth, "/api/denied", []string{"denyAll"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied for denyAll, got %d", result)
	}
}

func TestWebExpressionVoter_Authenticated(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}
	result := voter.Vote(context.Background(), auth, "/api/users", []string{"authenticated"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted for authenticated user, got %d", result)
	}

	result = voter.Vote(context.Background(), nil, "/api/users", []string{"authenticated"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied for nil authentication, got %d", result)
	}
}

func TestWebExpressionVoter_HasRole(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_ADMIN", "ROLE_USER"},
		authenticated: true,
	}

	result := voter.Vote(context.Background(), auth, "/api/admin", []string{"hasRole('ADMIN')"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted for hasRole('ADMIN'), got %d", result)
	}

	result = voter.Vote(context.Background(), auth, "/api/admin", []string{"hasRole('GUEST')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied for hasRole('GUEST'), got %d", result)
	}
}

func TestWebExpressionVoter_HasAnyRole(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	result := voter.Vote(context.Background(), auth, "/api/users", []string{"hasAnyRole('ADMIN','USER')"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted for hasAnyRole('ADMIN','USER'), got %d", result)
	}

	result = voter.Vote(context.Background(), auth, "/api/admin", []string{"hasAnyRole('ADMIN','GUEST')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied for hasAnyRole('ADMIN','GUEST'), got %d", result)
	}
}

func TestWebExpressionVoter_HasAuthority(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"read", "write"},
		authenticated: true,
	}

	result := voter.Vote(context.Background(), auth, "/api/data", []string{"hasAuthority('read')"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted for hasAuthority('read'), got %d", result)
	}

	result = voter.Vote(context.Background(), auth, "/api/data", []string{"hasAuthority('delete')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied for hasAuthority('delete'), got %d", result)
	}
}

func TestWebExpressionVoter_HasAnyAuthority(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"read"},
		authenticated: true,
	}

	result := voter.Vote(context.Background(), auth, "/api/data", []string{"hasAnyAuthority('read','write')"})
	if result != AccessGranted {
		t.Errorf("expected AccessGranted for hasAnyAuthority with matching, got %d", result)
	}

	result = voter.Vote(context.Background(), auth, "/api/data", []string{"hasAnyAuthority('write','delete')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied for hasAnyAuthority without matching, got %d", result)
	}
}

func TestWebExpressionVoter_EmptyAttributes(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	result := voter.Vote(context.Background(), auth, "/api/users", []string{})
	if result != AccessAbstain {
		t.Errorf("expected AccessAbstain for empty attributes, got %d", result)
	}
}

func TestWebExpressionVoter_NilAuthentication(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	result := voter.Vote(context.Background(), nil, "/api/users", []string{"hasRole('ADMIN')"})
	if result != AccessDenied {
		t.Errorf("expected AccessDenied for nil authentication, got %d", result)
	}
}

func TestWebExpressionVoter_UnsupportedExpression(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	result := voter.Vote(context.Background(), auth, "/api/users", []string{"unknownExpression"})
	if result != AccessAbstain {
		t.Errorf("expected AccessAbstain for unknown expression, got %d", result)
	}
}

func TestWebExpressionVoter_Supports(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	if !voter.Supports("anything") {
		t.Error("expected Supports to return true for any attribute")
	}
}

func TestExpressionBasedUrlRegistry_PermitAll(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/public/**").PermitAll()

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if len(rules[0].Patterns) != 1 || rules[0].Patterns[0] != "/api/public/**" {
		t.Errorf("expected pattern /api/public/**, got %v", rules[0].Patterns)
	}
	if len(rules[0].Attributes) != 1 || rules[0].Attributes[0] != "permitAll" {
		t.Errorf("expected attribute permitAll, got %v", rules[0].Attributes)
	}
}

func TestExpressionBasedUrlRegistry_HasRole(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/admin/**").HasRole("ADMIN")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasRole('ADMIN')" {
		t.Errorf("expected hasRole('ADMIN'), got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyRole(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/manager/**").HasAnyRole("ADMIN", "MANAGER")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasAnyRole('ADMIN','MANAGER')" {
		t.Errorf("expected hasAnyRole('ADMIN','MANAGER'), got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyAuthority_Single(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/data/**").HasAnyAuthority("read")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasAuthority('read')" {
		t.Errorf("expected hasAuthority('read'), got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_HasAnyAuthority(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/data/**").HasAnyAuthority("read", "write")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasAnyAuthority('read','write')" {
		t.Errorf("expected hasAnyAuthority('read','write'), got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_DenyAll(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/internal/**").DenyAll()

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "denyAll" {
		t.Errorf("expected denyAll, got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_Authenticated(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/secure/**").Authenticated()

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "authenticated" {
		t.Errorf("expected authenticated, got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_MultipleRules(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/public/**").PermitAll()
	authz.RequestMatchers("/api/admin/**").HasRole("ADMIN")
	authz.RequestMatchers("**").Authenticated()

	rules := getRegistryRules(authz)
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "permitAll" {
		t.Errorf("expected first rule permitAll, got %v", rules[0].Attributes[0])
	}
	if rules[1].Attributes[0] != "hasRole('ADMIN')" {
		t.Errorf("expected second rule hasRole('ADMIN'), got %v", rules[1].Attributes[0])
	}
	if rules[2].Attributes[0] != "authenticated" {
		t.Errorf("expected third rule authenticated, got %v", rules[2].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_AnyRequest(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.AnyRequest().Authenticated()

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if len(rules[0].Patterns) != 1 || rules[0].Patterns[0] != "**" {
		t.Errorf("expected pattern **, got %v", rules[0].Patterns)
	}
}

func TestExpressionBasedUrlRegistry_EmptyGet(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	rules := getRegistryRules(authz)
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for empty registry, got %d", len(rules))
	}
}

func TestExpressionBasedUrlRegistry_SingleAuthority(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/data/**").HasAnyAuthority("read")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasAuthority('read')" {
		t.Errorf("expected hasAuthority('read') for single authority, got %v", rules[0].Attributes[0])
	}
}

func TestExpressionBasedUrlRegistry_SingleRole(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/admin/**").HasAnyRole("ADMIN")

	rules := getRegistryRules(authz)
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Attributes[0] != "hasRole('ADMIN')" {
		t.Errorf("expected hasRole('ADMIN') for single role, got %v", rules[0].Attributes[0])
	}
}

func TestAuthorizeRequests(t *testing.T) {
	t.Parallel()

	authz := NewAuthorizeRequests()
	authz.RequestMatchers("/api/public/**").PermitAll()
	authz.RequestMatchers("/api/admin/**").HasRole("ADMIN")

	if authz == nil {
		t.Error("expected non-nil AuthorizeRequests")
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

func TestExtractExpressionArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attr     string
		prefix   string
		suffix   string
		expected string
	}{
		{
			name:     "hasRole",
			attr:     "hasRole('ADMIN')",
			prefix:   "hasRole('",
			suffix:   "')",
			expected: "ADMIN",
		},
		{
			name:     "hasAuthority",
			attr:     "hasAuthority('read')",
			prefix:   "hasAuthority('",
			suffix:   "')",
			expected: "read",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractExpressionArg(tt.attr, tt.prefix, tt.suffix)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestSplitExpressionArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attr     string
		prefix   string
		suffix   string
		expected []string
	}{
		{
			name:     "two roles",
			attr:     "hasAnyRole('ADMIN','USER')",
			prefix:   "hasAnyRole('",
			suffix:   "')",
			expected: []string{"ADMIN", "USER"},
		},
		{
			name:     "three authorities",
			attr:     "hasAnyAuthority('read','write','delete')",
			prefix:   "hasAnyAuthority('",
			suffix:   "')",
			expected: []string{"read", "write", "delete"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := splitExpressionArgs(tt.attr, tt.prefix, tt.suffix)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d args, got %d", len(tt.expected), len(result))
			}
			for i, arg := range result {
				if arg != tt.expected[i] {
					t.Errorf("expected arg[%d] = %q, got %q", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestJoinStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{
			name:     "empty",
			strs:     []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "single",
			strs:     []string{"a"},
			sep:      ",",
			expected: "a",
		},
		{
			name:     "multiple",
			strs:     []string{"a", "b", "c"},
			sep:      ",",
			expected: "a,b,c",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := joinStrings(tt.strs, tt.sep)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// getRegistryRules 通过类型断言从 AuthorizeRequests 获取规则。
func getRegistryRules(authz AuthorizeRequests) []UrlAuthorizationRule {
	ar := authz.(*authorizeRequests)
	return ar.registry.Get()
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

func BenchmarkWebExpressionVoter_Vote(b *testing.B) {
	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "admin",
		authorities:   []string{"ROLE_ADMIN"},
		authenticated: true,
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		voter.Vote(ctx, auth, "/api/admin", []string{"hasRole('ADMIN')"})
	}
}
