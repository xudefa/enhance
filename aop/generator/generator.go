package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Generator AOP 代理代码生成器
//
// 解析 Go 源码中的注解，匹配切面与代理目标，
// 生成代理类代码文件。
type Generator struct {
	parser   *Parser         // 源码解析器
	engine   *TemplateEngine // 模板引擎
	registry *Registry       // 代理注册表
}

// NewGenerator 创建代理代码生成器
func NewGenerator() (*Generator, error) {
	engine, err := NewTemplateEngine()
	if err != nil {
		return nil, err
	}

	return &Generator{
		parser:   NewParser(),
		engine:   engine,
		registry: NewRegistry(),
	}, nil
}

// Generate 扫描指定目录并生成代理代码
//
// 参数：
//   - dir: 需要扫描的源码目录
//   - mode: 生成模式（"simple" 简单代理, "aop" 运行时 AOP, "static" 静态 AOP 零反射）
func (g *Generator) Generate(dir string, mode string) error {
	if err := g.parser.ParseDir(dir); err != nil {
		return fmt.Errorf("failed to parse directory: %w", err)
	}

	aspects := g.parser.GetAspects()
	proxies := g.parser.GetProxies()
	interfaces := g.parser.GetInterfaces()
	funcs := g.parser.GetFuncs()

	if len(proxies) == 0 && len(interfaces) == 0 {
		return nil
	}

	for _, proxy := range proxies {
		if err := g.generateProxy(proxy, aspects, funcs, mode); err != nil {
			return fmt.Errorf("failed to generate proxy for %s: %w", proxy.Name, err)
		}
	}

	for _, intf := range interfaces {
		if err := g.generateInterfaceProxy(intf, aspects, funcs, mode); err != nil {
			return fmt.Errorf("failed to generate interface proxy for %s: %w", intf.Name, err)
		}
	}

	return nil
}

// generateProxy 为单个代理目标生成代理代码
func (g *Generator) generateProxy(proxy ProxyInfo, aspects []AspectInfo, funcs []AdviceInfo, mode string) error {
	matchedAspects := g.matchAspects(proxy, aspects, funcs)

	data := g.buildProxyTemplateData(proxy, matchedAspects, mode)

	output, err := g.engine.GenerateProxy(data, mode)
	if err != nil {
		return err
	}

	outputPath := g.getOutputPath(proxy.FilePath)
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return err
	}

	g.registry.Register(proxy.BeanID, outputPath)

	return nil
}

// generateInterfaceProxy 为接口生成代理代码
func (g *Generator) generateInterfaceProxy(intf InterfaceInfo, aspects []AspectInfo, funcs []AdviceInfo, mode string) error {
	matchedAspects := g.matchAspectsForInterface(intf, aspects, funcs)

	data := g.buildInterfaceProxyTemplateData(intf, matchedAspects, mode)

	output, err := g.engine.GenerateInterfaceProxy(data, mode)
	if err != nil {
		return err
	}

	outputPath := g.getOutputPath(intf.FilePath)
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return err
	}

	g.registry.Register(intf.BeanID, outputPath)

	return nil
}

// matchAspectsForInterface 匹配接口与切面通知
func (g *Generator) matchAspectsForInterface(intf InterfaceInfo, aspects []AspectInfo, funcs []AdviceInfo) []AspectTemplateData {
	var matched []AspectTemplateData

	for _, aspect := range aspects {
		for _, advice := range aspect.Advices {
			for _, target := range advice.Targets {
				if g.isMatchForInterface(target, intf) {
					matched = append(matched, AspectTemplateData{
						MethodName:       g.extractMethodName(target),
						AdviceType:       toTitleCase(string(advice.Type)),
						AdviceFunc:       g.buildAdviceFuncName(aspect.Name, advice.Method, advice.IsFunc),
						Order:            aspect.Order,
						AspectName:       aspect.Name,
						AspectMethodName: advice.Method,
					})
				}
			}
		}
	}

	for _, advice := range funcs {
		for _, target := range advice.Targets {
			if g.isMatchForInterface(target, intf) {
				matched = append(matched, AspectTemplateData{
					MethodName: g.extractMethodName(target),
					AdviceType: toTitleCase(string(advice.Type)),
					AdviceFunc: advice.FuncName,
					Order:      0,
				})
			}
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Order != matched[j].Order {
			return matched[i].Order < matched[j].Order
		}
		return matched[i].MethodName < matched[j].MethodName
	})

	return matched
}

// isMatchForInterface 判断目标表达式是否匹配接口
func (g *Generator) isMatchForInterface(target string, intf InterfaceInfo) bool {
	parts := strings.Split(target, ".")
	if len(parts) != 2 {
		return false
	}
	return parts[0] == intf.Name
}

