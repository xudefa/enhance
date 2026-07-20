package casbinxorm

import (
	"testing"

	"github.com/xudefa/enhance/security"
)

func TestCasbinXormConfigDefaults(t *testing.T) {
	cfg := &CasbinXormConfig{
		ModelType:        security.DefaultCasbinModelType,
		ModelPath:        security.DefaultCasbinModelPath,
		PolicyType:       DefaultCasbinXormPolicyType,
		AutoCreateTable:  DefaultCasbinXormAutoCreateTable,
		TableName:        DefaultCasbinXormTableName,
		DatabasePrefix:   DefaultCasbinXormDatabasePrefix,
		AutoLoad:         security.DefaultCasbinAutoLoad,
		AutoLoadInterval: security.DefaultCasbinAutoLoadInterval,
	}

	if cfg.PolicyType != "xorm" {
		t.Errorf("expected default policy-type 'xorm', got '%s'", cfg.PolicyType)
	}
	if cfg.AutoCreateTable != true {
		t.Errorf("expected default auto-create-table true, got %v", cfg.AutoCreateTable)
	}
	if cfg.TableName != "casbin_rule" {
		t.Errorf("expected default table-name 'casbin_rule', got '%s'", cfg.TableName)
	}
	if cfg.DatabasePrefix != "" {
		t.Errorf("expected default database-prefix '', got '%s'", cfg.DatabasePrefix)
	}
}

func TestValidateConfigFileModel(t *testing.T) {
	autoConfig := &CasbinXormAutoConfiguration{}

	cfg := &CasbinXormConfig{
		ModelType:       "file",
		ModelPath:       "config/casbin_model.conf",
		TableName:       "casbin_rule",
		AutoCreateTable: true,
	}

	err := autoConfig.validateConfig(cfg)
	if err != nil {
		t.Errorf("expected no error for valid file model config, got: %v", err)
	}
}

func TestValidateConfigStringModel(t *testing.T) {
	autoConfig := &CasbinXormAutoConfiguration{}

	cfg := &CasbinXormConfig{
		ModelType:       "string",
		ModelText:       "[request_definition]\nr = sub, obj, act",
		TableName:       "casbin_rule",
		AutoCreateTable: true,
	}

	err := autoConfig.validateConfig(cfg)
	if err != nil {
		t.Errorf("expected no error for valid string model config, got: %v", err)
	}
}

func TestValidateConfigMissingModelPath(t *testing.T) {
	autoConfig := &CasbinXormAutoConfiguration{}

	cfg := &CasbinXormConfig{
		ModelType: "file",
		ModelPath: "",
		TableName: "casbin_rule",
	}

	err := autoConfig.validateConfig(cfg)
	if err == nil {
		t.Error("expected error for missing model-path, got nil")
	}
}

func TestValidateConfigMissingModelText(t *testing.T) {
	autoConfig := &CasbinXormAutoConfiguration{}

	cfg := &CasbinXormConfig{
		ModelType: "string",
		ModelText: "",
		TableName: "casbin_rule",
	}

	err := autoConfig.validateConfig(cfg)
	if err == nil {
		t.Error("expected error for missing model-text, got nil")
	}
}

func TestValidateConfigInvalidModelType(t *testing.T) {
	autoConfig := &CasbinXormAutoConfiguration{}

	cfg := &CasbinXormConfig{
		ModelType: "invalid",
		TableName: "casbin_rule",
	}

	err := autoConfig.validateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid model-type, got nil")
	}
}

func TestValidateConfigMissingTableName(t *testing.T) {
	autoConfig := &CasbinXormAutoConfiguration{}

	cfg := &CasbinXormConfig{
		ModelType: "file",
		ModelPath: "config/casbin_model.conf",
		TableName: "",
	}

	err := autoConfig.validateConfig(cfg)
	if err == nil {
		t.Error("expected error for missing table-name, got nil")
	}
}

func TestConfigKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"CasbinXormEnabled", CasbinXormEnabled, "security.casbin.enabled"},
		{"CasbinXormPolicyType", CasbinXormPolicyType, "security.casbin.policy-type"},
		{"CasbinXormAutoCreateTable", CasbinXormAutoCreateTable, "security.casbin.auto-create-table"},
		{"CasbinXormTableName", CasbinXormTableName, "security.casbin.table-name"},
		{"CasbinXormDatabasePrefix", CasbinXormDatabasePrefix, "security.casbin.database-prefix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key != tt.expected {
				t.Errorf("expected key '%s', got '%s'", tt.expected, tt.key)
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected interface{}
	}{
		{"DefaultCasbinXormPolicyType", DefaultCasbinXormPolicyType, "xorm"},
		{"DefaultCasbinXormAutoCreateTable", DefaultCasbinXormAutoCreateTable, true},
		{"DefaultCasbinXormTableName", DefaultCasbinXormTableName, "casbin_rule"},
		{"DefaultCasbinXormDatabasePrefix", DefaultCasbinXormDatabasePrefix, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("expected default value %v, got %v", tt.expected, tt.value)
			}
		})
	}
}
