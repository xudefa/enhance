package environment

import (
	"testing"
)

func TestEnvironment_GetProperty_Plan(t *testing.T) {
	t.Parallel()

	env := NewMapEnvironment(map[string]string{
		"app.name": "test-app",
		"app.port": "8080",
	})

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"existing property", "app.name", "test-app"},
		{"missing property", "missing", ""},
		{"with default", "missing", "default"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := env.GetPropertyWithDefault(tt.key, tt.expected)
			if result != tt.expected {
				t.Errorf("GetPropertyWithDefault() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnvironment_GetActiveProfiles_Plan(t *testing.T) {
	t.Parallel()

	env := NewMapEnvironmentWithProfiles(
		map[string]string{},
		[]string{"dev", "test"},
	)

	profiles := env.GetActiveProfiles()
	if len(profiles) != 2 {
		t.Errorf("GetActiveProfiles() returned %d profiles, want 2", len(profiles))
	}
}

func TestEnvironment_AcceptsProfiles_Plan(t *testing.T) {
	t.Parallel()

	env := NewMapEnvironmentWithProfiles(
		map[string]string{},
		[]string{"dev", "test"},
	)

	tests := []struct {
		name     string
		profiles []string
		expected bool
	}{
		{"accepts dev", []string{"dev"}, true},
		{"accepts test", []string{"test"}, true},
		{"accepts prod", []string{"prod"}, false},
		{"accepts multiple", []string{"dev", "prod"}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := env.AcceptsProfiles(tt.profiles...)
			if result != tt.expected {
				t.Errorf("AcceptsProfiles() = %v, want %v", result, tt.expected)
			}
		})
	}
}
