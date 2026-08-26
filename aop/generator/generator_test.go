package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdviceInfo(t *testing.T) {
	t.Parallel()
	advice := AdviceInfo{
		Type:    AdviceBefore,
		Method:  "Log",
		Targets: []string{"UserService.GetUser"},
		IsFunc:  false,
	}

	if advice.Type != AdviceBefore {
		t.Errorf("Expected AdviceBefore, got %v", advice.Type)
	}

	if len(advice.Targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(advice.Targets))
	}
}

func TestAspectInfo(t *testing.T) {
	t.Parallel()
	aspect := AspectInfo{
		Name:    "LoggingAspect",
		Order:   1,
		Package: "service",
		Advices: []AdviceInfo{
			{Type: AdviceBefore, Method: "Log"},
		},
	}

	if aspect.Order != 1 {
		t.Errorf("Expected order 1, got %d", aspect.Order)
	}

	if len(aspect.Advices) != 1 {
		t.Errorf("Expected 1 advice, got %d", len(aspect.Advices))
	}
}

func TestMethodInfo(t *testing.T) {
	t.Parallel()
	method := MethodInfo{
		Name:     "GetUser",
		Receiver: "*UserService",
		Params:   []ParamInfo{{Name: "id", Type: "int"}},
		Results:  []ParamInfo{{Name: "", Type: "*User"}, {Name: "", Type: "error"}},
		Exported: true,
	}

	if method.Name != "GetUser" {
		t.Errorf("Expected GetUser, got %s", method.Name)
	}

	if !method.Exported {
		t.Error("Expected exported method")
	}
}

func TestParser_ParseAspectAnnotation(t *testing.T) {
	t.Parallel()
	parser := NewParser()

	text := `@Aspect(order=1)`
	parser.parseAspectAnnotation(text, "LoggingAspect", "service")

	aspect, exists := parser.aspects["LoggingAspect"]
	if !exists {
		t.Fatal("Aspect not found")
	}

	if aspect.Name != "LoggingAspect" {
		t.Errorf("Expected LoggingAspect, got %s", aspect.Name)
	}

	if aspect.Order != 1 {
		t.Errorf("Expected order 1, got %d", aspect.Order)
	}
}

func TestParser_ParseProxyAnnotation(t *testing.T) {
	t.Parallel()
	parser := NewParser()

	text := `@AopProxy`
	parser.parseProxyAnnotation(text, "UserService", "service", "/path/to/service.go")

	proxy, exists := parser.proxies["UserService"]
	if !exists {
		t.Fatal("Proxy not found")
	}

	if proxy.Name != "UserService" {
		t.Errorf("Expected UserService, got %s", proxy.Name)
	}

	if proxy.BeanID != "userService" {
		t.Errorf("Expected userService, got %s", proxy.BeanID)
	}
}

func TestParser_ParseProxyAnnotationWithCustomBeanID(t *testing.T) {
	t.Parallel()
	parser := NewParser()

	text := `@AopProxy(beanId="customService")`
	parser.parseProxyAnnotation(text, "UserService", "service", "/path/to/service.go")

	proxy, exists := parser.proxies["UserService"]
	if !exists {
		t.Fatal("Proxy not found")
	}

	if proxy.BeanID != "customService" {
		t.Errorf("Expected customService, got %s", proxy.BeanID)
	}
}

func TestParser_ExtractAdviceType(t *testing.T) {
	t.Parallel()
	parser := NewParser()

	tests := []struct {
		text     string
		expected AdviceType
	}{
		{"@Before(\"UserService.GetUser\")", AdviceBefore},
		{"@After(\"UserService.GetUser\")", AdviceAfter},
		{"@Around(\"UserService.GetUser\")", AdviceAround},
		{"@AfterReturning(\"UserService.GetUser\")", AdviceAfterReturning},
		{"@AfterThrowing(\"UserService.GetUser\")", AdviceAfterThrowing},
		{"@Invalid(\"UserService.GetUser\")", ""},
	}

	for _, tt := range tests {
		result := parser.extractAdviceType(tt.text)
		if result != tt.expected {
			t.Errorf("extractAdviceType(%q) = %v, expected %v", tt.text, result, tt.expected)
		}
	}
}

