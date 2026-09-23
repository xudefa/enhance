package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// renderReport 将审计结果渲染为 Markdown 报告。
func renderReport(root string, results []PackageResult, generated time.Time, stats Stats) string {
	var buf bytes.Buffer
	writeHeader(&buf, root, generated, stats)
	writeSummary(&buf, results)
	writeDetail(&buf, results)
	return buf.String()
}

func writeHeader(buf *bytes.Buffer, root string, generated time.Time, stats Stats) {
	buf.WriteString("# AI 可读性与可维护性审计报告\n\n")
	fmt.Fprintf(buf, "> 生成时间：%s\n", generated.Format("2006-01-02 15:04"))
	fmt.Fprintf(buf, "> 扫描根目录：`%s`\n", root)
	fmt.Fprintf(buf, "> 扫描包数：%d ｜ 文件数：%d ｜ 总发现：%d\n\n",
		stats.Packages, stats.Files, stats.TotalFindings)
	buf.WriteString("**严重程度分级**\n\n")
	buf.WriteString("- **Mechanical（机械可改）**：局部修改即可，可安全批量处理。\n")
	buf.WriteString("- **Breaking（需决策）**：可能影响调用方或需拆包，必须人工确认。\n\n")
}

func writeSummary(buf *bytes.Buffer, results []PackageResult) {
	buf.WriteString("## 汇总\n\n")
	buf.WriteString("| 包/区域 | 发现数 | 待决策 | 机械可改 |\n")
	buf.WriteString("|---|---|---|---|\n")
	counts := packageCounts(results)
	for _, dir := range sortedDirKeys(counts) {
		c := counts[dir]
		fmt.Fprintf(buf, "| `%s` | %d | %d | %d |\n", dir, c[0], c[1], c[2])
	}
	buf.WriteString("\n")
}

func writeDetail(buf *bytes.Buffer, results []PackageResult) {
	buf.WriteString("## 明细\n\n")
	for _, pkg := range results {
		if len(pkg.Findings) == 0 && len(pkg.TestFindings) == 0 {
			continue
		}
		fmt.Fprintf(buf, "### %s\n\n", pkg.Dir)
		breaking, mechanical := splitBySeverity(pkg.Findings)
		if len(breaking) > 0 {
			buf.WriteString("#### 待决策（Breaking）\n\n")
			renderFindings(buf, breaking)
		}
		if len(mechanical) > 0 {
			buf.WriteString("#### 机械可改（Mechanical）\n\n")
			renderFindings(buf, mechanical)
		}
		if len(pkg.TestFindings) > 0 {
			buf.WriteString("#### 测试文件\n\n")
			renderFindings(buf, pkg.TestFindings)
		}
		buf.WriteString("\n")
	}
}

func packageCounts(results []PackageResult) map[string][3]int {
	counts := map[string][3]int{}
	for _, pkg := range results {
		if len(pkg.Findings) == 0 && len(pkg.TestFindings) == 0 {
			continue
		}
		arr := counts[pkg.Dir]
		for _, f := range pkg.Findings {
			arr[0]++
			if f.Severity == SeverityBreaking {
				arr[1]++
			} else {
				arr[2]++
			}
		}
		arr[0] += len(pkg.TestFindings)
		arr[2] += len(pkg.TestFindings)
		counts[pkg.Dir] = arr
	}
	return counts
}

func sortedDirKeys(counts map[string][3]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func splitBySeverity(findings []Finding) ([]Finding, []Finding) {
	var breaking, mechanical []Finding
	for _, f := range findings {
		if f.Severity == SeverityBreaking {
			breaking = append(breaking, f)
		} else {
			mechanical = append(mechanical, f)
		}
	}
	sort.SliceStable(breaking, func(i, j int) bool {
		return byLocation(breaking[i], breaking[j])
	})
	sort.SliceStable(mechanical, func(i, j int) bool {
		return byLocation(mechanical[i], mechanical[j])
	})
	return breaking, mechanical
}

func byLocation(a, b Finding) bool {
	if a.File != b.File {
		return a.File < b.File
	}
	return a.Line < b.Line
}

func renderFindings(buf *bytes.Buffer, findings []Finding) {
	for _, f := range findings {
		fmt.Fprintf(buf, "- [ ] `%s:%d` **[%s] %s**\n", f.File, f.Line, f.Category, f.Message)
		if f.Suggestion != "" {
			fmt.Fprintf(buf, "  - 建议：%s\n", f.Suggestion)
		}
	}
}

// writeJSONReport 将审计结果以结构化 JSON 写出，供检查/评分脚本消费。
func writeJSONReport(path string, generated time.Time, stats Stats, results []PackageResult) error {
	report := struct {
		Generated string    `json:"generated"`
		Stats     Stats     `json:"stats"`
		Findings  []Finding `json:"findings"`
	}{
		Generated: generated.Format(time.RFC3339),
		Stats:     stats,
	}
	for _, pkg := range results {
		report.Findings = append(report.Findings, pkg.Findings...)
		report.Findings = append(report.Findings, pkg.TestFindings...)
	}
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0o644)
}
