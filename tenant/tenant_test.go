package tenant

import (
	"context"
	"errors"
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

// TestSubdomainResolver_EmptyHost_Coverage 测试空 Host 时的 SubdomainResolver
func TestSubdomainResolver_EmptyHost_Coverage(t *testing.T) {
	t.Parallel()
	resolver := NewSubdomainResolver("example.com")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = ""

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error for empty host")
	}
}

// TestJWTClaims_ContextRoundTrip_Coverage 测试 JWT Claims 的上下文存储和提取
func TestJWTClaims_ContextRoundTrip_Coverage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(ctx context.Context) context.Context
		wantOK  bool
		wantTID string
	}{
		{
			name: "claims stored and retrieved",
			setup: func(ctx context.Context) context.Context {
				return SetJWTClaims(ctx, map[string]any{"tid": "tenant-1"})
			},
			wantOK:  true,
			wantTID: "tenant-1",
		},
		{
			name:  "no claims in context",
			setup: func(ctx context.Context) context.Context { return ctx },
		},
		{
			name: "non-map value under claims key",
			setup: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, jwtClaimsKey{}, "not-a-map")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			claims, ok := ExtractJWTClaims(tt.setup(context.Background()))
			if ok != tt.wantOK {
				t.Fatalf("expected ok = %v, got %v", tt.wantOK, ok)
			}
			if !tt.wantOK {
				return
			}
			if got := claims["tid"]; got != tt.wantTID {
				t.Errorf("expected tid %q, got %v", tt.wantTID, got)
			}
		})
	}
}

// TestNewJWTExtractor_Coverage 测试创建 JWTExtractor
func TestNewJWTExtractor_Coverage(t *testing.T) {
	t.Parallel()
	parse := func(authHeader string) (map[string]any, error) {
		return map[string]any{"sub": authHeader}, nil
	}

	extractor := NewJWTExtractor(parse)
	if extractor == nil {
		t.Fatal("expected non-nil extractor")
	}
	if extractor.parse == nil {
		t.Fatal("expected parse function to be set")
	}
}

// TestJWTExtractor_Handle_Coverage 测试 JWTExtractor.Handle
func TestJWTExtractor_Handle_Coverage(t *testing.T) {
	t.Parallel()
	tests := testJWTExtractorHandleCases()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			extractor := NewJWTExtractor(tt.parse)

			var gotTID string
			handler := extractor.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims, _ := ExtractJWTClaims(r.Context())
				gotTID, _ = claims["tid"].(string)
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
			if tt.wantTID != "" && gotTID != tt.wantTID {
				t.Errorf("expected tid %q in downstream handler, got %q", tt.wantTID, gotTID)
			}
		})
	}
}

type jwtExtractorHandleCase struct {
	name       string
	authHeader string
	parse      func(string) (map[string]any, error)
	wantStatus int
	wantTID    string
}

func testJWTExtractorHandleCases() []jwtExtractorHandleCase {
	return []jwtExtractorHandleCase{
		{
			name:       "missing authorization header returns 401",
			parse:      func(string) (map[string]any, error) { return nil, nil },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "parse failure returns 401",
			authHeader: "Bearer invalid",
			parse:      func(string) (map[string]any, error) { return nil, errors.New("invalid token") },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid token passes claims downstream",
			authHeader: "Bearer valid",
			parse: func(string) (map[string]any, error) {
				return map[string]any{"tid": "tenant-9"}, nil
			},
			wantStatus: http.StatusOK,
			wantTID:    "tenant-9",
		},
	}
}

// TestJWTResolver_Resolve_Coverage 测试 JWTResolver.Resolve
func TestJWTResolver_Resolve_Coverage(t *testing.T) {
	t.Parallel()
	tests := testJWTResolverResolveCases()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resolver := NewJWTResolver(tt.claimName)
			if resolver == nil {
				t.Fatal("expected non-nil resolver")
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.hasClaims {
				req = req.WithContext(SetJWTClaims(req.Context(), tt.claims))
			}

			got, err := resolver.Resolve(req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err presence = %v, got %v", tt.wantErr, err)
			}
			if !tt.wantErr && got != tt.wantID {
				t.Errorf("expected tenant ID %q, got %q", tt.wantID, got)
			}
		})
	}
}

type jwtResolverResolveCase struct {
	name      string
	hasClaims bool
	claims    map[string]any
	claimName string
	wantID    string
	wantErr   bool
}

