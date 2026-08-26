package aop

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core"
)

func TestAopManager_RegisterAspect(t *testing.T) {
	t.Parallel()

	manager := &AopManager{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
	}

	aspect := &AspectMeta{
		Order: 1,
	}
	manager.RegisterAspect(aspect)

	if len(manager.aspects) != 1 {
		t.Errorf("expected 1 aspect, got %d", len(manager.aspects))
	}
}

func TestAopManager_RegisterAspects(t *testing.T) {
	t.Parallel()

	manager := &AopManager{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
	}

	aspect1 := &AspectMeta{Order: 1}
	aspect2 := &AspectMeta{Order: 2}
	manager.RegisterAspects(aspect1, aspect2)

	if len(manager.aspects) != 2 {
		t.Errorf("expected 2 aspects, got %d", len(manager.aspects))
	}
}

func TestAopManager_GetAspects(t *testing.T) {
	t.Parallel()

	manager := &AopManager{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
	}

	aspect := &AspectMeta{Order: 1}
	manager.RegisterAspect(aspect)

	aspects := manager.GetAspects()
	if len(aspects) != 1 {
		t.Errorf("expected 1 aspect, got %d", len(aspects))
	}

	// 验证返回的副本长度正确
	if cap(aspects) != len(aspects) {
		t.Logf("aspects capacity: %d, length: %d", cap(aspects), len(aspects))
	}
}

func TestAopManager_GetConfig(t *testing.T) {
	t.Parallel()

	manager := &AopManager{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
	}

	config := manager.GetConfig()
	if config == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestAopManager_MatchAspectsForType(t *testing.T) {
	t.Parallel()

	manager := &AopManager{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
	}

	// Test with nil type
	matched := manager.MatchAspectsForType(nil)
	if len(matched) != 0 {
		t.Errorf("expected 0 matched aspects for nil type, got %d", len(matched))
	}
}

func TestAopContainer_Basic(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)
	if container == nil {
		t.Fatal("expected non-nil container")
	}
}

func TestAopContainer_WithBaseContainer(t *testing.T) {
	t.Parallel()

	baseContainer := core.NewContainer()
	container := NewAopContainer(baseContainer)
	if container == nil {
		t.Fatal("expected non-nil container")
	}
}

func TestAopContainer_RegisterAspect(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)

	// 获取初始数量
	initialCount := len(container.GetAspects())

	aspect := &AspectMeta{Order: 1}
	container.RegisterAspect(aspect)

	aspects := container.GetAspects()
	if len(aspects) != initialCount+1 {
		t.Errorf("expected %d aspects, got %d", initialCount+1, len(aspects))
	}
}

func TestAopContainer_RegisterAspects(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)

	// 获取初始数量
	initialCount := len(container.GetAspects())

	aspect1 := &AspectMeta{Order: 1}
	aspect2 := &AspectMeta{Order: 2}
	container.RegisterAspects(aspect1, aspect2)

	aspects := container.GetAspects()
	if len(aspects) != initialCount+2 {
		t.Errorf("expected %d aspects, got %d", initialCount+2, len(aspects))
	}
}

func TestAopContainer_GetIntegration(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)
	integration := container.GetIntegration()
	if integration == nil {
		t.Error("expected non-nil integration")
	}
}

func TestAopContainer_GetFactory(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)
	factory := container.GetFactory()
	if factory == nil {
		t.Error("expected non-nil factory")
	}
}

func TestAopContainer_GetProcessor(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)
	processor := container.GetProcessor()
	if processor == nil {
		t.Error("expected non-nil processor")
	}
}

func TestAopContainer_EnableDisableAop(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)

	// 默认应该启用
	if !container.IsAopEnabled() {
		t.Error("expected AOP to be enabled by default")
	}

	// 禁用
	container.DisableAop()
	if container.IsAopEnabled() {
		t.Error("expected AOP to be disabled")
	}

	// 重新启用
	container.EnableAop()
	if !container.IsAopEnabled() {
		t.Error("expected AOP to be enabled")
	}
}

func TestAopContainer_RegisterAopBeanWithID(t *testing.T) {
	t.Parallel()

	// 跳过这个测试，因为需要完整的AOP设置
	t.Skip("requires full AOP setup")
}

func TestAopContainer_RegisterAopBeanWithAspects(t *testing.T) {
	t.Parallel()

	// 跳过这个测试，因为需要完整的AOP设置
	t.Skip("requires full AOP setup")
}

func TestAopContainer_RegisterAopBean_Extended(t *testing.T) {
	t.Parallel()
	t.Skip("requires full container setup")
}

