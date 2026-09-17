package tenant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeaderResolver(t *testing.T) {
	t.Parallel()
	resolver := NewHeaderResolver("X-Tenant-ID")

	// 测试正常情况
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Tenant-ID", "tenant-123")

	tenantID, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("Failed to resolve tenant: %v", err)
	}

	if tenantID != "tenant-123" {
		t.Errorf("expected tenant ID 'tenant-123', got %s", tenantID)
	}

	// 测试缺少请求头
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	_, err = resolver.Resolve(req2)
	if err == nil {
		t.Error("expected error when header is missing")
	}
}

func TestSubdomainResolver(t *testing.T) {
	t.Parallel()
	resolver := NewSubdomainResolver("example.com")

	// 测试正常情况
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Host = "tenant1.example.com"

	tenantID, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("Failed to resolve tenant: %v", err)
	}

	if tenantID != "tenant1" {
		t.Errorf("expected tenant ID 'tenant1', got %s", tenantID)
	}

	// 测试带端口的情况
	req2 := httptest.NewRequest("GET", "/api/test", nil)
	req2.Host = "tenant2.example.com:8080"

	tenantID2, err := resolver.Resolve(req2)
	if err != nil {
		t.Fatalf("Failed to resolve tenant: %v", err)
	}

	if tenantID2 != "tenant2" {
		t.Errorf("expected tenant ID 'tenant2', got %s", tenantID2)
	}

	// 测试不匹配的基础域名
	req3 := httptest.NewRequest("GET", "/api/test", nil)
	req3.Host = "tenant1.other.com"

	_, err = resolver.Resolve(req3)
	if err == nil {
		t.Error("expected error when base domain doesn't match")
	}
}

func TestPathResolver(t *testing.T) {
	t.Parallel()
	resolver := NewPathResolver(0)

	// 测试正常情况
	req := httptest.NewRequest("GET", "/tenant1/api/test", nil)

	tenantID, err := resolver.Resolve(req)
	if err != nil {
		t.Fatalf("Failed to resolve tenant: %v", err)
	}

	if tenantID != "tenant1" {
		t.Errorf("expected tenant ID 'tenant1', got %s", tenantID)
	}

	// 测试第二个段
	resolver2 := NewPathResolver(1)
	req2 := httptest.NewRequest("GET", "/api/tenant1/users", nil)

	tenantID2, err := resolver2.Resolve(req2)
	if err != nil {
		t.Fatalf("Failed to resolve tenant: %v", err)
	}

	if tenantID2 != "tenant1" {
		t.Errorf("expected tenant ID 'tenant1', got %s", tenantID2)
	}

	// 测试路径段不足
	req3 := httptest.NewRequest("GET", "/api", nil)
	_, err = resolver2.Resolve(req3)
	if err == nil {
		t.Error("expected error when path segment is missing")
	}
}

func TestTenantManager(t *testing.T) {
	t.Parallel()
	resolver := NewHeaderResolver("X-Tenant-ID")
	manager := NewTenantManager(resolver)

	// 注册租户
	tenant1 := &Tenant{
		ID:       "tenant-1",
		Name:     "Tenant 1",
		Database: "db_tenant1",
		Enabled:  true,
	}
	manager.RegisterTenant(tenant1)

	// 获取租户
	tenant, err := manager.GetTenant("tenant-1")
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	if tenant.Name != "Tenant 1" {
		t.Errorf("expected tenant name 'Tenant 1', got %s", tenant.Name)
	}

	// 设置当前租户
	err = manager.SetCurrentTenant("tenant-1")
	if err != nil {
		t.Fatalf("Failed to set current tenant: %v", err)
	}

	current := manager.GetCurrentTenant()
	if current == nil || current.ID != "tenant-1" {
		t.Fatalf("expected current tenant ID 'tenant-1', got %+v", current)
	}

	// 清除当前租户
	manager.ClearCurrentTenant()
	if manager.GetCurrentTenant() != nil {
		t.Error("expected current tenant to be cleared")
	}

	// 获取不存在的租户
	_, err = manager.GetTenant("nonexistent")
	if err == nil {
		t.Error("expected error when getting nonexistent tenant")
	}
}

func TestTenantMiddleware(t *testing.T) {
	t.Parallel()
	resolver := NewHeaderResolver("X-Tenant-ID")
	manager := NewTenantManager(resolver)

	// 注册租户
	manager.RegisterTenant(&Tenant{
		ID:      "tenant-1",
		Name:    "Tenant 1",
		Enabled: true,
	})

	middleware := NewTenantMiddleware(manager)

	// 测试正常请求
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant, ok := TenantFromContext(r.Context())
		if !ok {
			t.Error("expected tenant in context")
		}

		if tenant.ID != "tenant-1" {
			t.Errorf("expected tenant ID 'tenant-1', got %s", tenant.ID)
		}

		w.WriteHeader(http.StatusOK)
	})

	if code := testTenantMiddlewareServe(t, middleware, handler, "tenant-1"); code != http.StatusOK {
		t.Errorf("expected status 200, got %d", code)
	}

	// 测试缺少租户 ID
	if code := testTenantMiddlewareServe(t, middleware, handler, ""); code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", code)
	}

	// 测试禁用的租户
	manager.RegisterTenant(&Tenant{
		ID:      "tenant-disabled",
		Name:    "Disabled Tenant",
		Enabled: false,
	})

	if code := testTenantMiddlewareServe(t, middleware, handler, "tenant-disabled"); code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", code)
	}
}

func testTenantMiddlewareServe(t *testing.T, middleware TenantMiddleware, handler http.Handler, tenantID string) int {
	req := httptest.NewRequest("GET", "/api/test", nil)
	if tenantID != "" {
		req.Header.Set("X-Tenant-ID", tenantID)
	}
	rr := httptest.NewRecorder()

	middleware.Handle(handler).ServeHTTP(rr, req)
	return rr.Code
}

func TestTenantFromContext(t *testing.T) {
	t.Parallel()
	tenant := &Tenant{
		ID:   "tenant-1",
		Name: "Tenant 1",
	}

	ctx := context.WithValue(context.Background(), tenantContextKey{}, tenant)

	retrieved, ok := TenantFromContext(ctx)
	if !ok {
		t.Fatal("expected to retrieve tenant from context")
	}

	if retrieved.ID != "tenant-1" {
		t.Errorf("expected tenant ID 'tenant-1', got %s", retrieved.ID)
	}

	// 测试没有租户的 context
	ctx2 := context.Background()
	_, ok = TenantFromContext(ctx2)
	if ok {
		t.Error("expected not to retrieve tenant from context without tenant")
	}
}
