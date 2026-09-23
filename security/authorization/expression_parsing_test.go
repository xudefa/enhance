package authorization

import (
	"testing"
)

func TestExtractExpressionArg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attr     string
		prefix   string
		suffix   string
		expected string
	}{
		{
			name:     "hasRole",
			attr:     "hasRole('ADMIN')",
			prefix:   "hasRole('",
			suffix:   "')",
			expected: "ADMIN",
		},
		{
			name:     "hasAuthority",
			attr:     "hasAuthority('read')",
			prefix:   "hasAuthority('",
			suffix:   "')",
			expected: "read",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			extractedArg := extractExpressionArg(tt.attr, tt.prefix, tt.suffix)
			if extractedArg != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, extractedArg)
			}
		})
	}
}

func TestSplitExpressionArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attr     string
		prefix   string
		suffix   string
		expected []string
	}{
		{
			name:     "two roles",
			attr:     "hasAnyRole('ADMIN','USER')",
			prefix:   "hasAnyRole('",
			suffix:   "')",
			expected: []string{"ADMIN", "USER"},
		},
		{
			name:     "three authorities",
			attr:     "hasAnyAuthority('read','write','delete')",
			prefix:   "hasAnyAuthority('",
			suffix:   "')",
			expected: []string{"read", "write", "delete"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			parsedArgs := splitExpressionArgs(tt.attr, tt.prefix, tt.suffix)
			if len(parsedArgs) != len(tt.expected) {
				t.Fatalf("expected %d args, got %d", len(tt.expected), len(parsedArgs))
			}
			for i, arg := range parsedArgs {
				if arg != tt.expected[i] {
					t.Errorf("expected arg[%d] = %q, got %q", i, tt.expected[i], arg)
				}
			}
		})
	}
}

func TestJoinStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		strs     []string
		sep      string
		expected string
	}{
		{
			name:     "empty",
			strs:     []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "single",
			strs:     []string{"a"},
			sep:      ",",
			expected: "a",
		},
		{
			name:     "multiple",
			strs:     []string{"a", "b", "c"},
			sep:      ",",
			expected: "a,b,c",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			joinResult := joinStrings(tt.strs, tt.sep)
			if joinResult != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, joinResult)
			}
		})
	}
}
