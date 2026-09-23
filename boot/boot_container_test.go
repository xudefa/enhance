package boot

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/core"
)

func TestBoot_GetBean(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Name string
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(NewModule().
			Name("test-module").
			Bean(Provide(func(c core.Container) (TestBean, error) {
				return TestBean{Name: "test-bean"}, nil
			})).
			Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	bean, err := core.GetByName[TestBean](app.Container(), "")
	if err != nil {
		t.Fatalf("GetBean failed: %v", err)
	}
	if bean.Name != "test-bean" {
		t.Errorf("Expected Name 'test-bean', got %s", bean.Name)
	}
}

func TestBoot_GetBeanTyped(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Name string
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(NewModule().
			Name("test-module").
			Bean(Provide(func(c core.Container) (TestBean, error) {
				return TestBean{Name: "typed-bean"}, nil
			})).
			Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	bean, err := core.GetByName[TestBean](app.Container(), "")
	if err != nil {
		t.Fatalf("GetBeanTyped failed: %v", err)
	}
	if bean.Name != "typed-bean" {
		t.Errorf("Expected Name 'typed-bean', got %s", bean.Name)
	}
}

func TestBoot_GetBeanNotRegistered(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Name string
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	_, err = core.GetByName[TestBean](app.Container(), "")
	if err == nil {
		t.Error("Expected error for non-existent bean")
	}
}

func TestBoot_HasBean(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Name string
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(NewModule().
			Name("test-module").
			Bean(Provide(func(c core.Container) (TestBean, error) {
				return TestBean{Name: "test-bean"}, nil
			})).
			Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !core.Has[TestBean](app.Container(), "") {
		t.Error("Expected Bean to exist")
	}
}

func TestBoot_GetByType(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Name string
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(NewModule().
			Name("test-module").
			Bean(Provide(func(c core.Container) (TestBean, error) {
				return TestBean{Name: "test-bean"}, nil
			})).
			Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	ctx := app.Context()
	bean, err := ctx.GetByType(reflect.TypeOf(TestBean{}))
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}
	testBean, ok := bean.(TestBean)
	if !ok {
		t.Fatal("Expected TestBean type")
	}
	if testBean.Name != "test-bean" {
		t.Errorf("Expected Name 'test-bean', got %s", testBean.Name)
	}
}

func TestBoot_GetByType_NotFound(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Name string
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	ctx := app.Context()
	_, err = ctx.GetByType(reflect.TypeOf(TestBean{}))
	if err == nil {
		t.Error("Expected error when bean not found")
	}
}

func TestBoot_Register(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Name string
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(NewModule().
			Name("test-module").
			Bean(Provide(func(c core.Container) (TestBean, error) {
				return TestBean{Name: "test"}, nil
			})).
			Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	if !core.Has[TestBean](app.Container(), "") {
		t.Error("Expected TestBean to exist after Register")
	}
}

func TestBoot_ContainerAccessor(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Name string
	}
	app, err := NewApplication(
		WithAppName("test-app"),
		WithoutAutoConfig(),
		WithoutStarters(),
		WithModules(NewModule().
			Name("test-module").
			Bean(Provide(func(c core.Container) (TestBean, error) {
				return TestBean{Name: "accessor-test"}, nil
			})).
			Build(),
		),
	)
	if err != nil {
		t.Fatalf("NewApplication failed: %v", err)
	}
	if err := app.Start(); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer app.Stop()
	bean, err := core.GetByName[TestBean](app.Container(), "")
	if err != nil {
		t.Fatalf("GetBean failed: %v", err)
	}
	if bean.Name != "accessor-test" {
		t.Errorf("Expected Name 'accessor-test', got %s", bean.Name)
	}
}
