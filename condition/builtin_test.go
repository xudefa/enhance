package condition

import (
	"testing"
)

func TestOnPropertyOrDefault_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		key          string
		defaultValue string
		vals         []string
		want         string
	}{
		{"with value", "gin.enabled", "true", []string{"true"}, "OnPropertyOrDefault(gin.enabled=true, default=true)"},
		{"without value", "gin.enabled", "true", nil, "OnPropertyOrDefault(gin.enabled, default=true)"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := OnPropertyOrDefault(tt.key, tt.defaultValue, tt.vals...)
			if got := c.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOnPropertyOrDefault_Matches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		key          string
		defaultValue string
		expectedVal  string
		propValue    any
		propExists   bool
		want         bool
	}{
		{
			name:         "property exists and matches",
			key:          "gin.enabled",
			defaultValue: "true",
			expectedVal:  "true",
			propValue:    "true",
			propExists:   true,
			want:         true,
		},
		{
			name:         "property exists but not matches",
			key:          "gin.enabled",
			defaultValue: "true",
			expectedVal:  "true",
			propValue:    "false",
			propExists:   true,
			want:         false,
		},
		{
			name:         "property not exists use default matches",
			key:          "gin.enabled",
			defaultValue: "true",
			expectedVal:  "true",
			propValue:    nil,
			propExists:   false,
			want:         true,
		},
		{
			name:         "property not exists use default not matches",
			key:          "gin.enabled",
			defaultValue: "false",
			expectedVal:  "true",
			propValue:    nil,
			propExists:   false,
			want:         false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &mockConditionContext{
				envFn: func(key string) (any, bool) {
					if key == tt.key {
						return tt.propValue, tt.propExists
					}
					return nil, false
				},
			}
			c := OnPropertyOrDefault(tt.key, tt.defaultValue, tt.expectedVal)
			if got := c.Matches(ctx); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
