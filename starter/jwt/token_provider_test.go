package jwt

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewTokenProvider_Defaults(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider()

	if p.secretKey != "" {
		t.Errorf("expected empty secretKey, got %q", p.secretKey)
	}
	if p.expiration != time.Hour {
		t.Errorf("expected expiration 1h, got %v", p.expiration)
	}
	if p.issuer != "" {
		t.Errorf("expected empty issuer, got %q", p.issuer)
	}
	if p.audience != "" {
		t.Errorf("expected empty audience, got %q", p.audience)
	}
}

func TestTokenProvider_WithSecretKey(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(WithSecretKey("my-secret"))
	if p.secretKey != "my-secret" {
		t.Errorf("expected secretKey 'my-secret', got %q", p.secretKey)
	}
}

func TestTokenProvider_WithExpiration(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(WithExpiration(2 * time.Hour))
	if p.expiration != 2*time.Hour {
		t.Errorf("expected expiration 2h, got %v", p.expiration)
	}
}

func TestTokenProvider_WithRefreshExpiration(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(WithRefreshExpiration(48 * time.Hour))
	if p.refreshExpiration != 48*time.Hour {
		t.Errorf("expected refreshExpiration 48h, got %v", p.refreshExpiration)
	}
}

func TestTokenProvider_WithIssuer(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(WithIssuer("my-app"))
	if p.issuer != "my-app" {
		t.Errorf("expected issuer 'my-app', got %q", p.issuer)
	}
}

func TestTokenProvider_WithAudience(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(WithAudience("my-audience"))
	if p.audience != "my-audience" {
		t.Errorf("expected audience 'my-audience', got %q", p.audience)
	}
}

func TestGenerateToken_EmptySecret(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider()
	_, err := p.GenerateToken(context.Background(), "alice", []string{"ROLE_USER"})
	if !errors.Is(err, ErrEmptySecret) {
		t.Errorf("expected ErrEmptySecret, got %v", err)
	}
}

func TestParseToken_EmptySecret(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider()
	_, err := p.ParseToken(context.Background(), "some-token")
	if !errors.Is(err, ErrEmptySecret) {
		t.Errorf("expected ErrEmptySecret, got %v", err)
	}
}

func TestGenerateToken_Success(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(
		WithSecretKey("test-secret"),
		WithExpiration(time.Hour),
	)

	token, err := p.GenerateToken(context.Background(), "alice", []string{"ROLE_ADMIN", "ROLE_USER"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestParseToken_Success(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(
		WithSecretKey("test-secret"),
		WithExpiration(time.Hour),
		WithIssuer("test-issuer"),
		WithAudience("test-audience"),
	)

	token, err := p.GenerateToken(context.Background(), "alice", []string{"ROLE_ADMIN"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := p.ParseToken(context.Background(), token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if claims.Subject != "alice" {
		t.Errorf("Subject = %q, want 'alice'", claims.Subject)
	}
	if claims.Issuer != "test-issuer" {
		t.Errorf("Issuer = %q, want 'test-issuer'", claims.Issuer)
	}
	if claims.Audience != "test-audience" {
		t.Errorf("Audience = %q, want 'test-audience'", claims.Audience)
	}
	if len(claims.Authorities) != 1 || claims.Authorities[0] != "ROLE_ADMIN" {
		t.Errorf("Authorities = %v, want [ROLE_ADMIN]", claims.Authorities)
	}
	if claims.Expiration.Before(time.Now()) {
		t.Error("Expiration should be in the future")
	}
	if claims.IssuedAt.After(time.Now()) {
		t.Error("IssuedAt should be in the past or now")
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(WithSecretKey("test-secret"))

	_, err := p.ParseToken(context.Background(), "invalid.token.string")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	t.Parallel()
	p1 := NewTokenProvider(WithSecretKey("secret-1"))
	p2 := NewTokenProvider(WithSecretKey("secret-2"))

	token, err := p1.GenerateToken(context.Background(), "alice", []string{"ROLE_USER"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	_, err = p2.ParseToken(context.Background(), token)
	if err == nil {
		t.Error("expected error when parsing with wrong secret")
	}
}

func TestValidateToken_Success(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(
		WithSecretKey("test-secret"),
		WithExpiration(time.Hour),
	)

	token, err := p.GenerateToken(context.Background(), "alice", []string{"ROLE_USER"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	err = p.ValidateToken(context.Background(), token)
	if err != nil {
		t.Errorf("ValidateToken() error = %v", err)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(WithSecretKey("test-secret"))

	err := p.ValidateToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestRefreshToken_Success(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(
		WithSecretKey("test-secret"),
		WithExpiration(time.Hour),
		WithRefreshExpiration(48*time.Hour),
	)

	token, err := p.GenerateToken(context.Background(), "alice", []string{"ROLE_USER"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	newToken, err := p.RefreshToken(context.Background(), token)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}
	if newToken == "" {
		t.Error("expected non-empty refreshed token")
	}

	newClaims, err := p.ParseToken(context.Background(), newToken)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if newClaims.Subject != "alice" {
		t.Errorf("Subject = %q, want 'alice'", newClaims.Subject)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(WithSecretKey("test-secret"))

	_, err := p.RefreshToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestRefreshToken_DefaultExpiration(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(
		WithSecretKey("test-secret"),
		WithExpiration(time.Hour),
	)

	token, err := p.GenerateToken(context.Background(), "alice", []string{"ROLE_USER"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	newToken, err := p.RefreshToken(context.Background(), token)
	if err != nil {
		t.Fatalf("RefreshToken() error = %v", err)
	}

	newClaims, err := p.ParseToken(context.Background(), newToken)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if newClaims.Subject != "alice" {
		t.Errorf("Subject = %q, want 'alice'", newClaims.Subject)
	}
}

func TestTokenProvider_MultipleOptions(t *testing.T) {
	t.Parallel()
	p := NewTokenProvider(
		WithSecretKey("full-secret"),
		WithExpiration(2*time.Hour),
		WithRefreshExpiration(72*time.Hour),
		WithIssuer("multi-app"),
		WithAudience("multi-aud"),
	)

	if p.secretKey != "full-secret" {
		t.Errorf("secretKey = %q, want 'full-secret'", p.secretKey)
	}
	if p.expiration != 2*time.Hour {
		t.Errorf("expiration = %v, want 2h", p.expiration)
	}
	if p.refreshExpiration != 72*time.Hour {
		t.Errorf("refreshExpiration = %v, want 72h", p.refreshExpiration)
	}
	if p.issuer != "multi-app" {
		t.Errorf("issuer = %q, want 'multi-app'", p.issuer)
	}
	if p.audience != "multi-aud" {
		t.Errorf("audience = %q, want 'multi-aud'", p.audience)
	}
}
