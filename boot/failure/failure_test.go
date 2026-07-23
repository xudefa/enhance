package failure

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestDefaultFailureAnalyzer_Supports(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "port in use error",
			err:  fmt.Errorf("listen tcp :8080: bind: address already in use"),
			want: true,
		},
		{
			name: "permission denied error",
			err:  os.ErrPermission,
			want: true,
		},
		{
			name: "permission denied wrapped",
			err:  fmt.Errorf("open /etc/config: %w", os.ErrPermission),
			want: true,
		},
		{
			name: "file not found error",
			err:  fmt.Errorf("open application.json: file not found"),
			want: true,
		},
		{
			name: "no such file error",
			err:  fmt.Errorf("stat config.yaml: no such file or directory"),
			want: true,
		},
		{
			name: "unsupported error",
			err:  fmt.Errorf("some unrelated error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			analyzer := NewDefaultFailureAnalyzer()
			if got := analyzer.Supports(tt.err); got != tt.want {
				t.Errorf("Supports() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultFailureAnalyzer_Analyze(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		err            error
		wantDesc       string
		wantAction     string
		wantComponents []string
	}{
		{
			name:           "port in use",
			err:            fmt.Errorf("listen tcp :8080: bind: address already in use"),
			wantDesc:       "服务器端口已被占用，无法启动",
			wantAction:     "检查端口是否被其他进程占用，或更换端口",
			wantComponents: []string{"server", "network"},
		},
		{
			name:           "permission denied",
			err:            os.ErrPermission,
			wantDesc:       "权限不足，无法访问所需资源",
			wantAction:     "检查文件或目录的访问权限，或以适当权限运行应用",
			wantComponents: []string{"filesystem"},
		},
		{
			name:           "file not found",
			err:            fmt.Errorf("open application.json: file not found"),
			wantDesc:       "配置文件或资源文件未找到",
			wantAction:     "检查文件路径是否正确，确保文件存在",
			wantComponents: []string{"filesystem", "config"},
		},
		{
			name:           "no such file",
			err:            fmt.Errorf("stat config.yaml: no such file or directory"),
			wantDesc:       "配置文件或资源文件未找到",
			wantAction:     "检查文件路径是否正确，确保文件存在",
			wantComponents: []string{"filesystem", "config"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			analyzer := NewDefaultFailureAnalyzer()
			result := analyzer.Analyze(tt.err)
			if result == nil {
				t.Fatal("Analyze() returned nil")
			}
			if result.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", result.Description, tt.wantDesc)
			}
			if result.Action != tt.wantAction {
				t.Errorf("Action = %q, want %q", result.Action, tt.wantAction)
			}
			if result.Exception != tt.err {
				t.Errorf("Exception not preserved")
			}
			if len(result.Components) != len(tt.wantComponents) {
				t.Errorf("Components length = %d, want %d", len(result.Components), len(tt.wantComponents))
			} else {
				for i, c := range result.Components {
					if c != tt.wantComponents[i] {
						t.Errorf("Components[%d] = %q, want %q", i, c, tt.wantComponents[i])
					}
				}
			}
		})
	}
}

func TestDefaultFailureAnalyzer_Analyze_NilError(t *testing.T) {
	t.Parallel()

	analyzer := NewDefaultFailureAnalyzer()
	result := analyzer.Analyze(nil)
	if result != nil {
		t.Errorf("Analyze(nil) = %v, want nil", result)
	}
}

func TestDefaultFailureAnalyzer_Analyze_UnsupportedError(t *testing.T) {
	t.Parallel()

	analyzer := NewDefaultFailureAnalyzer()
	result := analyzer.Analyze(errors.New("unknown error"))
	if result != nil {
		t.Errorf("Analyze(unsupported) = %v, want nil", result)
	}
}

func TestGetSuggestions_NilAnalysis(t *testing.T) {
	t.Parallel()

	suggestions := GetSuggestions(nil)
	if suggestions != nil {
		t.Errorf("GetSuggestions(nil) = %v, want nil", suggestions)
	}
}

func TestGetSuggestions_NetworkComponent(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "port in use",
		Action:      "check port",
		Components:  []string{"server", "network"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) == 0 {
		t.Fatal("GetSuggestions() returned empty")
	}

	found := false
	for _, s := range suggestions {
		if containsSubstr(s, "lsof") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected suggestion containing 'lsof'")
	}
}

func TestGetSuggestions_FilesystemConfigComponent(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "config not found",
		Action:      "check path",
		Components:  []string{"filesystem", "config"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) == 0 {
		t.Fatal("GetSuggestions() returned empty")
	}

	found := false
	for _, s := range suggestions {
		if containsSubstr(s, "配置文件") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected suggestion about config file")
	}
}

func TestGetSuggestions_FilesystemOnly(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "permission denied",
		Action:      "check permissions",
		Components:  []string{"filesystem"},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) == 0 {
		t.Fatal("GetSuggestions() returned empty")
	}

	found := false
	for _, s := range suggestions {
		if containsSubstr(s, "chmod") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected suggestion containing 'chmod'")
	}
}

func TestGetSuggestions_DefaultComponent(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "unknown error",
		Action:      "check logs",
		Components:  []string{},
	}

	suggestions := GetSuggestions(analysis)
	if len(suggestions) == 0 {
		t.Fatal("GetSuggestions() returned empty")
	}

	found := false
	for _, s := range suggestions {
		if containsSubstr(s, "日志") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected default log suggestion")
	}
}

func TestFormatFailureAnalysis(t *testing.T) {
	t.Parallel()

	analysis := &FailureAnalysis{
		Description: "test description",
		Action:      "test action",
		Exception:   fmt.Errorf("test error"),
		Components:  []string{"server"},
	}

	result := FormatFailureAnalysis(analysis)
	if result == "" {
		t.Fatal("FormatFailureAnalysis() returned empty")
	}

	if !containsSubstr(result, "APPLICATION FAILED TO START") {
		t.Error("expected header in output")
	}
	if !containsSubstr(result, "test description") {
		t.Error("expected description in output")
	}
	if !containsSubstr(result, "test action") {
		t.Error("expected action in output")
	}
}

func TestFormatFailureAnalysis_Nil(t *testing.T) {
	t.Parallel()

	result := FormatFailureAnalysis(nil)
	if result != "" {
		t.Errorf("FormatFailureAnalysis(nil) = %q, want empty", result)
	}
}

func TestContainsComponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		components []string
		target     string
		want       bool
	}{
		{"present", []string{"a", "b", "c"}, "b", true},
		{"absent", []string{"a", "b", "c"}, "d", false},
		{"empty list", []string{}, "a", false},
		{"nil list", nil, "a", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := containsComponent(tt.components, tt.target); got != tt.want {
				t.Errorf("containsComponent() = %v, want %v", got, tt.want)
			}
		})
	}
}

// containsSubstr 检查字符串是否包含子串。
func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstrHelper(s, substr))
}

func containsSubstrHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
