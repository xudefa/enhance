package generator

import (
	"fmt"
	"go/ast"
	"strings"
)

// parseAspectAnnotation 解析 @Aspect 注解，提取切面信息
func (p *Parser) parseAspectAnnotation(text, structName, pkgName string) {
	order := 0

	if idx := strings.Index(text, "order="); idx >= 0 {
		start := idx + 6
		end := strings.IndexAny(text[start:], ",)")
		if end >= 0 {
			if _, err := fmt.Sscanf(text[start:start+end], "%d", &order); err != nil {
				fmt.Printf("[enhance] failed to parse aspect order from annotation: %v\n", err)
			}
		}
	}

	p.aspects[structName] = &AspectInfo{
		Name:    structName,
		Order:   order,
		Package: pkgName,
		Advices: []AdviceInfo{},
	}
}

// parseProxyAnnotation 解析 @AopProxy 注解，提取代理信息
func (p *Parser) parseProxyAnnotation(text, structName, pkgName, filePath string) {
	beanID := strings.ToLower(string(structName[0])) + structName[1:]

	if idx := strings.Index(text, "beanId="); idx >= 0 {
		start := idx + 7
		end := strings.IndexAny(text[start:], ",)")
		if end >= 0 {
			beanID = strings.Trim(text[start:start+end], `"`)
		}
	}

	p.proxies[structName] = &ProxyInfo{
		Name:     structName,
		Package:  pkgName,
		FilePath: filePath,
		Target:   structName,
		Methods:  []MethodInfo{},
		Aspects:  []AspectInfo{},
		BeanID:   beanID,
	}
}

// parseInterfaceAnnotation 解析 @ProxyInterface 注解，提取接口信息
func (p *Parser) parseInterfaceAnnotation(text, interfaceName, pkgName, filePath string) {
	beanID := strings.ToLower(string(interfaceName[0])) + interfaceName[1:]

	if idx := strings.Index(text, "beanId="); idx >= 0 {
		start := idx + 7
		end := strings.IndexAny(text[start:], ",)")
		if end >= 0 {
			beanID = strings.Trim(text[start:start+end], `"`)
		}
	}

	p.interfaces[interfaceName] = &InterfaceInfo{
		Name:     interfaceName,
		Package:  pkgName,
		FilePath: filePath,
		Methods:  []MethodInfo{},
		BeanID:   beanID,
		Aspects:  []AspectInfo{},
	}
}

// parseInterfaceMethods 解析接口定义中的方法
func (p *Parser) parseInterfaceMethods(intf *ast.InterfaceType, interfaceName, pkgName, filePath string) {
	if intf.Methods == nil {
		return
	}

	_, exists := p.interfaces[interfaceName]
	if !exists {
		return
	}

	for _, method := range intf.Methods.List {
		if len(method.Names) == 0 {
			continue
		}

		methodName := method.Names[0].Name
		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		methodInfo := MethodInfo{
			Name:     methodName,
			Receiver: interfaceName,
			Exported: methodName[0] >= 'A' && methodName[0] <= 'Z',
		}

		if funcType.Params != nil {
			for _, param := range funcType.Params.List {
				paramType := p.exprToString(param.Type)
				for _, name := range param.Names {
					methodInfo.Params = append(methodInfo.Params, ParamInfo{
						Name: name.Name,
						Type: paramType,
					})
				}
				if len(param.Names) == 0 {
					methodInfo.Params = append(methodInfo.Params, ParamInfo{
						Name: "",
						Type: paramType,
					})
				}
			}
		}

		if funcType.Results != nil {
			for _, result := range funcType.Results.List {
				resultType := p.exprToString(result.Type)
				for _, name := range result.Names {
					methodInfo.Results = append(methodInfo.Results, ParamInfo{
						Name: name.Name,
						Type: resultType,
					})
				}
				if len(result.Names) == 0 {
					methodInfo.Results = append(methodInfo.Results, ParamInfo{
						Name: "",
						Type: resultType,
					})
				}
			}
		}

		p.interfaces[interfaceName].Methods = append(p.interfaces[interfaceName].Methods, methodInfo)
	}
}

// parseFuncDecl 解析函数声明，提取代理方法或独立通知函数
func (p *Parser) parseFuncDecl(decl *ast.FuncDecl, pkgName string) {
	if decl.Recv == nil {
		p.parseStandaloneFunc(decl, pkgName)
		return
	}

	recvType := p.resolveRecvType(decl.Recv)
	if recvType == "" {
		return
	}

	aspect, isAspect := p.aspects[recvType]
	proxy, isProxy := p.proxies[recvType]
	intf, isInterface := p.interfaces[recvType]

	if !isAspect && !isProxy && !isInterface {
		return
	}

	methodName := decl.Name.Name

	if isProxy {
		method := p.parseMethodInfo(decl, methodName)
		proxy.Methods = append(proxy.Methods, method)
	} else if isInterface {
		method := p.parseMethodInfo(decl, methodName)
		intf.Methods = append(intf.Methods, method)
	}

	if isAspect && decl.Doc != nil {
		p.parseAspectMethod(decl, aspect, methodName, pkgName)
	}
}

