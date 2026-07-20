// Package zap 提供 Zap 高性能日志自动配置。
//
// Zap 是 Uber 开源的高性能结构化日志库。
//
// 功能特性：
//   - 自动配置 Zap 日志器
//   - 支持 JSON/Console 格式
//   - 支持日志级别配置
//   - 支持文件输出和日志轮转
//   - 实现 log.Logger 接口
//
// 配置示例：
//
//	{
//	  "log": {
//	    "zap": {
//	      "enabled": true,
//	      "level": "info",
//	      "format": "json",
//	      "output-path": "stdout"
//	    }
//	  }
//	}
//
// 使用示例：
//
//	logger := core.MustGetBean[log.Logger](app.Container())
//	logger.Info(context.Background(), "Application started",
//	    log.KeyValue{Key: "version", Value: "1.0.0"},
//	)
package zap
