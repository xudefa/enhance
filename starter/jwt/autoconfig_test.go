package jwt

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestJWTConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-jwt", environment.PriorityNormal, map[string]any{
		"security.jwt.enabled":                  "true",
		"security.jwt.secret-key":               "test-secret-key",
		"security.jwt.expires-duration":         "7200",
		"security.jwt.refresh-expires-duration": "172800",
		"security.jwt.issuer":                   "test-app",
	}))

	cfg := &JwtConfig{
		SecretKey:              DefaultJWTSecretKey,
		ExpiresDuration:        DefaultJWTExpiresDuration,
		RefreshExpiresDuration: DefaultJWTRefreshExpiresDuration,
		Issuer:                 DefaultJWTIssuer,
		ExcludePaths:           DefaultJWTExcludePaths,
		SigningMethod:          DefaultJWTSigningMethod,
	}

	err := env.BindPrefix("security.jwt", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected security.jwt.enabled to be true")
	}
	if cfg.SecretKey != "test-secret-key" {
		t.Errorf("expected secret-key 'test-secret-key', got '%s'", cfg.SecretKey)
	}
	if cfg.ExpiresDuration != 7200 {
		t.Errorf("expected expires-duration 7200, got %d", cfg.ExpiresDuration)
	}
	if cfg.Issuer != "test-app" {
		t.Errorf("expected issuer 'test-app', got '%s'", cfg.Issuer)
	}
}

func TestJWTConfig_DefaultValues(t *testing.T) {
	cfg := &JwtConfig{
		SecretKey:              DefaultJWTSecretKey,
		ExpiresDuration:        DefaultJWTExpiresDuration,
		RefreshExpiresDuration: DefaultJWTRefreshExpiresDuration,
		Issuer:                 DefaultJWTIssuer,
		ExcludePaths:           DefaultJWTExcludePaths,
		SigningMethod:          DefaultJWTSigningMethod,
	}

	if cfg.SecretKey != "enhanceJwtSecret" {
		t.Errorf("expected default secret-key 'enhanceJwtSecret', got '%s'", cfg.SecretKey)
	}
	if cfg.ExpiresDuration != 600 {
		t.Errorf("expected default expires-duration 600, got %d", cfg.ExpiresDuration)
	}
	if cfg.RefreshExpiresDuration != 3600 {
		t.Errorf("expected default refresh-expires-duration 3600, got %d", cfg.RefreshExpiresDuration)
	}
	if cfg.Issuer != "enhance" {
		t.Errorf("expected default issuer 'enhance', got '%s'", cfg.Issuer)
	}
	if cfg.SigningMethod != "HS256" {
		t.Errorf("expected default signing-method 'HS256', got '%s'", cfg.SigningMethod)
	}
}
