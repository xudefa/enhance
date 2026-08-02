package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/xudefa/enhance/log"
)

func TestScheduleStarter_Name(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{}
	if starter.Name() != "ScheduleStarter" {
		t.Errorf("expected name 'ScheduleStarter', got %s", starter.Name())
	}
}

func TestScheduleStarter_Dependencies(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{}
	deps := starter.Dependencies()

	if deps != nil {
		t.Errorf("expected nil dependencies, got %v", deps)
	}
}

func TestScheduleStarter_Start_NoScheduler(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{
		logger: log.NewLoggerBuilder().Build(),
	}

	// 使用 nil context 模拟 ApplicationContext
	err := starter.Start(nil)

	if err != nil {
		t.Fatalf("start should not fail without scheduler: %v", err)
	}
}

func TestScheduleStarter_Stop_NoScheduler(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{
		logger: log.NewLoggerBuilder().Build(),
	}

	err := starter.Stop(nil)

	if err != nil {
		t.Fatalf("stop should not fail without scheduler: %v", err)
	}
}

func TestScheduleStarter_GetCondition(t *testing.T) {
	t.Parallel()

	starter := &ScheduleStarter{}
	cond := starter.GetCondition()

	if cond == nil {
		t.Error("condition should not be nil")
	}
}

func TestScheduleConfig_DefaultValues(t *testing.T) {
	t.Parallel()

	cfg := &ScheduleConfig{
		Enabled:  true,
		PoolSize: DefaultSchedulePoolSize,
	}

	if !cfg.Enabled {
		t.Error("expected enabled to be true by default")
	}

	if cfg.PoolSize != DefaultSchedulePoolSize {
		t.Errorf("expected pool size %d, got %d", DefaultSchedulePoolSize, cfg.PoolSize)
	}
}

func TestScheduler_CalculateNextRun_FixedDelay(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	task := NewFixedDelayTask("test-delay", 5*time.Second, func(ctx context.Context) error {
		return nil
	})

	// 注册任务以设置 scheduledTask
	_ = scheduler.Register(task)

	nextRun := scheduler.calculateNextRun(task, time.Now())

	// 验证下次运行时间在延迟时间附近
	if nextRun.Before(time.Now().Add(4 * time.Second)) {
		t.Error("next run should be at least 4 seconds from now")
	}
}

func TestScheduler_CalculateNextRun_FixedRate(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	task := NewFixedRateTask("test-rate", 10*time.Second, func(ctx context.Context) error {
		return nil
	})

	// 注册任务以设置 scheduledTask
	_ = scheduler.Register(task)

	nextRun := scheduler.calculateNextRun(task, time.Now())

	// 验证下次运行时间在间隔时间附近
	if nextRun.Before(time.Now().Add(9 * time.Second)) {
		t.Error("next run should be at least 9 seconds from now")
	}
}

func TestScheduler_CalculateNextRun_CronTask(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	task := NewTask("test-cron", "30 * * * * *", func(ctx context.Context) error {
		return nil
	})

	// 注册任务以设置 scheduledTask
	_ = scheduler.Register(task)

	nextRun := scheduler.calculateNextRun(task, time.Now())

	// 验证下次运行时间在未来
	if nextRun.Before(time.Now()) {
		t.Error("next run should be in the future")
	}

	// 验证下次运行时间在 60 秒内（因为每分钟的第 30 秒执行）
	maxExpected := time.Now().Add(60 * time.Second)
	if nextRun.After(maxExpected) {
		t.Errorf("next run should be within 60 seconds, got %v", nextRun)
	}
}

func TestScheduler_CalculateNextRun_InvalidCron(t *testing.T) {
	t.Parallel()

	scheduler := NewScheduler(context.Background())

	// 创建一个带有无效 cron 表达式的任务（通过直接构造）
	task := &invalidCronTask{name: "invalid", cron: "invalid cron"}

	nextRun := scheduler.calculateNextRun(task, time.Now())

	// 验证对于无效 cron，返回一个默认时间
	if nextRun.Before(time.Now()) {
		t.Error("next run should be in the future even for invalid cron")
	}
}

// invalidCronTask 用于测试无效 cron 表达式的任务
type invalidCronTask struct {
	name string
	cron string
}

func (t *invalidCronTask) Name() string {
	return t.name
}

func (t *invalidCronTask) Cron() string {
	return t.cron
}

func (t *invalidCronTask) Execute(ctx context.Context) error {
	return nil
}
