package config

import (
	"testing"
	"time"
)

func TestNewWatchEvent(t *testing.T) {
	t.Parallel()
	before := time.Now()
	event := NewWatchEvent(EventModify, "app.name", "new-name", "file-loader")
	after := time.Now()

	if event.Type != EventModify {
		t.Errorf("expected Type '%s', got '%s'", EventModify, event.Type)
	}
	if event.Key != "app.name" {
		t.Errorf("expected Key 'app.name', got '%s'", event.Key)
	}
	if event.Value != "new-name" {
		t.Errorf("expected Value 'new-name', got '%v'", event.Value)
	}
	if event.Source != "file-loader" {
		t.Errorf("expected Source 'file-loader', got '%s'", event.Source)
	}
	if event.Timestamp.Before(before) || event.Timestamp.After(after) {
		t.Error("expected Timestamp to be between before and after")
	}
}

func TestNewWatchEvent_Delete(t *testing.T) {
	t.Parallel()
	event := NewWatchEvent(EventDelete, "app.port", nil, "config-center")
	if event.Type != EventDelete {
		t.Errorf("expected Type '%s', got '%s'", EventDelete, event.Type)
	}
	if event.Value != nil {
		t.Errorf("expected nil Value, got %v", event.Value)
	}
}

func TestNewWatchEvent_Create(t *testing.T) {
	t.Parallel()
	event := NewWatchEvent(EventCreate, "new.key", 42, "api")
	if event.Type != EventCreate {
		t.Errorf("expected Type '%s', got '%s'", EventCreate, event.Type)
	}
	if event.Value != 42 {
		t.Errorf("expected Value 42, got %v", event.Value)
	}
}

func TestConfigModelOptions(t *testing.T) {
	t.Parallel()
	model, err := New(nil,
		WithConfigName("myapp"),
		WithConfigFile("/etc/myapp/config.json"),
		WithConfigPath("/etc/myapp", "./config"),
		WithConfigType("yaml"),
		WithEnvironment("production"),
		WithEnvVariable("MYAPP_"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if model.ConfigName != "myapp" {
		t.Errorf("expected ConfigName 'myapp', got %s", model.ConfigName)
	}
	if model.ConfigFile != "/etc/myapp/config.json" {
		t.Errorf("expected ConfigFile '/etc/myapp/config.json', got %s", model.ConfigFile)
	}
	if len(model.ConfigPaths) != 2 || model.ConfigPaths[0] != "/etc/myapp" {
		t.Errorf("expected ConfigPaths ['/etc/myapp', './config'], got %v", model.ConfigPaths)
	}
	if model.ConfigType != "yaml" {
		t.Errorf("expected ConfigType 'yaml', got %s", model.ConfigType)
	}
	if model.Env != "production" {
		t.Errorf("expected Env 'production', got %s", model.Env)
	}
	if model.OptionName != "MYAPP_" {
		t.Errorf("expected OptionName 'MYAPP_', got %s", model.OptionName)
	}
}

func TestNew_WithLoadFunction(t *testing.T) {
	t.Parallel()
	called := false
	loadFn := func(m *ConfigModel) error {
		called = true
		m.Config["loaded"] = true
		return nil
	}

	model, err := New(loadFn, WithConfigName("test"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected load function to be called")
	}
	if model.Config["loaded"] != true {
		t.Error("expected loaded config to be set")
	}
}

func TestNew_LoadFunctionError(t *testing.T) {
	t.Parallel()
	loadFn := func(m *ConfigModel) error {
		return &testLoadError{msg: "load failed"}
	}

	_, err := New(loadFn)
	if err == nil {
		t.Fatal("expected error from load function")
	}
	if err.Error() != "load failed" {
		t.Errorf("expected error message 'load failed', got '%s'", err.Error())
	}
}

func TestNew_NilLoadFunction(t *testing.T) {
	t.Parallel()
	model, err := New(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model == nil {
		t.Error("expected non-nil model")
	}
	if model.Config == nil {
		t.Error("expected non-nil Config map")
	}
}

type testLoadError struct {
	msg string
}

func (e *testLoadError) Error() string { return e.msg }