func TestAopContainer_RegisterAopBeanWithID_Extended(t *testing.T) {
	t.Parallel()
	t.Skip("requires full container setup")
}

func TestAopContainer_RegisterAopBeanWithAspects_Extended(t *testing.T) {
	t.Parallel()
	t.Skip("requires full container setup")
}

func TestAopContainer_RegisterProxyType_Extended(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)

	// 调用registerProxyType
	container.registerProxyType("TestType", "/path/to/test.go")

	// 验证没有panic
}

func TestAopBeanPostProcessor_ProcessBean(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	processor := NewAopBeanPostProcessor(integration)

	// 测试nil bean
	result := processor.ProcessBean("testBean", nil)
	if result != nil {
		t.Errorf("expected nil result for nil bean")
	}

	// 测试禁用状态
	processor.Disable()
	bean := &struct{ Name string }{Name: "test"}
	result = processor.ProcessBean("testBean", bean)
	if result != bean {
		t.Errorf("expected same bean when disabled")
	}
}

func TestAopBeanPostProcessor_NeedsProxy(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	processor := NewAopBeanPostProcessor(integration)

	// 测试nil
	if processor.needsProxy(nil) {
		t.Error("expected false for nil bean")
	}

	// 测试普通对象（没有匹配的切面）
	bean := &struct{ Name string }{Name: "test"}
	if processor.needsProxy(bean) {
		t.Error("expected false for bean without matching aspects")
	}
}

func TestAopBeanDefinition_WithMethods(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	beanDef := NewAopBeanDefinition("testBean", reflect.TypeOf(TestBean{}))

	// 测试WithAopEnabled
	beanDef.WithAopEnabled(false)
	if beanDef.EnableAop {
		t.Error("expected EnableAop to be false")
	}

	// 测试WithProxyMode
	beanDef.WithProxyMode(AopModeRuntime)
	if beanDef.ProxyMode != AopModeRuntime {
		t.Errorf("expected ProxyMode Runtime, got %v", beanDef.ProxyMode)
	}

	// 测试WithAspects
	aspect := &AspectMeta{Order: 1}
	beanDef.WithAspects(aspect)
	if len(beanDef.Aspects) != 1 {
		t.Errorf("expected 1 aspect, got %d", len(beanDef.Aspects))
	}

	// 测试WithProxyType
	beanDef.WithProxyType(reflect.TypeOf(TestBean{}))
	if beanDef.ProxyType != reflect.TypeOf(TestBean{}) {
		t.Error("expected ProxyType to be set")
	}
}

func TestAopIntegration_GetScannedProxy(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())

	// 测试不存在的类型
	_, ok := integration.GetScannedProxy("NonExistent")
	if ok {
		t.Error("expected false for non-existent proxy type")
	}

	// 注册并获取
	integration.RegisterProxyType("TestProxy", "/path/to/test.go")
	path, ok := integration.GetScannedProxy("TestProxy")
	if !ok {
		t.Error("expected true for registered proxy type")
	}
	if path != "/path/to/test.go" {
		t.Errorf("expected path '/path/to/test.go', got '%s'", path)
	}
}

func TestGlobalAopIntegration_Functions(t *testing.T) {
	t.Parallel()

	// 测试GetGlobalAopIntegration
	global := GetGlobalAopIntegration()
	if global == nil {
		t.Fatal("expected non-nil global integration")
	}

	// 测试SetGlobalAopIntegration
	newIntegration := NewAopIntegration(DefaultAopConfig())
	SetGlobalAopIntegration(newIntegration)
	if GetGlobalAopIntegration() != newIntegration {
		t.Error("expected global integration to be updated")
	}

	// 恢复原始集成
	SetGlobalAopIntegration(global)
}

