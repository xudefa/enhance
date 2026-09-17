package tenant

import (
	"net/http/httptest"
	"testing"
)

func TestTenantIsolation(t *testing.T) {
	t.Parallel()
	resolver := NewHeaderResolver("X-Tenant-ID")
	manager := NewTenantManager(resolver)

	manager.RegisterTenant(&Tenant{
		ID:       "tenant-1",
		Name:     "Tenant 1",
		Database: "db_tenant1",
	})

	isolation := NewTenantIsolation(manager)

	// 测试数据库隔离
	db, err := isolation.IsolateDatabase("tenant-1")
	if err != nil {
		t.Fatalf("Failed to isolate database: %v", err)
	}

	if db != "db_tenant1" {
		t.Errorf("expected database 'db_tenant1', got %s", db)
	}

	// 测试模式隔离
	schema, err := isolation.IsolateSchema("tenant-1")
	if err != nil {
		t.Fatalf("Failed to isolate schema: %v", err)
	}

	if schema != "tenant_tenant-1" {
		t.Errorf("expected schema 'tenant_tenant-1', got %s", schema)
	}

	// 测试行级隔离
	rowID := isolation.IsolateRow("tenant-1")
	if rowID != "tenant-1" {
		t.Errorf("expected row ID 'tenant-1', got %s", rowID)
	}

	// 测试没有数据库的租户
	manager.RegisterTenant(&Tenant{
		ID:   "tenant-nodb",
		Name: "Tenant NoDB",
	})

	_, err = isolation.IsolateDatabase("tenant-nodb")
	if err == nil {
		t.Error("expected error when tenant has no database")
	}
}

func TestTenantRegistry(t *testing.T) {
	t.Parallel()
	registry := NewTenantRegistry()

	// 添加租户
	registry.Add(&Tenant{
		ID:   "tenant-1",
		Name: "Tenant 1",
	})

	registry.Add(&Tenant{
		ID:   "tenant-2",
		Name: "Tenant 2",
	})

	if registry.Count() != 2 {
		t.Errorf("expected 2 tenants, got %d", registry.Count())
	}

	// 获取租户
	tenant, err := registry.Get("tenant-1")
	if err != nil {
		t.Fatalf("Failed to get tenant: %v", err)
	}

	if tenant.Name != "Tenant 1" {
		t.Errorf("expected tenant name 'Tenant 1', got %s", tenant.Name)
	}

	// 列出租户
	tenants := registry.List()
	if len(tenants) != 2 {
		t.Errorf("expected 2 tenants in list, got %d", len(tenants))
	}

	// 移除租户
	registry.Remove("tenant-1")

	if registry.Count() != 1 {
		t.Errorf("expected 1 tenant after removal, got %d", registry.Count())
	}

	// 获取不存在的租户
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("expected error when getting nonexistent tenant")
	}
}

func TestTenantProvider(t *testing.T) {
	t.Parallel()
	resolver := NewHeaderResolver("X-Tenant-ID")
	manager := NewTenantManager(resolver)

	manager.RegisterTenant(&Tenant{
		ID:       "tenant-1",
		Name:     "Tenant 1",
		Database: "db_tenant1",
	})

	provider := NewTenantProvider(manager)

	// 测试没有当前租户
	if provider.GetCurrentTenantID() != "" {
		t.Error("expected empty tenant ID when no current tenant")
	}

	if provider.GetCurrentTenantName() != "" {
		t.Error("expected empty tenant name when no current tenant")
	}

	if provider.IsMultiTenant() {
		t.Error("expected not to be in multi-tenant mode")
	}

	// 设置当前租户
	_ = manager.SetCurrentTenant("tenant-1")

	if provider.GetCurrentTenantID() != "tenant-1" {
		t.Errorf("expected tenant ID 'tenant-1', got %s", provider.GetCurrentTenantID())
	}

	if provider.GetCurrentTenantName() != "Tenant 1" {
		t.Errorf("expected tenant name 'Tenant 1', got %s", provider.GetCurrentTenantName())
	}

	db, err := provider.GetCurrentTenantDatabase()
	if err != nil {
		t.Fatalf("Failed to get tenant database: %v", err)
	}

	if db != "db_tenant1" {
		t.Errorf("expected database 'db_tenant1', got %s", db)
	}

	if !provider.IsMultiTenant() {
		t.Error("expected to be in multi-tenant mode")
	}
}

func TestExtractSubdomain(t *testing.T) {
	t.Parallel()
	tests := []struct {
		host        string
		baseDomain  string
		expectedSub string
	}{
		{"tenant1.example.com", "example.com", "tenant1"},
		{"tenant1.example.com:8080", "example.com", "tenant1"},
		{"sub.tenant1.example.com", "example.com", "sub.tenant1"},
		{"example.com", "example.com", ""},
		{"other.com", "example.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			subdomain := extractSubdomain(tt.host, tt.baseDomain)
			if subdomain != tt.expectedSub {
				t.Errorf("extractSubdomain(%q, %q) = %q, expected %q",
					tt.host, tt.baseDomain, subdomain, tt.expectedSub)
			}
		})
	}
}

func TestSplitPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path         string
		expectedSegs []string
	}{
		{"/api/users", []string{"api", "users"}},
		{"/api", []string{"api"}},
		{"/", []string{}},
		{"", []string{}},
		{"/tenant1/api/users", []string{"tenant1", "api", "users"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			segments := splitPath(tt.path)
			if len(segments) != len(tt.expectedSegs) {
				t.Errorf("splitPath(%q) = %v, expected %v", tt.path, segments, tt.expectedSegs)
			}
		})
	}
}

func TestTenantManager_ResolveFromRequest(t *testing.T) {
	t.Parallel()
	resolver := NewHeaderResolver("X-Tenant-ID")
	manager := NewTenantManager(resolver)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Tenant-ID", "tenant-123")

	tenantID, err := manager.ResolveFromRequest(req)
	if err != nil {
		t.Fatalf("Failed to resolve from request: %v", err)
	}

	if tenantID != "tenant-123" {
		t.Errorf("expected tenant ID 'tenant-123', got %s", tenantID)
	}
}

func TestTenant_Enabled(t *testing.T) {
	t.Parallel()
	tenant1 := &Tenant{
		ID:      "tenant-1",
		Name:    "Tenant 1",
		Enabled: true,
	}

	tenant2 := &Tenant{
		ID:      "tenant-2",
		Name:    "Tenant 2",
		Enabled: false,
	}

	if !tenant1.Enabled {
		t.Error("expected tenant1 to be enabled")
	}

	if tenant2.Enabled {
		t.Error("expected tenant2 to be disabled")
	}
}

func TestTenant_Metadata(t *testing.T) {
	t.Parallel()
	tenant := &Tenant{
		ID:   "tenant-1",
		Name: "Tenant 1",
		Metadata: map[string]string{
			"plan":     "premium",
			"maxUsers": "100",
		},
	}

	if tenant.Metadata["plan"] != "premium" {
		t.Errorf("expected plan 'premium', got %s", tenant.Metadata["plan"])
	}

	if tenant.Metadata["maxUsers"] != "100" {
		t.Errorf("expected maxUsers '100', got %s", tenant.Metadata["maxUsers"])
	}
}
