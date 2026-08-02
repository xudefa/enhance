package context

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/config/refresh"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
	"github.com/xudefa/enhance/event"
	"github.com/xudefa/enhance/lifecycle"
)

// ApplicationContextBuilder 应用上下文构建器，支持链式配置
type ApplicationContextBuilder struct {
	container       core.Container
	env             *environment.Environment
	profiles        []string
	lifecycle       *lifecycle.LifecycleManager
	eventBus        *event.EventBusWithOrdering
	refreshOpts     []refresh.RefreshOption
	phaseListeners  []lifecycle.PhaseListener
	eventListeners  map[string][]event.EventListener
	beans           map[reflect.Type][]core.BeanOption
	refreshScopeMgr *refresh.RefreshScopeManager
}

// NewApplicationContextBuilder 创建应用上下文构建器
func NewApplicationContextBuilder() *ApplicationContextBuilder {
	return &ApplicationContextBuilder{
		eventListeners: make(map[string][]event.EventListener),
		beans:          make(map[reflect.Type][]core.BeanOption),
	}
}

// Container 设置IoC容器
func (b *ApplicationContextBuilder) Container(container core.Container) *ApplicationContextBuilder {
	b.container = container
	return b
}

// Environment 设置环境配置
func (b *ApplicationContextBuilder) Environment(env *environment.Environment) *ApplicationContextBuilder {
	b.env = env
	return b
}

// Profile 添加激活的 Profile
func (b *ApplicationContextBuilder) Profile(profile string) *ApplicationContextBuilder {
	b.profiles = append(b.profiles, profile)
	return b
}

// Profiles 批量添加激活的 Profiles
func (b *ApplicationContextBuilder) Profiles(profiles ...string) *ApplicationContextBuilder {
	b.profiles = append(b.profiles, profiles...)
	return b
}

// Lifecycle 设置生命周期管理器
func (b *ApplicationContextBuilder) Lifecycle(lifecycle *lifecycle.LifecycleManager) *ApplicationContextBuilder {
	b.lifecycle = lifecycle
	return b
}

// EventBus 设置事件总线
func (b *ApplicationContextBuilder) EventBus(eventBus *event.EventBusWithOrdering) *ApplicationContextBuilder {
	b.eventBus = eventBus
	return b
}

// RefreshScopeManager 设置刷新作用域管理器
func (b *ApplicationContextBuilder) RefreshScopeManager(mgr *refresh.RefreshScopeManager) *ApplicationContextBuilder {
	b.refreshScopeMgr = mgr
	return b
}

// WithRefreshOption 添加刷新配置选项
func (b *ApplicationContextBuilder) WithRefreshOption(opts ...refresh.RefreshOption) *ApplicationContextBuilder {
	b.refreshOpts = append(b.refreshOpts, opts...)
	return b
}

// WithPhaseListener 添加阶段监听器
func (b *ApplicationContextBuilder) WithPhaseListener(listener lifecycle.PhaseListener) *ApplicationContextBuilder {
	b.phaseListeners = append(b.phaseListeners, listener)
	return b
}

// WithEventListener 添加事件监听器
func (b *ApplicationContextBuilder) WithEventListener(eventType string, listener event.EventListener) *ApplicationContextBuilder {
	b.eventListeners[eventType] = append(b.eventListeners[eventType], listener)
	return b
}

// OnApplicationStarted 注册应用启动事件监听器
func (b *ApplicationContextBuilder) OnApplicationStarted(listener event.EventListener) *ApplicationContextBuilder {
	return b.WithEventListener(event.EventApplicationStarted, listener)
}

// OnApplicationReady 注册应用就绪事件监听器
func (b *ApplicationContextBuilder) OnApplicationReady(listener event.EventListener) *ApplicationContextBuilder {
	return b.WithEventListener(event.EventApplicationReady, listener)
}

// OnApplicationStopped 注册应用停止事件监听器
func (b *ApplicationContextBuilder) OnApplicationStopped(listener event.EventListener) *ApplicationContextBuilder {
	return b.WithEventListener(event.EventApplicationStopped, listener)
}

// Bean 注册Bean定义
func (b *ApplicationContextBuilder) Bean(t reflect.Type, opts ...core.BeanOption) *ApplicationContextBuilder {
	b.beans[t] = append(b.beans[t], opts...)
	return b
}

