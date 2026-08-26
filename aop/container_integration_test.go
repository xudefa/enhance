package aop

import (
	"reflect"
	"testing"
)

func TestNewAopBeanPostProcessor_WithIntegration(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	processor := NewAopBeanPostProcessor(integration)
	if processor == nil {
		t.Fatal("NewAopBeanPostProcessor() returned nil")
	}
	if processor.integration != integration {
		t.Error("processor should use provided integration")
	}
	if !processor.IsEnabled() {
		t.Error("processor should be enabled by default")
	}
}

func TestAopBeanPostProcessor_EnableDisableCycle(t *testing.T) {
	t.Parallel()

	processor := NewAopBeanPostProcessor(nil)

	if !processor.IsEnabled() {
		t.Error("should be enabled initially")
	}

	processor.Disable()
	if processor.IsEnabled() {
		t.Error("should be disabled after Disable()")
	}

	processor.Enable()
	if !processor.IsEnabled() {
		t.Error("should be enabled after Enable()")
	}
}

func TestAopBeanPostProcessor_ProcessBean_Disabled(t *testing.T) {
	t.Parallel()

	processor := NewAopBeanPostProcessor(nil)
	processor.Disable()

	bean := "original"
	result := processor.ProcessBean("testBean", bean)
	if result != bean {
		t.Error("should return original bean when disabled")
	}
}

func TestNewAopBeanDefinition_BuilderChain(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}
	beanType := reflect.TypeOf(TestBean{})

	aspect := &AspectMeta{Order: 1}
	proxyType := reflect.TypeOf((*interface{})(nil)).Elem()

	def := NewAopBeanDefinition("myBean", beanType).
		WithAopEnabled(false).
		WithProxyMode(AopModeRuntime).
		WithAspects(aspect).
		WithProxyType(proxyType)

	if def.BeanID != "myBean" {
		t.Errorf("BeanID = %q, want %q", def.BeanID, "myBean")
	}
	if def.EnableAop != false {
		t.Error("EnableAop should be false")
	}
	if def.ProxyMode != AopModeRuntime {
		t.Errorf("ProxyMode = %v, want %v", def.ProxyMode, AopModeRuntime)
	}
	if len(def.Aspects) != 1 {
		t.Errorf("Aspects length = %d, want 1", len(def.Aspects))
	}
	if def.ProxyType != proxyType {
		t.Error("ProxyType should match")
	}
	if def.TargetType != beanType {
		t.Error("TargetType should match")
	}
}

func TestNewAopBeanDefinition_MultipleAspects(t *testing.T) {
	t.Parallel()

	def := NewAopBeanDefinition("bean", reflect.TypeOf(""))
	a1 := &AspectMeta{Order: 1}
	a2 := &AspectMeta{Order: 2}
	def.WithAspects(a1).WithAspects(a2)

	if len(def.Aspects) != 2 {
		t.Errorf("expected 2 aspects, got %d", len(def.Aspects))
	}
}

func TestGlobalAopBeanPostProcessor_Exists(t *testing.T) {
	t.Parallel()

	if GlobalAopBeanPostProcessor == nil {
		t.Error("GlobalAopBeanPostProcessor should not be nil")
	}
}