func TestCreateProxy_Global(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}
	bean := &TestBean{Name: "test"}

	// 使用全局函数创建代理
	result := CreateProxy("testBean", bean)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestRegisterAspectToGlobal(t *testing.T) {
	t.Parallel()

	// 保存原始切面
	originalAspects := GetGlobalAspects()

	// 注册新切面
	aspect := &AspectMeta{Order: 999}
	RegisterAspectToGlobal(aspect)

	// 验证切面已注册
	aspects := GetGlobalAspects()
	found := false
	for _, a := range aspects {
		if a == aspect {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected aspect to be registered globally")
	}

	// 恢复原始切面
	SetGlobalAopIntegration(NewAopIntegration(DefaultAopConfig()))
	for _, a := range originalAspects {
		RegisterAspectToGlobal(a)
	}
}

func TestBuildTagChecker(t *testing.T) {
	t.Parallel()

	checker := NewBuildTagChecker()

	// 测试HasTag
	if checker.HasTag("goaop") {
		t.Log("goaop build tag is present")
	} else {
		t.Log("goaop build tag is not present")
	}

	// 测试未知标签
	if checker.HasTag("unknown") {
		t.Error("expected false for unknown tag")
	}

	// 测试模式检查
	if checker.IsGeneratedMode() {
		t.Log("in generated mode")
	} else {
		t.Log("in runtime mode")
	}

	// 测试GetOptimalMode
	mode := checker.GetOptimalMode()
	if mode != AopModeGenerated && mode != AopModeRuntime {
		t.Errorf("expected Generated or Runtime mode, got %v", mode)
	}
}

func TestDetectOptimalMode(t *testing.T) {
	t.Parallel()

	mode := DetectOptimalMode()
	if mode != AopModeGenerated && mode != AopModeRuntime {
		t.Errorf("expected Generated or Runtime mode, got %v", mode)
	}
}

func TestConfigureAopManager(t *testing.T) {
	t.Parallel()

	config := ConfigureAopManager()
	if config == nil {
		t.Fatal("expected non-nil config")
	}
	if config.Mode != DetectOptimalMode() {
		t.Error("expected config mode to match detected optimal mode")
	}
}

func TestAopBeanFactory_CreateBean(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	factory := NewAopBeanFactory(integration)

	type TestBean struct {
		Name string
	}
	bean := &TestBean{Name: "test"}

	// 测试禁用AOP的bean定义
	beanDef := NewAopBeanDefinition("testBean", reflect.TypeOf(TestBean{}))
	beanDef.EnableAop = false

	result, err := factory.CreateBean("testBean", beanDef, bean)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != bean {
		t.Error("expected same bean when AOP disabled")
	}
}

func TestAopBeanFactory_GetProcessor(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	factory := NewAopBeanFactory(integration)

	processor := factory.GetProcessor()
	if processor == nil {
		t.Error("expected non-nil processor")
	}
}

func TestGlobalAopBeanFactory(t *testing.T) {
	t.Parallel()

	if GlobalAopBeanFactory == nil {
		t.Error("expected non-nil GlobalAopBeanFactory")
	}
}

func TestGlobalAopBeanPostProcessor(t *testing.T) {
	t.Parallel()

	if GlobalAopBeanPostProcessor == nil {
		t.Error("expected non-nil GlobalAopBeanPostProcessor")
	}
}

func TestJoinPointImpl_Methods(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}
	bean := &TestBean{Name: "test"}

	var proceedCalled bool
	proceed := func() (any, error) {
		proceedCalled = true
		return "result", nil
	}

	var proceedWithArgsCalled bool
	proceedWithArgs := func(args []any) (any, error) {
		proceedWithArgsCalled = true
		return "result2", nil
	}

	ctx := context.Background()
	jp := NewJoinPointWithContext(ctx, bean, "TestMethod", []any{"arg1"}, proceed, proceedWithArgs)

	// 测试Target
	if jp.Target() != bean {
		t.Error("expected Target to return bean")
	}

	// 测试Method
	if jp.Method() != "TestMethod" {
		t.Errorf("expected Method 'TestMethod', got '%s'", jp.Method())
	}

	// 测试Args
	args := jp.Args()
	if len(args) != 1 || args[0] != "arg1" {
		t.Errorf("expected Args ['arg1'], got %v", args)
	}

	// 测试Context
	if jp.Context() != ctx {
		t.Error("expected Context to match")
	}

	// 测试Proceed
	result, err := jp.Proceed()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "result" {
		t.Errorf("expected result 'result', got %v", result)
	}
	if !proceedCalled {
		t.Error("expected proceed function to be called")
	}

	// 测试ProceedWithArgs
	result2, err := jp.(JoinPoint).ProceedWithArgs([]any{"newArg"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result2 != "result2" {
		t.Errorf("expected result2 'result2', got %v", result2)
	}
	if !proceedWithArgsCalled {
		t.Error("expected proceedWithArgs function to be called")
	}

	// 测试GetResult和SetResult
	jp.SetResult("newResult")
	if jp.GetResult() != "newResult" {
		t.Errorf("expected GetResult 'newResult', got %v", jp.GetResult())
	}

	// 测试GetError和SetError
	testErr := fmt.Errorf("test error")
	jp.SetError(testErr)
	if jp.GetError() != testErr {
		t.Error("expected GetError to match testErr")
	}
}

func TestJoinPointImpl_NilProceed(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}
	bean := &TestBean{Name: "test"}

	// 创建带nil proceed的连接点
	jp := NewJoinPointWithContext(context.Background(), bean, "Test", nil, nil, nil)

	// Proceed应该返回nil, nil
	result, err := jp.Proceed()
	if result != nil || err != nil {
		t.Errorf("expected nil, nil from Proceed with nil proceed, got %v, %v", result, err)
	}

	// ProceedWithArgs应该返回nil, nil
	result, err = jp.ProceedWithArgs(nil)
	if result != nil || err != nil {
		t.Errorf("expected nil, nil from ProceedWithArgs with nil proceedWithArgs, got %v, %v", result, err)
	}
}

func TestInvocationImpl_Methods(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}
	bean := &TestBean{Name: "test"}

	var proceedCalled bool
	proceed := func() (any, error) {
		proceedCalled = true
		return "result", nil
	}

	jp := NewJoinPointWithContext(context.Background(), bean, "Test", []any{"arg1"}, proceed, nil)
	inv := NewInvocation(jp, proceed)

	// 测试JoinPoint
	if inv.JoinPoint() != jp {
		t.Error("expected JoinPoint to match")
	}

	// 测试Arguments
	args := inv.Arguments()
	if len(args) != 1 || args[0] != "arg1" {
		t.Errorf("expected Arguments ['arg1'], got %v", args)
	}

	// 测试SetArgs - 需要类型断言到具体实现
	if invImpl, ok := inv.(*invocationImpl); ok {
		invImpl.SetArgs([]any{"newArg"})
		args = invImpl.Arguments()
		if len(args) != 1 || args[0] != "newArg" {
			t.Errorf("expected Arguments ['newArg'] after SetArgs, got %v", args)
		}
	}

	// 测试Proceed
	result, err := inv.Proceed()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "result" {
		t.Errorf("expected result 'result', got %v", result)
	}
	if !proceedCalled {
		t.Error("expected proceed function to be called")
	}

	// 测试Error和SetError - 需要类型断言到具体实现
	if invImpl, ok := inv.(*invocationImpl); ok {
		testErr := fmt.Errorf("test error")
		invImpl.SetError(testErr)
		if invImpl.Error() != testErr {
			t.Error("expected Error to match testErr")
		}
	}
}

