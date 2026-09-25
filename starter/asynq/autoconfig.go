package asynq

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/hibiken/asynq"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

var asynqAutoConfig = &AsynqAutoConfiguration{}

func init() {
	boot.RegisterAutoConfigWith(asynqAutoConfig,
		boot.WithConditions(
			condition.OnPropertyOrDefault(AsynqEnabled, ConditionTrue, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityTaskLayer)),
	)
	boot.RegisterStarter(asynqAutoConfig)
}

// AsynqAutoConfiguration Asynq 异步任务队列自动配置类。
// 负责初始化 Asynq 客户端和调度器并注册到 IoC 容器。
type AsynqAutoConfiguration struct {
	logger     log.Logger
	client     *asynq.Client
	scheduler  *asynq.Scheduler
	config     *AsynqConfig
	mu         sync.Mutex
	configured bool
	ctx        context.Context
}

// Configure 配置 Asynq 异步任务队列。
// 创建 Asynq 客户端和调度器实例，并注册到 IoC 容器。
// 同一应用上下文重复调用直接返回；新上下文（重启）会先释放旧资源再重新配置。
// 若配置中 Enabled=false，跳过客户端创建，直接返回。
func (c *AsynqAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.configured && c.ctx == ctx.Context() {
		return nil
	}

	c.closeResources()

	env := ctx.Environment()

	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	cfg, err := LoadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Asynq config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid Asynq config: %w", err)
	}

	c.config = cfg

	if !cfg.Enabled {
		c.logger.Info(ctx.Context(), "Asynq disabled by configuration")
		c.configured = true
		c.ctx = ctx.Context()
		return nil
	}

	opt := cfg.RedisClientOpt()

	c.client = asynq.NewClient(opt)

	if cfg.EnableScheduler {
		c.scheduler = asynq.NewScheduler(opt, nil)
	}

	if err := ctx.Container().RegisterInstance(c.client, reflect.TypeFor[*asynq.Client]()); err != nil {
		c.closeResources()
		return fmt.Errorf("failed to register Asynq Client: %w", err)
	}

	if c.scheduler != nil {
		if err := ctx.Container().RegisterInstance(c.scheduler, reflect.TypeFor[*asynq.Scheduler]()); err != nil {
			c.closeResources()
			return fmt.Errorf("failed to register Asynq Scheduler: %w", err)
		}
	}

	c.logger.Info(ctx.Context(), "Asynq task queue configured",
		log.KeyValue{Key: "config", Value: cfg.String()},
	)

	c.configured = true
	c.ctx = ctx.Context()

	return nil
}

// Start 启动 Asynq 调度器。
// 如果启用了调度器，启动定时任务调度。
func (c *AsynqAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.scheduler != nil {
		c.logger.Info(ctx.Context(), "Asynq scheduler starting")
		return c.scheduler.Run()
	}
	return nil
}

// Stop 停止 Asynq 客户端。
// 关闭客户端连接和调度器，释放资源。
// 多次调用安全（幂等）。
func (c *AsynqAutoConfiguration) Stop(ctx boot.ApplicationContext) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.configured {
		return nil
	}

	c.closeResources()
	c.configured = false
	c.logger.Info(ctx.Context(), "Asynq task queue stopped")

	return nil
}

// closeResources 关闭客户端和调度器，释放资源。
// 调用方必须持有 c.mu 锁。
func (c *AsynqAutoConfiguration) closeResources() {
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
	if c.scheduler != nil {
		c.scheduler.Shutdown()
		c.scheduler = nil
	}
	c.config = nil
}

// Name 返回启动器名称。
func (c *AsynqAutoConfiguration) Name() string {
	return StarterName
}

// Dependencies 返回依赖的其他启动器名称。
func (c *AsynqAutoConfiguration) Dependencies() []string {
	return nil
}

// GetCondition 返回启动器条件。
func (c *AsynqAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnPropertyOrDefault(AsynqEnabled, ConditionTrue, ConditionTrue)
}

// GetClient 获取 Asynq Client 实例。
func (c *AsynqAutoConfiguration) GetClient() *asynq.Client {
	return c.client
}

// GetScheduler 获取 Asynq Scheduler 实例。
func (c *AsynqAutoConfiguration) GetScheduler() *asynq.Scheduler {
	return c.scheduler
}

// GetConfig 获取当前 Asynq 配置。
// Configure 未执行时返回 nil。
func (c *AsynqAutoConfiguration) GetConfig() *AsynqConfig {
	return c.config
}

// IsConfigured 返回是否已完成配置。
func (c *AsynqAutoConfiguration) IsConfigured() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.configured
}

