package condition

import (
	"testing"
)

func TestOnProperty_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		key  string
		vals []string
		want string
	}{
		{"key only", "server.port", nil, "OnProperty(server.port)"},
		{"key and value", "server.port", []string{"8080"}, "OnProperty(server.port=8080)"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := OnProperty(tt.key, tt.vals...)
			if got := c.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOnMissingProperty_String(t *testing.T) {
	t.Parallel()
	c := OnMissingProperty("key")
	if got := c.String(); got != "OnMissingProperty(key)" {
		t.Errorf("String() = %q", got)
	}
}

func TestOnBean_String(t *testing.T) {
	t.Parallel()
	c := OnBean("db")
	if got := c.String(); got != "OnBean(db)" {
		t.Errorf("String() = %q", got)
	}
}

func TestOnMissingBean_String(t *testing.T) {
	t.Parallel()
	c := OnMissingBean("cache")
	if got := c.String(); got != "OnMissingBean(cache)" {
		t.Errorf("String() = %q", got)
	}
}

func TestOnProfile_String(t *testing.T) {
	t.Parallel()
	c := OnProfile("dev")
	if got := c.String(); got != "OnProfile(dev)" {
		t.Errorf("String() = %q", got)
	}
}

func TestOnProfile_NegateWithoutAcceptor(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) { return nil, false },
	}
	c := OnProfile("!prod")
	if !c.Matches(ctx) {
		t.Error("OnProfile(!prod) should match when env doesn't accept profiles")
	}
}

func TestOnModuleLoaded_String(t *testing.T) {
	t.Parallel()
	c := OnModuleLoaded("cache")
	if got := c.String(); got != "OnModuleLoaded(cache)" {
		t.Errorf("String() = %q", got)
	}
}

func TestOnMissingModule_String(t *testing.T) {
	t.Parallel()
	c := OnMissingModule("cache")
	if got := c.String(); got != "OnMissingModule(cache)" {
		t.Errorf("String() = %q", got)
	}
}

func TestCustom_Condition(t *testing.T) {
	t.Parallel()
	c := Custom("always-true", func(ctx ConditionContext) bool {
		return true
	})
	if !c.Matches(nil) {
		t.Error("Custom always-true should match")
	}
	if got := c.String(); got != "Custom(always-true)" {
		t.Errorf("String() = %q", got)
	}
}

func TestCustom_ConditionFalse(t *testing.T) {
	t.Parallel()
	c := Custom("always-false", func(ctx ConditionContext) bool {
		return false
	})
	if c.Matches(nil) {
		t.Error("Custom always-false should not match")
	}
}

func TestOnPropertyPrefix_NoPropertySource(t *testing.T) {
	t.Parallel()
	ctx := &mockConditionContext{
		envFn: func(key string) (any, bool) { return nil, false },
	}
	c := OnPropertyPrefix("app.")
	if c.Matches(ctx) {
		t.Error("OnPropertyPrefix should not match when env doesn't support PropertySources")
	}
	if got := c.String(); got != "OnPropertyPrefix(app.)" {
		t.Errorf("String() = %q", got)
	}
}