// parseStandaloneFunc 解析独立函数上的通知注解
func (p *Parser) parseStandaloneFunc(decl *ast.FuncDecl, pkgName string) {
	if decl.Doc == nil {
		return
	}

	funcName := decl.Name.Name

	for _, comment := range decl.Doc.List {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimSpace(text)

		adviceType := p.extractAdviceType(text)
		if adviceType == "" {
			continue
		}

		targets := p.extractTargets(text)
		if len(targets) == 0 {
			continue
		}

		p.funcs[funcName] = &AdviceInfo{
			Type:     adviceType,
			Method:   funcName,
			Targets:  targets,
			IsFunc:   true,
			FuncName: funcName,
			Package:  pkgName,
		}
	}
}

// parseAspectMethod 解析切面方法上的通知注解
func (p *Parser) parseAspectMethod(decl *ast.FuncDecl, aspect *AspectInfo, methodName, pkgName string) {
	for _, comment := range decl.Doc.List {
		text := strings.TrimSpace(comment.Text)
		text = strings.TrimPrefix(text, "//")
		text = strings.TrimSpace(text)

		adviceType := p.extractAdviceType(text)
		if adviceType == "" {
			continue
		}

		targets := p.extractTargets(text)
		if len(targets) == 0 {
			continue
		}

		aspect.Advices = append(aspect.Advices, AdviceInfo{
			Type:       adviceType,
			Method:     methodName,
			Targets:    targets,
			IsFunc:     false,
			FuncName:   "",
			Package:    pkgName,
			AspectName: aspect.Name,
		})
	}
}

// parseMethodInfo 解析方法签名信息
func (p *Parser) parseMethodInfo(decl *ast.FuncDecl, methodName string) MethodInfo {
	method := MethodInfo{
		Name:     methodName,
		Receiver: p.resolveRecvType(decl.Recv),
		Exported: decl.Name.IsExported(),
	}

	if decl.Type.Params != nil {
		for _, param := range decl.Type.Params.List {
			paramType := p.exprToString(param.Type)
			for _, name := range param.Names {
				method.Params = append(method.Params, ParamInfo{
					Name: name.Name,
					Type: paramType,
				})
			}
			if len(param.Names) == 0 {
				method.Params = append(method.Params, ParamInfo{
					Name: "",
					Type: paramType,
				})
			}
		}
	}

	if decl.Type.Results != nil {
		for _, result := range decl.Type.Results.List {
			resultType := p.exprToString(result.Type)
			for _, name := range result.Names {
				method.Results = append(method.Results, ParamInfo{
					Name: name.Name,
					Type: resultType,
				})
			}
			if len(result.Names) == 0 {
				method.Results = append(method.Results, ParamInfo{
					Name: "",
					Type: resultType,
				})
			}
		}
	}

	return method
}

// resolveRecvType 解析方法接收者类型名
func (p *Parser) resolveRecvType(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	expr := recv.List[0].Type
	switch t := expr.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// exprToString 将 AST 表达式转换为类型字符串
func (p *Parser) exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + p.exprToString(t.X)
	case *ast.ArrayType:
		return "[]" + p.exprToString(t.Elt)
	case *ast.SelectorExpr:
		return p.exprToString(t.X) + "." + t.Sel.Name
	case *ast.FuncType:
		return "func"
	case *ast.InterfaceType:
		return "any"
	case *ast.MapType:
		return "map[" + p.exprToString(t.Key) + "]" + p.exprToString(t.Value)
	default:
		return ""
	}
}

// extractAdviceType 从注解文本中提取通知类型
func (p *Parser) extractAdviceType(text string) AdviceType {
	text = strings.ToLower(text)
	if strings.HasPrefix(text, "@afterreturning") {
		return AdviceAfterReturning
	}
	if strings.HasPrefix(text, "@afterthrowing") {
		return AdviceAfterThrowing
	}
	if strings.HasPrefix(text, "@before") {
		return AdviceBefore
	}
	if strings.HasPrefix(text, "@after") {
		return AdviceAfter
	}
	if strings.HasPrefix(text, "@around") {
		return AdviceAround
	}
	return ""
}

// extractTargets 从注解文本中提取目标方法列表
func (p *Parser) extractTargets(text string) []string {
	start := strings.Index(text, "(")
	if start < 0 {
		return nil
	}
	end := strings.LastIndex(text, ")")
	if end <= start {
		return nil
	}

	paramsStr := text[start+1 : end]
	targets := strings.Split(paramsStr, ",")

	var result []string
	for _, target := range targets {
		target = strings.TrimSpace(target)
		target = strings.Trim(target, `"`)
		if target != "" {
			result = append(result, target)
		}
	}

	return result
}
