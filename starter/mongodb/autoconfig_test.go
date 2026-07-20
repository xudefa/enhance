package mongodb

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestMongoDBConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-mongodb", environment.PriorityNormal, map[string]any{
		"mongodb.enabled":         "true",
		"mongodb.host":            "192.168.1.100",
		"mongodb.port":            "27018",
		"mongodb.username":        "admin",
		"mongodb.password":        "password",
		"mongodb.database":        "testdb",
		"mongodb.max-pool-size":   "50",
		"mongodb.min-pool-size":   "5",
		"mongodb.connect-timeout": "15",
	}))

	cfg := &MongoDBConfig{
		Host:                   DefaultMongoDBHost,
		Port:                   DefaultMongoDBPort,
		Database:               DefaultMongoDBDatabase,
		AuthSource:             DefaultMongoDBAuthSource,
		MaxPoolSize:            DefaultMaxPoolSize,
		MinPoolSize:            DefaultMinPoolSize,
		ConnectTimeout:         DefaultConnectTimeout,
		ServerSelectionTimeout: DefaultServerSelectionTimeout,
	}

	err := env.BindPrefix("mongodb", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected mongodb.enabled to be true")
	}
	if cfg.Host != "192.168.1.100" {
		t.Errorf("expected host '192.168.1.100', got '%s'", cfg.Host)
	}
	if cfg.Port != 27018 {
		t.Errorf("expected port 27018, got %d", cfg.Port)
	}
	if cfg.Username != "admin" {
		t.Errorf("expected username 'admin', got '%s'", cfg.Username)
	}
	if cfg.Database != "testdb" {
		t.Errorf("expected database 'testdb', got '%s'", cfg.Database)
	}
	if cfg.MaxPoolSize != 50 {
		t.Errorf("expected max-pool-size 50, got %d", cfg.MaxPoolSize)
	}
}

func TestMongoDBConfig_DefaultValues(t *testing.T) {
	cfg := &MongoDBConfig{
		Host:                   DefaultMongoDBHost,
		Port:                   DefaultMongoDBPort,
		Database:               DefaultMongoDBDatabase,
		AuthSource:             DefaultMongoDBAuthSource,
		MaxPoolSize:            DefaultMaxPoolSize,
		MinPoolSize:            DefaultMinPoolSize,
		ConnectTimeout:         DefaultConnectTimeout,
		ServerSelectionTimeout: DefaultServerSelectionTimeout,
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 27017 {
		t.Errorf("expected default port 27017, got %d", cfg.Port)
	}
	if cfg.Database != "enhance" {
		t.Errorf("expected default database 'enhance', got '%s'", cfg.Database)
	}
	if cfg.MaxPoolSize != 100 {
		t.Errorf("expected default max-pool-size 100, got %d", cfg.MaxPoolSize)
	}
	if cfg.MinPoolSize != 10 {
		t.Errorf("expected default min-pool-size 10, got %d", cfg.MinPoolSize)
	}
}

func TestMongoDBConfig_BuildURI(t *testing.T) {
	autoConfig := &MongoDBAutoConfiguration{}

	// 无认证
	cfg := &MongoDBConfig{
		Host:     "localhost",
		Port:     27017,
		Database: "enhance",
	}
	uri := autoConfig.buildURI(cfg)
	expected := "mongodb://localhost:27017/enhance"
	if uri != expected {
		t.Errorf("expected URI '%s', got '%s'", expected, uri)
	}

	// 有认证
	cfgWithAuth := &MongoDBConfig{
		Host:       "localhost",
		Port:       27017,
		Username:   "admin",
		Password:   "secret",
		Database:   "enhance",
		AuthSource: "admin",
	}
	uri = autoConfig.buildURI(cfgWithAuth)
	expected = "mongodb://admin:secret@localhost:27017/enhance?authSource=admin"
	if uri != expected {
		t.Errorf("expected URI '%s', got '%s'", expected, uri)
	}
}
