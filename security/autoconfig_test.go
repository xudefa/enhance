package security

import (
	"testing"
)

func TestSecurityAutoConfiguration_HasFilter(t *testing.T) {
	t.Parallel()

	c := &SecurityAutoConfiguration{}

	filters := []SecurityFilter{
		NewCorsFilter(CorsConfig{AllowedOrigins: []string{"*"}}),
		NewRateLimitFilter(RateLimitConfig{Enabled: true}),
		NewAnonymousAuthenticationFilter(),
	}

	if !c.hasFilter(filters, (*CorsFilter)(nil)) {
		t.Error("expected hasFilter to find CorsFilter")
	}
	if !c.hasFilter(filters, (*RateLimitFilter)(nil)) {
		t.Error("expected hasFilter to find RateLimitFilter")
	}
	if c.hasFilter(filters, (*BasicAuthenticationFilter)(nil)) {
		t.Error("expected hasFilter to not find BasicAuthenticationFilter")
	}
}

func TestSecurityAutoConfiguration_HasFilter_Empty(t *testing.T) {
	t.Parallel()

	c := &SecurityAutoConfiguration{}

	if c.hasFilter(nil, (*CorsFilter)(nil)) {
		t.Error("expected hasFilter to return false for nil filters")
	}
	if c.hasFilter([]SecurityFilter{}, (*RateLimitFilter)(nil)) {
		t.Error("expected hasFilter to return false for empty filters")
	}
}

func TestParseStringSlice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", []string{}},
		{"single", "GET", []string{"GET"}},
		{"multiple", "GET,POST,PUT", []string{"GET", "POST", "PUT"}},
		{"with spaces", " GET , POST , PUT ", []string{"GET", "POST", "PUT"}},
		{"trailing comma", "GET,POST,", []string{"GET", "POST"}},
		{"empty parts", ",,,", []string{}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseStringSlice(tt.input)
			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d items, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("expected[%d] = %q, got %q", i, tt.expected[i], v)
				}
			}
		})
	}
}

func TestSecurityAutoConfiguration_TypeReference(t *testing.T) {
	t.Parallel()

	c := &SecurityAutoConfiguration{}
	if c == nil {
		t.Error("expected non-nil SecurityAutoConfiguration")
	}
}