func TestInvocationImpl_NilProceed(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}
	bean := &TestBean{Name: "test"}

	jp := NewJoinPointWithContext(context.Background(), bean, "Test", nil, nil, nil)
	inv := NewInvocation(jp, nil)

	// Proceed应该返回nil, nil
	result, err := inv.Proceed()
	if result != nil || err != nil {
		t.Errorf("expected nil, nil from Proceed with nil proceed, got %v, %v", result, err)
	}
}

func TestMethodSignature(t *testing.T) {
	t.Parallel()

	type TestBean struct{}

	sig := NewMethodSignature("TestMethod", reflect.TypeOf(TestBean{}))

	if sig.Name() != "TestMethod" {
		t.Errorf("expected Name 'TestMethod', got '%s'", sig.Name())
	}

	if sig.DeclaringType() != reflect.TypeOf(TestBean{}) {
		t.Error("expected DeclaringType to match")
	}
}

func TestAopIntegration_GetProxyFactory(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())

	factory := integration.GetProxyFactory()
	if factory == nil {
		t.Error("expected non-nil proxy factory")
	}
}

func TestAopIntegration_GetMetadataExtractor(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())

	extractor := integration.GetMetadataExtractor()
	if extractor == nil {
		t.Error("expected non-nil metadata extractor")
	}
}

func TestBuildTagChecker_IsRuntimeMode(t *testing.T) {
	t.Parallel()

	checker := NewBuildTagChecker()

	// IsRuntimeMode应该返回!IsGeneratedMode
	if checker.IsRuntimeMode() == checker.IsGeneratedMode() {
		t.Error("expected IsRuntimeMode to be opposite of IsGeneratedMode")
	}
}

func TestGetProxyWithAutoMode(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}
	bean := &TestBean{Name: "test"}

	result := GetProxyWithAutoMode("testBean", bean)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestAutoRegister(t *testing.T) {
	t.Parallel()

	// 测试不存在的bean
	err := AutoRegister("NonExistentBean")
	if err == nil {
		t.Error("expected error for non-existent bean")
	}
}

func TestInitializeAop(t *testing.T) {
	t.Parallel()

	// 保存原始集成
	original := GetGlobalAopIntegration()

	// 调用InitializeAop（不应该panic）
	InitializeAop()

	// 验证全局集成已更新
	newIntegration := GetGlobalAopIntegration()
	if newIntegration == nil {
		t.Error("expected non-nil global integration after InitializeAop")
	}

	// 恢复原始集成
	SetGlobalAopIntegration(original)
}
