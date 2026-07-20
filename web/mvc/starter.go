package mvc

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/xudefa/enhance/actuator"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
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
	s := &WebStarter{
		config: DefaultConfig(),
		name:   "web",
		logger: log.Build(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
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
	s.logger.Info(context.Background(), "开始配置 Web MVC 模块...")

	// 如果上下文为 nil，跳过依赖注入（用于测试场景）
	if ctx == nil || ctx.Container() == nil {
		return nil
	}

	// 直接操作原始控制器列表，避免副本问题
	mu.RLock()
	ctrlCount := len(controllers)
	mu.RUnlock()
	s.logger.Debug(context.Background(), "发现控制器", log.KeyValue{Key: "count", Value: ctrlCount})

	// 对每个控制器进行依赖注入
	for i := 0; i < ctrlCount; i++ {
		mu.RLock()
		ctrl := controllers[i]
		mu.RUnlock()
		if err := injectControllerDependencies(ctx, ctrl); err != nil {
			s.logger.Error(context.Background(), "控制器依赖注入失败",
				log.KeyValue{Key: "controller", Value: fmt.Sprintf("%T", ctrl)},
				log.KeyValue{Key: "error", Value: err.Error()},
			)
			return fmt.Errorf("failed to inject controller %T: %w", ctrl, err)
		}
		s.logger.Debug(context.Background(), "控制器依赖注入完成",
			log.KeyValue{Key: "controller", Value: fmt.Sprintf("%T", ctrl)},
		)
	}
	return nil
}

// injectControllerDependencies 注入控制器的依赖
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

	ctrlType := ctrlValue.Type()
	container := ctx.Container()

	// 遍历所有字段
	for i := 0; i < ctrlValue.NumField(); i++ {
		field := ctrlType.Field(i)
		fieldValue := ctrlValue.Field(i)

		// 跳过未导出的字段
		if !fieldValue.CanSet() {
			continue
		}

		// 跳过已经设置的字段
		switch fieldValue.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map:
			if !fieldValue.IsNil() {
				continue
			}
		}

		// 尝试从容器中获取依赖
		fieldType := field.Type
		beans, err := container.Get(fieldType)
		if err != nil || len(beans) == 0 {
			// 依赖不存在，跳过（可能是可选依赖）
			continue
		}

		// 设置字段值
		fieldValue.Set(reflect.ValueOf(beans[0]))
	}

	return nil
}

// Start 启动阶段调用
func (s *WebStarter) Start(ctx boot.ApplicationContext) error {
	s.logger.Info(context.Background(), "开始启动 Web 服务器...")

	// 创建默认路由器（如果未提供）
	if s.router == nil {
		return fmt.Errorf("router is required")
	}

	// 注册所有控制器
	controllers := GetControllers()
	for _, ctrl := range controllers {
		ctrl.Routes(s.router)
		s.logger.Info(context.Background(), "控制器已注册",
			log.KeyValue{Key: "controller", Value: fmt.Sprintf("%T", ctrl)},
		)
	}

	// 应用中间件
	for _, mw := range s.middlewares {
		s.router.Use(mw)
	}

	s.logger.Debug(context.Background(), "中间件已应用",
		log.KeyValue{Key: "count", Value: len(s.middlewares)},
	)

	// 注册 HttpEndpointRegistry,允许 Actuator 等模块自动挂载端点
	if ctx != nil && ctx.Container() != nil {
		if endpointRegistry, err := s.createEndpointRegistry(); err == nil && endpointRegistry != nil {
			if err := ctx.Container().RegisterInstance(endpointRegistry, reflect.TypeFor[actuator.HttpEndpointRegistry]()); err != nil {
				s.logger.Warn(context.Background(), "注册 HttpEndpointRegistry 失败,Actuator 端点将无法自动挂载",
					log.KeyValue{Key: "error", Value: err.Error()},
				)
			}
		}
	}

	// 创建默认服务器（如果未提供）
	if s.server == nil {
		return fmt.Errorf("server is required")
	}

	// 设置处理器
	if s.handler != nil {
		// 使用自定义处理器（如安全过滤器链）
		s.server.SetHandler(s.handler)
		s.logger.Debug(context.Background(), "使用自定义处理器")
	} else {
		// 使用默认路由器
		s.server.SetHandler(s.router)
		s.logger.Debug(context.Background(), "使用默认路由器")
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.logger.Info(context.Background(), "Web 服务器启动中",
		log.KeyValue{Key: "addr", Value: addr},
	)

	// 在后台启动服务器
	go func() {
		if err := s.server.Start(); err != nil && err != http.ErrServerClosed {
			s.logger.Error(context.Background(), "Web 服务器运行错误",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
		}
	}()

	return nil
}

// createEndpointRegistry 创建 HttpEndpointRegistry 实例
func (s *WebStarter) createEndpointRegistry() (actuator.HttpEndpointRegistry, error) {
	// 如果 router 实现了 http.Handler,创建一个基于 Handler 的注册表
	if handler, ok := s.router.(http.Handler); ok {
		return &handlerBasedEndpointRegistry{
			handler: handler,
			mux:     http.NewServeMux(),
		}, nil
	}

	return nil, fmt.Errorf("router does not implement http.Handler")
}

// handlerBasedEndpointRegistry 基于 http.Handler 的端点注册表
type handlerBasedEndpointRegistry struct {
	handler   http.Handler
	mux       *http.ServeMux
	endpoints map[string]bool
}

func (r *handlerBasedEndpointRegistry) RegisterEndpoint(method, path string, handler http.Handler) {
	if r.mux == nil || handler == nil {
		return
	}

	pattern := path
	if method != "" {
		pattern = method + " " + path
	}
	r.mux.Handle(pattern, handler)

	if r.endpoints == nil {
		r.endpoints = make(map[string]bool)
	}
	r.endpoints[path] = true
}

func (r *handlerBasedEndpointRegistry) RegisterEndpoints(endpoints []actuator.EndpointConfig) {
	for _, ep := range endpoints {
		r.RegisterEndpoint(ep.Method, ep.Path, ep.Handler)
	}
}

func (r *handlerBasedEndpointRegistry) HasEndpoint(path string) bool {
	_, exists := r.endpoints[path]
	return exists
}

// Stop 停止阶段调用
func (s *WebStarter) Stop(ctx boot.ApplicationContext) error {
	s.logger.Info(context.Background(), "开始停止 Web 服务器...")
	if s.server != nil {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.server.Stop(timeoutCtx); err != nil {
			s.logger.Error(context.Background(), "Web 服务器停止失败",
				log.KeyValue{Key: "error", Value: err.Error()},
			)
			return err
		}
		s.logger.Info(context.Background(), "Web 服务器已停止")
	}
	return nil
}

// GetCondition 返回启动条件
func (s *WebStarter) GetCondition() condition.Condition {
	return nil
}
