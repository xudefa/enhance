package main

import (
	"go/ast"
	"go/token"
)

// fileInfo 保存单个待审计 Go 文件的解析结果。
type fileInfo struct {
	path   string
	area   string
	isTest bool
	src    []byte
	fset   *token.FileSet
	file   *ast.File
}

// detect 对单个文件执行全部维度检测。
// 测试文件只检测命名与结构，不检测魔法数字、doc 与错误上下文。
func detect(f *fileInfo) []Finding {
	var findings []Finding
	findings = append(findings, analyzeBadNames(f)...)
	if !f.isTest {
		findings = append(findings, analyzeMagicNumbers(f)...)
		findings = append(findings, analyzeErrorContext(f)...)
		findings = append(findings, analyzeDocumentation(f)...)
	}
	findings = append(findings, analyzeMaintainability(f)...)
	return findings
}
