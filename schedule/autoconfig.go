// Package schedule 提供定时任务调度功能，用于 enhance 框架。
package schedule

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

// ScheduleAutoConfiguration 调度器自动配置类。
//
// 当配置文件中启用调度器时自动生效（schedule.enabled=true）。
// 负责创建调度器实例并注册到 IoC 容器中。
type ScheduleAutoConfiguration struct {
	logger log.Logger
}

// Configure 配置调度器。
//
// 该方法在自动配置阶段调用，负责：
//  1. 从 Environment 中读取调度器配置
//  2. 创建调度器实例
//  3. 注册 Scheduler 到 IoC 容器
func (c *ScheduleAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()
	env := ctx.Environment()

	// 从容器获取日志记录器，如果不存在则使用默认值
	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.NewLoggerBuilder().Build()
	}

	cfg, err := c.loadConfig(env)
	if err != nil {
		return fmt.Errorf("加载调度器配置失败: %w", err)
	}

	if !cfg.Enabled {
		return nil
	}

	scheduler := NewScheduler(
		WithPoolSize(cfg.PoolSize),
		WithLogger(c.logger),
	)

	// 注册调度器到 IoC 容器
	if err := container.RegisterInstance(scheduler, reflect.TypeOf(scheduler)); err != nil {
		return fmt.Errorf("注册调度器失败: %w", err)
	}

	c.logger.Info(context.Background(), "调度器已配置",
		log.KeyValue{Key: "pool_size", Value: cfg.PoolSize})

	return nil
}

// ScheduleConfig 调度器配置。
type ScheduleConfig struct {
	Enabled  bool `json:"enabled" mapstructure:"enabled"`
	PoolSize int  `json:"pool-size" mapstructure:"pool-size"`
}

// loadConfig 从 Environment 加载调度器配置。
func (c *ScheduleAutoConfiguration) loadConfig(env *environment.Environment) (*ScheduleConfig, error) {
	cfg := &ScheduleConfig{
		Enabled:  true,
		PoolSize: DefaultSchedulePoolSize,
	}

	if err := env.BindPrefix("schedule", cfg); err != nil {
		return nil, fmt.Errorf("绑定调度器配置失败: %w", err)
	}

	return cfg, nil
}

// ScheduleStarter 调度器启动器。
//
// 管理调度器的生命周期，在应用启动时启动调度器，
// 在应用停止时优雅关闭调度器。
type ScheduleStarter struct {
	scheduler *DefaultScheduler
	logger    log.Logger
}

// Name 返回启动器名称。
func (s *ScheduleStarter) Name() string {
	return "ScheduleStarter"
}

// Dependencies 返回依赖的其他启动器名称。
func (s *ScheduleStarter) Dependencies() []string {
	return nil
}

// Configure 配置阶段调用。
func (s *ScheduleStarter) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()

	// 从容器获取日志记录器
	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		s.logger = logger
	} else {
		s.logger = log.NewLoggerBuilder().Build()
	}

	// 从容器获取调度器
	bean, err := ctx.GetByType(reflect.TypeOf(&DefaultScheduler{}))
	if err != nil {
		return nil
	}

	s.scheduler = bean.(*DefaultScheduler)
	return nil
}

// Start 启动阶段调用。
func (s *ScheduleStarter) Start(ctx boot.ApplicationContext) error {
	if s.scheduler == nil {
		return nil
	}

	appCtx := context.Background()
	if err := s.scheduler.Start(appCtx); err != nil {
		return fmt.Errorf("启动调度器失败: %w", err)
	}

	s.logger.Info(context.Background(), "调度器已启动")
	return nil
}

// Stop 停止阶段调用。
func (s *ScheduleStarter) Stop(ctx boot.ApplicationContext) error {
	if s.scheduler == nil {
		return nil
	}

	appCtx := context.Background()
	shutdownCtx, cancel := context.WithTimeout(appCtx, 30*time.Second)
	defer cancel()

	if err := s.scheduler.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("关闭调度器失败: %w", err)
	}

	s.logger.Info(context.Background(), "调度器已停止")
	return nil
}

// GetCondition 返回启动条件。
func (s *ScheduleStarter) GetCondition() condition.Condition {
	return condition.OnProperty(ScheduleEnabled, ConditionTrue)
}

var _ boot.Starter = (*ScheduleStarter)(nil)

func init() {
	boot.RegisterAutoConfigWith(&ScheduleAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(ScheduleEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPriorityBusinessLayer)), // 业务层，在 Web 层之后执行
	)

	boot.RegisterStarter(&ScheduleStarter{})
}
