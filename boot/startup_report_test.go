package boot

import (
	"strings"
	"testing"
	"time"
)

func TestStartupReport_Basic(t *testing.T) {
	t.Parallel()
	report := NewStartupReport()
	report.SetAppInfo("test-app", "1.0.0")
	report.StartTiming()
	time.Sleep(10 * time.Millisecond)
	report.StopTiming()
	report.SetBeanCount(10)
	report.SetAutoConfigCount(5)
	report.SetStarterCount(3)
	report.SetModuleCount(2)
	report.AddStarter("GinStarter")
	report.AddStarter("GormStarter")
	report.AddModule("DatabaseModule")
	report.SetServerAddr("localhost:8080")
	report.SetActuatorAddr("localhost:8080")

	output := report.format()
	if output == "" {
		t.Error("format() should not return empty string")
	}

	// 验证关键信息是否包含
	if !strings.Contains(output, "test-app") {
		t.Error("output should contain app name")
	}
	if !strings.Contains(output, "1.0.0") {
		t.Error("output should contain version")
	}
	if !strings.Contains(output, "10 个 Bean") {
		t.Error("output should contain bean count")
	}
	if !strings.Contains(output, "GinStarter") {
		t.Error("output should contain starter name")
	}
	if !strings.Contains(output, "localhost:8080") {
		t.Error("output should contain server address")
	}
}

func TestStartupReport_Disabled(t *testing.T) {
	t.Parallel()
	report := NewStartupReport()
	report.Disable()
	report.SetAppInfo("test-app", "1.0.0")
	// 不应该 panic 或输出任何内容
	report.Print()
}

func TestStartupReport_Warnings(t *testing.T) {
	t.Parallel()
	report := NewStartupReport()
	report.AddWarning("unused bean: xxx")
	report.AddWarning("unused bean: yyy")

	output := report.format()
	if !strings.Contains(output, "unused bean: xxx") {
		t.Error("output should contain warning")
	}
}

func TestStartupReport_Empty(t *testing.T) {
	t.Parallel()
	report := NewStartupReport()
	report.SetAppInfo("empty-app", "0.0.1")
	report.StartTiming()
	report.StopTiming()

	output := report.format()
	if output == "" {
		t.Error("format() should not return empty string even with no data")
	}
}

func TestStartupReport_SortStrings(t *testing.T) {
	t.Parallel()
	report := NewStartupReport()
	input := []string{"c", "a", "b"}
	sorted := report.sortStrings(input)
	if sorted[0] != "a" || sorted[1] != "b" || sorted[2] != "c" {
		t.Errorf("sortStrings() = %v, want [a b c]", sorted)
	}
}

func TestGlobalStartupReport(t *testing.T) {
	t.Parallel()
	report := GetStartupReport()
	if report == nil {
		t.Error("GetStartupReport() should not return nil")
	}

	ResetStartupReport()
	report2 := GetStartupReport()
	if report2 == nil {
		t.Error("GetStartupReport() after reset should not return nil")
	}
}
