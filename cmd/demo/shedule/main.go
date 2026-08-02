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

	// 创建调度器
	scheduler := schedule.NewScheduler(
		context.Background(),
		schedule.WithPoolSize(5),
		schedule.WithErrorHandler(func(taskName string, err error) {
			fmt.Printf("[ERROR] Task %s failed: %v\n", taskName, err)
		}),
	)

	// 计数器用于演示
	var cronCount, fixedDelayCount, fixedRateCount int32

	// 注册 Cron 任务：每2秒执行一次
	cronTask := schedule.NewTask("cron-task", "0 */2 * * * *", func(ctx context.Context) error {
		count := atomic.AddInt32(&cronCount, 1)
		fmt.Printf("[Cron] Task executed %d times at %s\n", count, time.Now().Format("15:04:05"))
		return nil
	})

	if err := scheduler.Register(cronTask); err != nil {
		fmt.Printf("Failed to register cron task: %v\n", err)
		return
	}

	// 注册固定延迟任务：延迟1秒执行
	fixedDelayTask := schedule.NewFixedDelayTask("fixed-delay-task", 1*time.Second, func(ctx context.Context) error {
		count := atomic.AddInt32(&fixedDelayCount, 1)
		fmt.Printf("[FixedDelay] Task executed %d times at %s\n", count, time.Now().Format("15:04:05"))
		time.Sleep(500 * time.Millisecond) // 模拟任务执行时间
		return nil
	})

	if err := scheduler.Register(fixedDelayTask); err != nil {
		fmt.Printf("Failed to register fixed delay task: %v\n", err)
		return
	}

	// 注册固定频率任务：每3秒执行一次
	fixedRateTask := schedule.NewFixedRateTask("fixed-rate-task", 3*time.Second, func(ctx context.Context) error {
		count := atomic.AddInt32(&fixedRateCount, 1)
		fmt.Printf("[FixedRate] Task executed %d times at %s\n", count, time.Now().Format("15:04:05"))
		return nil
	})

	if err := scheduler.Register(fixedRateTask); err != nil {
		fmt.Printf("Failed to register fixed rate task: %v\n", err)
		return
	}

	// 启动调度器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := scheduler.Start(ctx); err != nil {
		fmt.Printf("Failed to start scheduler: %v\n", err)
		return
	}

	fmt.Println("Scheduler started. Press Ctrl+C to stop.")

	// 等待5秒后注销一个任务
	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("\n=== Unregistering fixed-rate-task ===")
		scheduler.Unregister("fixed-rate-task")
		fmt.Println("Task unregistered")
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n=== Shutting down scheduler ===")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := scheduler.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Scheduler stopped successfully")
	fmt.Printf("Final counts - Cron: %d, FixedDelay: %d, FixedRate: %d\n",
		atomic.LoadInt32(&cronCount),
		atomic.LoadInt32(&fixedDelayCount),
		atomic.LoadInt32(&fixedRateCount))
}
