// Package asynq 提供 Asynq 异步任务队列自动配置。
//
// Asynq 是基于 Redis 的异步任务队列库，支持任务重试和定时任务。
// 本模块提供自动配置支持，内置任务调度器。
//
// 功能特性：
//   - 自动配置 Asynq 客户端和调度器
//   - 支持任务重试
//   - 支持定时任务
//   - 支持任务优先级
//
// 配置示例：
//
//	{
//	  "asynq": {
//	    "enabled": true,
//	    "host": "localhost",
//	    "port": 6379,
//	    "enable-scheduler": false
//	  }
//	}
//
// 使用示例：
//
//	client := core.MustGetBean[*asynq.Client](app.Container())
//	task := asynq.NewTask("email:send", []byte(`{"to":"user@example.com"}`))
//	client.Enqueue(task)
package asynq

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/hibiken/asynq"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

// init 注册 Asynq 自动配置类。
// 当配置 asynq.enabled=true 时自动触发配置。
var asynqAutoConfig = &AsynqAutoConfiguration{}

func init() {
	boot.RegisterAutoConfigWith(asynqAutoConfig,
		boot.WithConditions(
			condition.OnProperty(AsynqEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityTaskLayer)),
	)
	boot.RegisterStarter(asynqAutoConfig)
}

// AsynqAutoConfiguration Asynq 异步任务队列自动配置类。
// 负责初始化 Asynq 客户端和调度器并注册到 IoC 容器。
type AsynqAutoConfiguration struct {
	logger    log.Logger       // 日志记录器
	client    *asynq.Client    // Asynq 客户端实例
	scheduler *asynq.Scheduler // Asynq 调度器实例
	config    *AsynqConfig     // Asynq 配置信息
	ctx       context.Context  // 应用上下文
}

// Configure 配置 Asynq 异步任务队列。
// 创建 Asynq 客户端和调度器实例，并注册到 IoC 容器。
func (c *AsynqAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	// 获取日志记录器，如果不存在则使用默认日志器
	if logger, err := core.GetByName[log.Logger](ctx.Container(), ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}

	// 加载配置
	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("failed to load Asynq config: %w", err)
	}

	c.config = cfg

	// 存储应用上下文
	c.ctx = ctx.Context()

	// 配置 Redis 连接选项
	opt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	}

	// 创建 Asynq 客户端
	c.client = asynq.NewClient(opt)

	// 如果启用调度器，创建调度器实例
	if cfg.EnableScheduler {
		c.scheduler = asynq.NewScheduler(opt, nil)
	}

	// 注册客户端实例到 IoC 容器
	if err := ctx.Container().RegisterInstance(c.client, reflect.TypeFor[*asynq.Client]()); err != nil {
		return fmt.Errorf("failed to register Asynq Client: %w", err)
	}

	// 如果启用了调度器，注册调度器实例到 IoC 容器
	if c.scheduler != nil {
		if err := ctx.Container().RegisterInstance(c.scheduler, reflect.TypeFor[*asynq.Scheduler]()); err != nil {
			return fmt.Errorf("failed to register Asynq Scheduler: %w", err)
		}
	}

	c.logger.Info(ctx.Context(), "Asynq task queue configured",
		log.KeyValue{Key: "host", Value: cfg.Host},
		log.KeyValue{Key: "port", Value: cfg.Port},
	)

	return nil
}

// Start 启动 Asynq 调度器。
// 如果启用了调度器，启动定时任务调度。
func (c *AsynqAutoConfiguration) Start(ctx boot.ApplicationContext) error {
	if c.scheduler != nil {
		return c.scheduler.Run()
	}
	return nil
}

// Stop 停止 Asynq 客户端。
// 关闭客户端连接和调度器，释放资源。
func (c *AsynqAutoConfiguration) Stop(ctx boot.ApplicationContext) error {
	var sCtx context.Context
	if ctx != nil {
		sCtx = ctx.Context()
	} else {
		sCtx = context.Background()
	}
	if c.client != nil {
		c.client.Close()
	}
	if c.scheduler != nil {
		c.scheduler.Shutdown()
	}
	c.logger.Info(sCtx, "Asynq task queue stopped")
	return nil
}

