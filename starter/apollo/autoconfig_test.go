package apollo

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestApolloConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-apollo", environment.PriorityNormal, map[string]any{
		"apollo.enabled":   "true",
		"apollo.app_id":    "test-app",
		"apollo.cluster":   "dev",
		"apollo.meta_addr": "http://192.168.1.100:8080",
		"apollo.namespace": "custom-namespace",
	}))

	cfg := &ApolloConfig{
		Cluster:        DefaultCluster,
		Namespace:      DefaultNamespace,
		IsBackupConfig: DefaultIsBackupConfig,
	}

	err := env.BindPrefix("apollo", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected apollo.enabled to be true")
	}
	if cfg.AppID != "test-app" {
		t.Errorf("expected app_id 'test-app', got '%s'", cfg.AppID)
	}
	if cfg.Cluster != "dev" {
		t.Errorf("expected cluster 'dev', got '%s'", cfg.Cluster)
	}
	if cfg.Namespace != "custom-namespace" {
		t.Errorf("expected namespace 'custom-namespace', got '%s'", cfg.Namespace)
	}
}

func TestApolloConfig_DefaultValues(t *testing.T) {
	cfg := &ApolloConfig{
		Cluster:        DefaultCluster,
		Namespace:      DefaultNamespace,
		IsBackupConfig: DefaultIsBackupConfig,
	}

	if cfg.Cluster != "default" {
		t.Errorf("expected default cluster 'default', got '%s'", cfg.Cluster)
	}
	if cfg.Namespace != "application" {
		t.Errorf("expected default namespace 'application', got '%s'", cfg.Namespace)
	}
	if !cfg.IsBackupConfig {
		t.Error("expected default is_backup_config to be true")
	}
}
