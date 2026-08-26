package tenant

import "testing"

func TestExtractSubdomain_Helpers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		host       string
		baseDomain string
		want       string
	}{
		{"basic subdomain", "tenant1.example.com", "example.com", "tenant1"},
		{"multi-level subdomain", "a.b.example.com", "example.com", "a.b"},
		{"no subdomain", "example.com", "example.com", ""},
		{"with port", "tenant1.example.com:8080", "example.com", "tenant1"},
		{"no match", "tenant1.other.com", "example.com", ""},
		{"empty host", "", "example.com", ""},
		{"empty base", "tenant1.example.com", "", ""},
		{"exact match minus dot", "example.com", "example.com", ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractSubdomain(tt.host, tt.baseDomain)
			if got != tt.want {
				t.Errorf("extractSubdomain(%q, %q) = %q, want %q", tt.host, tt.baseDomain, got, tt.want)
			}
		})
	}
}

func TestSplitPath_Helpers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		want []string
	}{
		{"empty", "", []string{}},
		{"root", "/", []string{}},
		{"single", "/health", []string{"health"}},
		{"multi", "/api/v1/users", []string{"api", "v1", "users"}},
		{"trailing slash", "/api/v1/", []string{"api", "v1"}},
		{"double slash", "/a//b", []string{"a", "b"}},
		{"no leading slash", "a/b/c", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitPath(tt.path)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d (%v vs %v)", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIndexOf_Helpers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		c    byte
		want int
	}{
		{"found", "hello", 'l', 2},
		{"not found", "hello", 'x', -1},
		{"first char", "hello", 'h', 0},
		{"last char", "hello", 'o', 4},
		{"empty", "", 'a', -1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := indexOf(tt.s, tt.c); got != tt.want {
				t.Errorf("indexOf(%q, %c) = %d, want %d", tt.s, tt.c, got, tt.want)
			}
		})
	}
}

func TestEndsWith_Helpers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		s      string
		suffix string
		want   bool
	}{
		{"basic match", "hello", "llo", true},
		{"full match", "hello", "hello", true},
		{"no match", "hello", "world", false},
		{"suffix longer", "hi", "hello", false},
		{"empty suffix", "hello", "", true},
		{"empty both", "", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := endsWith(tt.s, tt.suffix); got != tt.want {
				t.Errorf("endsWith(%q, %q) = %v, want %v", tt.s, tt.suffix, got, tt.want)
			}
		})
	}
}