// Name 返回启动器名称。
func (c *AsynqAutoConfiguration) Name() string {
	return "AsynqStarter"
}

// Dependencies 返回依赖的其他启动器名称。
func (c *AsynqAutoConfiguration) Dependencies() []string {
	return nil
}

// GetCondition 返回启动器条件。
func (c *AsynqAutoConfiguration) GetCondition() condition.Condition {
	return condition.OnProperty(AsynqEnabled, ConditionTrue)
}

// GetClient 获取 Asynq Client 实例。
// 返回底层的 *asynq.Client 实例，可用于高级操作。
func (c *AsynqAutoConfiguration) GetClient() *asynq.Client {
	return c.client
}

// GetScheduler 获取 Asynq Scheduler 实例。
// 返回底层的 *asynq.Scheduler 实例，可用于高级操作。
func (c *AsynqAutoConfiguration) GetScheduler() *asynq.Scheduler {
	return c.scheduler
}

// Enqueue 添加任务到队列。
// 立即执行的任务会被添加到队列中等待处理。
//
// 使用示例：
//
//	task := asynq.NewTask("email:send", payload)
//	info, err := asynq.Enqueue(task, asynq.MaxRetry(3))
func (c *AsynqAutoConfiguration) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task, opts...)
}

// EnqueueAt 在指定时间添加任务。
// 任务会在指定的时间点执行。
//
// 使用示例：
//
//	task := asynq.NewTask("report:generate", payload)
//	tomorrow := time.Now().Add(24 * time.Hour)
//	asynq.EnqueueAt(task, tomorrow)
func (c *AsynqAutoConfiguration) EnqueueAt(task *asynq.Task, t time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task, append(opts, asynq.ProcessAt(t))...)
}

// EnqueueIn 在指定延迟后添加任务。
// 任务会在指定的延迟后执行。
//
// 使用示例：
//
//	task := asynq.NewTask("notification:send", payload)
//	asynq.EnqueueIn(task, 5*time.Minute)
func (c *AsynqAutoConfiguration) EnqueueIn(task *asynq.Task, d time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task, append(opts, asynq.ProcessIn(d))...)
}

// AsynqConfig Asynq 异步任务队列配置。
// 包含 Asynq 客户端和调度器的所有可配置参数。
type AsynqConfig struct {
	Enabled         bool   `json:"enabled" mapstructure:"enabled"`                   // 是否启用 Asynq
	Host            string `json:"host" mapstructure:"host"`                         // Redis 主机地址
	Port            int    `json:"port" mapstructure:"port"`                         // Redis 端口
	Password        string `json:"password" mapstructure:"password"`                 // Redis 密码
	DB              int    `json:"db" mapstructure:"db"`                             // Redis 数据库
	PoolSize        int    `json:"pool-size" mapstructure:"pool-size"`               // 连接池大小
	EnableScheduler bool   `json:"enable-scheduler" mapstructure:"enable-scheduler"` // 是否启用调度器
}

// 配置常量。
const (
	AsynqEnabled           = "asynq.enabled" // 启用条件配置键
	DefaultAsynqHost       = "localhost"     // 默认 Redis 主机
	DefaultAsynqPort       = 6379            // 默认 Redis 端口
	DefaultAsynqDB         = 0               // 默认 Redis 数据库
	DefaultAsynqPoolSize   = 10              // 默认连接池大小
	DefaultEnableScheduler = false           // 默认不启用调度器
	ConditionTrue          = "true"          // 条件真值
)

// loadConfig 从 Environment 加载 Asynq 配置。
// 使用默认值初始化配置，然后从配置中心绑定用户自定义值。
func (c *AsynqAutoConfiguration) loadConfig(env *environment.Environment) (*AsynqConfig, error) {
	cfg := &AsynqConfig{
		Host:            DefaultAsynqHost,
		Port:            DefaultAsynqPort,
		DB:              DefaultAsynqDB,
		PoolSize:        DefaultAsynqPoolSize,
		EnableScheduler: DefaultEnableScheduler,
	}

	if err := env.BindPrefix("asynq", cfg); err != nil {
		return nil, fmt.Errorf("failed to bind Asynq config: %w", err)
	}

	return cfg, nil
}
