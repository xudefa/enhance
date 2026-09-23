package main

import (
	"go/parser"
	"go/token"
	"testing"
)

// parseSnippet 将源码片段解析为 fileInfo，便于单测。
func parseSnippet(t *testing.T, code string, isTest bool) *fileInfo {
	t.Helper()
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, "snippet.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	return &fileInfo{
		path:   "snippet.go",
		isTest: isTest,
		src:    []byte(code),
		fset:   fset,
		file:   parsed,
	}
}
