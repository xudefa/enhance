package banner

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

var stdoutMu sync.Mutex

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	stdoutMu.Lock()
	defer stdoutMu.Unlock()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
	}
	r.Close()
	return buf.String()
}

func TestTextBanner_Print(t *testing.T) {
	tests := []struct {
		name     string
		template string
		props    map[string]any
		version  string
		want     []string
	}{
		{
			name:     "with template and properties",
			template: "Hello {{name}} v{{version}}",
			props:    map[string]any{"name": "my-app"},
			version:  "1.0.0",
			want:     []string{"my-app", "1.0.0"},
		},
		{
			name:     "empty template",
			template: "",
			props:    nil,
			version:  "2.0.0",
			want:     []string{"2.0.0", "Application"},
		},
		{
			name:     "template only version",
			template: "App v{{version}}",
			props:    nil,
			version:  "3.0.0",
			want:     []string{"3.0.0"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := NewTextBanner(tt.template, tt.props)
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

func TestTextBanner_Mode(t *testing.T) {
	t.Parallel()

	b := NewTextBanner("test", nil)
	if b.Mode() != BannerModeConsole {
		t.Errorf("expected BannerModeConsole, got %v", b.Mode())
	}

	b.(*TextBanner).SetMode(BannerModeOff)
	if b.Mode() != BannerModeOff {
		t.Errorf("expected BannerModeOff, got %v", b.Mode())
	}
}

func TestTextBanner_OffMode(t *testing.T) {
	b := NewTextBanner("should not appear", nil)
	b.(*TextBanner).SetMode(BannerModeOff)

	output := captureStdout(t, func() {
		if err := b.Print("1.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "" {
		t.Errorf("expected empty output in off mode, got %q", output)
	}
}

func TestASCIIArtBanner_Print(t *testing.T) {
	tests := []struct {
		name    string
		art     string
		color   string
		version string
		want    []string
	}{
		{
			name:    "with art",
			art:     "  ___  \n /   \\ ",
			color:   "",
			version: "1.0.0",
			want:    []string{"___", "1.0.0"},
		},
		{
			name:    "empty art",
			art:     "",
			color:   "",
			version: "2.0.0",
			want:    []string{"2.0.0"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := NewASCIIArtBanner(tt.art, tt.color)
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

func TestASCIIArtBanner_Mode(t *testing.T) {
	t.Parallel()

	b := NewASCIIArtBanner("art", "red")
	if b.Mode() != BannerModeConsole {
		t.Errorf("expected BannerModeConsole, got %v", b.Mode())
	}

	b.(*ASCIIArtBanner).SetMode(BannerModeLog)
	if b.Mode() != BannerModeLog {
		t.Errorf("expected BannerModeLog, got %v", b.Mode())
	}
}

func TestASCIIArtBanner_OffMode(t *testing.T) {
	b := NewASCIIArtBanner("should not appear", "")
	b.(*ASCIIArtBanner).SetMode(BannerModeOff)

	output := captureStdout(t, func() {
		if err := b.Print("1.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if output != "" {
		t.Errorf("expected empty output in off mode, got %q", output)
	}
}

func TestLegacyBanner_Print(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		appName  string
		profiles []string
		version  string
		want     []string
	}{
		{
			name:     "with profiles",
			lines:    []string{"Line 1", "Line 2"},
			appName:  "my-app",
			profiles: []string{"dev", "test"},
			version:  "1.0.0",
			want:     []string{"Line 1", "Line 2", "my-app", "1.0.0", "dev", "test"},
		},
		{
			name:     "no profiles",
			lines:    []string{"Banner"},
			appName:  "app",
			profiles: []string{},
			version:  "2.0.0",
			want:     []string{"Banner", "app", "2.0.0", "default"},
		},
		{
			name:     "empty lines",
			lines:    []string{},
			appName:  "test-app",
			profiles: []string{"prod"},
			version:  "3.0.0",
			want:     []string{"test-app", "3.0.0", "prod"},
		},
		{
			name:     "default app name",
			lines:    []string{},
			appName:  "",
			profiles: []string{},
			version:  "4.0.0",
			want:     []string{"Application", "4.0.0"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
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

func TestLegacyBanner_Mode(t *testing.T) {
	t.Parallel()

	b := NewLegacyBanner(WithLines([]string{"test"}))
	if b.Mode() != BannerModeConsole {
		t.Errorf("expected BannerModeConsole, got %v", b.Mode())
	}

	b.(*LegacyBanner).SetMode(BannerModeOff)
	if b.Mode() != BannerModeOff {
		t.Errorf("expected BannerModeOff, got %v", b.Mode())
	}
}

func TestLegacyBanner_OffMode(t *testing.T) {
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

func TestBannerMode_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode BannerMode
		want string
	}{
		{BannerModeConsole, "console"},
		{BannerModeLog, "log"},
		{BannerModeOff, "off"},
		{BannerMode(99), "unknown"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("BannerMode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatProfiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		profiles []string
		want     string
	}{
		{"nil profiles", nil, "default"},
		{"empty profiles", []string{}, "default"},
		{"single profile", []string{"dev"}, "dev"},
		{"multiple profiles", []string{"dev", "test", "prod"}, "dev, test, prod"},
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
