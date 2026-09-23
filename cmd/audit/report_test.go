package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRenderReport_Structure(t *testing.T) {
	t.Parallel()
	results := []PackageResult{
		{
			Dir: "core",
			Findings: []Finding{
				{File: "core/a.go", Line: 12, Category: CategoryBadNames, Severity: SeverityMechanical, Message: "单字母局部变量", Suggestion: "改名"},
				{File: "core/a.go", Line: 5, Category: CategoryMaintainability, Severity: SeverityBreaking, Message: "参数超限", Suggestion: "选项模式"},
			},
		},
	}
	stats := Stats{Packages: 1, Files: 1, TotalFindings: 2}
	report := renderReport(".", results, time.Date(2026, 9, 12, 10, 0, 0, 0, time.Local), stats)

	for _, want := range []string{"# AI 可读性与可维护性审计报告", "## 汇总", "### core", "[bad_names]", "[maintainability]", "- [ ] `core/a.go:12`"} {
		if !strings.Contains(report, want) {
			t.Errorf("报告缺少 %q", want)
		}
	}
	// Breaking 应先于 Mechanical 渲染
	breakIdx := strings.Index(report, "#### 待决策（Breaking）")
	mecIdx := strings.Index(report, "#### 机械可改（Mechanical）")
	if breakIdx == -1 || mecIdx == -1 || breakIdx > mecIdx {
		t.Errorf("Breaking 分组应出现在 Mechanical 之前")
	}
}

func TestRenderReport_GroupsTestFindings(t *testing.T) {
	t.Parallel()
	results := []PackageResult{
		{
			Dir: "core",
			TestFindings: []Finding{
				{File: "core/a_test.go", Line: 3, Category: CategoryMaintainability, Severity: SeverityMechanical, Message: "文件超长", Suggestion: "拆分"},
			},
		},
	}
	report := renderReport(".", results, time.Now(), Stats{TotalFindings: 1})
	if !strings.Contains(report, "#### 测试文件") {
		t.Errorf("报告应含测试文件分组")
	}
}

func TestWriteJSONReport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.json")
	err := writeJSONReport(path, time.Now(), Stats{Packages: 1, Files: 1, TotalFindings: 1},
		[]PackageResult{{Dir: "x", Findings: []Finding{{File: "x/a.go", Line: 1, Category: CategoryErrorContext, Severity: SeverityMechanical, Message: "m", Suggestion: "s"}}}})
	if err != nil {
		t.Fatalf("writeJSONReport error: %v", err)
	}
	reportData, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(reportData, &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	findings, ok := parsed["findings"].([]any)
	if !ok || len(findings) == 0 {
		t.Fatalf("findings should be a non-empty array, got %T (%v)", parsed["findings"], parsed["findings"])
	}
	first, ok := findings[0].(map[string]any)
	if !ok {
		t.Fatalf("findings[0] should be an object, got %T", findings[0])
	}
	if got := first["category"]; got != "error_context" {
		t.Errorf("findings[0].category = %v, want error_context (lowercase key)", got)
	}
}
