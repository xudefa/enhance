package boot

import (
	"testing"
)

func TestNewApplication_RejectsInvalidOptionType(t *testing.T) {
	t.Parallel()
	app, err := NewApplication(nil)
	if err != nil {
		t.Fatalf("NewApplication(nil) error = %v", err)
	}
	if app == nil {
		t.Fatal("expected non-nil app")
	}
	_ = app.Stop()
}

func TestWithModulesOption_AsBootOption(t *testing.T) {
	t.Parallel()
	mod := NamedModule("test", Module{})
	app, err := NewApplication(WithModulesOption(Module{}), WithAppName("x"), WithModules(mod))
	if err != nil {
		t.Fatalf("NewApplication error = %v", err)
	}
	if len(app.config.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(app.config.Modules))
	}
	_ = app.Stop()
}