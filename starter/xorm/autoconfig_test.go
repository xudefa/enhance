package xorm

import (
	"testing"
)

func TestXormConfigDefaults(t *testing.T) {
	cfg := &XormConfig{
		Type:            DefaultXORMType,
		Host:            DefaultXORMHost,
		Port:            DefaultXORMPort,
		Username:        DefaultXORMUsername,
		Password:        DefaultXORMPassword,
		Database:        DefaultXORMDatabase,
		Charset:         DefaultXORMCharset,
		MaxOpenConns:    DefaultXORMMaxOpenConns,
		MaxIdleConns:    DefaultXORMMaxIdleConns,
		ConnMaxLifetime: DefaultXORMConnMaxLifetime,
		ShowSQL:         DefaultXORMShowSQL,
	}

	if cfg.Type != "mysql" {
		t.Errorf("expected default type 'mysql', got '%s'", cfg.Type)
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
	if cfg.ShowSQL != false {
		t.Errorf("expected default show-sql false, got %v", cfg.ShowSQL)
	}
}

func TestBuildDSNMySQL(t *testing.T) {
	autoConfig := &XormAutoConfiguration{}
	cfg := &XormConfig{
		Type:     "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "testdb",
		Charset:  "utf8mb4",
	}

	dsn := autoConfig.buildDSN(cfg)
	expected := "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}

func TestBuildDSNPostgres(t *testing.T) {
	autoConfig := &XormAutoConfiguration{}
	cfg := &XormConfig{
		Type:     "postgres",
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "password",
		Database: "testdb",
	}

	dsn := autoConfig.buildDSN(cfg)
	expected := "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable"
	if dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}

func TestBuildDSNSQLite(t *testing.T) {
	autoConfig := &XormAutoConfiguration{}
	cfg := &XormConfig{
		Type:     "sqlite3",
		Database: "./test.db",
	}

	dsn := autoConfig.buildDSN(cfg)
	expected := "./test.db"
	if dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}

func TestBuildDSNDefault(t *testing.T) {
	autoConfig := &XormAutoConfiguration{}
	cfg := &XormConfig{
		Type:     "unknown",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		Password: "password",
		Database: "testdb",
		Charset:  "utf8mb4",
	}

	dsn := autoConfig.buildDSN(cfg)
	expected := "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	if dsn != expected {
		t.Errorf("expected DSN '%s', got '%s'", expected, dsn)
	}
}

func TestConfigKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"XORMEnabled", XORMEnabled, "db.xorm.enabled"},
		{"XORMType", XORMType, "db.xorm.type"},
		{"XORMHost", XORMHost, "db.xorm.host"},
		{"XORMPort", XORMPort, "db.xorm.port"},
		{"XORMUsername", XORMUsername, "db.xorm.username"},
		{"XORMPassword", XORMPassword, "db.xorm.password"},
		{"XORMDatabase", XORMDatabase, "db.xorm.database"},
		{"XORMCharset", XORMCharset, "db.xorm.charset"},
		{"XORMMaxOpenConns", XORMMaxOpenConns, "db.xorm.max-open-conns"},
		{"XORMMaxIdleConns", XORMMaxIdleConns, "db.xorm.max-idle-conns"},
		{"XORMConnMaxLifetime", XORMConnMaxLifetime, "db.xorm.conn-max-lifetime"},
		{"XORMShowSQL", XORMShowSQL, "db.xorm.show-sql"},
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
		{"DefaultXORMType", DefaultXORMType, "mysql"},
		{"DefaultXORMHost", DefaultXORMHost, "localhost"},
		{"DefaultXORMPort", DefaultXORMPort, 3306},
		{"DefaultXORMUsername", DefaultXORMUsername, "scott"},
		{"DefaultXORMPassword", DefaultXORMPassword, "123456"},
		{"DefaultXORMDatabase", DefaultXORMDatabase, "demo"},
		{"DefaultXORMCharset", DefaultXORMCharset, "utf8mb4"},
		{"DefaultXORMMaxOpenConns", DefaultXORMMaxOpenConns, 100},
		{"DefaultXORMMaxIdleConns", DefaultXORMMaxIdleConns, 10},
		{"DefaultXORMConnMaxLifetime", DefaultXORMConnMaxLifetime, 3600},
		{"DefaultXORMShowSQL", DefaultXORMShowSQL, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value != tt.expected {
				t.Errorf("expected default value %v, got %v", tt.expected, tt.value)
			}
		})
	}
}
