package main

import (
	"go/ast"
	"go/token"
	"strings"
)

// analyzeErrorContext 检测错误处理中可能丢失上下文的情况。
// 规则1：裸返回 err；规则2：fmt.Errorf 用 %v/%s 格式化 error 参数。
func analyzeErrorContext(f *fileInfo) []Finding {
	if f.isTest {
		return nil
	}
	var findings []Finding
	ast.Inspect(f.file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.ReturnStmt:
			for _, res := range node.Results {
				if ident, ok := res.(*ast.Ident); ok && isErrorName(ident.Name) {
					findings = append(findings, errorFinding(f, ident.Pos(),
						"裸返回 "+ident.Name+"，可能丢失错误上下文",
						"改用 fmt.Errorf 包装上下文：fmt.Errorf(\"...: %w\", "+ident.Name+")"))
				}
			}
		case *ast.CallExpr:
			if isFmtErrorf(node) && misusesPercentW(node) {
				findings = append(findings, errorFinding(f, node.Pos(),
					"fmt.Errorf 使用 %v/%s 格式化 error，丢失错误链",
					"将错误参数改为 %w 以保留 errors.Is/As 能力"))
			}
		}
		return true
	})
	return findings
}

// isErrorName 判断标识符是否表示 error 值。
// 遵循 Go 命名约定：error 变量通常是 err 或 err<大写>（如 errNotFound）；
// errs/errmsg 等非常规命名的变量不做 error 推断，避免误报。
func isErrorName(name string) bool {
	if name == "err" || name == "error" || name == "e" {
		return true
	}
	if len(name) <= 3 || !strings.HasPrefix(name, "err") {
		return false
	}
	return name[3] >= 'A' && name[3] <= 'Z'
}

func isFmtErrorf(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Errorf" {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "fmt"
}

// misusesPercentW 判断 fmt.Errorf 是否用 %v/%s 接收了 error 参数。
func misusesPercentW(call *ast.CallExpr) bool {
	if len(call.Args) < 2 {
		return false
	}
	format, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return false
	}
	text := strings.Trim(format.Value, "`\"")
	if strings.Contains(text, "%w") || !strings.ContainsAny(text, "%v%s") {
		return false
	}
	for _, arg := range call.Args[1:] {
		if ident, ok := arg.(*ast.Ident); ok && isErrorName(ident.Name) {
			return true
		}
	}
	return false
}

func errorFinding(f *fileInfo, pos token.Pos, msg, sug string) Finding {
	return Finding{
		File: f.path, Line: f.fset.Position(pos).Line,
		Category:   CategoryErrorContext,
		Severity:   SeverityMechanical,
		Message:    msg,
		Suggestion: sug,
	}
}
