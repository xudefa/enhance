package mvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	coreioc "github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/web/core"
)

// WebStarter Web MVC 启动器
//
// 负责初始化 HTTP 服务器、注册控制器和中间件。
// 实现 boot.Starter 接口，在应用启动时自动执行。
type WebStarter struct {
	server      core.Server
	router      core.Router
	config      Config
	middlewares []core.MiddlewareFunc
	name        string
	handler     http.Handler // 自定义处理器（如安全过滤器链）
	logger      log.Logger
	errCh       chan error     // 服务器启动/运行错误（可观测）
	wg          sync.WaitGroup // 服务器 goroutine 跟踪
}

// WebStarterOption 是 WebStarter 配置选项函数。
type WebStarterOption func(*WebStarter)

// WithConfig 设置 Web 配置。
func WithConfig(config Config) WebStarterOption {
	return func(s *WebStarter) {
		s.config = config
	}
}

// WithName 设置启动器名称。
func WithName(name string) WebStarterOption {
	return func(s *WebStarter) {
		s.name = name
	}
}

// WithLogger 设置日志记录器。
func WithLogger(logger log.Logger) WebStarterOption {
	return func(s *WebStarter) {
		s.logger = logger
	}
}

// WithServer 设置服务器实现。
func WithServer(server core.Server) WebStarterOption {
	return func(s *WebStarter) {
		s.server = server
	}
}

// WithRouter 设置路由器实现。
func WithRouter(router core.Router) WebStarterOption {
	return func(s *WebStarter) {
		s.router = router
	}
}

// WithMiddlewares 设置中间件列表。
func WithMiddlewares(middlewares []core.MiddlewareFunc) WebStarterOption {
	return func(s *WebStarter) {
		s.middlewares = middlewares
	}
}

// WithHandler 设置自定义处理器。
func WithHandler(handler http.Handler) WebStarterOption {
	return func(s *WebStarter) {
		s.handler = handler
	}
}

// NewWebStarter 创建新的 Web 启动器
func NewWebStarter(opts ...WebStarterOption) *WebStarter {
	starter := &WebStarter{
		config: DefaultConfig(),
		name:   "web",
		logger: log.Build(),
	}

	for _, opt := range opts {
		opt(starter)
	}

	return starter
}

// SetRouter 设置路由器（支持扩展，可替换为 gin/hertz 等）
func (s *WebStarter) SetRouter(router core.Router) {
	s.router = router
}

// SetServer 设置服务器（支持扩展，可替换为 gin/hertz 等）
func (s *WebStarter) SetServer(server core.Server) {
	s.server = server
}

// SetHandler 设置自定义处理器（如安全过滤器链）
// 如果设置了自定义处理器，Start() 将使用它而不是默认的 router
func (s *WebStarter) SetHandler(handler http.Handler) {
	s.handler = handler
}

// SetMiddlewares 设置中间件列表
func (s *WebStarter) SetMiddlewares(middlewares []core.MiddlewareFunc) {
	s.middlewares = middlewares
}

// AddMiddleware 添加中间件
func (s *WebStarter) AddMiddleware(middleware core.MiddlewareFunc) {
	s.middlewares = append(s.middlewares, middleware)
}

// Name 返回启动器名称
func (s *WebStarter) Name() string {
	return s.name
}

// Dependencies 返回依赖的其他启动器名称
func (s *WebStarter) Dependencies() []string {
	return nil
}

// Configure 配置阶段调用
func (s *WebStarter) Configure(ctx boot.ApplicationContext) error {
	if ctx == nil || ctx.Container() == nil {
		return nil
	}
	sCtx := ctx.Context()
	s.logger.Info(sCtx, "开始配置 Web MVC 模块...")

	// 获取控制器快照，避免 TOCTOU 竞态
	mu.RLock()
	snapshot := make([]core.Controller, len(controllers))
	copy(snapshot, controllers)
	mu.RUnlock()
	s.logger.Debug(sCtx, "发现控制器", log.KeyValue{Key: "count", Value: len(snapshot)})

	// 对每个控制器进行依赖注入
	for _, ctrl := range snapshot {
		if err := injectControllerDependencies(ctx, ctrl); err != nil {
			s.logger.Error(sCtx, "控制器依赖注入失败",
				log.KeyValue{Key: "controller", Value: fmt.Sprintf("%T", ctrl)},
				log.KeyValue{Key: "error", Value: err.Error()},
			)
			return fmt.Errorf("failed to inject controller %T: %w", ctrl, err)
		}
		s.logger.Debug(sCtx, "控制器依赖注入完成",
			log.KeyValue{Key: "controller", Value: fmt.Sprintf("%T", ctrl)},
		)
	}
	return nil
}

// injectControllerDependencies 注入控制器的依赖
//
// 控制器必须以指针注册（如 &MyController{}），
// 值类型控制器无法回写注入结果，将返回错误。
func injectControllerDependencies(ctx boot.ApplicationContext, ctrl core.Controller) error {
	ctrlValue := reflect.ValueOf(ctrl)

	// 如果是指针，获取元素
	if ctrlValue.Kind() == reflect.Ptr {
		ctrlValue = ctrlValue.Elem()
	}

	// 必须是结构体
	if ctrlValue.Kind() != reflect.Struct {
		return nil
	}

	// 值类型控制器不可寻址，注入结果无法回写到原始实例
	if !ctrlValue.CanAddr() {
		return fmt.Errorf("controller %T is a value type; register it as a pointer to enable dependency injection", ctrl)
	}

	ctrlType := ctrlValue.Type()
	container := ctx.Container()

	for i := 0; i < ctrlValue.NumField(); i++ {
		injectSingleField(container, ctrlType, ctrlValue, i)
	}

	return nil
}

