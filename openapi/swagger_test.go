package openapi

import (
	"testing"
)

func TestSwaggerUIHTML(t *testing.T) {
	t.Parallel()
	html := swaggerUIHTML("http://localhost:8080/openapi.json")
	if html == "" {
		t.Fatal("expected non-empty HTML")
	}
	if !containsSubstring(html, "Swagger UI") {
		t.Error("expected HTML to contain 'Swagger UI'")
	}
	if !containsSubstring(html, "http://localhost:8080/openapi.json") {
		t.Error("expected HTML to contain the openapi.json URL")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && searchSubstring(s, substr))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
