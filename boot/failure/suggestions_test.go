package failure

import (
	"strings"
	"testing"
)

func TestGetSuggestions_FilesystemAndConfigDetail(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "config file not found",
		Action:      "check config path",
		Components:  []string{"filesystem", "config"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) != 3 {
		t.Errorf("expected 3 suggestions, got %d", len(suggestions))
	}

	if !strings.Contains(suggestions[0], "配置文件路径") {
		t.Errorf("expected config path suggestion, got %q", suggestions[0])
	}
}

func TestGetSuggestions_FilesystemOnlyDetail(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "permission denied",
		Action:      "check permissions",
		Components:  []string{"filesystem"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) != 3 {
		t.Errorf("expected 3 suggestions, got %d", len(suggestions))
	}

	if !strings.Contains(suggestions[0], "文件或目录路径") {
		t.Errorf("expected file path suggestion, got %q", suggestions[0])
	}
}

func TestGetSuggestions_UnknownDetail(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "something failed",
		Action:      "check logs",
		Components:  []string{"unknown"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) != 1 {
		t.Errorf("expected 1 suggestion, got %d", len(suggestions))
	}

	if !strings.Contains(suggestions[0], "查看错误日志") {
		t.Errorf("expected log suggestion, got %q", suggestions[0])
	}
}

func TestGetSuggestions_EmptyComponentsDetail(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "unknown failure",
		Action:      "investigate",
		Components:  []string{},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) != 1 {
		t.Errorf("expected 1 suggestion, got %d", len(suggestions))
	}
}

func TestFormatFailureAnalysis_Detail(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "Port 8080 is already in use",
		Action:      "Change the port",
		Components:  []string{"network"},
	}

	result := FormatFailureAnalysis(analysis)

	if !strings.Contains(result, "APPLICATION FAILED TO START") {
		t.Error("expected failure header")
	}
	if !strings.Contains(result, analysis.Description) {
		t.Error("expected description in output")
	}
	if !strings.Contains(result, analysis.Action) {
		t.Error("expected action in output")
	}
	if !strings.Contains(result, "Suggestions:") {
		t.Error("expected suggestions section")
	}
}

func TestFormatFailureAnalysis_WithoutMatchingDetail(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "Unknown error",
		Action:      "Check logs",
		Components:  []string{"unknown"},
	}

	result := FormatFailureAnalysis(analysis)

	if !strings.Contains(result, "Check logs") {
		t.Error("expected action in output")
	}
	if !strings.Contains(result, "Suggestions:") {
		t.Error("expected suggestions section")
	}
}

func TestContainsComponentInternal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		components []string
		target     string
		want       bool
	}{
		{"found", []string{"a", "b", "c"}, "b", true},
		{"not found", []string{"a", "c"}, "b", false},
		{"empty", []string{}, "a", false},
		{"nil", nil, "a", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := containsComponent(tt.components, tt.target)
			if got != tt.want {
				t.Errorf("containsComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}
