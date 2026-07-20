package swagger

import (
	"testing"

	"github.com/xudefa/enhance/config/environment"
)

func TestSwaggerConfig_LoadConfig(t *testing.T) {
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-swagger", environment.PriorityNormal, map[string]any{
		"swagger.enabled": "true",
		"swagger.host":    "127.0.0.1",
		"swagger.port":    "9090",
		"swagger.url":     "/api-docs/*",
		"swagger.title":   "My API",
	}))

	cfg := &SwaggerConfig{
		Host:  DefaultSwaggerHost,
		Port:  DefaultSwaggerPort,
		URL:   DefaultSwaggerURL,
		Title: DefaultSwaggerTitle,
	}

	err := env.BindPrefix("swagger", cfg)
	if err != nil {
		t.Fatalf("绑定配置失败: %v", err)
	}

	if !cfg.Enabled {
		t.Error("expected swagger.enabled to be true")
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host '127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.URL != "/api-docs/*" {
		t.Errorf("expected url '/api-docs/*', got '%s'", cfg.URL)
	}
}

func TestSwaggerConfig_DefaultValues(t *testing.T) {
	cfg := &SwaggerConfig{
		Host:  DefaultSwaggerHost,
		Port:  DefaultSwaggerPort,
		URL:   DefaultSwaggerURL,
		Title: DefaultSwaggerTitle,
	}

	if cfg.Host != "localhost" {
		t.Errorf("expected default host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}
	if cfg.URL != "/swagger/*" {
		t.Errorf("expected default url '/swagger/*', got '%s'", cfg.URL)
	}
}
