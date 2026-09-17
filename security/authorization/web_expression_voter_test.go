package authorization

import (
	"context"
	"testing"
)

func TestWebExpressionVoter_PermitAll(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	auth := &mockAuthentication{
		principal:     "user",
		authorities:   []string{"ROLE_USER"},
		authenticated: true,
	}

	voteResult := voter.Vote(context.Background(), auth, "/api/public", []string{"permitAll"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted for permitAll, got %d", voteResult)
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

	voteResult := voter.Vote(context.Background(), auth, "/api/denied", []string{"denyAll"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied for denyAll, got %d", voteResult)
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
	voteResult := voter.Vote(context.Background(), auth, "/api/users", []string{"authenticated"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted for authenticated user, got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), nil, "/api/users", []string{"authenticated"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied for nil authentication, got %d", voteResult)
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

	voteResult := voter.Vote(context.Background(), auth, "/api/admin", []string{"hasRole('ADMIN')"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted for hasRole('ADMIN'), got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), auth, "/api/admin", []string{"hasRole('GUEST')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied for hasRole('GUEST'), got %d", voteResult)
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

	voteResult := voter.Vote(context.Background(), auth, "/api/users", []string{"hasAnyRole('ADMIN','USER')"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted for hasAnyRole('ADMIN','USER'), got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), auth, "/api/admin", []string{"hasAnyRole('ADMIN','GUEST')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied for hasAnyRole('ADMIN','GUEST'), got %d", voteResult)
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

	voteResult := voter.Vote(context.Background(), auth, "/api/data", []string{"hasAuthority('read')"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted for hasAuthority('read'), got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), auth, "/api/data", []string{"hasAuthority('delete')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied for hasAuthority('delete'), got %d", voteResult)
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

	voteResult := voter.Vote(context.Background(), auth, "/api/data", []string{"hasAnyAuthority('read','write')"})
	if voteResult != AccessGranted {
		t.Errorf("expected AccessGranted for hasAnyAuthority with matching, got %d", voteResult)
	}

	voteResult = voter.Vote(context.Background(), auth, "/api/data", []string{"hasAnyAuthority('write','delete')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied for hasAnyAuthority without matching, got %d", voteResult)
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

	voteResult := voter.Vote(context.Background(), auth, "/api/users", []string{})
	if voteResult != AccessAbstain {
		t.Errorf("expected AccessAbstain for empty attributes, got %d", voteResult)
	}
}

func TestWebExpressionVoter_NilAuthentication(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()

	voteResult := voter.Vote(context.Background(), nil, "/api/users", []string{"hasRole('ADMIN')"})
	if voteResult != AccessDenied {
		t.Errorf("expected AccessDenied for nil authentication, got %d", voteResult)
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

	voteResult := voter.Vote(context.Background(), auth, "/api/users", []string{"unknownExpression"})
	if voteResult != AccessAbstain {
		t.Errorf("expected AccessAbstain for unknown expression, got %d", voteResult)
	}
}

func TestWebExpressionVoter_Supports(t *testing.T) {
	t.Parallel()

	voter := NewWebExpressionVoter()
	if !voter.Supports("anything") {
		t.Error("expected Supports to return true for any attribute")
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
