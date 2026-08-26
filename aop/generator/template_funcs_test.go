package generator

import (
	"strings"
	"testing"
)

func TestUniqueAdviceFuncs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  []AspectTemplateData
		expect int
	}{
		{
			name:   "empty",
			input:  nil,
			expect: 0,
		},
		{
			name: "no duplicates",
			input: []AspectTemplateData{
				{AdviceFunc: "a"},
				{AdviceFunc: "b"},
			},
			expect: 2,
		},
		{
			name: "with duplicates",
			input: []AspectTemplateData{
				{AdviceFunc: "a"},
				{AdviceFunc: "b"},
				{AdviceFunc: "a"},
			},
			expect: 2,
		},
		{
			name: "empty AdviceFunc skipped",
			input: []AspectTemplateData{
				{AdviceFunc: ""},
				{AdviceFunc: "a"},
			},
			expect: 1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := uniqueAdviceFuncs(tt.input)
			if len(got) != tt.expect {
				t.Errorf("uniqueAdviceFuncs() len = %d, want %d", len(got), tt.expect)
			}
		})
	}
}

func TestBuildMethodTemplateData_NoParamsNoReturn(t *testing.T) {
	t.Parallel()
	m := MethodInfo{Name: "Do"}
	data := buildMethodTemplateData(m)
	if data.Name != "Do" {
		t.Errorf("Name = %q", data.Name)
	}
	if data.HasParams || data.HasReturnValue {
		t.Error("expected no params and no return")
	}
	if !data.NoReturn {
		t.Error("expected NoReturn=true")
	}
	if data.ParamsStr != "" {
		t.Errorf("ParamsStr = %q, want empty", data.ParamsStr)
	}
}

func TestBuildMethodTemplateData_WithParamsAndError(t *testing.T) {
	t.Parallel()
	m := MethodInfo{
		Name: "Save",
		Params: []ParamInfo{
			{Name: "ctx", Type: "context.Context"},
			{Name: "data", Type: "string"},
		},
		Results: []ParamInfo{
			{Name: "", Type: "error"},
		},
	}
	data := buildMethodTemplateData(m)
	if !data.HasParams {
		t.Error("expected HasParams=true")
	}
	if !data.HasError {
		t.Error("expected HasError=true")
	}
	if !data.HasSingleErrorReturn {
		t.Error("expected HasSingleErrorReturn=true")
	}
	if data.HasMultipleReturns {
		t.Error("expected HasMultipleReturns=false")
	}
	if data.ArgsList != "ctx, data" {
		t.Errorf("ArgsList = %q, want %q", data.ArgsList, "ctx, data")
	}
}

func TestBuildMethodTemplateData_MultipleReturns(t *testing.T) {
	t.Parallel()
	m := MethodInfo{
		Name: "Get",
		Results: []ParamInfo{
			{Name: "", Type: "string"},
			{Name: "", Type: "error"},
		},
	}
	data := buildMethodTemplateData(m)
	if !data.HasMultipleReturns {
		t.Error("expected HasMultipleReturns=true")
	}
	if !data.HasReturnValue {
		t.Error("expected HasReturnValue=true")
	}
	if data.ResultsStr != "(string, error)" {
		t.Errorf("ResultsStr = %q", data.ResultsStr)
	}
}

func TestBuildMethodTemplateData_SingleReturn(t *testing.T) {
	t.Parallel()
	m := MethodInfo{
		Name: "Name",
		Results: []ParamInfo{
			{Name: "", Type: "string"},
		},
	}
	data := buildMethodTemplateData(m)
	if !data.HasSingleReturn {
		t.Error("expected HasSingleReturn=true")
	}
	if data.SingleReturnType != "string" {
		t.Errorf("SingleReturnType = %q, want %q", data.SingleReturnType, "string")
	}
}

func TestBuildStaticMethodTemplateData(t *testing.T) {
	t.Parallel()
	method := MethodInfo{
		Name:   "GetUser",
		Params: []ParamInfo{{Name: "id", Type: "int64"}},
		Results: []ParamInfo{
			{Name: "", Type: "*User"},
			{Name: "", Type: "error"},
		},
	}
	aspects := []AspectTemplateData{
		{MethodName: "GetUser", AdviceType: "Before", AdviceFunc: "logBefore"},
		{MethodName: "GetUser", AdviceType: "AfterReturning", AdviceFunc: "logAfterReturning"},
		{MethodName: "Other", AdviceType: "Before", AdviceFunc: "logOther"},
	}
	data := buildStaticMethodTemplateData(method, aspects)
	if !data.HasAspects {
		t.Error("expected HasAspects=true")
	}
	if !data.HasBeforeAdvices {
		t.Error("expected HasBeforeAdvices=true")
	}
	if len(data.AfterReturningAdvices) != 1 {
		t.Errorf("expected 1 AfterReturning, got %d", len(data.AfterReturningAdvices))
	}
}

func TestSafeTupleAccessVariants(t *testing.T) {
	t.Parallel()
	got := safeTupleAccess(0, "string")
	if !strings.Contains(got, "string") || !strings.Contains(got, "tuple[0]") {
		t.Errorf("safeTupleAccess = %q", got)
	}
}

