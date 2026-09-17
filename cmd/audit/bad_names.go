package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

// allowedSingleLetters 允许的单字母局部变量：container/client 惯例缩写。
var allowedSingleLetters = map[string]bool{
	"c": true,
}

// allowedTestLetters 测试文件允许的参数缩写。
var allowedTestLetters = map[string]bool{
	"t": true,
	"b": true,
}

// vagueIdentifiers 含义不清晰的通用命名黑名单。
var vagueIdentifiers = map[string]bool{
	"data": true, "tmp": true, "result": true, "info": true,
	"obj": true, "item": true, "val": true, "ret": true,
}

// analyzeBadNames 检测单字母与模糊命名的局部变量。
// 排除：i/j/k 循环变量、允许的惯例缩写、测试参数 t/b、空标识符。
func analyzeBadNames(f *fileInfo) []Finding {
	var findings []Finding
	for _, decl := range f.file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Body == nil {
			continue
		}
		loopVars := collectLoopVars(funcDecl.Body)
		declared, declPos := collectShortDecls(funcDecl.Body, loopVars, f.isTest)
		uses := countUses(funcDecl.Body, declared, declPos)
		appendBadNameFindings(f, uses, declPos, &findings)
	}
	return findings
}

// collectLoopVars 收集循环变量的位置集合，避免误报。
func collectLoopVars(body *ast.BlockStmt) map[token.Pos]bool {
	vars := map[token.Pos]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.RangeStmt:
			markIdent(vars, node.Key)
			markIdent(vars, node.Value)
		case *ast.ForStmt:
			if assign, ok := node.Init.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE {
				for _, lhs := range assign.Lhs {
					markIdent(vars, lhs)
				}
			}
		}
		return true
	})
	return vars
}

func markIdent(set map[token.Pos]bool, expr ast.Expr) {
	if ident, ok := expr.(*ast.Ident); ok {
		set[ident.Pos()] = true
	}
}

// collectShortDecls 收集函数体内 := 与 var 声明的可疑短名。
func collectShortDecls(body *ast.BlockStmt, loopVars map[token.Pos]bool, isTest bool) (map[string]bool, map[string]token.Pos) {
	declared := map[string]bool{}
	declPos := map[string]token.Pos{}
	collect := func(name string, pos token.Pos) {
		if name == "_" || loopVars[pos] {
			return
		}
		if len(name) == 1 {
			if allowedSingleLetters[name] {
				return
			}
			if isTest && allowedTestLetters[name] {
				return
			}
		} else if !vagueIdentifiers[name] {
			return
		}
		if _, exists := declared[name]; !exists {
			declared[name] = true
			declPos[name] = pos
		}
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			if node.Tok != token.DEFINE {
				return true
			}
			for _, lhs := range node.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					collect(ident.Name, ident.Pos())
				}
			}
		case *ast.GenDecl:
			if node.Tok != token.VAR {
				return true
			}
			for _, spec := range node.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					collect(name.Name, name.Pos())
				}
			}
		}
		return true
	})
	return declared, declPos
}

// countUses 统计声明之后各名称在函数体内的使用次数（含嵌套闭包，启发式）。
func countUses(body *ast.BlockStmt, declared map[string]bool, declPos map[string]token.Pos) map[string]int {
	uses := map[string]int{}
	ast.Inspect(body, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok || !declared[ident.Name] {
			return true
		}
		if ident.Pos() == declPos[ident.Name] {
			return true
		}
		uses[ident.Name]++
		return true
	})
	return uses
}

func appendBadNameFindings(f *fileInfo, uses map[string]int, declPos map[string]token.Pos, findings *[]Finding) {
	for name, count := range uses {
		minUses := 2
		if len(name) > 1 {
			minUses = 1
		}
		if count < minUses {
			continue
		}
		line := f.fset.Position(declPos[name]).Line
		if len(name) == 1 {
			*findings = append(*findings, Finding{
				File: f.path, Line: line,
				Category:   CategoryBadNames,
				Severity:   SeverityMechanical,
				Message:    fmt.Sprintf("单字母局部变量 %q 使用 %d 次", name, count),
				Suggestion: "改用语义化名称（如 server、nextCursor）",
			})
			continue
		}
		*findings = append(*findings, Finding{
			File: f.path, Line: line,
			Category:   CategoryBadNames,
			Severity:   SeverityMechanical,
			Message:    fmt.Sprintf("模糊命名变量 %q 使用 %d 次", name, count),
			Suggestion: "改用能表达业务含义的名称",
		})
	}
}
