package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAudit_EndToEnd(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "core/sub/a.go", `package sub

// F 返回一个魔法数字。
func F(x int) bool { return x > 60 }
`)
	writeFile(t, root, "starter/gin/b.go", "package gin\nfunc B() int { n := 5\n n = n + 1\n return n }\n")
	writeFile(t, root, "core/sub/a_test.go", "package sub\nimport \"testing\"\nfunc TestA(t *testing.T) {}\n")

	results, stats, err := audit(root)
	if err != nil {
		t.Fatalf("audit 失败: %v", err)
	}
	if stats.Files != 3 {
		t.Errorf("期望 3 个文件，得到 %d", stats.Files)
	}
	var hasMagic bool
	for _, pkg := range results {
		for _, fd := range pkg.Findings {
			if fd.Category == "magic_numbers" && strings.Contains(fd.Message, "60") {
				hasMagic = true
			}
		}
	}
	if !hasMagic {
		t.Fatalf("未在 core/sub 发现魔法数字 60: %+v", results)
	}
}

func TestAudit_MergesByPackage(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "pkg/a.go", "package pkg\nfunc X() int { return 0 }\n")
	writeFile(t, root, "pkg/b.go", "package pkg\nfunc Y() int { n := 1\n n = n + 1\n return n }\n")
	results, _, err := audit(root)
	if err != nil {
		t.Fatalf("audit 失败: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("两个文件应聚合到一个包: %+v", results)
	}
}

func TestClassifyArea(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		rel  string
		want string
	}{
		{"starter 子包", "starter/gin/gin.go", "starter"},
		{"starter 顶层文件", "starter/builder.go", "starter"},
		{"examples 子包", "examples/redis/main.go", "example"},
		{"examples 顶层文件", "examples/quickstart.go", "example"},
		{"tooling", "cmd/audit/main.go", "tooling"},
		{"框架包", "core/container.go", "framework"},
		{"框架子包", "web/mvc/router.go", "framework"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := classifyArea(tt.rel); got != tt.want {
				t.Errorf("classifyArea(%q) = %q, want %q", tt.rel, got, tt.want)
			}
		})
	}
}

func TestAudit_GroupsTopLevelStarterAsArea(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "starter/builder.go", "package starter\nfunc New() int { return 7 }\n")
	writeFile(t, root, "starter/gin/gin.go", "package gin\nfunc New() int { return 8 }\n")
	results, _, err := audit(root)
	if err != nil {
		t.Fatalf("audit 失败: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("starter 顶层文件与子包应合并为同一区域: %+v", results)
	}
	if results[0].Dir != "starter" {
		t.Errorf("期望合并为 starter, 得到 %q", results[0].Dir)
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写文件失败: %v", err)
	}
}
