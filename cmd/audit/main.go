package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// main 是审计 CLI 的入口：扫描、渲染并写出 Markdown 报告。
func main() {
	dir := flag.String("dir", ".", "扫描根目录")
	out := flag.String("out", "docs/AI_READABILITY_AUDIT.md", "报告输出路径")
	jsonOut := flag.String("json", "docs/audit-report.json", "JSON 报告输出路径")
	flag.Parse()

	root, err := filepath.Abs(*dir)
	if err != nil {
		fatal("解析目录失败: %v", err)
	}

	results, stats, err := audit(root)
	if err != nil {
		fatal("审计执行失败: %v", err)
	}

	report := renderReport(*dir, results, time.Now(), stats)
	printReport(report, *out)
	if err := writeJSONReport(*jsonOut, time.Now(), stats, results); err != nil {
		fatal("写入 JSON 报告失败: %v", err)
	}
}

// printReport 将报告输出到 stdout 或写入指定文件。
func printReport(report, out string) {
	if out == "" {
		fmt.Print(report)
		return
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		fatal("创建目录失败: %v", err)
	}
	if err := os.WriteFile(out, []byte(report), 0o644); err != nil {
		fatal("写入报告失败: %v", err)
	}
	fmt.Printf("报告已生成：%s\n", out)
}

// fatal 输出错误信息并以非零状态码退出。
func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "audit: "+format+"\n", args...)
	os.Exit(1)
}
