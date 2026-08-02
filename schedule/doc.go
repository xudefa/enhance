// Package schedule 提供定时任务调度功能，用于 enhance 框架。
//
// 该模块提供灵活的定时任务调度机制，支持基于Cron 表达式、固定延迟/固定频率执行等。
// 参考 Spring 的 Scheduler 设计。
//
// # 架构设计
//
//   - Task: 任务接口，定义任务执行逻辑
//   - Scheduler: 调度器接口，负责任务调度
//   - SchedulerOption: 调度器选项函数
//   - CronExpression: Cron 表达式解析器
//   - DefaultScheduler: 默认调度器实现
//
// # 核心功能
//
//   - Cron 表达式: 支持标准 Cron 表达式定义执行时间
//   - 固定延迟: 支持固定延迟执行任务
//   - 固定频率: 支持固定频率执行任务
//   - 并发控制: 支持任务并发执行控制
//   - 错误处理: 支持任务执行失败后的重试
//
// # Cron 表达式
//
// 支持 6 位 Cron 表达式：秒 分 时 日 月 周
//
//   - "0 */5 * * * *": 每 5 分钟执行
//   - "0 0 */1 * * *": 每小时执行
//   - "0 0 0 * * *": 每天零点执行
//   - "0 0 0 * * MON-FRI": 工作日零点执行
//
// # 配置选项
//
//   - WithPoolSize: 设置任务执行池大小（最大并发数）
//   - WithErrorHandler: 设置执行错误处理函数
package schedule

import (
	"context"
)

// Task 定时任务接口。
//
// 每个定时任务包含名称、Cron 表达式和执行逻辑。
// 通过 NewTask 创建基于函数的任务实例。
type Task interface {
	// Name 返回任务名称，用于注册和查找。
	Name() string
	// Cron 返回任务的 cron 表达式（6 字段 Spring 风格）。
	Cron() string
	// Execute 执行任务逻辑。
	Execute(ctx context.Context) error
}

// Scheduler 定时任务调度器接口。
type Scheduler interface {
	// Start 启动调度器，开始触发定时任务。
	Start(ctx context.Context) error
	// Shutdown 优雅关闭，等待正在执行的任务完成。
	Shutdown(ctx context.Context) error
	// Register 注册定时任务，任务名唯一，重复返回 error。
	Register(task Task) error
	// Unregister 注销定时任务。
	Unregister(name string) bool
	// IsRunning 返回调度器是否正在运行。
	IsRunning() bool
	// RegisteredTasks 返回所有已注册任务。
	RegisteredTasks() []Task
}

// 调度器配置键常量。
const (
	// ScheduleEnabled 调度器启用开关。
	ScheduleEnabled = "schedule.enabled"
	// SchedulePoolSize 调度器线程池大小。
	SchedulePoolSize = "schedule.pool-size"
	// ScheduleScanAnnotations 是否扫描注解注册任务。
	ScheduleScanAnnotations = "schedule.scan-annotations"
)

// 默认值常量。
const (
	// DefaultSchedulePoolSize 默认线程池大小。
	DefaultSchedulePoolSize = 10
	// DefaultScheduleScanAnnotations 默认是否扫描注解。
	DefaultScheduleScanAnnotations = true
	// ConditionTrue 条件判断的真值常量。
	ConditionTrue = "true"
)
