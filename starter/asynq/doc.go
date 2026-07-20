// Package asynq 提供 Asynq 异步任务队列自动配置。
//
// Asynq 是基于 Redis 的异步任务队列库。
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