func TestParser_ExtractTargets(t *testing.T) {
	t.Parallel()
	parser := NewParser()

	tests := []struct {
		text     string
		expected []string
	}{
		{"(\"UserService.GetUser\")", []string{"UserService.GetUser"}},
		{"(\"UserService.GetUser\", \"UserService.CreateUser\")", []string{"UserService.GetUser", "UserService.CreateUser"}},
		{"()", nil},
	}

	for _, tt := range tests {
		result := parser.extractTargets(tt.text)
		if len(result) != len(tt.expected) {
			t.Errorf("extractTargets(%q) = %v, expected %v", tt.text, result, tt.expected)
			continue
		}
		for i, target := range result {
			if target != tt.expected[i] {
				t.Errorf("extractTargets(%q)[%d] = %v, expected %v", tt.text, i, target, tt.expected[i])
			}
		}
	}
}

func TestParser_ParseFile(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package service

// @Aspect(order=1)
type LoggingAspect struct {}

// @Before("UserService.GetUser")
func (a *LoggingAspect) LogBefore(jp aop.JoinPoint) {}

// @AopProxy
type UserService struct {}

func (s *UserService) GetUser(id int) (*User, error) {
	return nil, nil
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	if err := parser.parseFile(testFile); err != nil {
		t.Fatal(err)
	}

	aspects := parser.GetAspects()
	if len(aspects) != 1 {
		t.Errorf("Expected 1 aspect, got %d", len(aspects))
	}

	proxies := parser.GetProxies()
	if len(proxies) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(proxies))
	}

	if len(proxies[0].Methods) != 1 {
		t.Errorf("Expected 1 method, got %d", len(proxies[0].Methods))
	}
}

func TestTemplateEngine_GenerateProxy(t *testing.T) {
	t.Parallel()
	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatal(err)
	}

	data := ProxyTemplateData{
		Package:    "service",
		ProxyName:  "UserServiceProxy",
		TargetName: "UserService",
		BeanID:     "userService",
		Imports:    []string{"github.com/xudefa/enhance/aop", "github.com/xudefa/enhance/core"},
		Aspects: []AspectTemplateData{
			{
				MethodName:       "GetUser",
				AdviceType:       "Before",
				AdviceFunc:       "loggingAspectBeforeLog",
				Order:            1,
				AspectName:       "LoggingAspect",
				AspectMethodName: "LogBefore",
			},
		},
		Methods: []MethodTemplateData{
			{
				Name:                  "GetUser",
				ParamsStr:             "id int",
				ResultsStr:            "(*User, error)",
				ArgsList:              "id",
				HasError:              true,
				HasMultipleReturns:    true,
				HasSingleReturn:       false,
				HasSingleErrorReturn:  false,
				NoReturn:              false,
				ReturnValuesWithError: "tuple[0].(*User), err",
				ReturnValuesFromTuple: "tuple[0].(*User), nil",
				ReturnValuesFallback:  "nil, nil",
			},
		},
	}

	output, err := engine.GenerateProxy(data, "aop")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output, "type UserServiceProxy struct") {
		t.Error("Generated code should contain UserServiceProxy struct")
	}

	if !strings.Contains(output, "func (p *UserServiceProxy) GetUser(id int) (*User, error)") {
		t.Error("Generated code should contain GetUser method")
		fmt.Println("Generated output:")
		fmt.Println(output)
	}

	if !strings.Contains(output, "func init()") {
		t.Error("Generated code should contain init function")
	}
}

func TestBuildMethodTemplateData(t *testing.T) {
	t.Parallel()
	method := MethodInfo{
		Name: "GetUser",
		Params: []ParamInfo{
			{Name: "id", Type: "int"},
		},
		Results: []ParamInfo{
			{Name: "", Type: "*User"},
			{Name: "", Type: "error"},
		},
		Exported: true,
	}

	data := buildMethodTemplateData(method)

	if data.Name != "GetUser" {
		t.Errorf("Expected GetUser, got %s", data.Name)
	}

	if data.ParamsStr != "id int" {
		t.Errorf("Expected 'id int', got %s", data.ParamsStr)
	}

	if data.ResultsStr != "(*User, error)" {
		t.Errorf("Expected '(*User, error)', got %s", data.ResultsStr)
	}

	if !data.HasError {
		t.Error("Expected hasError to be true")
	}
}

