package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xudefa/enhance/schedule"
)

func main() {
	fmt.Println("=== Schedule Builder Example ===")

	// 使用 Builder 模式创建调度器
	scheduler := schedule.NewSchedulerBuilder().
		PoolSize(10).
		ErrorHandler(func(taskName string, err error) {
			fmt.Printf("[ERROR] Task %s failed: %v\n", taskName, err)
		}).
		WithCronTask("cleanup-task", "0 0 */1 * * *", func(ctx context.Context) error {
			fmt.Printf("[Cleanup] Running cleanup at %s\n", time.Now().Format("15:04:05"))
			return nil
		}).
		WithFixedDelayTask("monitor-task", 2*time.Second, func(ctx context.Context) error {
			fmt.Printf("[Monitor] System check at %s\n", time.Now().Format("15:04:05"))
			return nil
		}).
		WithFixedRateTask("report-task", 5*time.Second, func(ctx context.Context) error {
			fmt.Printf("[Report] Generating report at %s\n", time.Now().Format("15:04:05"))
			return nil
		}).
		Build()

		// 启动调度器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := scheduler.Start(ctx); err != nil {
		fmt.Printf("Failed to start scheduler: %v\n", err)
		return
	}

	fmt.Println("Scheduler started. Press Ctrl+C to stop.")

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
}
