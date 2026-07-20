package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/xudefa/enhance/config/environment"
)

func TestValidatorConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-validator", environment.PriorityNormal, map[string]any{
		"validator.enabled":                  "true",
		"validator.enable-custom-validators": "false",
	}))

	cfg := &ValidatorConfig{
		EnableCustomValidators: DefaultEnableCustomValidators,
	}

	err := env.BindPrefix("validator", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected validator.enabled to be true")
	}
	if cfg.EnableCustomValidators {
		t.Error("expected enable-custom-validators to be false")
	}
}

func TestValidatorConfig_DefaultValues(t *testing.T) {
	cfg := &ValidatorConfig{
		EnableCustomValidators: DefaultEnableCustomValidators,
	}

	if !cfg.EnableCustomValidators {
		t.Error("expected default enable-custom-validators to be true")
	}
}

func TestValidator_ValidateStruct(t *testing.T) {
	v := NewTestValidator()

	type TestUser struct {
		Name  string `validate:"required,min=2"`
		Email string `validate:"required,email"`
	}

	// 测试有效数据
	validUser := TestUser{
		Name:  "John",
		Email: "john@example.com",
	}
	if err := v.Validate(validUser); err != nil {
		t.Errorf("expected no error for valid user, got %v", err)
	}

	// 测试无效数据
	invalidUser := TestUser{
		Name:  "A",
		Email: "invalid-email",
	}
	if err := v.Validate(invalidUser); err == nil {
		t.Error("expected error for invalid user, got nil")
	}
}

func TestValidator_ValidateVar(t *testing.T) {
	v := NewTestValidator()

	// 测试有效邮箱
	if err := v.ValidateVar("test@example.com", "email"); err != nil {
		t.Errorf("expected no error for valid email, got %v", err)
	}

	// 测试无效邮箱
	if err := v.ValidateVar("invalid-email", "email"); err == nil {
		t.Error("expected error for invalid email, got nil")
	}
}

// NewTestValidator 创建测试用的验证器
func NewTestValidator() *ValidatorAutoConfiguration {
	v := validator.New()
	return &ValidatorAutoConfiguration{
		validate: v,
	}
}
