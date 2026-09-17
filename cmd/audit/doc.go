// Package main 提供 enhance 项目的 AI 可读性与可维护性审计工具。
//
// 用法:
//
//	go run ./cmd/audit -dir . -out docs/AI_READABILITY_AUDIT.md
//
// 检测维度:
//   - bad_names: 单字母/模糊变量命名
//   - magic_numbers: 魔法数字与魔法字符串
//   - error_context: 错误信息上下文
//   - documentation: 导出符号 doc 注释
//   - maintainability: 文件/函数/接口结构超限
package main