// matchAspects 匹配代理目标与切面通知
func (g *Generator) matchAspects(proxy ProxyInfo, aspects []AspectInfo, funcs []AdviceInfo) []AspectTemplateData {
	var matched []AspectTemplateData

	for _, aspect := range aspects {
		for _, advice := range aspect.Advices {
			for _, target := range advice.Targets {
				if g.isMatch(target, proxy) {
					matched = append(matched, AspectTemplateData{
						MethodName:       g.extractMethodName(target),
						AdviceType:       toTitleCase(string(advice.Type)),
						AdviceFunc:       g.buildAdviceFuncName(aspect.Name, advice.Method, advice.IsFunc),
						Order:            aspect.Order,
						AspectName:       aspect.Name,
						AspectMethodName: advice.Method,
					})
				}
			}
		}
	}

	for _, advice := range funcs {
		for _, target := range advice.Targets {
			if g.isMatch(target, proxy) {
				matched = append(matched, AspectTemplateData{
					MethodName: g.extractMethodName(target),
					AdviceType: toTitleCase(string(advice.Type)),
					AdviceFunc: advice.FuncName,
					Order:      0,
				})
			}
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Order != matched[j].Order {
			return matched[i].Order < matched[j].Order
		}
		return matched[i].MethodName < matched[j].MethodName
	})

	return matched
}

// isMatch 判断目标表达式是否匹配代理
func (g *Generator) isMatch(target string, proxy ProxyInfo) bool {
	parts := strings.Split(target, ".")
	if len(parts) != 2 {
		return false
	}
	return parts[0] == proxy.Name
}

// extractMethodName 从目标表达式中提取方法名
func (g *Generator) extractMethodName(target string) string {
	parts := strings.Split(target, ".")
	if len(parts) == 2 {
		return parts[1]
	}
	return target
}

// buildAdviceFuncName 构建通知函数名
func (g *Generator) buildAdviceFuncName(aspectName, methodName string, isFunc bool) string {
	if isFunc {
		return methodName
	}
	return strings.ToLower(aspectName[:1]) + aspectName[1:] + methodName
}

// buildProxyTemplateData 构建代理模板数据
func (g *Generator) buildProxyTemplateData(proxy ProxyInfo, aspects []AspectTemplateData, mode string) ProxyTemplateData {
	var imports []string

	switch mode {
	case "static":
		imports = []string{
			"github.com/xudefa/enhance/core",
			"reflect",
		}
	case "aop":
		imports = []string{
			"github.com/xudefa/enhance/aop",
			"github.com/xudefa/enhance/core",
			"reflect",
		}
	default:
		imports = []string{
			"github.com/xudefa/enhance/core",
			"reflect",
		}
	}

	methods := make([]MethodTemplateData, 0, len(proxy.Methods))
	for _, method := range proxy.Methods {
		if !method.Exported {
			continue
		}
		if mode == "static" {
			methods = append(methods, buildStaticMethodTemplateData(method, aspects))
			continue
		}
		methods = append(methods, buildMethodTemplateData(method))
	}

	return ProxyTemplateData{
		Package:      proxy.Package,
		ProxyName:    proxy.Name + "Proxy",
		TargetName:   proxy.Target,
		BeanID:       proxy.BeanID,
		Imports:      imports,
		Aspects:      aspects,
		Methods:      methods,
		Dependencies: []string{},
	}
}

// buildInterfaceProxyTemplateData 构建接口代理模板数据
func (g *Generator) buildInterfaceProxyTemplateData(intf InterfaceInfo, aspects []AspectTemplateData, mode string) ProxyTemplateData {
	var imports []string

	switch mode {
	case "static":
		imports = []string{
			"github.com/xudefa/enhance/core",
			"reflect",
		}
	case "aop":
		imports = []string{
			"github.com/xudefa/enhance/aop",
			"github.com/xudefa/enhance/core",
			"reflect",
		}
	default:
		imports = []string{
			"github.com/xudefa/enhance/core",
			"reflect",
		}
	}

	methods := make([]MethodTemplateData, 0, len(intf.Methods))
	for _, method := range intf.Methods {
		if !method.Exported {
			continue
		}
		if mode == "static" {
			methods = append(methods, buildStaticMethodTemplateData(method, aspects))
			continue
		}
		methods = append(methods, buildMethodTemplateData(method))
	}

	return ProxyTemplateData{
		Package:      intf.Package,
		ProxyName:    intf.Name + "Proxy",
		TargetName:   intf.Name,
		BeanID:       intf.BeanID,
		Imports:      imports,
		Aspects:      aspects,
		Methods:      methods,
		Dependencies: []string{},
	}
}

// getOutputPath 获取代理文件输出路径
func (g *Generator) getOutputPath(sourcePath string) string {
	dir := filepath.Dir(sourcePath)
	base := filepath.Base(sourcePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return filepath.Join(dir, name+"_proxy.go")
}

// Clean 清除指定目录下所有生成的代理文件（*_proxy.go）
func (g *Generator) Clean(dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "_proxy.go") {
			return os.Remove(path)
		}
		return nil
	})
}

// GetRegistry 获取代理注册表
func (g *Generator) GetRegistry() *Registry {
	return g.registry
}

// toTitleCase 将字符串首字母转为大写
func toTitleCase(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
