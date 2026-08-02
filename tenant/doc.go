// Package tenant 提供多租户支持，用于 enhance 框架。
//
// 该模块提供多租户架构支持，包括租户上下文管理、租户隔离、租户数据源切换等功能。
// 参考 Saas 多租户架构的设计理念。
//
// # 架构设计
//
//   - Tenant: 租户结构体，包含租户基本信息
//   - TenantResolver: 租户解析器接口，从请求中解析租户
//   - TenantManager: 租户管理器接口，管理租户生命周期
//   - TenantMiddleware: 租户中间件接口，自动设置租户上下文
//   - TenantIsolation: 租户隔离器接口，提供数据隔离功能
//   - TenantRegistry: 租户注册表接口，管理所有租户
//   - TenantProvider: 租户提供者接口，提供便捷访问方法
//
// # 核心功能
//
//   - 租户上下文: 提供线程安全的租户上下文管理
//   - 租户解析: 支持从域名、请求头、URL 参数等解析租户
//   - 租户隔离: 支持数据级和逻辑级的租户隔离
//   - 数据源切换: 支持按租户动态切换数据源
//
// # 使用方式
//
// 创建租户管理器：
//
//	resolver := tenant.NewHeaderResolver("X-Tenant-ID")
//	manager := tenant.NewTenantManager(resolver)
//
// 注册租户：
//
//	manager.RegisterTenant(&tenant.Tenant{
//	    ID:       "tenant-1",
//	    Name:     "租户 1",
//	    Database: "tenant_1_db",
//	    Enabled:  true,
//	})
//
// 使用租户中间件：
//
//	middleware := tenant.NewTenantMiddleware(manager)
//	handler := middleware.Handle(nextHandler)
//
// 从 Context 获取租户：
//
//	tenant, ok := tenant.TenantFromContext(ctx)
//
// # 多租户架构模式
//
//   - 数据库级别隔离：每个租户独立数据库
//   - Schema 级别隔离：每个租户独立 Schema
//   - 数据级别隔离：所有租户共享数据库，通过 tenant_id 字段隔离
package tenant

import (
	"net/http"
)

// Tenant 租户结构体。
//
// 包含租户的基本信息和配置。
type Tenant struct {
	// ID 租户 ID。
	ID string
	// Name 租户名称。
	Name string
	// Domain 租户域名。
	Domain string
	// Database 租户数据库。
	Database string
	// Enabled 是否启用。
	Enabled bool
	// Metadata 租户元数据。
	Metadata map[string]string
}

// TenantResolver 租户解析器接口。
//
// 从 HTTP 请求中解析租户 ID。
// 支持多种解析策略：请求头、子域名、JWT、路径等。
type TenantResolver interface {
	// Resolve 解析租户 ID。
	Resolve(req *http.Request) (string, error)
}

// TenantManager 租户管理器接口。
//
// 管理租户的注册、查询和当前租户上下文。
type TenantManager interface {
	// RegisterTenant 注册租户。
	RegisterTenant(tenant *Tenant)

	// GetTenant 获取租户。
	GetTenant(tenantID string) (*Tenant, error)

	// SetCurrentTenant 设置当前租户。
	//
	// 注意：当前租户是进程级共享状态，并发请求会互相覆盖。
	// 多请求场景请使用 context（SetTenantToContext / TenantFromContext）传递租户。
	SetCurrentTenant(tenantID string) error

	// GetCurrentTenant 获取当前租户。
	//
	// 注意：当前租户是进程级共享状态，并发请求之间可能互相覆盖。
	// 请优先使用 TenantFromContext 从请求 context 中获取租户。
	GetCurrentTenant() *Tenant

	// ClearCurrentTenant 清除当前租户。
	ClearCurrentTenant()

	// ResolveFromRequest 从 HTTP 请求解析租户。
	ResolveFromRequest(req *http.Request) (string, error)
}

// TenantMiddleware 租户中间件接口。
//
// 自动从请求中解析租户并设置上下文。
type TenantMiddleware interface {
	// Handle 处理 HTTP 请求。
	Handle(next http.Handler) http.Handler
}

// TenantIsolation 租户隔离器接口。
//
// 提供租户数据隔离功能，支持数据库级、Schema 级和行级隔离。
type TenantIsolation interface {
	// IsolateDatabase 数据库隔离。
	IsolateDatabase(tenantID string) (string, error)

	// IsolateSchema 模式隔离。
	IsolateSchema(tenantID string) (string, error)

	// IsolateRow 行级隔离。
	IsolateRow(tenantID string) string
}

// TenantRegistry 租户注册表接口。
//
// 管理所有租户的注册和查询。
type TenantRegistry interface {
	// Add 添加租户。
	Add(tenant *Tenant)

	// Remove 移除租户。
	Remove(tenantID string)

	// Get 获取租户。
	Get(tenantID string) (*Tenant, error)

	// List 列出所有租户。
	List() []*Tenant

	// Count 获取租户数量。
	Count() int
}

// TenantProvider 租户提供者接口。
//
// 提供获取当前租户的便捷方法。
type TenantProvider interface {
	// GetCurrentTenantID 获取当前租户 ID。
	GetCurrentTenantID() string

	// GetCurrentTenantName 获取当前租户名称。
	GetCurrentTenantName() string

	// GetCurrentTenantDatabase 获取当前租户数据库。
	GetCurrentTenantDatabase() (string, error)

	// IsMultiTenant 检查是否为多租户模式。
	IsMultiTenant() bool
}
