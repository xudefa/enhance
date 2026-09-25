package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/xudefa/enhance/schedule"
)

func main() {
	fmt.Println("=== Schedule Example ===")

	scheduler := newScheduler()
	counters := &taskCounters{}

	if err := registerDemoTasks(scheduler, counters); err != nil {
		fmt.Printf("Failed to register demo tasks: %v\n", err)
		return
	}

	if err := startScheduler(scheduler); err != nil {
		fmt.Printf("Failed to start scheduler: %v\n", err)
		return
	}

	fmt.Println("Scheduler started. Press Ctrl+C to stop.")

	go stopFixedRateTaskLater(scheduler)

	if err := waitForShutdown(scheduler); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Scheduler stopped successfully")
	fmt.Printf("Final counts - Cron: %d, FixedDelay: %d, FixedRate: %d\n",
		atomic.LoadInt32(&counters.cron),
		atomic.LoadInt32(&counters.fixedDelay),
		atomic.LoadInt32(&counters.fixedRate))
}

// newScheduler 创建带错误处理器与固定线程池的调度器。
func newScheduler() schedule.Scheduler {
	return schedule.NewScheduler(
		context.Background(),
		schedule.WithPoolSize(5),
		schedule.WithErrorHandler(func(taskName string, err error) {
			fmt.Printf("[ERROR] Task %s failed: %v\n", taskName, err)
		}),
	)
}

// taskCounters 统计三类任务各自的执行次数。
type taskCounters struct {
	cron       int32
	fixedDelay int32
	fixedRate  int32
}

// registerDemoTasks 注册 Cron、固定延迟与固定频率三个演示任务。
func registerDemoTasks(scheduler schedule.Scheduler, counters *taskCounters) error {
	cronTask := schedule.NewTask("cron-task", "0 */2 * * * *", func(ctx context.Context) error {
		count := atomic.AddInt32(&counters.cron, 1)
		fmt.Printf("[Cron] Task executed %d times at %s\n", count, time.Now().Format("15:04:05"))
		return nil
	})
	if err := scheduler.Register(cronTask); err != nil {
		return fmt.Errorf("failed to register cron task: %w", err)
	}

	fixedDelayTask := schedule.NewFixedDelayTask("fixed-delay-task", 1*time.Second, func(ctx context.Context) error {
		count := atomic.AddInt32(&counters.fixedDelay, 1)
		fmt.Printf("[FixedDelay] Task executed %d times at %s\n", count, time.Now().Format("15:04:05"))
		time.Sleep(500 * time.Millisecond) // 模拟任务执行时间
		return nil
	})
	if err := scheduler.Register(fixedDelayTask); err != nil {
		return fmt.Errorf("failed to register fixed delay task: %w", err)
	}

	fixedRateTask := schedule.NewFixedRateTask("fixed-rate-task", 3*time.Second, func(ctx context.Context) error {
		count := atomic.AddInt32(&counters.fixedRate, 1)
		fmt.Printf("[FixedRate] Task executed %d times at %s\n", count, time.Now().Format("15:04:05"))
		return nil
	})
	if err := scheduler.Register(fixedRateTask); err != nil {
		return fmt.Errorf("failed to register fixed rate task: %w", err)
	}
	return nil
}

// startScheduler 启动调度器。
func startScheduler(scheduler schedule.Scheduler) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return scheduler.Start(ctx)
}

// stopFixedRateTaskLater 5 秒后注销 fixed-rate-task 演示任务。
func stopFixedRateTaskLater(scheduler schedule.Scheduler) {
	time.Sleep(5 * time.Second)
	fmt.Println("\n=== Unregistering fixed-rate-task ===")
	scheduler.Unregister("fixed-rate-task")
	fmt.Println("Task unregistered")
}

// waitForShutdown 等待中断信号并优雅关闭调度器。
func waitForShutdown(scheduler schedule.Scheduler) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n=== Shutting down scheduler ===")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	return scheduler.Shutdown(shutdownCtx)
}
