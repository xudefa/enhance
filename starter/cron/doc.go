// Package cron 提供定时任务自动配置。
//
// Cron 是 Go 语言最流行的定时任务库。
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

// ==================== 配置键常量 ====================

const (
	// Cron 配置
	CronEnabled    = "cron.enabled"
	CronWithLogger = "cron.with-logger"

	// 日志字段常量
	LogFieldDetails = "details"
)

// ==================== 默认值常量 ====================

const (
	// Cron 默认值
	DefaultWithLogger = false

	// 条件值常量
	ConditionTrue = "true"
)
