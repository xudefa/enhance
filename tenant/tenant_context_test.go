package tenant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// TestSetTenantToContext_RoundTrip 验证通过 context 存取租户。
func TestSetTenantToContext_RoundTrip(t *testing.T) {
	t.Parallel()
	tenant := &Tenant{ID: "tenant-1", Name: "Tenant 1"}

	ctx := SetTenantToContext(context.Background(), tenant)
	got, ok := TenantFromContext(ctx)
	if !ok {
		t.Fatal("expected tenant in context")
	}
	if got != tenant {
		t.Errorf("expected same tenant pointer, got %v", got)
	}
}

// TestTenantContext_Isolation 验证基于 context 的租户上下文按请求隔离（回归测试）。
//
// 背景：SetCurrentTenant/GetCurrentTenant 是进程级共享状态，并发请求会互相覆盖，
// 导致租户串扰。推荐使用 context 传递租户，每个请求拥有独立的租户上下文。
func TestTenantContext_Isolation(t *testing.T) {
	t.Parallel()
	manager := NewTenantManager(nil)
	manager.RegisterTenant(&Tenant{ID: "tenant-1", Name: "T1"})
	manager.RegisterTenant(&Tenant{ID: "tenant-2", Name: "T2"})

	var wg sync.WaitGroup
	for _, id := range []string{"tenant-1", "tenant-2"} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			for k := 0; k < 100; k++ {
				tnt, err := manager.GetTenant(id)
				if err != nil {
					t.Errorf("GetTenant(%s): %v", id, err)
					return
				}
				ctx := SetTenantToContext(context.Background(), tnt)
				got, ok := TenantFromContext(ctx)
				if !ok || got.ID != id {
					t.Errorf("上下文中的租户发生串扰: got %v, want %s", got, id)
					return
				}
			}
		}(id)
	}
	wg.Wait()
}

// TestTenantMiddleware_ConcurrentRequests 并发请求通过中间件时各自获取正确的租户。
//
// 中间件内部使用 context 传递租户，验证其不依赖进程级 currentTenant。
func TestTenantMiddleware_ConcurrentRequests(t *testing.T) {
	t.Parallel()
	resolver := NewHeaderResolver("X-Tenant-ID")
	manager := NewTenantManager(resolver)
	manager.RegisterTenant(&Tenant{ID: "tenant-1", Name: "T1", Enabled: true})
	manager.RegisterTenant(&Tenant{ID: "tenant-2", Name: "T2", Enabled: true})
	middleware := NewTenantMiddleware(manager)

	var wg sync.WaitGroup
	for _, id := range []string{"tenant-1", "tenant-2"} {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tenant, ok := TenantFromContext(r.Context())
				if !ok {
					t.Errorf("请求 %s 的 context 中没有租户", id)
					return
				}
				if tenant.ID != id {
					t.Errorf("请求 %s 获取到租户 %s（串扰）", id, tenant.ID)
				}
			})
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-Tenant-ID", id)
			middleware.Handle(handler).ServeHTTP(httptest.NewRecorder(), req)
		}(id)
	}
	wg.Wait()
}

// TestTenantManager_ConcurrentCurrentTenant 并发读写进程级 currentTenant 不产生数据竞争（-race 验证）。
func TestTenantManager_ConcurrentCurrentTenant(t *testing.T) {
	t.Parallel()
	manager := NewTenantManager(nil)
	manager.RegisterTenant(&Tenant{ID: "tenant-1", Name: "T1"})
	manager.RegisterTenant(&Tenant{ID: "tenant-2", Name: "T2"})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = manager.SetCurrentTenant("tenant-1")
		}()
		go func() {
			defer wg.Done()
			cur := manager.GetCurrentTenant()
			if cur != nil && cur.ID != "tenant-1" && cur.ID != "tenant-2" {
				t.Errorf("unexpected tenant: %v", cur.ID)
			}
		}()
	}
	wg.Wait()
	manager.ClearCurrentTenant()
}
