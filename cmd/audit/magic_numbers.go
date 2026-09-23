package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
)

// minMagicValue 绝对值小于该值的整数字面量视为合法小值。
const minMagicValue = 10

// knownPorts 常见服务端口，不作为魔法数字报告。
var knownPorts = map[int64]bool{
	80: true, 443: true, 3000: true, 3306: true, 5432: true,
	6379: true, 8080: true, 8443: true, 9092: true, 9200: true, 27017: true,
}

// analyzeMagicNumbers 检测条件比较、switch 分支与返回值中的魔法数字。
func analyzeMagicNumbers(f *fileInfo) []Finding {
	var findings []Finding
	skip := map[token.Pos]bool{}
	markTimeRelated(f, skip)
	visited := map[token.Pos]bool{}
	ast.Inspect(f.file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.BinaryExpr:
			if isComparableOp(node.Op) {
				checkComparison(node, magicCheckContext{skip: skip, f: f, findings: &findings, visited: visited})
			}
		case *ast.CaseClause:
			for _, expr := range node.List {
				checkLiteral(expr, "switch 分支常量", magicCheckContext{skip: skip, f: f, findings: &findings, visited: visited})
			}
		case *ast.ReturnStmt:
			for _, expr := range node.Results {
				checkLiteral(expr, "返回值", magicCheckContext{skip: skip, f: f, findings: &findings, visited: visited})
			}
		}
		return true
	})
	return findings
}

func isComparableOp(op token.Token) bool {
	switch op {
	case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
		return true
	}
	return false
}

// markTimeRelated 将与 time 常量运算的数字位置收集为跳过集合。
func markTimeRelated(f *fileInfo, skip map[token.Pos]bool) {
	ast.Inspect(f.file, func(n ast.Node) bool {
		bin, ok := n.(*ast.BinaryExpr)
		if !ok || (bin.Op != token.MUL && bin.Op != token.QUO) {
			return true
		}
		if containsSelectorExpr(bin.X, "time") {
			collectNumericPos(bin.Y, skip)
		}
		if containsSelectorExpr(bin.Y, "time") {
			collectNumericPos(bin.X, skip)
		}
		return true
	})
}

// containsSelectorExpr 判断表达式内是否引用 pkg 包的 SelectorExpr。
func containsSelectorExpr(expr ast.Expr, pkg string) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if ident, ok2 := sel.X.(*ast.Ident); ok2 && ident.Name == pkg {
			found = true
			return false
		}
		return true
	})
	return found
}

// collectNumericPos 记录表达式内的整数字面量位置（含一元负号）。
func collectNumericPos(expr ast.Expr, skip map[token.Pos]bool) {
	switch exprNode := expr.(type) {
	case *ast.BasicLit:
		if exprNode.Kind == token.INT {
			skip[exprNode.Pos()] = true
		}
	case *ast.UnaryExpr:
		skip[exprNode.Pos()] = true
		collectNumericPos(exprNode.X, skip)
	}
}

// literalValue 提取表达式中的整数字面量（支持一元负号）。
func literalValue(expr ast.Expr) (int64, bool) {
	switch exprNode := expr.(type) {
	case *ast.BasicLit:
		if exprNode.Kind == token.INT {
			parsedVal, err := strconv.ParseInt(exprNode.Value, 0, 64)
			return parsedVal, err == nil
		}
	case *ast.UnaryExpr:
		if exprNode.Op == token.SUB {
			parsedVal, ok := literalValue(exprNode.X)
			return -parsedVal, ok
		}
	}
	return 0, false
}

func isMagic(v int64) bool {
	abs := v
	if abs < 0 {
		abs = -abs
	}
	if abs < minMagicValue {
		return false
	}
	if isHTTPStatus(abs) || knownPorts[abs] || isPowerOfTwo(abs) {
		return false
	}
	return true
}

// isHTTPStatus 100-599 视为 HTTP 状态码。
func isHTTPStatus(v int64) bool { return v >= 100 && v <= 599 }

func isPowerOfTwo(v int64) bool {
	if v <= 0 {
		return false
	}
	return v&(v-1) == 0
}

// magicCheckContext 魔法数字检查上下文。
type magicCheckContext struct {
	skip     map[token.Pos]bool
	f        *fileInfo
	findings *[]Finding
	visited  map[token.Pos]bool
}

func checkComparison(expr ast.Expr, ctx magicCheckContext) {
	ast.Inspect(expr, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		checkLiteral(n, "条件比较", ctx)
		return true
	})
}

func checkLiteral(n ast.Node, usage string, ctx magicCheckContext) {
	if n == nil {
		return
	}
	if ctx.skip[n.Pos()] {
		return
	}
	expr, ok := n.(ast.Expr)
	if !ok {
		return
	}
	numVal, ok := literalValue(expr)
	if !ok || !isMagic(numVal) {
		return
	}
	addMagicFinding(literalPos(expr), numVal, usage, ctx.f, ctx.findings, ctx.visited)
}

// literalPos 负数取数字字面量本身的位置，作为唯一位置去重。
func literalPos(expr ast.Expr) token.Pos {
	if unaryExpr, ok := expr.(*ast.UnaryExpr); ok && unaryExpr.Op == token.SUB {
		return unaryExpr.X.Pos()
	}
	return expr.Pos()
}

func addMagicFinding(pos token.Pos, value int64, usage string, f *fileInfo, findings *[]Finding, visited map[token.Pos]bool) {
	if visited[pos] {
		return
	}
	visited[pos] = true
	*findings = append(*findings, Finding{
		File:       f.path,
		Line:       f.fset.Position(pos).Line,
		Category:   CategoryMagicNumbers,
		Severity:   SeverityMechanical,
		Message:    fmt.Sprintf("魔法数字 %d 出现在%s", value, usage),
		Suggestion: "提取为命名常量（const）并附注释说明含义",
	})
}
