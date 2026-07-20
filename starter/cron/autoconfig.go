// Package cron 提供定时任务自动配置。
//
// Cron 是 Go 语言最流行的定时任务库，支持秒级定时任务。
// 本模块提供自动配置支持，内置任务恢复机制。
//
// 功能特性：
//   - 自动配置 Cron 调度器
//   - 支持秒级定时任务
//   - 支持 Cron 表达式
//   - 任务恢复机制
//
// 配置示例：
//
//	{
//	  "cron": {
//	    "enabled": true,
//	    "with-logger": false
//	  }
//	}
//
// 使用示例：
//
//	c := core.MustGetBean[*cron.Cron](app.Container())
//	c.AddFunc("0 30 * * * *", func() {
//	    fmt.Println("Every hour on the half hour")
//	})
package cron

import (
	"context"
	"fmt"
	"reflect"

	"github.com/robfig/cron/v3"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

// init 注册 Cron 自动配置类。
// 当配置 cron.enabled=true 时自动触发配置。
func init() {
	boot.RegisterAutoConfigWith(&CronAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(CronEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityTaskLayer)),
	)
}

// CronAutoConfiguration 定时任务自动配置类。
// 负责初始化 cron.Cron 调度器并注册到 IoC 容器。
type CronAutoConfiguration struct {
	logger log.Logger  // 日志记录器
	cron   *cron.Cron  // Cron 调度器实例
	config *CronConfig // Cron 配置信息
}

// Configure 配置定时任务。
// 创建 cron.Cron 调度器实例，配置日志和恢复机制，并注册到 IoC 容器。
func (c *CronAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
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
		return fmt.Errorf("加载 Cron 配置失败: %w", err)
	}

	c.config = cfg

	// 配置 Cron 选项
	opts := []cron.Option{
		cron.WithSeconds(), // 启用秒级定时任务
	}

	// 如果启用日志，配置自定义日志适配器
	if cfg.WithLogger {
		opts = append(opts, cron.WithLogger(&cronLogger{logger: c.logger}))
	}

	// 创建 Cron 调度器实例
	c.cron = cron.New(opts...)

	// 注册 Cron 实例到 IoC 容器
	if err := ctx.Container().RegisterInstance(c.cron, reflect.TypeFor[*cron.Cron]()); err != nil {
		return fmt.Errorf("注册 Cron 实例失败: %w", err)
	}

	c.logger.Info(context.Background(), "Cron 定时任务已配置")

	return nil
}

// Start 启动定时任务。
// 启动 Cron 调度器，开始执行已注册的定时任务。
// 注意：启动前需要确保已添加至少一个定时任务。
func (c *CronAutoConfiguration) Start() {
	if c.cron != nil {
		c.cron.Start()
		c.logger.Info(context.Background(), "Cron 定时任务已启动")
	}
}

// Stop 停止定时任务。
// 停止 Cron 调度器，等待正在执行的任务完成后退出。
func (c *CronAutoConfiguration) Stop() {
	if c.cron != nil {
		c.cron.Stop()
		c.logger.Info(context.Background(), "Cron 定时任务已停止")
	}
}

// GetCron 获取 Cron 实例。
// 返回底层的 *cron.Cron 实例，可用于高级操作。
func (c *CronAutoConfiguration) GetCron() *cron.Cron {
	return c.cron
}

// AddJob 添加定时任务。
// 使用 Cron 表达式定义任务执行时间。
//
// # Cron 表达式格式：秒 分 时 日 月 周
//
// 使用示例：
//
//	// 每分钟执行
//	cron.AddJob("0 * * * * *", myJob)
//
//	// 每小时 30 分执行
//	cron.AddJob("0 30 * * * *", myJob)
//
//	// 每天凌晨 2 点执行
//	cron.AddJob("0 0 2 * * *", myJob)
func (c *CronAutoConfiguration) AddJob(spec string, job cron.Job) (cron.EntryID, error) {
	return c.cron.AddJob(spec, job)
}

// AddFunc 添加定时任务函数。
// 使用 Cron 表达式定义任务执行时间，直接传入函数。
//
// 使用示例：
//
//	cron.AddFunc("0 * * * * *", func() {
//	    fmt.Println("每分钟执行")
//	})
func (c *CronAutoConfiguration) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	return c.cron.AddFunc(spec, cmd)
}

// Remove 移除定时任务。
// 根据任务 ID 移除已注册的定时任务。
func (c *CronAutoConfiguration) Remove(id cron.EntryID) {
	c.cron.Remove(id)
}

// CronConfig 定时任务配置。
// 包含 Cron 调度器的所有可配置参数。
type CronConfig struct {
	Enabled    bool `json:"enabled" mapstructure:"enabled"`         // 是否启用 Cron
	WithLogger bool `json:"with-logger" mapstructure:"with-logger"` // 是否启用日志
}

// cronLogger 实现 cron.Logger 接口。
// 将 Cron 的日志输出适配到 enhance 的日志系统。
type cronLogger struct {
	logger log.Logger
}

// Info 输出信息级别日志。
func (l *cronLogger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(context.Background(), msg, log.KeyValue{Key: "details", Value: keysAndValues})
}

// Error 输出错误级别日志。
func (l *cronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.logger.Error(context.Background(), msg, log.KeyValue{Key: "error", Value: err.Error()})
}

// loadConfig 从 Environment 加载 Cron 配置。
// 使用默认值初始化配置，然后从配置中心绑定用户自定义值。
func (c *CronAutoConfiguration) loadConfig(env *environment.Environment) (*CronConfig, error) {
	cfg := &CronConfig{
		WithLogger: DefaultWithLogger,
	}

	if err := env.BindPrefix("cron", cfg); err != nil {
		return nil, fmt.Errorf("绑定 Cron 配置失败: %w", err)
	}

	return cfg, nil
}
