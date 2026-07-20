package ent

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestEntConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-ent", environment.PriorityNormal, map[string]any{
		"ent.enabled":  "true",
		"ent.driver":   "postgres",
		"ent.dsn":      "postgres://user:pass@localhost/db",
		"ent.database": "testdb",
	}))

	cfg := &EntConfig{
		Driver:   DefaultDriver,
		DSN:      DefaultDSN,
		Database: DefaultDatabase,
	}

	err := env.BindPrefix("ent", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected ent.enabled to be true")
	}
	if cfg.Driver != "postgres" {
		t.Errorf("expected driver 'postgres', got '%s'", cfg.Driver)
	}
	if cfg.Database != "testdb" {
		t.Errorf("expected database 'testdb', got '%s'", cfg.Database)
	}
}

func TestEntConfig_DefaultValues(t *testing.T) {
	cfg := &EntConfig{
		Driver:   DefaultDriver,
		DSN:      DefaultDSN,
		Database: DefaultDatabase,
	}

	if cfg.Driver != "mysql" {
		t.Errorf("expected default driver 'mysql', got '%s'", cfg.Driver)
	}
	if cfg.Database != "enhance" {
		t.Errorf("expected default database 'enhance', got '%s'", cfg.Database)
	}
}
