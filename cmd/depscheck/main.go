// Command depscheck detects dependency cycles across go.work modules.
//
// It parses go.work to discover all modules, scans Go source files for
// imports that reference other workspace modules, builds a directed
// dependency graph, and reports any cycles found using DFS.
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go/parser"
	"go/token"
)

// moduleInfo holds a module's directory and its declared module path.
type moduleInfo struct {
	dir  string // relative directory, e.g. "./starter/chi"
	path string // module path, e.g. "github.com/xudefa/enhance/starter/chi"
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ depscheck: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	workFile, err := findGoWork()
	if err != nil {
		return err
	}

	modules, err := parseGoWork(workFile)
	if err != nil {
		return err
	}

	if len(modules) == 0 {
		fmt.Println("✅ 未发现 go.work 模块，跳过依赖检查")
		return nil
	}

	// Build a set of all module paths for quick lookup.
	modPaths := make(map[string]bool, len(modules))
	for _, m := range modules {
		modPaths[m.path] = true
	}

	// Build adjacency list: module dir -> list of module dirs it depends on.
	graph := make(map[string][]string, len(modules))
	for _, m := range modules {
		deps, err := collectImports(m, modPaths)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ 警告: 扫描 %s 失败: %v\n", m.dir, err)
			continue
		}
		graph[m.dir] = deps
	}

	// Detect cycles using DFS.
	cycles := detectCycles(graph, modules)

	if len(cycles) > 0 {
		fmt.Println("❌ 发现依赖环:")
		for _, cycle := range cycles {
			fmt.Printf("   %s\n", strings.Join(cycle, " → "))
		}
		return fmt.Errorf("发现 %d 个依赖环", len(cycles))
	}

	fmt.Println("✅ 依赖方向检查完成")
	return nil
}

// findGoWork locates the go.work file by walking up from cwd.
func findGoWork() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取工作目录失败: %w", err)
	}
	for {
		path := filepath.Join(dir, "go.work")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("未找到 go.work 文件")
		}
		dir = parent
	}
}

// parseGoWork reads go.work and returns all use directives.
func parseGoWork(path string) ([]moduleInfo, error) {
	fileHandle, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开 go.work 失败: %w", err)
	}
	defer fileHandle.Close()

	var modules []moduleInfo
	workDir := filepath.Dir(path)
	inUse := false

	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "use (" {
			inUse = true
			continue
		}
		if inUse && line == ")" {
			inUse = false
			continue
		}

		dir := line
		if inUse {
			dir = strings.TrimSpace(line)
		}
		if dir == "" || dir == "use" || strings.HasPrefix(dir, "go ") {
			continue
		}
		// Strip trailing comments.
		if idx := strings.Index(dir, "//"); idx >= 0 {
			dir = strings.TrimSpace(dir[:idx])
		}
		// Handle single-line use.
		if strings.HasPrefix(dir, "use ") {
			dir = strings.TrimSpace(strings.TrimPrefix(dir, "use "))
		}
		// Strip quotes if present.
		dir = strings.Trim(dir, "\"")

		modPath, err := readModulePath(filepath.Join(workDir, dir, "go.mod"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ 跳过 %s: %v\n", dir, err)
			continue
		}
		modules = append(modules, moduleInfo{dir: dir, path: modPath})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取 go.work 失败: %w", err)
	}
	return modules, nil
}

// readModulePath extracts the module directive from a go.mod file.
func readModulePath(gomod string) (string, error) {
	fileHandle, err := os.Open(gomod)
	if err != nil {
		return "", err
	}
	defer fileHandle.Close()

	scanner := bufio.NewScanner(fileHandle)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	return "", fmt.Errorf("未找到 module 声明: %s", gomod)
}

// collectImports parses all .go files in a module dir and returns
// imports that reference other workspace modules.
func collectImports(m moduleInfo, modPaths map[string]bool) ([]string, error) {
	var deps []string
	seen := make(map[string]bool)

	err := filepath.Walk(m.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() && path != m.dir {
			// Skip hidden directories and vendor.
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "vendor" || base == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		return collectImportsFromFile(path, m.path, modPaths, seen, &deps)
	})
	return deps, err
}

// collectImportsFromFile parses a single Go file and appends workspace imports.
func collectImportsFromFile(filePath, modPath string, modPaths map[string]bool, seen map[string]bool, deps *[]string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil // skip unparseable files
	}

	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		// Find which workspace module this import belongs to.
		for mod := range modPaths {
			if path == mod || strings.HasPrefix(path, mod+"/") {
				if !seen[mod] && mod != modPath {
					seen[mod] = true
					*deps = append(*deps, mod)
				}
				break
			}
		}
	}
	return nil
}

// detectCycles performs DFS-based cycle detection on the dependency graph.
// Returns all cycles found, where each cycle is a path of module dirs.
func detectCycles(graph map[string][]string, modules []moduleInfo) [][]string {
	const (
		white = 0 // unvisited
		gray  = 1 // in current DFS path
		black = 2 // fully processed
	)

	color := make(map[string]int, len(modules))
	parent := make(map[string]string, len(modules))
	var cycles [][]string

	var dfs func(node string)
	dfs = func(node string) {
		color[node] = gray
		for _, neighbor := range graph[node] {
			switch color[neighbor] {
			case gray:
				// Found a cycle — reconstruct the path.
				cycle := []string{neighbor, node}
				cur := node
				for cur != neighbor {
					cur = parent[cur]
					if cur == "" {
						break
					}
					cycle = append(cycle, cur)
				}
				// Reverse to get forward order.
				for i, j := 0, len(cycle)-1; i < j; i, j = i+1, j-1 {
					cycle[i], cycle[j] = cycle[j], cycle[i]
				}
				cycles = append(cycles, cycle)
			case white:
				parent[neighbor] = node
				dfs(neighbor)
			}
		}
		color[node] = black
	}

	for _, m := range modules {
		if color[m.dir] == white {
			dfs(m.dir)
		}
	}
	return cycles
}
