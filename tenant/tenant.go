// Package tenant 提供多租户支持，用于 enhance 框架。
package tenant

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

type tenantContextKey struct{}

type jwtClaimsKey struct{}

// headerResolverImpl TenantResolver 接口的基于请求头的实现。
type headerResolverImpl struct {
	headerName string
}

// NewHeaderResolver 创建基于请求头的租户解析器。
func NewHeaderResolver(headerName string) TenantResolver {
	return &headerResolverImpl{
		headerName: headerName,
	}
}

// Resolve 实现 TenantResolver 接口。
func (r *headerResolverImpl) Resolve(req *http.Request) (string, error) {
	tenantID := req.Header.Get(r.headerName)
	if tenantID == "" {
		return "", fmt.Errorf("tenant ID not found in header %s", r.headerName)
	}
	return tenantID, nil
}

// subdomainResolverImpl TenantResolver 接口的基于子域名的实现。
type subdomainResolverImpl struct {
	baseDomain string
}

// NewSubdomainResolver 创建基于子域名的租户解析器。
func NewSubdomainResolver(baseDomain string) TenantResolver {
	return &subdomainResolverImpl{
		baseDomain: baseDomain,
	}
}

// Resolve 实现 TenantResolver 接口。
func (r *subdomainResolverImpl) Resolve(req *http.Request) (string, error) {
	host := req.Host
	if host == "" {
		return "", fmt.Errorf("host is empty")
	}

	// 提取子域名
	subdomain := extractSubdomain(host, r.baseDomain)
	if subdomain == "" {
		return "", fmt.Errorf("subdomain not found for host %s", host)
	}

	return subdomain, nil
}

// jwtResolverImpl TenantResolver 接口的基于 JWT 的实现。
type jwtResolverImpl struct {
	claimName string
}

// NewJWTResolver 创建基于 JWT 的租户解析器。
func NewJWTResolver(claimName string) TenantResolver {
	return &jwtResolverImpl{
		claimName: claimName,
	}
}

// Resolve 实现 TenantResolver 接口。
func (r *jwtResolverImpl) Resolve(req *http.Request) (string, error) {
	// 从 context 中获取 JWT claims
	claims, ok := req.Context().Value(jwtClaimsKey{}).(map[string]any)
	if !ok {
		return "", fmt.Errorf("JWT claims not found in context")
	}

	tenantID, ok := claims[r.claimName].(string)
	if !ok {
		return "", fmt.Errorf("tenant ID not found in JWT claims")
	}

	return tenantID, nil
}

// pathResolverImpl TenantResolver 接口的基于路径的实现。
type pathResolverImpl struct {
	segmentIndex int
}

// NewPathResolver 创建基于路径的租户解析器。
func NewPathResolver(segmentIndex int) TenantResolver {
	return &pathResolverImpl{
		segmentIndex: segmentIndex,
	}
}

// Resolve 实现 TenantResolver 接口。
func (r *pathResolverImpl) Resolve(req *http.Request) (string, error) {
	path := req.URL.Path
	segments := splitPath(path)

	if r.segmentIndex >= len(segments) {
		return "", fmt.Errorf("tenant ID not found in path")
	}

	return segments[r.segmentIndex], nil
}

// tenantManagerImpl TenantManager 接口的默认实现。
type tenantManagerImpl struct {
	mu            sync.RWMutex
	resolver      TenantResolver
	tenants       map[string]*Tenant
	currentTenant *Tenant
}

// NewTenantManager 创建租户管理器。
func NewTenantManager(resolver TenantResolver) TenantManager {
	return &tenantManagerImpl{
		resolver: resolver,
		tenants:  make(map[string]*Tenant),
	}
}

// RegisterTenant 注册租户。
func (m *tenantManagerImpl) RegisterTenant(tenant *Tenant) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tenants[tenant.ID] = tenant
}

// GetTenant 获取租户。
func (m *tenantManagerImpl) GetTenant(tenantID string) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant %s not found", tenantID)
	}

	return tenant, nil
}

// SetCurrentTenant 设置当前租户。
func (m *tenantManagerImpl) SetCurrentTenant(tenantID string) error {
	tenant, err := m.GetTenant(tenantID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentTenant = tenant
	return nil
}

// GetCurrentTenant 获取当前租户。
func (m *tenantManagerImpl) GetCurrentTenant() *Tenant {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.currentTenant
}

// ClearCurrentTenant 清除当前租户。
func (m *tenantManagerImpl) ClearCurrentTenant() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentTenant = nil
}

// ResolveFromRequest 从 HTTP 请求解析租户。
func (m *tenantManagerImpl) ResolveFromRequest(req *http.Request) (string, error) {
	return m.resolver.Resolve(req)
}

// tenantMiddlewareImpl TenantMiddleware 接口的默认实现。
type tenantMiddlewareImpl struct {
	manager TenantManager
}

// NewTenantMiddleware 创建租户中间件。
func NewTenantMiddleware(manager TenantManager) TenantMiddleware {
	return &tenantMiddlewareImpl{
		manager: manager,
	}
}

