package main

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"
)

// analyzeDocumentation 检测缺失 doc 注释的导出标识符，以及英文注释。
func analyzeDocumentation(f *fileInfo) []Finding {
	var findings []Finding
	for _, decl := range f.file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			analyzeFuncDoc(f, d, &findings)
		case *ast.GenDecl:
			analyzeGroupDoc(f, d, &findings)
		}
	}
	return findings
}

func analyzeFuncDoc(f *fileInfo, d *ast.FuncDecl, findings *[]Finding) {
	name := d.Name.Name
	if !d.Name.IsExported() || name == "init" || name == "main" || isTestFunc(name) {
		return
	}
	if d.Doc == nil {
		*findings = append(*findings, docFinding(f, d.Pos(),
			"导出函数 "+name+" 缺少 doc 注释",
			"在函数上方添加 // "+name+" ... 描述职责与参数"))
		return
	}
	if isEnglishComment(d.Doc.Text()) {
		*findings = append(*findings, docFinding(f, d.Pos(),
			"导出函数 "+name+" 的注释为英文（规范要求中文）",
			"将注释改为中文"))
	}
}

func analyzeGroupDoc(f *fileInfo, d *ast.GenDecl, findings *[]Finding) {
	groupDoc := d.Doc
	for _, spec := range d.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			if s.Name.IsExported() && s.Doc == nil && groupDoc == nil {
				*findings = append(*findings, docFinding(f, s.Pos(),
					"导出类型 "+s.Name.Name+" 缺少 doc 注释",
					"在类型上方添加 // "+s.Name.Name+" 描述"))
			}
		case *ast.ValueSpec:
			for _, name := range s.Names {
				if name.IsExported() && s.Doc == nil && groupDoc == nil {
					*findings = append(*findings, docFinding(f, name.Pos(),
						"导出标识符 "+name.Name+" 缺少 doc 注释",
						"在声明上方添加注释"))
				}
			}
		}
	}
}

// isTestFunc 判断是否为测试/基准/示例函数。
func isTestFunc(name string) bool {
	return strings.HasPrefix(name, "Test") ||
		strings.HasPrefix(name, "Benchmark") ||
		strings.HasPrefix(name, "Example")
}

// isEnglishComment 判断注释文本是否为纯英文（无 CJK 字符且词数足够）。
func isEnglishComment(text string) bool {
	hasCJK := false
	wordCount := 0
	inWord := false
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			hasCJK = true
			break
		}
		if unicode.IsLetter(r) {
			if !inWord {
				wordCount++
				inWord = true
			}
			continue
		}
		inWord = false
	}
	return !hasCJK && wordCount >= 5
}

func docFinding(f *fileInfo, pos token.Pos, msg, sug string) Finding {
	return Finding{
		File: f.path, Line: f.fset.Position(pos).Line,
		Category:   CategoryDocumentation,
		Severity:   SeverityMechanical,
		Message:    msg,
		Suggestion: sug,
	}
}
