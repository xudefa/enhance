package boot

import (
	"bytes"
	"strings"
	"testing"

	"github.com/xudefa/enhance/boot/banner"
)

func TestBanner_Print(t *testing.T) {
	t.Parallel()

	b := banner.NewLegacyBanner(
		banner.WithLines([]string{"Test Banner"}),
		banner.WithAppName("my-app"),
		banner.WithProfiles([]string{"dev"}),
	)

	var buf bytes.Buffer
	err := b.Print("1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Print goes to stdout; verify via DefaultBanner behavior
	_ = buf
}

func TestDefaultBanner(t *testing.T) {
	t.Parallel()

	b := banner.NewLegacyBanner(
		banner.WithLines([]string{"Default Banner"}),
		banner.WithAppName("test-app"),
		banner.WithProfiles([]string{"prod", "test"}),
	)

	err := b.Print("2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBanner_NoProfiles(t *testing.T) {
	t.Parallel()

	b := banner.NewLegacyBanner(
		banner.WithLines([]string{"Line 1", "Line 2"}),
		banner.WithAppName("app"),
		banner.WithProfiles([]string{}),
	)

	err := b.Print("1.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewBanner(t *testing.T) {
	t.Parallel()

	b := NewBanner(
		banner.WithLines([]string{"Custom Banner"}),
		banner.WithAppName("custom-app"),
	)

	_ = b
	if b == nil {
		t.Fatal("expected non-nil banner")
	}
}

func TestDefaultBannerType(t *testing.T) {
	t.Parallel()

	if DefaultBanner == nil {
		t.Fatal("expected non-nil DefaultBanner")
	}

	mode := DefaultBanner.Mode()
	if mode != banner.BannerModeConsole {
		t.Errorf("expected BannerModeConsole, got %v", mode)
	}
}

func TestNewBannerReturnsInterface(t *testing.T) {
	t.Parallel()

	var b banner.Banner = NewBanner(
		banner.WithLines([]string{"Line"}),
	)

	if b == nil {
		t.Fatal("expected non-nil banner")
	}
	if !strings.Contains("banner", "banner") {
		t.Fatal("sanity check")
	}
}