// Enqueue 添加任务到队列。
func (c *AsynqAutoConfiguration) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if c.client == nil {
		return nil, ErrClientNotConfigured
	}
	return c.client.Enqueue(task, opts...)
}

// EnqueueAt 在指定时间添加任务。
func (c *AsynqAutoConfiguration) EnqueueAt(task *asynq.Task, t time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.enqueueWithOptions(task, asynq.ProcessAt(t), opts)
}

// EnqueueIn 在指定延迟后添加任务。
func (c *AsynqAutoConfiguration) EnqueueIn(task *asynq.Task, d time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.enqueueWithOptions(task, asynq.ProcessIn(d), opts)
}

// enqueueWithOptions 合并额外选项并提交任务。
func (c *AsynqAutoConfiguration) enqueueWithOptions(task *asynq.Task, extra asynq.Option, opts []asynq.Option) (*asynq.TaskInfo, error) {
	if c.client == nil {
		return nil, ErrClientNotConfigured
	}
	allOpts := make([]asynq.Option, 0, len(opts)+1)
	allOpts = append(allOpts, opts...)
	allOpts = append(allOpts, extra)
	return c.client.Enqueue(task, allOpts...)
}

// BuildAddr 构建 Redis 连接地址。
func (cfg *AsynqConfig) BuildAddr() string {
	return fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
}

// RedisClientOpt 根据 Asynq 配置构建 asynq.RedisClientOpt。
func (cfg *AsynqConfig) RedisClientOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.BuildAddr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	}
}

// ApplyDefaults 为零值字段补充默认值。
// 适用于 JSON/YAML 反序列化后补全配置。
func (cfg *AsynqConfig) ApplyDefaults() {
	if cfg.Host == "" {
		cfg.Host = DefaultAsynqHost
	}
	if cfg.Port == 0 {
		cfg.Port = DefaultAsynqPort
	}
	if cfg.PoolSize == 0 {
		cfg.PoolSize = DefaultAsynqPoolSize
	}
	if cfg.Concurrency == 0 {
		cfg.Concurrency = DefaultConcurrency
	}
	if cfg.RetryLimit == 0 {
		cfg.RetryLimit = DefaultRetryLimit
	}
}

// String 返回配置的可读表示，便于日志和调试。
// Password 字段脱敏显示。
func (cfg *AsynqConfig) String() string {
	pwd := "****"
	if cfg.Password == "" {
		pwd = ""
	}
	return fmt.Sprintf("AsynqConfig{Enabled=%v, Host=%s, Port=%d, Password=%s, DB=%d, PoolSize=%d, EnableScheduler=%v, Concurrency=%d, RetryLimit=%d, Timeout=%v}",
		cfg.Enabled, cfg.Host, cfg.Port, pwd, cfg.DB, cfg.PoolSize, cfg.EnableScheduler, cfg.Concurrency, cfg.RetryLimit, cfg.Timeout)
}

// maxRedisPort Redis 端口的最大值（TCP 端口范围为 1-65535）。
const maxRedisPort = 65535

// Validate 校验 Asynq 配置合法性。
func (cfg *AsynqConfig) Validate() error {
	if cfg.Host == "" {
		return ErrInvalidHost
	}
	if cfg.Port <= 0 || cfg.Port > maxRedisPort {
		return ErrInvalidPort
	}
	if cfg.DB < 0 {
		return ErrInvalidDB
	}
	if cfg.PoolSize < 0 {
		return ErrInvalidPoolSize
	}
	if cfg.Concurrency < 0 {
		return ErrInvalidConcurrency
	}
	if cfg.RetryLimit < 0 {
		return ErrInvalidRetryLimit
	}
	if cfg.Timeout < 0 {
		return ErrInvalidTimeout
	}
	return nil
}

// NewAsynqConfig 创建带默认值的 Asynq 配置。
func NewAsynqConfig() *AsynqConfig {
	return &AsynqConfig{
		Host:            DefaultAsynqHost,
		Port:            DefaultAsynqPort,
		DB:              DefaultAsynqDB,
		PoolSize:        DefaultAsynqPoolSize,
		EnableScheduler: DefaultEnableScheduler,
		Concurrency:     DefaultConcurrency,
		RetryLimit:      DefaultRetryLimit,
		Timeout:         DefaultTimeout,
	}
}

// LoadConfig 从 Environment 加载 Asynq 配置。
// 使用默认值初始化配置，然后从配置中心绑定用户自定义值。
func LoadConfig(env *environment.Environment) (*AsynqConfig, error) {
	cfg := NewAsynqConfig()

	if err := env.BindPrefix(ConfigPrefix, cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Asynq config: %w", err)
	}

	return cfg, nil
}
