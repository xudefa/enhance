// Package retry 提供独立的重试机制支持。
//
// 从 event/deadletter.go 和 web/server/retry.go 抽取而来，
// 为事件处理、HTTP 请求等场景提供统一的重试策略和执行器。
//
// 核心组件：
//   - RetryPolicy: 重试策略配置（最大次数、退避策略、延迟计算）
//   - Executor: 重试执行器（支持 jitter 防惊群、上下文取消、回调通知）
//   - BackoffStrategy: 退避策略（None/Fixed/Linear/Exponential）
//
// 使用示例：
//
//	policy := retry.ExponentialBackoff(3, 100*time.Millisecond, 10*time.Second)
//	executor := retry.NewExecutor(policy)
//
//	result, err := executor.Execute(ctx, func(ctx context.Context) (any, error) {
//	    return httpClient.Get(ctx, url)
//	})
package retry