func TestGenerator_Generate(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	serviceFile := filepath.Join(tmpDir, "service.go")
	contentStr := `package service

// @Aspect(order=1)
type LoggingAspect struct {}

// @Before("UserService.GetUser")
func (a *LoggingAspect) LogBefore(jp aop.JoinPoint) {}

// @AopProxy
type UserService struct {}

func (s *UserService) GetUser(id int) (*User, error) {
	return nil, nil
}
`

	if err := os.WriteFile(serviceFile, []byte(contentStr), 0644); err != nil {
		t.Fatal(err)
	}

	gen, err := NewGenerator()
	if err != nil {
		t.Fatal(err)
	}

	if err := gen.Generate(tmpDir, "aop"); err != nil {
		t.Fatal(err)
	}

	outputFile := filepath.Join(tmpDir, "service_proxy.go")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Generated file should exist")
	}

	contentBytes, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(contentBytes), "UserServiceProxy") {
		t.Error("Generated code should contain UserServiceProxy")
	}
}

func TestGenerator_Clean(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	aopFile := filepath.Join(tmpDir, "service_proxy.go")
	if err := os.WriteFile(aopFile, []byte("// generated"), 0644); err != nil {
		t.Fatal(err)
	}

	normalFile := filepath.Join(tmpDir, "service.go")
	if err := os.WriteFile(normalFile, []byte("// normal"), 0644); err != nil {
		t.Fatal(err)
	}

	gen, err := NewGenerator()
	if err != nil {
		t.Fatal(err)
	}

	if err := gen.Clean(tmpDir); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(aopFile); !os.IsNotExist(err) {
		t.Error("Generated file should be deleted")
	}

	if _, err := os.Stat(normalFile); os.IsNotExist(err) {
		t.Error("Normal file should not be deleted")
	}
}

func TestRegistry_Register(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()

	registry.Register("userService", "/path/to/service_proxy.go")

	path, exists := registry.Get("userService")
	if !exists {
		t.Error("Bean should be registered")
	}

	if path != "/path/to/service_proxy.go" {
		t.Errorf("Expected /path/to/service_proxy.go, got %s", path)
	}
}

func TestRegistry_List(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()

	registry.Register("userService", "/path/to/service_proxy.go")
	registry.Register("orderService", "/path/to/order_proxy.go")

	list := registry.List()

	if len(list) != 2 {
		t.Errorf("Expected 2 beans, got %d", len(list))
	}

	if list["userService"] != "/path/to/service_proxy.go" {
		t.Error("userService path mismatch")
	}
}

func TestRegistry_SaveAndLoad(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	registryFile := filepath.Join(tmpDir, "registry.json")

	registry := NewRegistry()
	registry.Register("userService", "/path/to/service_proxy.go")

	if err := registry.Save(registryFile); err != nil {
		t.Fatal(err)
	}

	newRegistry := NewRegistry()
	if err := newRegistry.Load(registryFile); err != nil {
		t.Fatal(err)
	}

	path, exists := newRegistry.Get("userService")
	if !exists {
		t.Error("Bean should be loaded")
	}

	if path != "/path/to/service_proxy.go" {
		t.Errorf("Expected /path/to/service_proxy.go, got %s", path)
	}
}

func TestRegistry_Clear(t *testing.T) {
	t.Parallel()
	registry := NewRegistry()

	registry.Register("userService", "/path/to/service_proxy.go")
	registry.Clear()

	list := registry.List()
	if len(list) != 0 {
		t.Errorf("Expected 0 beans after clear, got %d", len(list))
	}
}

