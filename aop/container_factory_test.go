package aop

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core"
)

func TestNewAopBeanFactory_WithIntegration(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	factory := NewAopBeanFactory(integration)
	if factory == nil {
		t.Fatal("NewAopBeanFactory() returned nil")
	}
	if factory.integration != integration {
		t.Error("factory should use provided integration")
	}
}

func TestNewAopBeanFactory_NilIntegration(t *testing.T) {
	t.Parallel()

	factory := NewAopBeanFactory(nil)
	if factory == nil {
		t.Fatal("NewAopBeanFactory(nil) returned nil")
	}
	if factory.integration == nil {
		t.Error("factory.integration should use global integration when nil passed")
	}
}

func TestAopBeanFactory_CreateBean_NilDef(t *testing.T) {
	t.Parallel()

	factory := NewAopBeanFactory(NewAopIntegration(DefaultAopConfig()))
	bean := "test"

	result, err := factory.CreateBean("testBean", nil, bean)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != bean {
		t.Error("should return original target when beanDef is nil")
	}
}

func TestAopBeanFactory_CreateBean_Disabled(t *testing.T) {
	t.Parallel()

	factory := NewAopBeanFactory(NewAopIntegration(DefaultAopConfig()))
	bean := "test"

	beanDef := NewAopBeanDefinition("testBean", reflect.TypeOf(""))
	beanDef.EnableAop = false

	result, err := factory.CreateBean("testBean", beanDef, bean)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != bean {
		t.Error("should return original target when AOP disabled")
	}
}

func TestAopBeanFactory_CreateBean_WithAspects(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	factory := NewAopBeanFactory(integration)

	type TestBean struct {
		Name string
	}
	bean := &TestBean{Name: "test"}

	aspect := &AspectMeta{
		PointCut: MatchByClassName("TestBean"),
		Advice:   Before(func(jp JoinPoint) {}),
		Order:    1,
	}

	beanDef := NewAopBeanDefinition("testBean", reflect.TypeOf(TestBean{}))
	beanDef.WithAspects(aspect)

	result, err := factory.CreateBean("testBean", beanDef, bean)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestAopBeanFactory_GetProcessor_Custom(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	factory := NewAopBeanFactory(integration)

	processor := factory.GetProcessor()
	if processor == nil {
		t.Fatal("GetProcessor() returned nil")
	}
	if processor.integration != integration {
		t.Error("processor should use same integration")
	}
}

func TestAopBeanFactory_GetProcessor_Global(t *testing.T) {
	t.Parallel()

	factory := NewAopBeanFactory(nil)
	processor := factory.GetProcessor()
	if processor == nil {
		t.Fatal("GetProcessor() from global should not be nil")
	}
}

func TestAopBeanFactory_RegisterBean(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	factory := NewAopBeanFactory(NewAopIntegration(DefaultAopConfig()))

	beanDef := NewAopBeanDefinition("testBean", reflect.TypeOf(""))
	beanDef.ProxyType = reflect.TypeOf("")

	err := factory.RegisterBean(container, beanDef, "target")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGlobalAopBeanFactory_Exists(t *testing.T) {
	t.Parallel()

	if GlobalAopBeanFactory == nil {
		t.Error("GlobalAopBeanFactory should not be nil")
	}
}
