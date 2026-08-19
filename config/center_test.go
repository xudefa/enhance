package config

import (
	"testing"
	"time"
)

func TestWithDataID(t *testing.T) {
	t.Parallel()
	cfg := &ConfigCenterConfig{}
	opt := WithDataID("my-config")
	opt(cfg)
	if cfg.DataID != "my-config" {
		t.Errorf("expected DataID 'my-config', got %s", cfg.DataID)
	}
}

func TestWithGroup(t *testing.T) {
	t.Parallel()
	cfg := &ConfigCenterConfig{}
	opt := WithGroup("DEFAULT_GROUP")
	opt(cfg)
	if cfg.Group != "DEFAULT_GROUP" {
		t.Errorf("expected Group 'DEFAULT_GROUP', got %s", cfg.Group)
	}
}

func TestWithFormat(t *testing.T) {
	t.Parallel()
	cfg := &ConfigCenterConfig{}
	opt := WithFormat("/config/myapp/")
	opt(cfg)
	if cfg.Prefix != "/config/myapp/" {
		t.Errorf("expected Prefix '/config/myapp/', got %s", cfg.Prefix)
	}
}

func TestWithProfiles(t *testing.T) {
	t.Parallel()
	cfg := &ConfigCenterConfig{}
	opt := WithProfiles([]string{"dev", "staging"})
	opt(cfg)
	if cfg.Namespace != "dev,staging" {
		t.Errorf("expected Namespace 'dev,staging', got %s", cfg.Namespace)
	}
}

func TestWithProfiles_Empty(t *testing.T) {
	t.Parallel()
	cfg := &ConfigCenterConfig{}
	opt := WithProfiles([]string{})
	opt(cfg)
	if cfg.Namespace != "" {
		t.Errorf("expected empty Namespace, got %s", cfg.Namespace)
	}
}

func TestWithRemoteSource_Nacos(t *testing.T) {
	t.Parallel()
	cfg := WithRemoteSource("nacos", []string{"127.0.0.1:8848"},
		WithDataID("app-config"),
		WithGroup("DEFAULT_GROUP"),
	)

	if cfg.DataID != "app-config" {
		t.Errorf("expected DataID 'app-config', got %s", cfg.DataID)
	}
	if cfg.Group != "DEFAULT_GROUP" {
		t.Errorf("expected Group 'DEFAULT_GROUP', got %s", cfg.Group)
	}
	if len(cfg.Endpoints) != 1 || cfg.Endpoints[0] != "127.0.0.1:8848" {
		t.Errorf("expected Endpoints ['127.0.0.1:8848'], got %v", cfg.Endpoints)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected Timeout 10s, got %v", cfg.Timeout)
	}
}

func TestWithRemoteSource_Etcd(t *testing.T) {
	t.Parallel()
	cfg := WithRemoteSource("etcd", []string{"127.0.0.1:2379"},
		WithFormat("/config/myapp/"),
	)

	if cfg.Prefix != "/config/myapp/" {
		t.Errorf("expected Prefix '/config/myapp/', got %s", cfg.Prefix)
	}
	if len(cfg.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(cfg.Endpoints))
	}
}

func TestWithRemoteSource_NoOptions(t *testing.T) {
	t.Parallel()
	cfg := WithRemoteSource("consul", []string{"127.0.0.1:8500"})

	if len(cfg.Endpoints) != 1 || cfg.Endpoints[0] != "127.0.0.1:8500" {
		t.Errorf("expected Endpoints ['127.0.0.1:8500'], got %v", cfg.Endpoints)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected default Timeout 10s, got %v", cfg.Timeout)
	}
}
