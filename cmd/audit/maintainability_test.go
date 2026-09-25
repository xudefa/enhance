package main

import (
	"strings"
	"testing"
)

func TestAnalyzeMaintainability_LongFile(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, "package p\n"+strings.Repeat("// filler\n", 520), false)
	findings := analyzeMaintainability(f)
	if len(findings) == 0 {
		t.Fatalf("超长文件应被报告")
	}
	found := false
	for _, fd := range findings {
		if strings.Contains(fd.Message, "行") && fd.Category == CategoryMaintainability {
			found = true
		}
	}
	if !found {
		t.Fatalf("未找到文件超长发现: %+v", findings)
	}
}

func TestAnalyzeMaintainability_LongFunc(t *testing.T) {
	t.Parallel()
	var sb strings.Builder
	sb.WriteString("package p\nfunc F() int {\n")
	for i := 0; i < 90; i++ {
		sb.WriteString("\tvalue := 0\n")
	}
	sb.WriteString("\treturn 0\n}\n")
	f := parseSnippet(t, sb.String(), false)
	findings := analyzeMaintainability(f)
	found := false
	for _, fd := range findings {
		if strings.Contains(fd.Message, "80 行") {
			found = true
		}
	}
	if !found {
		t.Fatalf("未找到超长函数发现: %+v", findings)
	}
}

func TestAnalyzeMaintainability_ShortFunc(t *testing.T) {
	t.Parallel()
	var sb strings.Builder
	sb.WriteString("package p\nfunc F() int {\n")
	for i := 0; i < 60; i++ {
		sb.WriteString("\tvalue := 0\n")
	}
	sb.WriteString("\treturn 0\n}\n")
	f := parseSnippet(t, sb.String(), false)
	if findings := analyzeMaintainability(f); len(findings) != 0 {
		t.Fatalf("60 行函数不应报告（规范 ≤80 行）: %+v", findings)
	}
}

func TestAnalyzeMaintainability_LongTestFile(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, "package p\n"+strings.Repeat("func TestX() {}\n", 1100), true)
	if findings := analyzeMaintainability(f); len(findings) == 0 {
		t.Fatalf("超 1000 行的测试文件应被报告")
	}
}

func TestAnalyzeMaintainability_TestFileWithinLimit(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, "package p\n"+strings.Repeat("func TestX() {}\n", 600), true)
	if findings := analyzeMaintainability(f); len(findings) != 0 {
		t.Fatalf("600 行测试文件不应报告（测试 ≤1000 行）: %+v", findings)
	}
}

func TestAnalyzeMaintainability_TooManyParams(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F(a int, b int, c int, d int, e int) int { return 0 }`, false)
	findings := analyzeMaintainability(f)
	if len(findings) != 1 {
		t.Fatalf("期望 1 条参数超限，得到 %+v", findings)
	}
	if findings[0].Severity != SeverityBreaking {
		t.Errorf("导出函数参数超限应为 Breaking: %v", findings[0].Severity)
	}
}

func TestAnalyzeMaintainability_TooManyParamsUnexported(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func f(a int, b int, c int, d int, e int) int { return 0 }`, false)
	findings := analyzeMaintainability(f)
	if len(findings) != 1 || findings[0].Severity != SeverityMechanical {
		t.Fatalf("非导出函数参数超限应为 Mechanical: %+v", findings)
	}
}

func TestAnalyzeMaintainability_ShortFile(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, "package p\nfunc F() int { return 0 }\n", false)
	if findings := analyzeMaintainability(f); len(findings) != 0 {
		t.Fatalf("正常文件不应报告: %+v", findings)
	}
}
