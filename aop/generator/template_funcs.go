package generator

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// ProxyTemplateData 代理模板数据
type ProxyTemplateData struct {
	Package      string               // 包名
	ProxyName    string               // 代理类名
	TargetName   string               // 目标结构体名
	BeanID       string               // Bean 标识
	Imports      []string             // 导入列表
	Aspects      []AspectTemplateData // 切面数据列表
	Methods      []MethodTemplateData // 方法数据列表
	Dependencies []string             // 依赖列表
}

// AspectTemplateData 切面模板数据
type AspectTemplateData struct {
	MethodName       string // 目标方法名
	AdviceType       string // 通知类型（Before/After/Around 等）
	AdviceFunc       string // 通知函数名
	Order            int    // 切面优先级
	AspectName       string // 切面名称
	AspectMethodName string // 切面方法名
}

// MethodTemplateData 方法模板数据
type MethodTemplateData struct {
	Name                  string              // 方法名
	ParamsStr             string              // 参数列表字符串
	ResultsStr            string              // 返回值列表字符串
	ArgsList              string              // 参数名列表字符串
	HasError              bool                // 是否包含 error 返回值
	HasMultipleReturns    bool                // 是否有多个返回值
	HasSingleReturn       bool                // 是否仅有单个非 error 返回值
	HasSingleErrorReturn  bool                // 是否仅有单个 error 返回值
	NoReturn              bool                // 是否无返回值
	SingleReturnType      string              // 单返回值类型
	ReturnValues          string              // 返回值表达式
	ReturnValuesWithError string              // 含 error 的返回值表达式
	ReturnValuesFromTuple string              // 从元组解析的返回值表达式
	ReturnValuesFallback  string              // 返回值回退表达式
	HasAspects            bool                // 是否有切面增强
	HasReturnValue        bool                // 是否有返回值（用于静态模板）
	HasParams             bool                // 是否有参数（用于静态模板）
	HasBeforeAdvices      bool                // 是否有 Before 通知
	HasAroundAdvices      bool                // 是否有 Around 通知
	FirstAroundAdviceFunc string              // 第一个 Around 通知函数名
	BeforeAdvices         []AdviceBindingData // Before 通知列表
	AroundAdvices         []AdviceBindingData // Around 通知列表
	AfterAdvices          []AdviceBindingData // After 通知列表
	AfterReturningAdvices []AdviceBindingData // AfterReturning 通知列表
	AfterThrowingAdvices  []AdviceBindingData // AfterThrowing 通知列表
}

// AdviceBindingData 通知绑定数据
type AdviceBindingData struct {
	AdviceFunc string // 通知函数名
	HasParams  bool   // 是否有参数
}

// AdviceAdapterTemplateData 通知适配器模板数据
type AdviceAdapterTemplateData struct {
	Package  string              // 包名
	Adapters []AdviceAdapterData // 适配器列表
}

// AdviceAdapterData 通知适配器数据
type AdviceAdapterData struct {
	FuncName   string // 适配器函数名
	AspectType string // 切面类型名
	MethodName string // 切面方法名
	IsAround   bool   // 是否为环绕通知
	HasReturn  bool   // 是否有返回值
}

// TemplateEngine 代码模板引擎
//
// 管理代理代码生成的 Go 模板，支持简单代理、AOP 增强代理、静态 AOP 代理和静态接口代理四种模式。
type TemplateEngine struct {
	simpleProxyTemplate          *template.Template // 简单代理模板
	aopProxyTemplate             *template.Template // AOP 增强代理模板
	staticAopProxyTemplate       *template.Template // 静态 AOP 代理模板（零反射）
	staticInterfaceProxyTemplate *template.Template // 静态接口代理模板（零反射）
	adapterTemplate              *template.Template // 通知适配器模板
}

