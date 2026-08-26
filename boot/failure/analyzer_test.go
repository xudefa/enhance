package failure

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestNewDefaultFailureAnalyzer(t *testing.T) {
	t.Parallel()
	analyzer := NewDefaultFailureAnalyzer()
	if analyzer == nil {
		t.Fatal("expected non-nil analyzer")
	}
}

func TestDefaultFailureAnalyzer_Supports(t *testing.T) {
	t.Parallel()
	analyzer := NewDefaultFailureAnalyzer()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"port in use", errors.New("listen tcp :8080: bind: address already in use"), true},
		{"permission denied", errors.New("open /etc/passwd: permission denied"), true},
		{"os.ErrPermission", os.ErrPermission, true},
		{"file not found", errors.New("open config.yaml: file not found"), true},
		{"no such file", errors.New("open config.yaml: no such file or directory"), true},
		{"other error", errors.New("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := analyzer.Supports(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDefaultFailureAnalyzer_Analyze_Nil(t *testing.T) {
	t.Parallel()
	analyzer := NewDefaultFailureAnalyzer()
	result := analyzer.Analyze(nil)
	if result != nil {
		t.Error("expected nil analysis for nil error")
	}
}

func TestDefaultFailureAnalyzer_Analyze_PortInUse(t *testing.T) {
	t.Parallel()
	analyzer := NewDefaultFailureAnalyzer()
	err := errors.New("listen tcp :8080: bind: address already in use")
	result := analyzer.Analyze(err)

	if result == nil {
		t.Fatal("expected non-nil analysis")
	}
	if result.Description == "" {
		t.Error("expected non-empty description")
	}
	if result.Action == "" {
		t.Error("expected non-empty action")
	}
	if len(result.Components) == 0 {
		t.Error("expected non-empty components")
	}
	if result.Exception != err {
		t.Error("expected exception to match input error")
	}
}

func TestDefaultFailureAnalyzer_Analyze_PermissionDenied(t *testing.T) {
	t.Parallel()
	analyzer := NewDefaultFailureAnalyzer()
	err := errors.New("open /etc/passwd: permission denied")
	result := analyzer.Analyze(err)

	if result == nil {
		t.Fatal("expected non-nil analysis")
	}
	if !containsComponent(result.Components, "filesystem") {
		t.Error("expected filesystem component")
	}
}

func TestDefaultFailureAnalyzer_Analyze_FileNotFound(t *testing.T) {
	t.Parallel()
	analyzer := NewDefaultFailureAnalyzer()
	err := errors.New("open config.yaml: no such file or directory")
	result := analyzer.Analyze(err)

	if result == nil {
		t.Fatal("expected non-nil analysis")
	}
	if !containsComponent(result.Components, "filesystem") {
		t.Error("expected filesystem component")
	}
	if !containsComponent(result.Components, "config") {
		t.Error("expected config component")
	}
}

func TestDefaultFailureAnalyzer_Analyze_Unsupported(t *testing.T) {
	t.Parallel()
	analyzer := NewDefaultFailureAnalyzer()
	err := errors.New("some unsupported error")
	result := analyzer.Analyze(err)

	if result != nil {
		t.Error("expected nil analysis for unsupported error")
	}
}

func TestGetSuggestions_Nil(t *testing.T) {
	t.Parallel()
	result := GetSuggestions(nil)
	if result != nil {
		t.Error("expected nil suggestions for nil analysis")
	}
}

func TestGetSuggestions_Network(t *testing.T) {
	t.Parallel()
	analysis := &FailureAnalysis{
		Description: "Port in use",
		Action:      "Change port",
		Components:  []string{"server", "network"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) == 0 {
		t.Fatal("expected non-empty suggestions")
	}
	if len(suggestions) < 3 {
		t.Errorf("expected at least 3 suggestions, got %d", len(suggestions))
	}
}

func TestGetSuggestions_FilesystemConfig(t *testing.T) {
	t.Parallel()
	analysis := &FailureAnalysis{
		Description: "File not found",
		Action:      "Check path",
		Components:  []string{"filesystem", "config"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) < 3 {
		t.Errorf("expected at least 3 suggestions, got %d", len(suggestions))
	}
}

func TestGetSuggestions_Filesystem(t *testing.T) {
	t.Parallel()
	analysis := &FailureAnalysis{
		Description: "Permission denied",
		Action:      "Check permissions",
		Components:  []string{"filesystem"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) < 3 {
		t.Errorf("expected at least 3 suggestions, got %d", len(suggestions))
	}
}

func TestGetSuggestions_Default(t *testing.T) {
	t.Parallel()
	analysis := &FailureAnalysis{
		Description: "Unknown error",
		Action:      "Check logs",
		Components:  []string{"other"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) != 1 {
		t.Errorf("expected 1 suggestion, got %d", len(suggestions))
	}
}

func TestFormatFailureAnalysis_Nil(t *testing.T) {
	t.Parallel()
	result := FormatFailureAnalysis(nil)
	if result != "" {
		t.Errorf("expected empty string, got %s", result)
	}
}

func TestFormatFailureAnalysis_WithSuggestions(t *testing.T) {
	t.Parallel()
	analysis := &FailureAnalysis{
		Description: "Port in use",
		Action:      "Change port",
		Components:  []string{"server", "network"},
	}

	result := FormatFailureAnalysis(analysis)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	if !containsSubstring(result, "APPLICATION FAILED TO START") {
		t.Error("expected header in output")
	}
	if !containsSubstring(result, "Port in use") {
		t.Error("expected description in output")
	}
	if !containsSubstring(result, "Change port") {
		t.Error("expected action in output")
	}
	if !containsSubstring(result, "Suggestions:") {
		t.Error("expected suggestions section in output")
	}
}

func TestFormatFailureAnalysis_WithoutSuggestions(t *testing.T) {
	t.Parallel()
	analysis := &FailureAnalysis{
		Description: "Unknown error",
		Action:      "Check logs",
		Components:  []string{"other"},
	}

	result := FormatFailureAnalysis(analysis)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	if !containsSubstring(result, "APPLICATION FAILED TO START") {
		t.Error("expected header in output")
	}
}

func TestContainsComponent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		components []string
		target     string
		expected   bool
	}{
		{"found", []string{"a", "b", "c"}, "b", true},
		{"not found", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"nil slice", nil, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := containsComponent(tt.components, tt.target)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && len(substr) > 0 && searchSubstring(s, substr)))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFailureAnalysis_Struct(t *testing.T) {
	t.Parallel()
	err := errors.New("test error")
	analysis := &FailureAnalysis{
		Description: "Test description",
		Action:      "Test action",
		Exception:   err,
		Components:  []string{"component1", "component2"},
	}

	if analysis.Description != "Test description" {
		t.Errorf("expected description 'Test description', got %s", analysis.Description)
	}
	if analysis.Action != "Test action" {
		t.Errorf("expected Action 'Test action', got %s", analysis.Action)
	}
	if analysis.Exception != err {
		t.Error("expected Exception to match")
	}
	if len(analysis.Components) != 2 {
		t.Errorf("expected 2 Components, got %d", len(analysis.Components))
	}
}

func TestDefaultFailureAnalyzer_Analyze_OsErrPermission(t *testing.T) {
	t.Parallel()
	analyzer := NewDefaultFailureAnalyzer()
	result := analyzer.Analyze(os.ErrPermission)

	if result == nil {
		t.Fatal("expected non-nil analysis for os.ErrPermission")
	}
	if !containsComponent(result.Components, "filesystem") {
		t.Error("expected filesystem component")
	}
}

func TestGetSuggestions_MultipleComponents(t *testing.T) {
	t.Parallel()
	analysis := &FailureAnalysis{
		Description: "Complex error",
		Action:      "Fix it",
		Components:  []string{"network", "filesystem", "config"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) == 0 {
		t.Fatal("expected non-empty suggestions")
	}
	// Should match network component first
	if len(suggestions) < 3 {
		t.Errorf("expected at least 3 suggestions for network, got %d", len(suggestions))
	}
}

func TestFormatFailureAnalysis_EmptyComponents(t *testing.T) {
	t.Parallel()
	analysis := &FailureAnalysis{
		Description: "Unknown error",
		Action:      "Check logs",
		Components:  []string{},
	}

	result := FormatFailureAnalysis(analysis)
	if result == "" {
		t.Fatal("expected non-empty result")
	}
	if !containsSubstring(result, "APPLICATION FAILED TO START") {
		t.Error("expected header in output")
	}
}

func TestDefaultFailureAnalyzer_Supports_WrappedError(t *testing.T) {
	t.Parallel()
	analyzer := NewDefaultFailureAnalyzer()
	innerErr := errors.New("address already in use")
	wrappedErr := fmt.Errorf("failed to start server: %w", innerErr)

	result := analyzer.Supports(wrappedErr)
	if !result {
		t.Error("expected to support wrapped error with port in use message")
	}
}