// Build 构建应用上下文
func (b *ApplicationContextBuilder) Build() (ApplicationContext, error) {
	// 创建默认容器
	if b.container == nil {
		b.container = core.NewContainer()
	}

	// 创建默认环境
	if b.env == nil {
		b.env = environment.NewEnvironment()
	}

	// 应用 Profiles 到环境配置
	for _, profile := range b.profiles {
		b.env.AddActiveProfile(profile)
	}

	// 创建默认事件总线
	if b.eventBus == nil {
		b.eventBus = event.NewEventBusWithOrdering()
	}

	// 创建应用上下文
	ctx := NewApplicationContext(b.container, b.env, b.refreshOpts...)

	// 替换事件总线（如果自定义）
	if b.eventBus != nil {
		ctx.events = b.eventBus
		ctx.asyncPublisher = event.NewAsyncPublisher(b.eventBus, event.WithWorkerCount(5))
	}

	// 替换生命周期管理器（如果自定义）
	if b.lifecycle != nil {
		ctx.lifecycle = b.lifecycle
	}

	// 替换刷新作用域管理器（如果自定义）
	if b.refreshScopeMgr != nil {
		ctx.refreshScopeMgr = b.refreshScopeMgr
	}

	// 注册阶段监听器
	for _, listener := range b.phaseListeners {
		ctx.lifecycle.AddListener(listener)
	}

	// 注册事件监听器
	for eventType, listeners := range b.eventListeners {
		for _, listener := range listeners {
			ctx.events.Subscribe(eventType, listener)
		}
	}

	// 注册Beans
	for t, opts := range b.beans {
		def := registry.BeanDef{
			Type: t,
		}
		for _, opt := range opts {
			opt(&def)
		}
		if err := ctx.container.RegisterBean(def); err != nil {
			beanID := ctx.container.Generate(t)
			return nil, fmt.Errorf("failed to register bean %s: %w", beanID, err)
		}
	}

	return ctx, nil
}

// MustBuild 构建应用上下文，失败则panic
func (b *ApplicationContextBuilder) MustBuild() ApplicationContext {
	ctx, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build application context: %v", err))
	}
	return ctx
}

// ApplicationContextHelper 应用上下文辅助工具
type ApplicationContextHelper struct {
	ctx ApplicationContext
}

// NewApplicationContextHelper 创建应用上下文辅助工具
func NewApplicationContextHelper(ctx ApplicationContext) *ApplicationContextHelper {
	return &ApplicationContextHelper{ctx: ctx}
}

// GetBeanByType 按类型获取Bean
func (h *ApplicationContextHelper) GetBeanByType(t reflect.Type) (any, error) {
	instances, err := h.ctx.Container().Get(t)
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return nil, core.ErrBeanNotFound
	}
	return instances[0], nil
}

// GetBeanByTypeOrDefault 按类型获取Bean，如果不存在返回默认值
func (h *ApplicationContextHelper) GetBeanByTypeOrDefault(t reflect.Type, defaultVal any) any {
	val, err := h.ctx.Container().Get(t)
	if err != nil || len(val) == 0 {
		return defaultVal
	}
	return val[0]
}

// HasBeanByType 检查指定类型的Bean是否存在
func (h *ApplicationContextHelper) HasBeanByType(t reflect.Type) bool {
	instances, err := h.ctx.Container().Get(t)
	return err == nil && len(instances) > 0
}

// GetProperty 获取字符串类型属性
func (h *ApplicationContextHelper) GetProperty(key string, defaultVal string) string {
	return h.ctx.Environment().GetString(key, defaultVal)
}

// GetIntProperty 获取整数类型属性
func (h *ApplicationContextHelper) GetIntProperty(key string, defaultVal int) int {
	return h.ctx.Environment().GetInt(key, defaultVal)
}

// GetBoolProperty 获取布尔类型属性
func (h *ApplicationContextHelper) GetBoolProperty(key string, defaultVal bool) bool {
	return h.ctx.Environment().GetBool(key, defaultVal)
}

// IsRunning 检查应用是否运行中
func (h *ApplicationContextHelper) IsRunning() bool {
	return h.ctx.Lifecycle().GetPhase() == lifecycle.PhaseRunning
}

// GetPhase 获取当前生命周期阶段
func (h *ApplicationContextHelper) GetPhase() lifecycle.ApplicationPhase {
	return h.ctx.Lifecycle().GetPhase()
}

// GetActiveProfiles 获取激活的Profiles
func (h *ApplicationContextHelper) GetActiveProfiles() []string {
	return h.ctx.Environment().GetActiveProfiles()
}

