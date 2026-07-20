package boot

import (
	"strings"
	"testing"
)

func TestConditionEvaluationReport_String(t *testing.T) {
	t.Parallel()
	report := NewConditionEvaluationReport()

	report.RecordPositiveMatch("ActuatorAutoConfiguration", []ConditionResult{
		{
			Condition: "OnProperty",
			Matched:   true,
			Message:   "@ConditionalOnProperty (actuator.enabled=true) matched",
		},
	})

	report.RecordNegativeMatch("RedisAutoConfiguration", []ConditionResult{
		{
			Condition: "OnClass",
			Matched:   false,
			Message:   "@ConditionalOnClass (redis.Client) not found",
		},
	})

	report.RecordExclusion("SecurityAutoConfiguration")
	report.RecordUnconditional("CoreAutoConfiguration")

	output := report.String()

	// 验证报告包含所有部分
	if !strings.Contains(output, "AUTO-CONFIGURATION REPORT") {
		t.Error("Report should contain header")
	}

	if !strings.Contains(output, "Positive matches") {
		t.Error("Report should contain positive matches section")
	}

	if !strings.Contains(output, "ActuatorAutoConfiguration matched") {
		t.Error("Report should contain ActuatorAutoConfiguration")
	}

	if !strings.Contains(output, "Negative matches") {
		t.Error("Report should contain negative matches section")
	}

	if !strings.Contains(output, "RedisAutoConfiguration did not match") {
		t.Error("Report should contain RedisAutoConfiguration")
	}

	if !strings.Contains(output, "Exclusions") {
		t.Error("Report should contain exclusions section")
	}

	if !strings.Contains(output, "SecurityAutoConfiguration") {
		t.Error("Report should contain excluded configuration")
	}

	if !strings.Contains(output, "Unconditional classes") {
		t.Error("Report should contain unconditional classes section")
	}

	if !strings.Contains(output, "CoreAutoConfiguration") {
		t.Error("Report should contain unconditional configuration")
	}
}

func TestConditionEvaluationReport_EmptySections(t *testing.T) {
	t.Parallel()
	report := NewConditionEvaluationReport()
	output := report.String()

	// 验证空报告包含 "(none)" 标记
	if !strings.Contains(output, "(none)") {
		t.Error("Empty report should contain '(none)' markers")
	}
}

func TestConditionEvaluationReport_Sorting(t *testing.T) {
	t.Parallel()
	report := NewConditionEvaluationReport()

	// 添加无序的匹配项
	report.RecordPositiveMatch("ZAutoConfiguration", []ConditionResult{})
	report.RecordPositiveMatch("AAutoConfiguration", []ConditionResult{})
	report.RecordPositiveMatch("MAutoConfiguration", []ConditionResult{})

	output := report.String()

	// 验证按字母顺序排序
	aIdx := strings.Index(output, "AAutoConfiguration")
	mIdx := strings.Index(output, "MAutoConfiguration")
	zIdx := strings.Index(output, "ZAutoConfiguration")

	if aIdx > mIdx || mIdx > zIdx {
		t.Error("Report should sort configurations alphabetically")
	}
}

func TestGlobalReportFunctions(t *testing.T) {
	// 测试全局开关
	origEnabled := IsAutoConfigReportEnabled()
	defer func() {
		if origEnabled {
			EnableAutoConfigReport()
		} else {
			DisableAutoConfigReport()
		}
	}()
	DisableAutoConfigReport()
	if IsAutoConfigReportEnabled() {
		t.Error("Report should be disabled by default")
	}

	EnableAutoConfigReport()
	if !IsAutoConfigReportEnabled() {
		t.Error("Report should be enabled after EnableAutoConfigReport()")
	}

	// 测试全局报告实例
	report := GetAutoConfigReport()
	if report == nil {
		t.Error("GetAutoConfigReport should return non-nil instance")
	}

	// 测试重置
	ResetAutoConfigReport()
	report = GetAutoConfigReport()
	if report == nil {
		t.Error("GetAutoConfigReport should return non-nil instance after reset")
	}
}

func TestTypeName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected string
	}{
		{"string", "string"},
		{123, "int"},
		{struct{}{}, "struct {}"},
	}

	for _, tt := range tests {
		result := typeName(tt.input)
		if result != tt.expected {
			t.Errorf("typeName(%T) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}
