package environment

import (
	"os"
	"testing"
)

func TestGetProfileActive_FromArgs(t *testing.T) {
	t.Parallel()
	args := []string{"--profile=dev", "--other=value"}
	profile := GetProfileActive(args)
	if profile != "dev" {
		t.Errorf("expected 'dev', got %s", profile)
	}
}

func TestGetProfileActive_FromEnv(t *testing.T) {
	t.Parallel()
	_ = os.Setenv("GO_BOOT_PROFILE", "test")
	defer func() { _ = os.Unsetenv("GO_BOOT_PROFILE") }()

	args := []string{"--other=value"}
	profile := GetProfileActive(args)
	if profile != "test" {
		t.Errorf("expected 'test', got %s", profile)
	}
}

func TestGetProfileActive_ArgsPriority(t *testing.T) {
	t.Parallel()
	_ = os.Setenv("GO_BOOT_PROFILE", "env-profile")
	defer func() { _ = os.Unsetenv("GO_BOOT_PROFILE") }()

	args := []string{"--profile=arg-profile"}
	profile := GetProfileActive(args)
	if profile != "arg-profile" {
		t.Errorf("expected 'arg-profile', got %s", profile)
	}
}

func TestGetProfileActive_None(t *testing.T) {
	t.Parallel()
	_ = os.Unsetenv("GO_BOOT_PROFILE")

	args := []string{"--other=value"}
	profile := GetProfileActive(args)
	if profile != "" {
		t.Errorf("expected empty string, got %s", profile)
	}
}

func TestParseProfiles_ProfileFunc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single profile",
			input:    "dev",
			expected: []string{"dev"},
		},
		{
			name:     "multiple profiles",
			input:    "dev,test",
			expected: []string{"dev", "test"},
		},
		{
			name:     "with spaces",
			input:    "dev, test , prod",
			expected: []string{"dev", "test", "prod"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "with empty items",
			input:    "dev,,test",
			expected: []string{"dev", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ParseProfiles(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d profiles, got %d", len(tt.expected), len(result))
			}
			for i, p := range tt.expected {
				if result[i] != p {
					t.Errorf("profile[%d]: expected %s, got %s", i, p, result[i])
				}
			}
		})
	}
}