// IsDev 检查是否为开发环境
func (h *ApplicationContextHelper) IsDev() bool {
	return h.ctx.Environment().AcceptsProfile("dev")
}

// IsProd 检查是否为生产环境
func (h *ApplicationContextHelper) IsProd() bool {
	return h.ctx.Environment().AcceptsProfile("prod")
}

// PublishEvent 发布事件
func (h *ApplicationContextHelper) PublishEvent(eventType string) {
	h.ctx.EventPublisher().Publish(&event.BaseEvent{
		EventType: eventType,
	})
}

// PublishStarted 发布应用启动事件
func (h *ApplicationContextHelper) PublishStarted() {
	h.ctx.EventPublisher().Publish(&event.BaseEvent{
		EventType: event.EventApplicationStarted,
	})
}

// PublishReady 发布应用就绪事件
func (h *ApplicationContextHelper) PublishReady() {
	h.ctx.EventPublisher().Publish(&event.BaseEvent{
		EventType: event.EventApplicationReady,
	})
}

// PublishStopped 发布应用停止事件
func (h *ApplicationContextHelper) PublishStopped() {
	h.ctx.EventPublisher().Publish(&event.BaseEvent{
		EventType: event.EventApplicationStopped,
	})
}

// Invoke 调用函数并自动注入依赖
func (h *ApplicationContextHelper) Invoke(fn any) error {
	rv := reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		return fmt.Errorf("Invoke: fn must be a function")
	}

	fnType := rv.Type()
	numIn := fnType.NumIn()
	if numIn == 0 {
		return extractInvokeError(rv.Call(nil), fnType)
	}

	args := make([]reflect.Value, numIn)
	for i := 0; i < numIn; i++ {
		paramType := fnType.In(i)
		instances, err := h.ctx.Container().Get(paramType)
		if err != nil || len(instances) == 0 {
			return fmt.Errorf("Invoke: cannot resolve parameter %d of type %s", i+1, paramType)
		}
		args[i] = reflect.ValueOf(instances[0])
	}

	return extractInvokeError(rv.Call(args), fnType)
}

// ApplicationRunner 应用运行器，简化应用的启动和停止
type ApplicationRunner struct {
	ctx ApplicationContext
}

// NewApplicationRunner 创建应用运行器
func NewApplicationRunner(ctx ApplicationContext) *ApplicationRunner {
	return &ApplicationRunner{ctx: ctx}
}

// Run 运行应用，阻塞直到应用停止
func (r *ApplicationRunner) Run() error {
	if err := r.ctx.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	slog.Info("application started successfully")

	// 等待停止信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan
	slog.Info("received signal, shutting down...", "signal", sig)

	return r.Stop()
}

// Stop 停止应用
func (r *ApplicationRunner) Stop() error {
	if err := r.ctx.Stop(); err != nil {
		return fmt.Errorf("failed to stop application: %w", err)
	}

	slog.Info("application stopped successfully")
	return nil
}

// Context 获取应用上下文
func (r *ApplicationRunner) Context() ApplicationContext {
	return r.ctx
}

// CreateApplicationContext 创建并配置应用上下文的便捷函数
func CreateApplicationContext(opts ...func(*ApplicationContextBuilder)) (ApplicationContext, error) {
	builder := NewApplicationContextBuilder()

	for _, opt := range opts {
		opt(builder)
	}

	return builder.Build()
}

// WithContainer 设置容器的选项函数
func WithContainer(container core.Container) func(*ApplicationContextBuilder) {
	return func(b *ApplicationContextBuilder) {
		b.Container(container)
	}
}

// WithEnvironment 设置环境的选项函数
func WithEnvironment(env *environment.Environment) func(*ApplicationContextBuilder) {
	return func(b *ApplicationContextBuilder) {
		b.Environment(env)
	}
}

// WithBean 注册Bean的选项函数
func WithBean(t reflect.Type, opts ...core.BeanOption) func(*ApplicationContextBuilder) {
	return func(b *ApplicationContextBuilder) {
		b.Bean(t, opts...)
	}
}

// WithProfile 添加Profile的选项函数
func WithProfile(profile string) func(*ApplicationContextBuilder) {
	return func(b *ApplicationContextBuilder) {
		b.Profile(profile)
	}
}

// WithProfiles 批量添加Profiles的选项函数
func WithProfiles(profiles ...string) func(*ApplicationContextBuilder) {
	return func(b *ApplicationContextBuilder) {
		b.Profiles(profiles...)
	}
}
