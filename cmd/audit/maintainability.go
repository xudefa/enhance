package main

import (
	"bytes"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

const (
	// 以下为 CODING_STYLE.md 规定的结构阈值。
	maxFileLines         = 500
	maxDocLines          = 500
	maxFuncLines         = 50
	maxBreakingFuncLines = 100
	maxParams            = 4
	maxMethodsPerType    = 200
)

// analyzeMaintainability 检测文件、函数、参数等结构性超限。
func analyzeMaintainability(f *fileInfo) []Finding {
	var findings []Finding
	fileStart := f.file.Pos()
	lineCount := bytes.Count(f.src, []byte{'\n'}) + 1
	if lineCount > maxFileLines {
		findings = append(findings, maintainFinding(f, fileStart,
			"文件共 "+strconv.Itoa(lineCount)+" 行，超过 "+strconv.Itoa(maxFileLines)+" 行限制",
			"拆分到多个小文件（按实现类或职责）"))
	}
	if strings.HasSuffix(f.path, "doc.go") && lineCount > maxDocLines {
		findings = append(findings, maintainFinding(f, fileStart,
			"doc.go 共 "+strconv.Itoa(lineCount)+" 行，超过 "+strconv.Itoa(maxDocLines)+" 行",
			"创建 types.go 分担类型定义"))
	}
	methodsPerType := map[string]int{}
	ast.Inspect(f.file, func(n ast.Node) bool {
		if funcDecl, ok := n.(*ast.FuncDecl); ok {
			analyzeFunc(f, funcDecl, &findings, methodsPerType)
		}
		return true
	})
	appendTypeOverloads(f, methodsPerType, &findings)
	return findings
}

func analyzeFunc(f *fileInfo, d *ast.FuncDecl, findings *[]Finding, methodsPerType map[string]int) {
	if d.Body == nil {
		return
	}
	start := f.fset.Position(d.Body.Pos()).Line
	end := f.fset.Position(d.Body.End()).Line
	bodyLines := end - start + 1

	target := "函数 " + d.Name.Name
	if d.Recv != nil && len(d.Recv.List) > 0 {
		recv := receiverTypeName(d.Recv.List[0].Type)
		if recv != "" {
			methodsPerType[recv]++
			target = "方法 " + recv + "." + d.Name.Name
		}
	}
	if bodyLines > maxBreakingFuncLines {
		*findings = append(*findings, maintainFinding(f, d.Pos(),
			target+" 方法体 "+strconv.Itoa(bodyLines)+" 行（超过 100 行）",
			"按职责拆分并提取子函数", SeverityBreaking))
	} else if bodyLines > maxFuncLines {
		*findings = append(*findings, maintainFinding(f, d.Pos(),
			target+" 方法体 "+strconv.Itoa(bodyLines)+" 行，超过 50 行",
			"提取子函数以降低圈复杂度"))
	}
	paramCount := countParams(d.Type.Params)
	if paramCount > maxParams {
		severity := SeverityMechanical
		if d.Name.IsExported() {
			severity = SeverityBreaking
		}
		*findings = append(*findings, maintainFinding(f, d.Pos(),
			target+" 参数 "+strconv.Itoa(paramCount)+" 个，超过 4 个",
			"使用函数式选项模式或结构体参数", severity))
	}
}

// countParams 计算函数参数总数（含无名参数）。
func countParams(fields *ast.FieldList) int {
	if fields == nil {
		return 0
	}
	total := 0
	for _, field := range fields.List {
		if len(field.Names) == 0 {
			total++
			continue
		}
		total += len(field.Names)
	}
	return total
}

// receiverTypeName 提取接收器类型名（支持指针与泛型接收器）。
func receiverTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return receiverTypeName(e.X)
	case *ast.IndexExpr:
		return receiverTypeName(e.X)
	}
	return ""
}

func appendTypeOverloads(f *fileInfo, methodsPerType map[string]int, findings *[]Finding) {
	for typeName, count := range methodsPerType {
		if count > maxMethodsPerType {
			*findings = append(*findings, maintainFinding(f, f.file.Pos(),
				"类型 "+typeName+" 有 "+strconv.Itoa(count)+" 个方法，超过 "+strconv.Itoa(maxMethodsPerType),
				"拆分职责到多个类型"))
		}
	}
}

// maintainFinding 构造结构类发现，默认 Mechanical。
func maintainFinding(f *fileInfo, pos token.Pos, msg, suggestion string, severities ...Severity) Finding {
	severity := SeverityMechanical
	if len(severities) > 0 {
		severity = severities[0]
	}
	return Finding{
		File: f.path, Line: f.fset.Position(pos).Line,
		Category:   CategoryMaintainability,
		Severity:   severity,
		Message:    msg,
		Suggestion: suggestion,
	}
}
