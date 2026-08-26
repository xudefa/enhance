package banner

import (
	"strings"
	"testing"
)

func TestLegacyBanner_Print_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		lines    []string
		appName  string
		profiles []string
		version  string
		want     []string
	}{
		{
			name:     "with all options",
			lines:    []string{"Line 1"},
			appName:  "my-app",
			profiles: []string{"dev", "test"},
			version:  "1.0.0",
			want:     []string{"Line 1", "my-app", "1.0.0", "dev", "test"},
		},
		{
			name:     "no profiles shows default",
			lines:    []string{},
			appName:  "app",
			profiles: []string{},
			version:  "2.0.0",
			want:     []string{"default", "app"},
		},
		{
			name:     "nil profiles shows default",
			lines:    []string{},
			appName:  "app",
			profiles: nil,
			version:  "3.0.0",
			want:     []string{"default"},
		},
		{
			name:     "empty appName uses Application",
			lines:    []string{},
			appName:  "",
			profiles: []string{},
			version:  "4.0.0",
			want:     []string{"Application"},
		},
		{
			name:     "multiple lines",
			lines:    []string{"A", "B", "C"},
			appName:  "test",
			profiles: []string{"prod"},
			version:  "5.0.0",
			want:     []string{"A", "B", "C"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := NewLegacyBanner(
				WithLines(tt.lines),
				WithAppName(tt.appName),
				WithProfiles(tt.profiles),
			)
			output := captureStdout(t, func() {
				if err := b.Print(tt.version); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("output %q does not contain %q", output, want)
				}
			}
		})
	}
}

func TestLegacyBanner_Print_LogMode(t *testing.T) {
	t.Parallel()

	b := NewLegacyBanner(WithLines([]string{"test"}), WithAppName("app"))
	b.(*LegacyBanner).SetMode(BannerModeLog)

	output := captureStdout(t, func() {
		if err := b.Print("1.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[enhance]") {
		t.Errorf("expected log prefix, got %q", output)
	}
}

func TestLegacyBanner_Print_OffMode(t *testing.T) {
	t.Parallel()

	b := NewLegacyBanner(WithLines([]string{"should not appear"}))
	b.(*LegacyBanner).SetMode(BannerModeOff)

	output := captureStdout(t, func() {
		if err := b.Print("1.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "" {
		t.Errorf("expected empty output in off mode, got %q", output)
	}
}

func TestLegacyBanner_SetMode(t *testing.T) {
	t.Parallel()

	b := NewLegacyBanner(WithLines([]string{"test"}))
	if b.Mode() != BannerModeConsole {
		t.Errorf("expected BannerModeConsole, got %v", b.Mode())
	}

	b.(*LegacyBanner).SetMode(BannerModeLog)
	if b.Mode() != BannerModeLog {
		t.Errorf("expected BannerModeLog, got %v", b.Mode())
	}
}

func TestLegacyBanner_DefaultAppName(t *testing.T) {
	t.Parallel()

	b := NewLegacyBanner(WithLines([]string{}))
	output := captureStdout(t, func() {
		_ = b.Print("1.0.0")
	})

	if !strings.Contains(output, ":: Application ::") {
		t.Errorf("expected default app name 'Application', got %q", output)
	}
}

func TestFormatProfiles_LegacyTest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		profiles []string
		want     string
	}{
		{"nil", nil, "default"},
		{"empty", []string{}, "default"},
		{"single", []string{"dev"}, "dev"},
		{"multiple", []string{"a", "b", "c"}, "a, b, c"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formatProfiles(tt.profiles); got != tt.want {
				t.Errorf("formatProfiles() = %q, want %q", got, tt.want)
			}
		})
	}
}