func TestTemplateEngine_NewTemplateEngine(t *testing.T) {
	t.Parallel()
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine: %v", err)
	}
	if engine == nil {
		t.Fatal("engine should not be nil")
	}
}

func TestTemplateEngine_GenerateProxy_Simple(t *testing.T) {
	t.Parallel()
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine: %v", err)
	}
	data := ProxyTemplateData{
		Package:    "testpkg",
		ProxyName:  "TestProxy",
		TargetName: "TestTarget",
		BeanID:     "testProxy",
		Imports:    []string{"fmt"},
		Methods: []MethodTemplateData{
			{
				Name:       "Do",
				ParamsStr:  "x int",
				ResultsStr: "string",
				ArgsList:   "x",
				NoReturn:   false,
			},
		},
		IsInterface: false,
	}
	got, err := engine.GenerateProxy(data, "simple")
	if err != nil {
		t.Fatalf("GenerateProxy: %v", err)
	}
	if !strings.Contains(got, "testpkg") {
		t.Error("expected package name in output")
	}
	if !strings.Contains(got, "TestProxy") {
		t.Error("expected proxy type in output")
	}
}

func TestTemplateEngine_GenerateProxy_AOP(t *testing.T) {
	t.Parallel()
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine: %v", err)
	}
	data := ProxyTemplateData{
		Package:    "testpkg",
		ProxyName:  "AopProxy",
		TargetName: "Target",
		BeanID:     "aopProxy",
		Methods:    []MethodTemplateData{},
	}
	got, err := engine.GenerateProxy(data, "aop")
	if err != nil {
		t.Fatalf("GenerateProxy: %v", err)
	}
	if !strings.Contains(got, "AopProxy") {
		t.Error("expected proxy type")
	}
}

func TestTemplateEngine_GenerateProxy_Static(t *testing.T) {
	t.Parallel()
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine: %v", err)
	}
	data := ProxyTemplateData{
		Package:    "testpkg",
		ProxyName:  "StaticProxy",
		TargetName: "Target",
		BeanID:     "staticProxy",
		Methods:    []MethodTemplateData{},
	}
	got, err := engine.GenerateProxy(data, "static")
	if err != nil {
		t.Fatalf("GenerateProxy: %v", err)
	}
	if !strings.Contains(got, "StaticProxy") {
		t.Error("expected proxy type")
	}
}

func TestTemplateEngine_GenerateInterfaceProxy_Simple(t *testing.T) {
	t.Parallel()
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine: %v", err)
	}
	data := ProxyTemplateData{
		Package:    "testpkg",
		ProxyName:  "IfaceProxy",
		TargetName: "MyInterface",
		BeanID:     "ifaceProxy",
		Methods:    []MethodTemplateData{},
		IsInterface: true,
	}
	got, err := engine.GenerateInterfaceProxy(data, "simple")
	if err != nil {
		t.Fatalf("GenerateInterfaceProxy: %v", err)
	}
	if !strings.Contains(got, "IfaceProxy") {
		t.Error("expected proxy type")
	}
}

func TestTemplateEngine_GenerateAdviceAdapterVariants(t *testing.T) {
	t.Parallel()
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine: %v", err)
	}
	data := AdviceAdapterTemplateData{
		Package: "testpkg",
		Adapters: []AdviceAdapterData{
			{FuncName: "adaptLog", AspectType: "LogAspect", MethodName: "Log", IsAround: false, HasReturn: false},
			{FuncName: "adaptAround", AspectType: "TracingAspect", MethodName: "Trace", IsAround: true, HasReturn: true},
		},
	}
	got, err := engine.GenerateAdviceAdapter(data)
	if err != nil {
		t.Fatalf("GenerateAdviceAdapter: %v", err)
	}
	if !strings.Contains(got, "adaptLog") || !strings.Contains(got, "adaptAround") {
		t.Error("expected adapter function names in output")
	}
}

func TestProxyTemplateData_Fields(t *testing.T) {
	t.Parallel()
	d := ProxyTemplateData{
		Package:     "pkg",
		ProxyName:   "Proxy",
		TargetName:  "Target",
		BeanID:      "bean1",
		IsInterface: true,
	}
	if d.Package != "pkg" || d.BeanID != "bean1" || !d.IsInterface {
		t.Errorf("unexpected: %+v", d)
	}
}

func TestMethodTemplateData_Fields(t *testing.T) {
	t.Parallel()
	d := MethodTemplateData{
		Name:                  "Do",
		HasError:              true,
		HasMultipleReturns:    true,
		HasReturnValue:        true,
		HasParams:             true,
		HasBeforeAdvices:      true,
		FirstAroundAdviceFunc: "aroundFn",
	}
	if !d.HasError || !d.HasMultipleReturns || !d.HasReturnValue || !d.HasParams || !d.HasBeforeAdvices {
		t.Errorf("unexpected: %+v", d)
	}
	if d.FirstAroundAdviceFunc != "aroundFn" {
		t.Errorf("FirstAroundAdviceFunc = %q", d.FirstAroundAdviceFunc)
	}
}
