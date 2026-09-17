package boot

import (
	"testing"
)

func TestBoot_WithExcludeMultiple_Extended(t *testing.T) {
	t.Parallel()

	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutStarters(),
		WithExclude("Config1", "Config2", "Config3", "Config4"),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}

	if len(app.config.ExcludedAutoConfigs) != 4 {
		t.Errorf("Expected 4 excluded configs, got %d", len(app.config.ExcludedAutoConfigs))
	}
}