func testJWTResolverResolveCases() []jwtResolverResolveCase {
	return []jwtResolverResolveCase{
		{
			name:      "no claims in request context",
			claimName: "tid",
			wantErr:   true,
		},
		{
			name:      "claim key missing from claims map",
			hasClaims: true,
			claims:    map[string]any{"other": "x"},
			claimName: "tid",
			wantErr:   true,
		},
		{
			name:      "claim value is not a string",
			hasClaims: true,
			claims:    map[string]any{"tid": 42},
			claimName: "tid",
			wantErr:   true,
		},
		{
			name:      "claim resolved successfully",
			hasClaims: true,
			claims:    map[string]any{"tid": "tenant-1"},
			claimName: "tid",
			wantID:    "tenant-1",
		},
	}
}

// TestTenantManager_SetCurrentTenant_NotFound_Coverage 测试设置不存在的当前租户
func TestTenantManager_SetCurrentTenant_NotFound_Coverage(t *testing.T) {
	t.Parallel()
	manager := NewTenantManager(NewHeaderResolver("X-Tenant-ID"))

	if err := manager.SetCurrentTenant("ghost"); err == nil {
		t.Fatal("expected error when setting unregistered current tenant")
	}
	if manager.GetCurrentTenant() != nil {
		t.Error("expected no current tenant after failed set")
	}
}

// TestTenantMiddleware_UnknownTenant_Coverage 测试未知租户中间件
func TestTenantMiddleware_UnknownTenant_Coverage(t *testing.T) {
	t.Parallel()
	manager := NewTenantManager(NewHeaderResolver("X-Tenant-ID"))
	middleware := NewTenantMiddleware(manager)

	called := false
	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "ghost")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for unknown tenant, got %d", rr.Code)
	}
	if called {
		t.Error("expected downstream handler not to be invoked")
	}
}

// TestTenantIsolation_UnknownTenantErrors_Coverage 测试未知租户隔离错误
func TestTenantIsolation_UnknownTenantErrors_Coverage(t *testing.T) {
	t.Parallel()
	isolation := NewTenantIsolation(NewTenantManager(NewHeaderResolver("X-Tenant-ID")))

	tests := []struct {
		name string
		act  func(TenantIsolation, string) (string, error)
	}{
		{"IsolateDatabase", TenantIsolation.IsolateDatabase},
		{"IsolateSchema", TenantIsolation.IsolateSchema},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.act(isolation, "ghost")
			if err == nil {
				t.Fatalf("expected error for unknown tenant via %s", tt.name)
			}
		})
	}
}

// TestTenantProvider_GetCurrentTenantDatabase_Errors_Coverage 测试 GetCurrentTenantDatabase 错误处理
func TestTenantProvider_GetCurrentTenantDatabase_Errors_Coverage(t *testing.T) {
	t.Parallel()
	tests := testTenantProviderGetCurrentTenantDatabaseErrorCases()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			manager := NewTenantManager(NewHeaderResolver("X-Tenant-ID"))
			tenant := &Tenant{ID: "tenant-1", Name: "Tenant 1"}
			if tt.registerDB {
				tenant.Database = "db_tenant1"
			}
			manager.RegisterTenant(tenant)
			if tt.setCurrent {
				if err := manager.SetCurrentTenant("tenant-1"); err != nil {
					t.Fatalf("failed to set current tenant: %v", err)
				}
			}

			provider := NewTenantProvider(manager)
			db, err := provider.GetCurrentTenantDatabase()

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err presence = %v, got %v", tt.wantErr, err)
			}
			if !tt.wantErr && db != tt.wantDatabase {
				t.Errorf("expected database %q, got %q", tt.wantDatabase, db)
			}
		})
	}
}

type tenantProviderGetCurrentTenantDatabaseCase struct {
	name         string
	registerDB   bool
	setCurrent   bool
	wantErr      bool
	wantDatabase string
}

func testTenantProviderGetCurrentTenantDatabaseErrorCases() []tenantProviderGetCurrentTenantDatabaseCase {
	return []tenantProviderGetCurrentTenantDatabaseCase{
		{
			name:    "no current tenant",
			wantErr: true,
		},
		{
			name:       "current tenant without database",
			registerDB: false,
			setCurrent: true,
			wantErr:    true,
		},
		{
			name:         "current tenant with database",
			registerDB:   true,
			setCurrent:   true,
			wantDatabase: "db_tenant1",
		},
	}
}

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