// NewTemplateEngine 创建代码模板引擎
func NewTemplateEngine() (*TemplateEngine, error) {
	simpleTmpl, err := template.New("simpleProxy").Parse(simpleProxyTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse simple proxy template: %w", err)
	}

	aopTmpl, err := template.New("aopProxy").Parse(aopProxyTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AOP proxy template: %w", err)
	}

	staticTmpl, err := template.New("staticAopProxy").Parse(staticAopProxyTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse static AOP proxy template: %w", err)
	}

	staticInterfaceTmpl, err := template.New("staticInterfaceProxy").Parse(staticInterfaceProxyTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse static interface proxy template: %w", err)
	}

	adapterTmpl, err := template.New("adapter").Parse(adviceAdapterTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse adapter template: %w", err)
	}

	return &TemplateEngine{
		simpleProxyTemplate:          simpleTmpl,
		aopProxyTemplate:             aopTmpl,
		staticAopProxyTemplate:       staticTmpl,
		staticInterfaceProxyTemplate: staticInterfaceTmpl,
		adapterTemplate:              adapterTmpl,
	}, nil
}

// GenerateProxy 根据模板数据生成代理代码
//
// mode 参数: "simple" 使用简单代理模板, "aop" 使用 AOP 增强模板, "static" 使用静态 AOP 模板（零反射）。
func (e *TemplateEngine) GenerateProxy(data ProxyTemplateData, mode string) (string, error) {
	var buf bytes.Buffer
	var tmpl *template.Template

	switch mode {
	case "static":
		tmpl = e.staticAopProxyTemplate
	case "aop":
		tmpl = e.aopProxyTemplate
	default:
		tmpl = e.simpleProxyTemplate
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute proxy template: %w", err)
	}
	return buf.String(), nil
}

// GenerateInterfaceProxy 根据模板数据生成接口代理代码
//
// mode 参数: "simple" 使用简单代理模板, "aop" 使用 AOP 增强模板, "static" 使用静态接口代理模板（零反射）。
func (e *TemplateEngine) GenerateInterfaceProxy(data ProxyTemplateData, mode string) (string, error) {
	var buf bytes.Buffer
	var tmpl *template.Template

	switch mode {
	case "static":
		tmpl = e.staticInterfaceProxyTemplate
	case "aop":
		tmpl = e.aopProxyTemplate
	default:
		tmpl = e.simpleProxyTemplate
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute interface proxy template: %w", err)
	}
	return buf.String(), nil
}

// GenerateAdviceAdapter 生成通知适配器代码
func (e *TemplateEngine) GenerateAdviceAdapter(data AdviceAdapterTemplateData) (string, error) {
	var buf bytes.Buffer
	if err := e.adapterTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute adapter template: %w", err)
	}
	return buf.String(), nil
}

