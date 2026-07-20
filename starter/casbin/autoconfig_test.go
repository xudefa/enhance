package casbin

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/security"
)

func TestCasbinConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-casbin", environment.PriorityNormal, map[string]any{
		"security.casbin.enabled":     "true",
		"security.casbin.model-type":  "file",
		"security.casbin.model-path":  "/etc/casbin/model.conf",
		"security.casbin.policy-type": "file",
		"security.casbin.policy-path": "/etc/casbin/policy.csv",
	}))

	cfg := &CasbinConfig{
		ModelType:  security.DefaultCasbinModelType,
		ModelPath:  security.DefaultCasbinModelPath,
		PolicyType: security.DefaultCasbinPolicyType,
		PolicyPath: security.DefaultCasbinPolicyPath,
	}

	err := env.BindPrefix("security.casbin", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected security.casbin.enabled to be true")
	}
	if cfg.ModelPath != "/etc/casbin/model.conf" {
		t.Errorf("expected model-path '/etc/casbin/model.conf', got '%s'", cfg.ModelPath)
	}
	if cfg.PolicyPath != "/etc/casbin/policy.csv" {
		t.Errorf("expected policy-path '/etc/casbin/policy.csv', got '%s'", cfg.PolicyPath)
	}
}

func TestCasbinConfig_DefaultValues(t *testing.T) {
	cfg := &CasbinConfig{
		ModelType:  security.DefaultCasbinModelType,
		ModelPath:  security.DefaultCasbinModelPath,
		PolicyType: security.DefaultCasbinPolicyType,
		PolicyPath: security.DefaultCasbinPolicyPath,
	}

	if cfg.ModelType != "file" {
		t.Errorf("expected default model-type 'file', got '%s'", cfg.ModelType)
	}
	if cfg.ModelPath != "config/casbin_model.conf" {
		t.Errorf("expected default model-path 'config/casbin_model.conf', got '%s'", cfg.ModelPath)
	}
	if cfg.PolicyType != "file" {
		t.Errorf("expected default policy-type 'file', got '%s'", cfg.PolicyType)
	}
	if cfg.PolicyPath != "config/casbin_policy.csv" {
		t.Errorf("expected default policy-path 'config/casbin_policy.csv', got '%s'", cfg.PolicyPath)
	}
}