// Handle 处理 HTTP 请求。
func (m *tenantMiddlewareImpl) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 解析租户
		tenantID, err := m.manager.ResolveFromRequest(r)
		if err != nil {
			http.Error(w, "Tenant ID required", http.StatusBadRequest)
			return
		}

		// 获取租户信息
		tenant, err := m.manager.GetTenant(tenantID)
		if err != nil {
			http.Error(w, "Invalid tenant", http.StatusForbidden)
			return
		}

		// 检查租户是否启用
		if !tenant.Enabled {
			http.Error(w, "Tenant is disabled", http.StatusForbidden)
			return
		}

		// 设置当前租户
		_ = m.manager.SetCurrentTenant(tenantID)
		defer m.manager.ClearCurrentTenant()

		// 将租户信息添加到 context
		ctx := context.WithValue(r.Context(), tenantContextKey{}, tenant)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TenantFromContext 从 context 获取租户。
func TenantFromContext(ctx context.Context) (*Tenant, bool) {
	tenant, ok := ctx.Value(tenantContextKey{}).(*Tenant)
	return tenant, ok
}

// tenantIsolationImpl TenantIsolation 接口的默认实现。
type tenantIsolationImpl struct {
	manager TenantManager
}

// NewTenantIsolation 创建租户隔离器。
func NewTenantIsolation(manager TenantManager) TenantIsolation {
	return &tenantIsolationImpl{
		manager: manager,
	}
}

// IsolateDatabase 数据库隔离。
func (i *tenantIsolationImpl) IsolateDatabase(tenantID string) (string, error) {
	tenant, err := i.manager.GetTenant(tenantID)
	if err != nil {
		return "", err
	}

	if tenant.Database == "" {
		return "", fmt.Errorf("tenant %s has no database configured", tenantID)
	}

	return tenant.Database, nil
}

// IsolateSchema 模式隔离。
func (i *tenantIsolationImpl) IsolateSchema(tenantID string) (string, error) {
	_, err := i.manager.GetTenant(tenantID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tenant_%s", tenantID), nil
}

// IsolateRow 行级隔离。
func (i *tenantIsolationImpl) IsolateRow(tenantID string) string {
	return tenantID
}

// tenantRegistryImpl TenantRegistry 接口的默认实现。
type tenantRegistryImpl struct {
	mu      sync.RWMutex
	tenants map[string]*Tenant
}

// NewTenantRegistry 创建租户注册表。
func NewTenantRegistry() TenantRegistry {
	return &tenantRegistryImpl{
		tenants: make(map[string]*Tenant),
	}
}

// Add 添加租户。
func (r *tenantRegistryImpl) Add(tenant *Tenant) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tenants[tenant.ID] = tenant
}

// Remove 移除租户。
func (r *tenantRegistryImpl) Remove(tenantID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.tenants, tenantID)
}

// Get 获取租户。
func (r *tenantRegistryImpl) Get(tenantID string) (*Tenant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenant, exists := r.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant %s not found", tenantID)
	}

	return tenant, nil
}

// List 列出所有租户。
func (r *tenantRegistryImpl) List() []*Tenant {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tenants := make([]*Tenant, 0, len(r.tenants))
	for _, tenant := range r.tenants {
		tenants = append(tenants, tenant)
	}

	return tenants
}

// Count 获取租户数量。
func (r *tenantRegistryImpl) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.tenants)
}

// tenantProviderImpl TenantProvider 接口的默认实现。
type tenantProviderImpl struct {
	manager TenantManager
}

// NewTenantProvider 创建租户提供者。
func NewTenantProvider(manager TenantManager) TenantProvider {
	return &tenantProviderImpl{
		manager: manager,
	}
}

// GetCurrentTenantID 获取当前租户 ID。
func (p *tenantProviderImpl) GetCurrentTenantID() string {
	tenant := p.manager.GetCurrentTenant()
	if tenant == nil {
		return ""
	}
	return tenant.ID
}

// GetCurrentTenantName 获取当前租户名称。
func (p *tenantProviderImpl) GetCurrentTenantName() string {
	tenant := p.manager.GetCurrentTenant()
	if tenant == nil {
		return ""
	}
	return tenant.Name
}

// GetCurrentTenantDatabase 获取当前租户数据库。
func (p *tenantProviderImpl) GetCurrentTenantDatabase() (string, error) {
	tenant := p.manager.GetCurrentTenant()
	if tenant == nil {
		return "", fmt.Errorf("no current tenant")
	}

	if tenant.Database == "" {
		return "", fmt.Errorf("tenant has no database configured")
	}

	return tenant.Database, nil
}

// IsMultiTenant 检查是否为多租户模式。
func (p *tenantProviderImpl) IsMultiTenant() bool {
	return p.manager.GetCurrentTenant() != nil
}

// extractSubdomain 提取子域名。
func extractSubdomain(host, baseDomain string) string {
	// 移除端口号
	if idx := indexOf(host, ':'); idx != -1 {
		host = host[:idx]
	}

	// 检查是否以 baseDomain 结尾
	if !endsWith(host, "."+baseDomain) {
		return ""
	}

	// 提取子域名
	subdomain := host[:len(host)-len("."+baseDomain)]
	return subdomain
}

// splitPath 分割路径。
func splitPath(path string) []string {
	// 移除开头的 /
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	if path == "" {
		return []string{}
	}

	segments := make([]string, 0)
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if start < i {
				segments = append(segments, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		segments = append(segments, path[start:])
	}

	return segments
}

// indexOf 查找字符位置。
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// endsWith 检查是否以指定字符串结尾。
func endsWith(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
