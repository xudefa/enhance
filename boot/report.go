package boot

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
)

// ConditionEvaluationReport 条件评估报告
//
// 记录所有自动配置的匹配情况，包括正面匹配、负面匹配、排除项和无条件类。
type ConditionEvaluationReport struct {
	mu              sync.RWMutex
	positiveMatches []AutoConfigMatchResult
	negativeMatches []AutoConfigMatchResult
	exclusions      []string
	unconditional   []string
}

// AutoConfigMatchResult 自动配置匹配结果
type AutoConfigMatchResult struct {
	// Name 自动配置名称
	Name string
	// Matched 是否匹配成功
	Matched bool
	// Conditions 条件列表及匹配结果
	Conditions []ConditionResult
}

// ConditionResult 单个条件的匹配结果
type ConditionResult struct {
	// Condition 条件描述（如 "@ConditionalOnProperty"）
	Condition string
	// Matched 是否匹配
	Matched bool
	// Message 详细信息
	Message string
}

// NewConditionEvaluationReport 创建条件评估报告
func NewConditionEvaluationReport() *ConditionEvaluationReport {
	return &ConditionEvaluationReport{
		positiveMatches: make([]AutoConfigMatchResult, 0),
		negativeMatches: make([]AutoConfigMatchResult, 0),
		exclusions:      make([]string, 0),
		unconditional:   make([]string, 0),
	}
}

// RecordPositiveMatch 记录正面匹配
func (r *ConditionEvaluationReport) RecordPositiveMatch(name string, conditions []ConditionResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.positiveMatches = append(r.positiveMatches, AutoConfigMatchResult{
		Name:       name,
		Matched:    true,
		Conditions: conditions,
	})
}

// RecordNegativeMatch 记录负面匹配
func (r *ConditionEvaluationReport) RecordNegativeMatch(name string, conditions []ConditionResult) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.negativeMatches = append(r.negativeMatches, AutoConfigMatchResult{
		Name:       name,
		Matched:    false,
		Conditions: conditions,
	})
}

// RecordExclusion 记录排除项
func (r *ConditionEvaluationReport) RecordExclusion(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.exclusions = append(r.exclusions, name)
}

// RecordUnconditional 记录无条件类
func (r *ConditionEvaluationReport) RecordUnconditional(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.unconditional = append(r.unconditional, name)
}

// Print 打印报告到标准输出
func (r *ConditionEvaluationReport) Print() {
	fmt.Println(r.String())
}

// String 返回报告的字符串表示
func (r *ConditionEvaluationReport) String() string {
	r.mu.RLock()
	positiveMatches := r.positiveMatches
	negativeMatches := r.negativeMatches
	exclusions := r.exclusions
	unconditional := r.unconditional
	r.mu.RUnlock()

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("============================\n")
	sb.WriteString("AUTO-CONFIGURATION REPORT\n")
	sb.WriteString("============================\n\n")

	// 正匹配
	sb.WriteString("Positive matches:\n")
	sb.WriteString("-----------------\n")
	if len(positiveMatches) > 0 {
		for _, match := range r.sortByName(positiveMatches) {
			fmt.Fprintf(&sb, "   %s matched:\n", match.Name)
			for _, cond := range match.Conditions {
				fmt.Fprintf(&sb, "      - %s matched\n", cond.Message)
			}
		}
	} else {
		sb.WriteString("   (none)\n")
	}
	sb.WriteString("\n")

	// 负匹配
	sb.WriteString("Negative matches:\n")
	sb.WriteString("-----------------\n")
	if len(negativeMatches) > 0 {
		for _, match := range r.sortByName(negativeMatches) {
			fmt.Fprintf(&sb, "   %s did not match:\n", match.Name)
			for _, cond := range match.Conditions {
				if !cond.Matched {
					fmt.Fprintf(&sb, "      - %s not matched\n", cond.Message)
				}
			}
		}
	} else {
		sb.WriteString("   (none)\n")
	}
	sb.WriteString("\n")

	// Exclusions
	sb.WriteString("Exclusions:\n")
	sb.WriteString("-----------\n")
	if len(exclusions) > 0 {
		for _, ex := range exclusions {
			fmt.Fprintf(&sb, "   - %s\n", ex)
		}
	} else {
		sb.WriteString("   (none)\n")
	}
	sb.WriteString("\n")

	// 无条件类
	sb.WriteString("Unconditional classes:\n")
	sb.WriteString("----------------------\n")
	if len(unconditional) > 0 {
		for _, name := range unconditional {
			fmt.Fprintf(&sb, "   - %s\n", name)
		}
	} else {
		sb.WriteString("   (none)\n")
	}

	return sb.String()
}

func (r *ConditionEvaluationReport) sortByName(matches []AutoConfigMatchResult) []AutoConfigMatchResult {
	sorted := make([]AutoConfigMatchResult, len(matches))
	copy(sorted, matches)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})
	return sorted
}

// globalReportEnabled 全局报告开关
var globalReportEnabled atomic.Bool

// EnableAutoConfigReport 启用自动配置报告
func EnableAutoConfigReport() {
	globalReportEnabled.Store(true)
}

// DisableAutoConfigReport 禁用自动配置报告
func DisableAutoConfigReport() {
	globalReportEnabled.Store(false)
}

// IsAutoConfigReportEnabled 检查自动配置报告是否启用
func IsAutoConfigReportEnabled() bool {
	return globalReportEnabled.Load()
}

// globalReport 全局报告实例
var globalReport atomic.Pointer[ConditionEvaluationReport]

func init() {
	globalReport.Store(NewConditionEvaluationReport())
}

// GetAutoConfigReport 获取全局报告实例
func GetAutoConfigReport() *ConditionEvaluationReport {
	return globalReport.Load()
}

// ResetAutoConfigReport 重置全局报告
func ResetAutoConfigReport() {
	globalReport.Store(NewConditionEvaluationReport())
}
