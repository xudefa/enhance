package boot

import (
	"fmt"
	"os"
	"strings"
)

// reportError 通过 FailureAnalyzer 输出友好错误提示并返回结构化错误
func (b *Boot) reportError(phase string, err error) *bootError {
	bootErr := &bootError{
		code:     ErrCodeUnknown,
		message:  err.Error(),
		phase:    phase,
		original: err,
	}

	// 使用 FailureAnalyzer 分析
	report := globalAnalyzerRegistry.Analyze(err)
	if report != nil {
		bootErr.analyzed = report.Description
		bootErr.suggestions = report.PossibleSolutions
		fmt.Fprintf(os.Stderr, "\n%s\n", formatFailure(report))
	}

	return bootErr
}

// collectAutoConfigReport 收集自动配置匹配情况
func (b *Boot) collectAutoConfigReport(allEntries []AutoConfigEntry, matchedEntries []AutoConfigEntry) {
	report := GetAutoConfigReport()

	// 构建已匹配的集合
	matchedSet := make(map[string]bool)
	for _, entry := range matchedEntries {
		name := typeName(entry.Config)
		matchedSet[name] = true
	}

	// 遍历所有条目，记录匹配结果
	for _, entry := range allEntries {
		name := typeName(entry.Config)

		// 构建条件结果
		conditions := make([]ConditionResult, 0, len(entry.Conditions))
		for _, cond := range entry.Conditions {
			ctx := newConditionCtx(b.ctx)
			matched := cond.Matches(ctx)
			conditions = append(conditions, ConditionResult{
				Condition: fmt.Sprintf("%T", cond),
				Matched:   matched,
				Message:   cond.String(),
			})
		}

		// 无条件时记录为无条件类
		if len(entry.Conditions) == 0 {
			report.RecordUnconditional(name)
			continue
		}

		// 根据是否匹配记录到正面或负面
		if matchedSet[name] {
			report.RecordPositiveMatch(name, conditions)
		} else {
			report.RecordNegativeMatch(name, conditions)
		}
	}
}

// typeName 获取类型的简短名称
func typeName(v any) string {
	t := fmt.Sprintf("%T", v)
	// 移除包路径，只保留类型名
	if idx := strings.LastIndex(t, "."); idx >= 0 {
		return t[idx+1:]
	}
	return t
}

// starterMatches 检查启动器条件是否匹配
func (b *Boot) starterMatches(s Starter) bool {
	cond := s.GetCondition()
	if cond == nil {
		return true
	}
	return cond.Matches(newConditionCtx(b.ctx))
}

// moduleMatches 检查模块条件是否全部匹配
func (b *Boot) moduleMatches(m Module) bool {
	conds := m.conditions
	if len(conds) == 0 {
		return true
	}
	ctx := newConditionCtx(b.ctx)
	for _, cond := range conds {
		if !cond.Matches(ctx) {
			return false
		}
	}
	return true
}

// deduplicateStarters 去重启动器列表（按名称去重，保留第一个）
func deduplicateStarters(starters []Starter) []Starter {
	seen := make(map[string]bool)
	result := make([]Starter, 0, len(starters))
	for _, s := range starters {
		name := s.Name()
		if !seen[name] {
			seen[name] = true
			result = append(result, s)
		}
	}
	return result
}