// injectSingleField 尝试从容器注入单个控制器字段。
func injectSingleField(container coreioc.Container, ctrlType reflect.Type, ctrlValue reflect.Value, i int) {
	field := ctrlType.Field(i)
	fieldValue := ctrlValue.Field(i)

	// 跳过未导出的字段
	if !fieldValue.CanSet() {
		return
	}

	// 仅注入带有 inject 标签的字段
	if _, ok := field.Tag.Lookup("inject"); !ok {
		return
	}

	// 跳过已经设置的字段
	if isAlreadySet(fieldValue) {
		return
	}

	// 尝试从容器中获取依赖
	beans, err := container.Get(field.Type)
	if err != nil || len(beans) == 0 {
		return
	}

	// 注入值必须与字段类型兼容，避免类型不匹配时 panic
	bean := reflect.ValueOf(beans[0])
	if bean.IsValid() && bean.Type().AssignableTo(field.Type) {
		fieldValue.Set(bean)
	}
}

// isAlreadySet 判断指针/接口/切片/map 类型的字段是否已有值。
func isAlreadySet(fieldValue reflect.Value) bool {
	switch fieldValue.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
		return !fieldValue.IsNil()
	}
	return false
}

// Start 启动阶段调用
func (s *WebStarter) Start(ctx boot.ApplicationContext) error {
	sCtx := s.resolveContext(ctx)
	s.logger.Info(sCtx, "开始启动 Web 服务器...")

	// 创建默认路由器（如果未提供）
	if s.router == nil {
		return fmt.Errorf("router is required")
	}

	// 应用中间件并注册所有控制器（中间件必须先于控制器注册，路由注册时会快照当前中间件链）
	s.applyMiddlewaresAndRegisterControllers(sCtx)

	s.logger.Debug(sCtx, "中间件已应用",
		log.KeyValue{Key: "count", Value: len(s.middlewares)},
	)

	// 创建默认服务器（如果未提供）
	if s.server == nil {
		return fmt.Errorf("server is required")
	}

	s.applyServerContext(sCtx)

	// 设置处理器
	handler, err := s.resolveHandler(sCtx)
	if err != nil {
		return err
	}
	s.server.SetHandler(handler)

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.logger.Info(sCtx, "Web 服务器启动中",
		log.KeyValue{Key: "addr", Value: addr},
	)

	// 在后台启动服务器，错误可观测
	s.startServerInBackground()

	return nil
}

// resolveContext 获取应用上下文，context 为 nil 时返回 background。
func (s *WebStarter) resolveContext(ctx boot.ApplicationContext) context.Context {
	if ctx != nil {
		return ctx.Context()
	}
	return context.Background()
}

// applyMiddlewaresAndRegisterControllers 应用中间件并注册所有控制器。
func (s *WebStarter) applyMiddlewaresAndRegisterControllers(sCtx context.Context) {
	for _, mw := range s.middlewares {
		s.router.Use(mw)
	}

	controllers := GetControllers()
	for _, ctrl := range controllers {
		ctrl.Routes(s.router)
		s.logger.Info(sCtx, "控制器已注册",
			log.KeyValue{Key: "controller", Value: fmt.Sprintf("%T", ctrl)},
		)
	}
}

// applyServerContext 若服务器支持 SetContext 则传递应用上下文。
func (s *WebStarter) applyServerContext(sCtx context.Context) {
	type contextSetter interface {
		SetContext(ctx context.Context)
	}
	if cs, ok := s.server.(contextSetter); ok {
		cs.SetContext(sCtx)
	}
}

// resolveHandler 返回使用的 HTTP 处理器：优先使用自定义 handler，否则使用 router。
func (s *WebStarter) resolveHandler(sCtx context.Context) (http.Handler, error) {
	if s.handler != nil {
		s.logger.Debug(sCtx, "使用自定义处理器")
		return s.handler, nil
	}
	rt, ok := s.router.(http.Handler)
	if !ok {
		return nil, fmt.Errorf("router %T does not implement http.Handler", s.router)
	}
	s.logger.Debug(sCtx, "使用默认路由器")
	return rt, nil
}

// startServerInBackground 在后台启动服务器。
func (s *WebStarter) startServerInBackground() {
	s.errCh = make(chan error, 1)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case s.errCh <- err:
			default:
			}
		}
	}()
}

// Stop 停止阶段调用
func (s *WebStarter) Stop(ctx boot.ApplicationContext) error {
	var sCtx context.Context
	if ctx != nil {
		sCtx = ctx.Context()
	} else {
		sCtx = context.Background()
	}
	s.logger.Info(sCtx, "开始停止 Web 服务器...")
	if s.server != nil {
		timeoutCtx, cancel := context.WithTimeout(sCtx, 10*time.Second)
		defer cancel()
		if err := s.server.Stop(timeoutCtx); err != nil {
			s.logger.Error(sCtx, "Web 服务器停止失败",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
			return fmt.Errorf("failed to stop web server: %w", err)
		}
		s.logger.Info(sCtx, "Web 服务器已停止")
	}
	// 停止后等待服务器 goroutine 退出
	s.wg.Wait()
	return nil
}

// Wait 阻塞直到服务器退出（正常关闭返回 nil）。
//
// 若服务器启动失败（如端口被占用），返回该错误。
func (s *WebStarter) Wait() error {
	if s.errCh == nil {
		return nil
	}
	select {
	case err := <-s.errCh:
		return err
	}
}

// GetCondition 返回启动条件
func (s *WebStarter) GetCondition() condition.Condition {
	return nil
}
