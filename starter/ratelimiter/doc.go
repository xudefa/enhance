// Package ratelimiter 提供限流器自动配置。
//
// RateLimiter 是 Go 标准库提供的令牌桶限流器。
//
// 功能特性：
//   - 自动配置限流器
//   - 支持令牌桶算法
//   - 支持可配置速率和突发
//   - 支持阻塞和非阻塞模式
//
// 配置示例：
//
//	{
//	  "ratelimiter": {
//	    "enabled": true,
//	    "rate": 10.0,
//	    "burst": 20
//	  }
//	}
//
// 使用示例：
//
//	limiter := core.MustGetBean[*rate.Limiter](app.Container())
//	if !limiter.Allow() {
//	    return errors.New("请求过于频繁")
//	}
package ratelimiter
