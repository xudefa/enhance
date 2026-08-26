package banner

import (
	"strings"
	"testing"
)

func TestASCIIArtBanner_Print_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		art     string
		color   string
		version string
		want    []string
	}{
		{
			name:    "with art and version",
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
		{
			name:    "empty version",
			art:     "test art",
			color:   "",
			version: "",
			want:    []string{"test art"},
		},
		{
			name:    "both empty",
			art:     "",
			color:   "",
			version: "",
			want:    []string{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

func TestASCIIArtBanner_Print_LogMode(t *testing.T) {
	t.Parallel()

	b := NewASCIIArtBanner("art", "")
	b.(*ASCIIArtBanner).SetMode(BannerModeLog)

	output := captureStdout(t, func() {
		if err := b.Print("1.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[enhance]") {
		t.Errorf("expected log prefix, got %q", output)
	}
	if !strings.Contains(output, "1.0.0") {
		t.Errorf("expected version in output, got %q", output)
	}
}

func TestASCIIArtBanner_Print_OffMode(t *testing.T) {
	t.Parallel()

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

func TestASCIIArtBanner_Mode_Default(t *testing.T) {
	t.Parallel()

	b := NewASCIIArtBanner("art", "red")
	if b.Mode() != BannerModeConsole {
		t.Errorf("expected BannerModeConsole, got %v", b.Mode())
	}
}

func TestASCIIArtBanner_SetMode(t *testing.T) {
	t.Parallel()

	b := NewASCIIArtBanner("art", "")
	impl := b.(*ASCIIArtBanner)

	impl.SetMode(BannerModeLog)
	if b.Mode() != BannerModeLog {
		t.Errorf("expected BannerModeLog, got %v", b.Mode())
	}

	impl.SetMode(BannerModeOff)
	if b.Mode() != BannerModeOff {
		t.Errorf("expected BannerModeOff, got %v", b.Mode())
	}
}

func TestASCIIArtBanner_MultilineArt(t *testing.T) {
	t.Parallel()

	b := NewASCIIArtBanner("line1\nline2\nline3", "")
	output := captureStdout(t, func() {
		if err := b.Print("1.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "line1") {
		t.Errorf("expected line1 in output")
	}
	if !strings.Contains(output, "line3") {
		t.Errorf("expected line3 in output")
	}
}
