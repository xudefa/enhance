package banner

import (
	"strings"
	"testing"
)

func TestTextBanner_Print_WithTemplate(t *testing.T) {
	t.Parallel()

	b := NewTextBanner("Hello {{name}} v{{version}}", map[string]any{"name": "my-app"})
	output := captureStdout(t, func() {
		if err := b.Print("1.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "my-app") {
		t.Errorf("expected 'my-app' in output, got %q", output)
	}
	if !strings.Contains(output, "1.0.0") {
		t.Errorf("expected '1.0.0' in output, got %q", output)
	}
}

func TestTextBanner_Print_EmptyTemplate(t *testing.T) {
	t.Parallel()

	b := NewTextBanner("", nil)
	output := captureStdout(t, func() {
		if err := b.Print("2.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "2.0.0") {
		t.Errorf("expected '2.0.0' in output, got %q", output)
	}
	if !strings.Contains(output, "Application") {
		t.Errorf("expected 'Application' in output, got %q", output)
	}
}

func TestTextBanner_Print_NilProperties(t *testing.T) {
	t.Parallel()

	b := NewTextBanner("App v{{version}}", nil)
	output := captureStdout(t, func() {
		if err := b.Print("3.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "3.0.0") {
		t.Errorf("expected '3.0.0' in output, got %q", output)
	}
}

func TestTextBanner_Print_LogMode(t *testing.T) {
	t.Parallel()

	b := NewTextBanner("test", nil)
	b.(*TextBanner).SetMode(BannerModeLog)

	output := captureStdout(t, func() {
		if err := b.Print("1.0.0"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "[enhance]") {
		t.Errorf("expected log prefix, got %q", output)
	}
}

func TestTextBanner_Print_OffMode(t *testing.T) {
	t.Parallel()

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

func TestTextBanner_Mode_Default(t *testing.T) {
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

func TestTextBanner_Render_MultiplePlaceholders(t *testing.T) {
	t.Parallel()

	b := NewTextBanner("{{a}}-{{b}}-{{c}}", map[string]any{
		"a": "x",
		"b": "y",
		"c": "z",
	})

	output := captureStdout(t, func() {
		_ = b.Print("1.0.0")
	})

	if !strings.Contains(output, "x-y-z") {
		t.Errorf("expected 'x-y-z' in output, got %q", output)
	}
}

func TestTextBanner_Render_VersionOverwrite(t *testing.T) {
	t.Parallel()

	b := NewTextBanner("v{{version}}", map[string]any{
		"version": "old",
	})

	output := captureStdout(t, func() {
		_ = b.Print("new")
	})

	if !strings.Contains(output, "vnew") {
		t.Errorf("expected 'vnew' in output, got %q", output)
	}
}
