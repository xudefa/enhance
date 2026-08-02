package gorm

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestGormConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-gorm", environment.PriorityNormal, map[string]any{
		"db.gorm.enabled":           "true",
		"db.gorm.host":              "192.168.1.100",
		"db.gorm.port":              "3307",
		"db.gorm.username":          "testuser",
		"db.gorm.password":          "testpass",
		"db.gorm.database":          "testdb",
		"db.gorm.charset":           "utf8",
		"db.gorm.max-open-conns":    "50",
		"db.gorm.max-idle-conns":    "5",
		"db.gorm.conn-max-lifetime": "1800",
	}))

	cfg := &GormConfig{
		Host:            DefaultGORMHost,
		Port:            DefaultGORMPort,
		Username:        DefaultGORMUsername,
		Password:        DefaultGORMPassword,
		Database:        DefaultGORMDatabase,
		Charset:         DefaultGORMCharset,
		MaxOpenConns:    DefaultGORMMaxOpenConns,
		MaxIdleConns:    DefaultGORMMaxIdleConns,
		ConnMaxLifetime: DefaultGORMConnMaxLifetime,
	}

	err := env.BindPrefix("db.gorm", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected db.gorm.enabled to be true")
	}
	if cfg.Host != "192.168.1.100" {
		t.Errorf("expected host '192.168.1.100', got '%s'", cfg.Host)
	}
	if cfg.Port != 3307 {
		t.Errorf("expected port 3307, got %d", cfg.Port)
	}
	if cfg.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", cfg.Username)
	}
	if cfg.Database != "testdb" {
		t.Errorf("expected database 'testdb', got '%s'", cfg.Database)
	}
	if cfg.MaxOpenConns != 50 {
		t.Errorf("expected max-open-conns 50, got %d", cfg.MaxOpenConns)
	}
}

func TestGormConfig_DefaultValues(t *testing.T) {
	cfg := &GormConfig{
		Host:            DefaultGORMHost,
		Port:            DefaultGORMPort,
		Username:        DefaultGORMUsername,
		Password:        DefaultGORMPassword,
		Database:        DefaultGORMDatabase,
		Charset:         DefaultGORMCharset,
		MaxOpenConns:    DefaultGORMMaxOpenConns,
		MaxIdleConns:    DefaultGORMMaxIdleConns,
		ConnMaxLifetime: DefaultGORMConnMaxLifetime,
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 3306 {
		t.Errorf("expected default port 3306, got %d", cfg.Port)
	}
	if cfg.MaxOpenConns != 100 {
		t.Errorf("expected default max-open-conns 100, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 10 {
		t.Errorf("expected default max-idle-conns 10, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 3600 {
		t.Errorf("expected default conn-max-lifetime 3600, got %d", cfg.ConnMaxLifetime)
	}
}

func TestGormConfig_BuildDSN(t *testing.T) {
	cfg := &GormConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "root",
		Database: "enhance",
		Charset:  "utf8mb4",
	}

	autoConfig := &GormAutoConfiguration{}
	dsn := autoConfig.buildDSN(cfg)

	expected := "root:root@tcp(localhost:3306)/enhance?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}