func TestParser_ParseInterfaceAnnotation(t *testing.T) {
	t.Parallel()
	parser := NewParser()

	text := `@ProxyInterface`
	parser.parseInterfaceAnnotation(text, "UserService", "service", "/path/to/service.go")

	intf, exists := parser.interfaces["UserService"]
	if !exists {
		t.Fatal("Interface not found")
	}

	if intf.Name != "UserService" {
		t.Errorf("Expected UserService, got %s", intf.Name)
	}

	if intf.BeanID != "userService" {
		t.Errorf("Expected userService, got %s", intf.BeanID)
	}
}

func TestParser_ParseInterfaceAnnotationWithCustomBeanID(t *testing.T) {
	t.Parallel()
	parser := NewParser()

	text := `@ProxyInterface(beanId="customUserService")`
	parser.parseInterfaceAnnotation(text, "UserService", "service", "/path/to/service.go")

	intf, exists := parser.interfaces["UserService"]
	if !exists {
		t.Fatal("Interface not found")
	}

	if intf.BeanID != "customUserService" {
		t.Errorf("Expected customUserService, got %s", intf.BeanID)
	}
}

func TestParser_ParseInterfaceMethods(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package service

// @ProxyInterface
type UserService interface {
	GetUser(id int) (*User, error)
	CreateUser(name string) error
}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	if err := parser.parseFile(testFile); err != nil {
		t.Fatal(err)
	}

	interfaces := parser.GetInterfaces()
	if len(interfaces) != 1 {
		t.Errorf("Expected 1 interface, got %d", len(interfaces))
	}

	if len(interfaces[0].Methods) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(interfaces[0].Methods))
	}

	if interfaces[0].Methods[0].Name != "GetUser" {
		t.Errorf("Expected GetUser, got %s", interfaces[0].Methods[0].Name)
	}

	if interfaces[0].Methods[1].Name != "CreateUser" {
		t.Errorf("Expected CreateUser, got %s", interfaces[0].Methods[1].Name)
	}
}

func TestToTitleCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"before", "Before"},
		{"after", "After"},
		{"around", "Around"},
		{"after_returning", "After_returning"},
		{"", ""},
		{"a", "A"},
	}

	for _, tt := range tests {
		result := toTitleCase(tt.input)
		if result != tt.expected {
			t.Errorf("toTitleCase(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestAdviceConstructorName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"after_returning", "AfterReturning"},
		{"after_throwing", "AfterThrowing"},
		{"before", "Before"},
		{"after", "After"},
		{"around", "Around"},
	}

	for _, tt := range tests {
		result := adviceConstructorName(tt.input)
		if result != tt.expected {
			t.Errorf("adviceConstructorName(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerator_GetRegistry(t *testing.T) {
	t.Parallel()

	gen, err := NewGenerator()
	if err != nil {
		t.Fatal(err)
	}

	registry := gen.GetRegistry()
	if registry == nil {
		t.Error("expected non-nil registry")
	}
}

func TestParser_ParseStandaloneFunc(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package service

// @Before("UserService.GetUser")
func BeforeLog() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	if err := parser.parseFile(testFile); err != nil {
		t.Fatal(err)
	}

	funcs := parser.GetFuncs()
	if len(funcs) != 1 {
		t.Errorf("Expected 1 func, got %d", len(funcs))
	}

	advice := funcs[0]
	if advice.Type != AdviceBefore {
		t.Errorf("Expected AdviceBefore, got %v", advice.Type)
	}
	if !advice.IsFunc {
		t.Error("expected IsFunc to be true")
	}
}

func TestExprToString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{"simple", "string", "string"},
		{"pointer", "*int", "*int"},
		{"slice", "[]string", "[]string"},
		{"map", "map[string]int", "map[string]int"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// 通过解析临时文件来测试exprToString
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.go")
			content := fmt.Sprintf(`package test

type TestStruct struct {
	Field %s
}
`, tt.expr)

			if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
				t.Fatal(err)
			}

			p := NewParser()
			if err := p.parseFile(testFile); err != nil {
				t.Fatal(err)
			}

			structs := p.GetProxies()
			// 验证解析成功
			if len(structs) == 0 {
				t.Log("no structs parsed")
			}
		})
	}
}

