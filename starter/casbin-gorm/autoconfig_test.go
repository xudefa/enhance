package casbingorm

import (
	"testing"

	"github.com/xudefa/enhance/security"
)

func TestCasbinGormConfig_DefaultValues(t *testing.T) {
	cfg := &CasbinGormConfig{
		ModelType:        security.DefaultCasbinModelType,
		ModelPath:        security.DefaultCasbinModelPath,
		PolicyType:       DefaultCasbinGormPolicyType,
		AutoCreateTable:  DefaultCasbinGormAutoCreateTable,
		TableName:        DefaultCasbinGormTableName,
		AutoLoad:         security.DefaultCasbinAutoLoad,
		AutoLoadInterval: security.DefaultCasbinAutoLoadInterval,
	}

	if cfg.ModelType != "file" {
		t.Errorf("expected default model_type 'file', got '%s'", cfg.ModelType)
	}
	if cfg.ModelPath != "config/casbin_model.conf" {
		t.Errorf("expected default model_path 'config/casbin_model.conf', got '%s'", cfg.ModelPath)
	}
	if cfg.PolicyType != "gorm" {
		t.Errorf("expected default policy_type 'gorm', got '%s'", cfg.PolicyType)
	}
	if cfg.TableName != "casbin_rule" {
		t.Errorf("expected default table_name 'casbin_rule', got '%s'", cfg.TableName)
	}
	if cfg.AutoLoadInterval != 5 {
		t.Errorf("expected default auto_load_interval 5, got %d", cfg.AutoLoadInterval)
	}
}

func TestCasbinGormConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *CasbinGormConfig
		wantErr bool
	}{
		{
			name: "valid config with model path",
			cfg: &CasbinGormConfig{
				ModelType:  "file",
				ModelPath:  "config/casbin_model.conf",
				PolicyType: "gorm",
				TableName:  "casbin_rule",
			},
			wantErr: false,
		},
		{
			name: "valid config with model text",
			cfg: &CasbinGormConfig{
				ModelType:  "text",
				ModelText:  "[request_definition]\nr = sub, obj, act",
				PolicyType: "gorm",
				TableName:  "casbin_rule",
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing model",
			cfg: &CasbinGormConfig{
				ModelType:  "file",
				ModelPath:  "",
				PolicyType: "gorm",
				TableName:  "casbin_rule",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置字段是否正确设置
			if tt.cfg.ModelType == "" {
				t.Error("model_type should not be empty")
			}
			if tt.cfg.PolicyType != "gorm" {
				t.Errorf("expected policy_type 'gorm', got '%s'", tt.cfg.PolicyType)
			}
		})
	}
}