// buildMethodTemplateData 从方法信息构建方法模板数据
func buildMethodTemplateData(method MethodInfo) MethodTemplateData {
	var params []string
	for _, param := range method.Params {
		if param.Name != "" {
			params = append(params, fmt.Sprintf("%s %s", param.Name, param.Type))
			continue
		}
		params = append(params, param.Type)
	}

	var results []string
	for _, result := range method.Results {
		if result.Name != "" {
			results = append(results, fmt.Sprintf("%s %s", result.Name, result.Type))
			continue
		}
		results = append(results, result.Type)
	}

	var resultsStr string
	if len(results) > 1 {
		resultsStr = fmt.Sprintf("(%s)", strings.Join(results, ", "))
	} else if len(results) == 1 {
		resultsStr = results[0]
	}

	var argsList []string
	for _, param := range method.Params {
		argsList = append(argsList, param.Name)
	}

	hasError := false
	for _, result := range method.Results {
		if result.Type == "error" {
			hasError = true
			break
		}
	}

	var returnValues []string
	var returnValuesWithError []string
	var returnValuesFromTuple []string
	var returnValuesFallback []string
	for i, result := range method.Results {
		if result.Type == "error" {
			returnValuesWithError = append(returnValuesWithError, "err")
			returnValuesFallback = append(returnValuesFallback, "nil")
			continue
		}
		if hasError {
			returnValuesWithError = append(returnValuesWithError, fmt.Sprintf("tuple[%d].(%s)", i, result.Type))
		} else {
			returnValues = append(returnValues, fmt.Sprintf("tuple[%d].(%s)", i, result.Type))
		}
		returnValuesFromTuple = append(returnValuesFromTuple, fmt.Sprintf("tuple[%d].(%s)", i, result.Type))
		returnValuesFallback = append(returnValuesFallback, "nil")
	}

	if len(returnValues) == 0 {
		returnValues = append(returnValues, "result")
	}

	if len(returnValuesFallback) == 0 {
		for range method.Results {
			returnValuesFallback = append(returnValuesFallback, "nil")
		}
	}

	if len(returnValuesFromTuple) == 0 {
		returnValuesFromTuple = returnValuesFallback
	}

	if hasError {
		returnValuesFromTuple = append(returnValuesFromTuple, "nil")
	}

	hasMultipleReturns := len(method.Results) > 1
	hasSingleReturn := len(method.Results) == 1 && !hasError
	hasSingleErrorReturn := len(method.Results) == 1 && hasError
	noReturn := len(method.Results) == 0
	singleReturnType := ""
	if hasSingleReturn {
		singleReturnType = method.Results[0].Type
	}

	return MethodTemplateData{
		Name:                  method.Name,
		ParamsStr:             strings.Join(params, ", "),
		ResultsStr:            resultsStr,
		ArgsList:              strings.Join(argsList, ", "),
		HasError:              hasError,
		HasMultipleReturns:    hasMultipleReturns,
		HasSingleReturn:       hasSingleReturn,
		HasSingleErrorReturn:  hasSingleErrorReturn,
		NoReturn:              noReturn,
		SingleReturnType:      singleReturnType,
		ReturnValues:          strings.Join(returnValues, ", "),
		ReturnValuesWithError: strings.Join(returnValuesWithError, ", "),
		ReturnValuesFromTuple: strings.Join(returnValuesFromTuple, ", "),
		ReturnValuesFallback:  strings.Join(returnValuesFallback, ", "),
		HasReturnValue:        len(method.Results) > 0,
		HasParams:             len(method.Params) > 0,
	}
}

// buildStaticMethodTemplateData 构建静态 AOP 代理模板数据
//
// 根据切面信息为每个方法生成对应的通知绑定数据。
func buildStaticMethodTemplateData(method MethodInfo, aspects []AspectTemplateData) MethodTemplateData {
	base := buildMethodTemplateData(method)

	var beforeAdvices, aroundAdvices, afterAdvices, afterReturningAdvices, afterThrowingAdvices []AdviceBindingData

	for _, aspect := range aspects {
		if aspect.MethodName != method.Name {
			continue
		}

		binding := AdviceBindingData{
			AdviceFunc: aspect.AdviceFunc,
			HasParams:  len(method.Params) > 0,
		}

		switch aspect.AdviceType {
		case "Before":
			beforeAdvices = append(beforeAdvices, binding)
		case "Around":
			aroundAdvices = append(aroundAdvices, binding)
		case "After":
			afterAdvices = append(afterAdvices, binding)
		case "AfterReturning":
			afterReturningAdvices = append(afterReturningAdvices, binding)
		case "AfterThrowing":
			afterThrowingAdvices = append(afterThrowingAdvices, binding)
		}
	}

	base.HasAspects = len(beforeAdvices)+len(aroundAdvices)+len(afterAdvices)+len(afterReturningAdvices)+len(afterThrowingAdvices) > 0
	base.HasBeforeAdvices = len(beforeAdvices) > 0
	base.HasAroundAdvices = len(aroundAdvices) > 0
	if len(aroundAdvices) > 0 {
		base.FirstAroundAdviceFunc = aroundAdvices[0].AdviceFunc
	}
	base.BeforeAdvices = beforeAdvices
	base.AroundAdvices = aroundAdvices
	base.AfterAdvices = afterAdvices
	base.AfterReturningAdvices = afterReturningAdvices
	base.AfterThrowingAdvices = afterThrowingAdvices

	return base
}