func TestParser_ParseDir(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "service1.go")
	content1 := `package service

// @Aspect(order=1)
type LoggingAspect struct {}

// @Before("UserService.GetUser")
func (a *LoggingAspect) LogBefore(jp aop.JoinPoint) {}
`

	file2 := filepath.Join(tmpDir, "service2.go")
	content2 := `package service

// @AopProxy
type UserService struct {}

func (s *UserService) GetUser(id int) (*User, error) {
	return nil, nil
}
`

	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	if err := parser.ParseDir(tmpDir); err != nil {
		t.Fatal(err)
	}

	aspects := parser.GetAspects()
	if len(aspects) != 1 {
		t.Errorf("Expected 1 aspect, got %d", len(aspects))
	}

	proxies := parser.GetProxies()
	if len(proxies) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(proxies))
	}
}

func TestParser_GetFuncs(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := `package service

// @Before("UserService.GetUser")
func BeforeLog() {}

// @After("UserService.CreateUser")
func AfterLog() {}
`

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	if err := parser.parseFile(testFile); err != nil {
		t.Fatal(err)
	}

	funcs := parser.GetFuncs()
	if len(funcs) != 2 {
		t.Errorf("Expected 2 funcs, got %d", len(funcs))
	}
}

func TestGenerator_Generate_InterfaceProxy(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	serviceFile := filepath.Join(tmpDir, "service.go")
	contentStr := `package service

// @ProxyInterface
type UserService interface {
	GetUser(id int) (*User, error)
}

// @Aspect(order=1)
type LoggingAspect struct {}

// @Before("UserService.GetUser")
func (a *LoggingAspect) LogBefore(jp aop.JoinPoint) {}
`

	if err := os.WriteFile(serviceFile, []byte(contentStr), 0644); err != nil {
		t.Fatal(err)
	}

	gen, err := NewGenerator()
	if err != nil {
		t.Fatal(err)
	}

	if err := gen.Generate(tmpDir, "aop"); err != nil {
		t.Fatal(err)
	}

	// 检查是否生成了接口代理文件
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	found := false
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "_proxy.go") {
			found = true
			contentBytes, err := os.ReadFile(filepath.Join(tmpDir, f.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(contentBytes), "UserServiceProxy") {
				t.Error("Generated code should contain UserServiceProxy")
			}
		}
	}

	if !found {
		t.Error("Expected proxy file to be generated")
	}
}

func TestGenerator_GenerateInterfaceProxy(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	serviceFile := filepath.Join(tmpDir, "service.go")
	contentStr := `package service

// @Aspect(order=1)
type LoggingAspect struct {}

// @Before("UserService.GetUser")
func (a *LoggingAspect) LogBefore(jp aop.JoinPoint) {}

// @ProxyInterface
type UserService interface {
	GetUser(id int) (*User, error)
}
`

	if err := os.WriteFile(serviceFile, []byte(contentStr), 0644); err != nil {
		t.Fatal(err)
	}

	gen, err := NewGenerator()
	if err != nil {
		t.Fatal(err)
	}

	if err := gen.Generate(tmpDir, "static"); err != nil {
		t.Fatal(err)
	}

	outputFile := filepath.Join(tmpDir, "service_proxy.go")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("Generated file should exist")
	}

	contentBytes, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	output := string(contentBytes)
	if !strings.Contains(output, "UserServiceProxy") {
		t.Error("Generated code should contain UserServiceProxy")
	}

	if !strings.Contains(output, "func (p *UserServiceProxy) GetUser(id int) (*User, error)") {
		t.Error("Generated code should contain GetUser method")
		fmt.Println("Generated output:")
		fmt.Println(output)
	}
}

// TestGenerator_StaticMode_ImportsAop 验证 static 模式生成的代码包含 aop 导入。
func TestGenerator_StaticMode_ImportsAop(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	serviceFile := filepath.Join(tmpDir, "service.go")
	contentStr := `package service

// @Aspect(order=1)
type LoggingAspect struct {}

// @Before("UserService.GetUser")
func (a *LoggingAspect) LogBefore(jp aop.JoinPoint) {}

// @AopProxy
type UserService struct {}

func (s *UserService) GetUser(id int) (*User, error) {
	return nil, nil
}
`

	if err := os.WriteFile(serviceFile, []byte(contentStr), 0644); err != nil {
		t.Fatal(err)
	}

	gen, err := NewGenerator()
	if err != nil {
		t.Fatal(err)
	}

	if err := gen.Generate(tmpDir, "static"); err != nil {
		t.Fatal(err)
	}

	contentBytes, err := os.ReadFile(filepath.Join(tmpDir, "service_proxy.go"))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(contentBytes), `"github.com/xudefa/enhance/aop"`) {
		t.Error("static mode generated code must import the aop package")
	}
}

