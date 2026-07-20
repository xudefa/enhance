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

	// 使用 Helper 工具
	scheduler := schedule.NewScheduler()
	helper := schedule.NewScheduleHelper(scheduler)

	if err := helper.RegisterCronTask("backup-task", "0 30 * * * *", func(ctx context.Context) error {
		fmt.Printf("[Backup] Running backup at %s\n", time.Now().Format("15:04:05"))
		return nil
	}); err != nil {
		fmt.Printf("Failed to register backup task: %v\n", err)
		return
	}

	fmt.Printf("Registered tasks: %d\n", helper.GetTaskCount())
	fmt.Printf("Has cleanup-task: %v\n", helper.HasTask("cleanup-task"))

	// 启动调度器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := scheduler.Start(ctx); err != nil {
		fmt.Printf("Failed to start scheduler: %v\n", err)
		return
	}

	fmt.Println("Scheduler started. Press Ctrl+C to stop.")

	// 等待10秒后注销monitor-task
	go func() {
		time.Sleep(10 * time.Second)
		fmt.Println("\n=== Unregistering monitor-task ===")
		if helper.UnregisterTask("monitor-task") {
			fmt.Println("Task unregistered successfully")
		} else {
			fmt.Println("Task not found")
		}
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
}
