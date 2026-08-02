package exception

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// TestResolverChain_ConcurrentAddResolver 验证并发添加解析器不丢更新（回归测试）。
//
// 背景：AddResolver 在锁外排序，排序结果回写时可能覆盖并发添加的解析器（lost update）。
func TestResolverChain_ConcurrentAddResolver(t *testing.T) {
	t.Parallel()
	chain := NewResolverChain()

	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(order int) {
			defer wg.Done()
			chain.AddResolver(&mockResolver{order: order})
		}(i)
	}
	wg.Wait()

	resolvers := chain.GetResolvers()
	if len(resolvers) != n {
		t.Fatalf("expected %d resolvers, got %d (lost update)", n, len(resolvers))
	}
}

// TestResolverChain_Resolve_NoDeadlockOnReentrant 验证 Resolve 不在持有锁时调用用户代码（回归测试）。
//
// 背景：Resolve 持有读锁时调用用户 Supports/Resolve，若用户代码重入链
// （如调用 AddResolver），RWMutex 不可重入导致死锁。
func TestResolverChain_Resolve_NoDeadlockOnReentrant(t *testing.T) {
	t.Parallel()
	chain := NewResolverChain()
	resolver := &mockResolver{
		order:    1,
		supports: func(err error) bool { return true },
		resolve: func(ctx context.Context, err error) *ErrorResponse {
			chain.AddResolver(&mockResolver{order: 2})
			return NewErrorResponse(200, "ok", "", "", nil)
		},
	}
	chain.AddResolver(resolver)

	done := make(chan struct{})
	go func() {
		chain.Resolve(context.Background(), errors.New("err"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Resolve deadlocked when user code re-entered the chain")
	}
}