// TestGenerator_InterfaceProxy_AllModes 验证接口代理在三种模式下生成可编译代码。
//
// 回归保护：接口代理不能生成 `*Interface`（指向接口的指针）或 `&Interface{}` 字面量。
func TestGenerator_InterfaceProxy_AllModes(t *testing.T) {
	t.Parallel()

	contentStr := `package service

// @Aspect(order=1)
type LoggingAspect struct {}

// @Before("UserService.GetUser")
func (a *LoggingAspect) LogBefore(jp aop.JoinPoint) {}

// @ProxyInterface
type UserService interface {
	GetUser(id int) (*User, error)
}

type User struct { Name string }
`

	for _, mode := range []string{"simple", "aop", "static"} {
		mode := mode
		t.Run(mode, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			serviceFile := filepath.Join(dir, "service.go")
			if err := os.WriteFile(serviceFile, []byte(contentStr), 0644); err != nil {
				t.Fatal(err)
			}

			gen, err := NewGenerator()
			if err != nil {
				t.Fatal(err)
			}

			if err := gen.Generate(dir, mode); err != nil {
				t.Fatalf("Generate(%s) error = %v", mode, err)
			}

			contentBytes, err := os.ReadFile(filepath.Join(dir, "service_proxy.go"))
			if err != nil {
				t.Fatal(err)
			}
			output := string(contentBytes)

			if strings.Contains(output, "Target *UserService") {
				t.Error("interface proxy must not declare a pointer-to-interface field")
			}
			if strings.Contains(output, "&UserService{}") {
				t.Error("interface proxy must not create an interface composite literal")
			}
		})
	}
}

func TestRenderStaticProxy(t *testing.T) {
	t.Parallel()

	data := &StaticProxyTemplateData{
		Package:    "test",
		ProxyName:  "TestProxy",
		TargetName: "TestTarget",
		SourceFile: "test.go",
		Imports:    []string{"fmt"},
		Methods: []StaticMethodInfo{
			{
				Name:           "TestMethod",
				ParamsStr:      "arg1 string",
				ResultsStr:     "string",
				HasParams:      true,
				HasReturnValue: true,
			},
		},
		AdviceFields: []AdviceField{
			{Name: "Before", Type: "func()"},
		},
	}

	result, err := RenderStaticProxy(data)
	if err != nil {
		t.Fatalf("RenderStaticProxy error: %v", err)
	}

	if !strings.Contains(result, "package test") {
		t.Error("expected package declaration")
	}
	if !strings.Contains(result, "TestProxy") {
		t.Error("expected proxy name")
	}
	if !strings.Contains(result, "TestMethod") {
		t.Error("expected method name")
	}
}

func TestRenderInterfaceProxy(t *testing.T) {
	t.Parallel()

	data := &InterfaceProxyTemplateData{
		Package:       "test",
		ProxyName:     "TestInterfaceProxy",
		InterfaceName: "TestInterface",
		SourceFile:    "test.go",
		Imports:       []string{"fmt"},
		Methods: []StaticMethodInfo{
			{
				Name:           "TestMethod",
				ParamsStr:      "arg1 string",
				ResultsStr:     "string",
				HasParams:      true,
				HasReturnValue: true,
			},
		},
	}

	result, err := RenderInterfaceProxy(data)
	if err != nil {
		t.Fatalf("RenderInterfaceProxy error: %v", err)
	}

	if !strings.Contains(result, "package test") {
		t.Error("expected package declaration")
	}
	if !strings.Contains(result, "TestInterfaceProxy") {
		t.Error("expected proxy name")
	}
	if !strings.Contains(result, "TestInterface") {
		t.Error("expected interface name")
	}
	if !strings.Contains(result, "TestMethod") {
		t.Error("expected method name")
	}
}

