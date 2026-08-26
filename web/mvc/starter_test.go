package mvc

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	wcore "github.com/xudefa/enhance/web/core"
)

// dummyRouter 实现 wcore.Router 接口的测试桩。
type dummyRouter struct{}

func (d *dummyRouter) GET(path string, handler wcore.HandlerFunc)    {}
func (d *dummyRouter) POST(path string, handler wcore.HandlerFunc)   {}
func (d *dummyRouter) PUT(path string, handler wcore.HandlerFunc)    {}
func (d *dummyRouter) DELETE(path string, handler wcore.HandlerFunc) {}
func (d *dummyRouter) PATCH(path string, handler wcore.HandlerFunc)  {}
func (d *dummyRouter) Handle(method, path string, handler wcore.HandlerFunc) {
}
func (d *dummyRouter) Group(prefix string) wcore.Router                 { return d }
func (d *dummyRouter) Use(middleware wcore.MiddlewareFunc)              {}
func (d *dummyRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

// failingServer 模拟启动失败的服务器。
type failingServer struct{}

func (f *failingServer) Start() error                                   { return fmt.Errorf("bind failed") }
func (f *failingServer) Stop(ctx context.Context) error                 { return nil }
func (f *failingServer) SetHandler(handler http.Handler)                {}
func (f *failingServer) Use(middleware func(http.Handler) http.Handler) {}

// mockBean 用于测试依赖注入
type mockBean struct {
	Name string
}

// mockControllerWithInject 带依赖注入的控制器
type mockControllerWithInject struct {
	Bean *mockBean `inject:"true"`
}

func (m *mockControllerWithInject) HandleGet() {}

func (m *mockControllerWithInject) Routes(router wcore.Router) {
	router.GET("/mock", nil)
}

// mockApplicationContext 模拟 ApplicationContext
type mockApplicationContext struct {
	container   core.Container
	ctx         context.Context
	environment *environment.Environment
}

func (m *mockApplicationContext) Container() core.Container {
	return m.container
}

func (m *mockApplicationContext) Context() context.Context {
	return m.ctx
}

func (m *mockApplicationContext) Environment() *environment.Environment {
	return m.environment
}

func (m *mockApplicationContext) Register(t reflect.Type, opts ...core.BeanOption) error {
	return nil
}

func (m *mockApplicationContext) GetByType(t reflect.Type) (any, error) {
	return nil, nil
}

func (m *mockApplicationContext) EventBus() boot.EventBusResult {
	return nil
}

func TestWebStarter_StartErrorObservable(t *testing.T) {
	t.Parallel()
	s := &WebStarter{
		router: &dummyRouter{},
		server: &failingServer{},
		config: Config{Host: "127.0.0.1", Port: 0},
		logger: log.Build(),
	}
	if err := s.Start(nil); err != nil {
		t.Fatalf("Start error = %v", err)
	}
	select {
	case err := <-s.errCh:
		if err == nil {
			t.Fatal("expected non-nil start error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for start error")
	}
}

func TestWebStarter_Stop(t *testing.T) {
	t.Parallel()

	// 测试没有server的情况
	s := &WebStarter{
		router: &dummyRouter{},
		config: Config{Host: "127.0.0.1", Port: 0},
		logger: log.Build(),
	}

	err := s.Stop(nil)
	if err != nil {
		t.Errorf("unexpected error when stopping without server: %v", err)
	}
}

func TestWebStarter_Wait(t *testing.T) {
	t.Parallel()

	// 测试没有errCh的情况
	s := &WebStarter{
		router: &dummyRouter{},
		logger: log.Build(),
	}

	err := s.Wait()
	if err != nil {
		t.Errorf("unexpected error when waiting without errCh: %v", err)
	}
}

func TestWebStarter_Configure_Success(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	err := core.Register[*mockBean](container,
		core.WithFactory[*mockBean](func(c ...any) (any, error) {
			return &mockBean{Name: "test"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to register bean: %v", err)
	}

	ctx := &mockApplicationContext{
		container: container,
		ctx:       context.Background(),
	}

	s := &WebStarter{
		logger: log.Build(),
	}

	// 注册控制器
	RegisterController(&mockControllerWithInject{})

	defer func() {
		mu.Lock()
		controllers = controllers[:0]
		mu.Unlock()
	}()

	err = s.Configure(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWebStarter_Configure_WithNilContainer(t *testing.T) {
	t.Parallel()

	ctx := &mockApplicationContext{
		container: nil,
		ctx:       context.Background(),
	}

	s := &WebStarter{
		logger: log.Build(),
	}

	err := s.Configure(ctx)
	if err != nil {
		t.Errorf("unexpected error with nil container: %v", err)
	}
}

// testControllerForInject 用于测试依赖注入的控制器
type testControllerForInject struct {
	Bean *mockBean `inject:"true"`
}

func (t testControllerForInject) Routes(router wcore.Router) {}

func TestInjectControllerDependencies_ValueType(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	err := core.Register[*mockBean](container,
		core.WithFactory[*mockBean](func(c ...any) (any, error) {
			return &mockBean{Name: "test"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to register bean: %v", err)
	}

	ctx := &mockApplicationContext{
		container: container,
		ctx:       context.Background(),
	}

	// 值类型控制器 - 传递值类型（不是指针），会报错因为无法寻址
	ctrl := testControllerForInject{}
	err = injectControllerDependencies(ctx, ctrl)
	if err == nil {
		t.Error("expected error for value type controller")
	}
}

// nonStructController 非结构体类型控制器
type nonStructController int

func (n *nonStructController) Routes(router wcore.Router) {}

func TestInjectControllerDependencies_NonStructType(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	err := core.Register[*mockBean](container,
		core.WithFactory[*mockBean](func(c ...any) (any, error) {
			return &mockBean{Name: "test"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to register bean: %v", err)
	}

	ctx := &mockApplicationContext{
		container: container,
		ctx:       context.Background(),
	}

	// 非结构体类型
	ctrl := nonStructController(42)
	err = injectControllerDependencies(ctx, &ctrl)
	if err != nil {
		t.Errorf("unexpected error for non-struct type: %v", err)
	}
}

// noInjectTagController 没有inject标签的控制器
type noInjectTagController struct {
	Bean *mockBean
}

func (n *noInjectTagController) Routes(router wcore.Router) {}

func TestInjectControllerDependencies_NoInjectTag(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	err := core.Register[*mockBean](container,
		core.WithFactory[*mockBean](func(c ...any) (any, error) {
			return &mockBean{Name: "test"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to register bean: %v", err)
	}

	ctx := &mockApplicationContext{
		container: container,
		ctx:       context.Background(),
	}

	ctrl := &noInjectTagController{}
	err = injectControllerDependencies(ctx, ctrl)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// alreadySetController 已经设置字段的控制器
type alreadySetController struct {
	Bean *mockBean `inject:"true"`
}

func (a *alreadySetController) Routes(router wcore.Router) {}

func TestInjectControllerDependencies_AlreadySet(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	err := core.Register[*mockBean](container,
		core.WithFactory[*mockBean](func(c ...any) (any, error) {
			return &mockBean{Name: "test"}, nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to register bean: %v", err)
	}

	ctx := &mockApplicationContext{
		container: container,
		ctx:       context.Background(),
	}

	ctrl := &alreadySetController{
		Bean: &mockBean{Name: "already set"},
	}
	err = injectControllerDependencies(ctx, ctrl)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// beanNotFoundController Bean未找到的控制器
type beanNotFoundController struct {
	Bean *mockBean `inject:"true"`
}

func (b *beanNotFoundController) Routes(router wcore.Router) {}

func TestInjectControllerDependencies_BeanNotFound(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()

	ctx := &mockApplicationContext{
		container: container,
		ctx:       context.Background(),
	}

	ctrl := &beanNotFoundController{}
	err := injectControllerDependencies(ctx, ctrl)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ctrl.Bean != nil {
		t.Error("expected bean to be nil")
	}
}

// typeMismatchController 类型不匹配的控制器
type typeMismatchController struct {
	Bean *mockBean `inject:"true"`
}

func (t *typeMismatchController) Routes(router wcore.Router) {}

func TestInjectControllerDependencies_TypeMismatch(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	err := core.Register[string](container,
		core.WithFactory[string](func(c ...any) (any, error) {
			return "string bean", nil
		}),
	)
	if err != nil {
		t.Fatalf("failed to register bean: %v", err)
	}

	ctx := &mockApplicationContext{
		container: container,
		ctx:       context.Background(),
	}

	ctrl := &typeMismatchController{}
	err = injectControllerDependencies(ctx, ctrl)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ctrl.Bean != nil {
		t.Error("expected bean to be nil due to type mismatch")
	}
}

func TestWebStarter_Stop_WithServer(t *testing.T) {
	t.Parallel()

	s := &WebStarter{
		router: &dummyRouter{},
		server: &failingServer{},
		config: Config{Host: "127.0.0.1", Port: 0},
		logger: log.Build(),
	}

	ctx := &mockApplicationContext{
		ctx: context.Background(),
	}

	err := s.Stop(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
