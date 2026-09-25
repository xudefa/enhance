package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PackageResult 聚合单个包/区域的全部发现。
type PackageResult struct {
	Dir          string
	Findings     []Finding
	TestFindings []Finding
}

// Stats 汇总统计信息。
type Stats struct {
	Packages      int
	Files         int
	TotalFindings int
}

// audit 扫描 root 下的全部 .go 文件并执行检测。
func audit(root string) ([]PackageResult, Stats, error) {
	groups := map[string]*PackageResult{}
	var stats Stats
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		fi, err := buildFileInfo(root, path)
		if err != nil {
			return fmt.Errorf("build file info for %q: %w", path, err)
		}
		if fi == nil {
			return nil // 文件解析失败，已输出跳过信息
		}
		key := groupKey(fi.area, fi.path)
		pkg := groups[key]
		if pkg == nil {
			pkg = &PackageResult{Dir: key}
			groups[key] = pkg
		}
		stats.Files++
		findings := detect(fi)
		stats.TotalFindings += len(findings)
		if fi.isTest {
			pkg.TestFindings = append(pkg.TestFindings, findings...)
			return nil
		}
		pkg.Findings = append(pkg.Findings, findings...)
		return nil
	})
	if err != nil {
		return nil, Stats{}, fmt.Errorf("walk directory %q: %w", root, err)
	}
	keys := sortedKeys(groups)
	results := make([]PackageResult, 0, len(keys))
	for _, key := range keys {
		results = append(results, *groups[key])
	}
	stats.Packages = len(results)
	return results, stats, nil
}

// buildFileInfo 读取并解析单个 Go 文件。
func buildFileInfo(root, path string) (*fileInfo, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return nil, fmt.Errorf("resolve relative path for %q: %w", path, err)
	}
	rel = filepath.ToSlash(rel)
	fset := token.NewFileSet()
	filename := filepath.Base(path)
	parsed, parseErr := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "跳过无法解析的文件 %s: %v\n", rel, parseErr)
		return nil, nil
	}
	if parsed == nil {
		return nil, nil
	}
	return &fileInfo{
		path:   rel,
		area:   classifyArea(rel),
		isTest: strings.HasSuffix(filename, "_test.go"),
		src:    src,
		fset:   fset,
		file:   parsed,
	}, nil
}

// classifyArea 按路径划分报告区域：顶层 starter/examples/cmd 目录各自归区。
func classifyArea(rel string) string {
	switch {
	case strings.HasPrefix(rel, "starter/"):
		return "starter"
	case strings.HasPrefix(rel, "examples/"):
		return "example"
	case strings.HasPrefix(rel, "cmd/"):
		return "tooling"
	default:
		return "framework"
	}
}

// groupKey 计算报告分组键：框架包按目录，其余按区域。
func groupKey(area, rel string) string {
	if area == "framework" {
		return filepath.ToSlash(filepath.Dir(rel))
	}
	return area
}

func shouldSkipDir(path string) bool {
	base := filepath.Base(path)
	if base == "." || base == "/" || base == "" {
		return false
	}
	switch base {
	case ".git", ".idea", ".vscode", "vendor", "node_modules":
		return true
	}
	return strings.HasPrefix(base, ".")
}

func sortedKeys(groups map[string]*PackageResult) []string {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
