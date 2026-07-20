package generator

import (
	"fmt"
	"strings"
)

// GoGenerateDirective 解析 //go:generate 指令
//
// 支持以下格式：
//   - //go:generate enhance aop gen -type=UserService
//   - //go:generate enhance aop gen -type=UserService,OrderService
//   - //go:generate enhance aop gen -type=UserService -output=proxy.go
//   - //go:generate enhance aop gen -interface=ServiceInterface
//   - //go:generate enhance aop gen -all
type GoGenerateDirective struct {
	Types      []string // 需要生成代理的结构体类型
	Interfaces []string // 需要生成代理的接口类型
	Output     string   // 输出文件名（默认 *_aop.go）
	Package    string   // 包名
	Mode       string   // 生成模式：simple, aop, static
	All        bool     // 是否为所有类型生成代理
}

// ParseGoGenerate 解析 //go:generate 指令
func ParseGoGenerate(comment string) (*GoGenerateDirective, error) {
	// 去除 //go:generate 前缀
	prefix := "//go:generate"
	if !strings.HasPrefix(comment, prefix) {
		return nil, fmt.Errorf("not a go:generate directive")
	}

	rest := strings.TrimSpace(comment[len(prefix):])

	// 检查是否是 enhance aop gen 命令
	if !strings.HasPrefix(rest, "enhance aop gen") {
		return nil, fmt.Errorf("not an enhance aop gen directive")
	}

	rest = strings.TrimSpace(rest[len("enhance aop gen"):])

	directive := &GoGenerateDirective{
		Mode: "static", // 默认使用静态模式（零反射）
	}

	// 解析参数
	args := parseArgs(rest)

	if types, ok := args["type"]; ok {
		directive.Types = splitAndTrim(types, ",")
	}
	if interfaces, ok := args["interface"]; ok {
		directive.Interfaces = splitAndTrim(interfaces, ",")
	}
	if output, ok := args["output"]; ok {
		directive.Output = output
	}
	if pkg, ok := args["package"]; ok {
		directive.Package = pkg
	}
	if mode, ok := args["mode"]; ok {
		directive.Mode = mode
	}
	if _, ok := args["all"]; ok {
		directive.All = true
	}

	return directive, nil
}

// parseArgs 解析命令行参数
func parseArgs(s string) map[string]string {
	args := make(map[string]string)
	parts := strings.Fields(s)

	for _, part := range parts {
		if !strings.HasPrefix(part, "-") {
			continue
		}
		part = part[1:] // 去除 -
		if idx := strings.Index(part, "="); idx >= 0 {
			args[part[:idx]] = part[idx+1:]
		} else {
			args[part] = ""
		}
	}

	return args
}

// splitAndTrim 分割字符串并去除空格
func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// HasTargets 检查是否有需要生成的目标
func (d *GoGenerateDirective) HasTargets() bool {
	return d.All || len(d.Types) > 0 || len(d.Interfaces) > 0
}