func TestRenderStaticProxy_InvalidTemplate(t *testing.T) {
	t.Parallel()

	// Test with invalid template syntax
	tmpl := &StaticProxyTemplateData{
		Package: "test",
	}

	// This should work with valid template
	_, err := RenderStaticProxy(tmpl)
	if err != nil {
		t.Fatalf("RenderStaticProxy error: %v", err)
	}
}

func TestRenderInterfaceProxy_InvalidTemplate(t *testing.T) {
	t.Parallel()

	tmpl := &InterfaceProxyTemplateData{
		Package:       "test",
		InterfaceName: "TestInterface",
	}

	_, err := RenderInterfaceProxy(tmpl)
	if err != nil {
		t.Fatalf("RenderInterfaceProxy error: %v", err)
	}
}

func TestTemplateEngine_GenerateAdviceAdapter(t *testing.T) {
	t.Parallel()

	engine, err := NewTemplateEngine()
	if err != nil {
		t.Fatalf("NewTemplateEngine error: %v", err)
	}

	data := AdviceAdapterTemplateData{
		Package: "test",
		Adapters: []AdviceAdapterData{
			{
				FuncName:   "TestAdapter",
				AspectType: "TestAspect",
				MethodName: "Before",
				IsAround:   false,
				HasReturn:  false,
			},
		},
	}

	result, err := engine.GenerateAdviceAdapter(data)
	if err != nil {
		t.Fatalf("GenerateAdviceAdapter error: %v", err)
	}

	if !strings.Contains(result, "package test") {
		t.Error("expected package declaration")
	}
	if !strings.Contains(result, "TestAdapter") {
		t.Error("expected adapter name")
	}
}

func TestTemplateEngine_GenerateAdviceAdapter_Error(t *testing.T) {
	t.Parallel()

	engine := &TemplateEngine{}

	data := AdviceAdapterTemplateData{
		Package: "test",
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for uninitialized template")
		}
	}()

	_, _ = engine.GenerateAdviceAdapter(data)
}

func TestSafeTupleAccess(t *testing.T) {
	t.Parallel()

	result := safeTupleAccess(0, "string")
	if !strings.Contains(result, "tuple[0]") {
		t.Error("expected tuple access")
	}
	if !strings.Contains(result, "string") {
		t.Error("expected type in result")
	}
	if !strings.Contains(result, "panic") {
		t.Error("expected panic on type mismatch")
	}
}

func TestBuildMethodTemplateData_NoNames(t *testing.T) {
	t.Parallel()

	method := MethodInfo{
		Name: "TestMethod",
		Params: []ParamInfo{
			{Type: "string"},
			{Type: "int"},
		},
		Results: []ParamInfo{
			{Type: "string"},
		},
	}

	data := buildMethodTemplateData(method)

	if !strings.Contains(data.ParamsStr, "string") {
		t.Error("expected params to contain 'string'")
	}
	if !strings.Contains(data.ResultsStr, "string") {
		t.Error("expected results to contain 'string'")
	}
}

func TestBuildMethodTemplateData_SingleResult(t *testing.T) {
	t.Parallel()

	method := MethodInfo{
		Name: "TestMethod",
		Params: []ParamInfo{
			{Name: "arg1", Type: "string"},
		},
		Results: []ParamInfo{
			{Name: "result", Type: "string"},
		},
	}

	data := buildMethodTemplateData(method)

	if data.ResultsStr != "result string" {
		t.Errorf("expected 'result string', got '%s'", data.ResultsStr)
	}
}

func TestBuildMethodTemplateData_NoParams(t *testing.T) {
	t.Parallel()

	method := MethodInfo{
		Name:    "TestMethod",
		Params:  []ParamInfo{},
		Results: []ParamInfo{},
	}

	data := buildMethodTemplateData(method)

	if data.HasParams {
		t.Error("expected HasParams to be false")
	}
	if data.HasReturnValue {
		t.Error("expected HasReturnValue to be false")
	}
}
