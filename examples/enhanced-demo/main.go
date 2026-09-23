// Package main 演示 enhance 框架 P0/P1/P2 优化后的新特性。
//
// 本示例展示：
//   - P0: 约定优于配置（无需配置即可启用功能）
//   - P0: 开发者体验优化（友好的启动报告）
//   - P1: 模块化组合 API（DSL 风格模块定义）
//   - P2: 插件系统（独立生命周期和依赖管理）
package main

import (
	"fmt"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
)

// ==================== 全局变量 ====================

// 示例：使用 DSL 风格定义模块
var DatabaseModule = boot.ModuleDef("database").
	DependsOn("config").
	Beans(
		boot.Provide(func(c core.Container) (*Database, error) {
			return NewDatabase("localhost:5432"), nil
		}),
		boot.Provide(func(c core.Container) (*UserRepository, error) {
			db, err := core.GetByName[*Database](c, "")
			if err != nil {
				return nil, fmt.Errorf("failed to get Database for UserRepository: %w", err)
			}
			return NewUserRepository(db), nil
		}),
	).
	Invoke(func(c core.Container) error {
		db, err := core.GetByName[*Database](c, "")
		if err != nil {
			return fmt.Errorf("failed to get Database for migration: %w", err)
		}
		return db.Migrate()
	}).
	Build()

// WebModule Web 模块定义，提供 Web 服务相关 Bean。
var WebModule = boot.ModuleDef("web").
	DependsOn("database").
	Beans(
		boot.Provide(func(c core.Container) (*UserService, error) {
			repo, err := core.GetByName[*UserRepository](c, "")
			if err != nil {
				return nil, fmt.Errorf("failed to get UserRepository for UserService: %w", err)
			}
			return NewUserService(repo), nil
		}),
	).
	Build()

// 示例：使用模块分组组织相关模块
var BackendModules = boot.ModuleGroup("backend").
	Include(DatabaseModule, WebModule).
	Build()

// ==================== 类型定义 ====================

// 模拟类型
type Database struct{ url string }

// UserRepository 用户仓储，负责用户数据的持久化操作。
type UserRepository struct{ db *Database }

// UserService 用户服务，提供用户业务逻辑处理。
type UserService struct{ repo *UserRepository }

// 示例：自定义插件
type MonitoringPlugin struct {
	enabled bool
}

// ==================== 函数定义 ====================

// Database 相关函数
func NewDatabase(url string) *Database { return &Database{url: url} }

// Migrate 执行数据库迁移操作。
func (d *Database) Migrate() error {
	fmt.Println("  [database] 执行数据库迁移...")
	return nil
}

// UserRepository 相关函数
func NewUserRepository(db *Database) *UserRepository {
	return &UserRepository{db: db}
}

// UserService 相关函数
func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

// MonitoringPlugin 接口实现
// Name 返回插件名称。
func (p *MonitoringPlugin) Name() string { return "monitoring" }

// Version 返回插件版本。
func (p *MonitoringPlugin) Version() string { return "1.0.0" }

// Dependencies 返回插件依赖列表。
func (p *MonitoringPlugin) Dependencies() []string { return []string{"database", "web"} }

// Init 初始化监控插件。
func (p *MonitoringPlugin) Init(ctx boot.PluginContext) error {
	fmt.Println("  [monitoring] 初始化监控系统...")
	p.enabled = true
	return nil
}

// Start 启动监控插件。
func (p *MonitoringPlugin) Start(ctx boot.PluginContext) error {
	fmt.Println("  [monitoring] 启动监控系统...")
	return nil
}

// Stop 停止监控插件。
func (p *MonitoringPlugin) Stop(ctx boot.PluginContext) error {
	fmt.Println("  [monitoring] 停止监控系统...")
	p.enabled = false
	return nil
}

// ==================== 主函数 ====================

func main() {
	// 示例 1: 使用新 API 创建应用
	// 约定优于配置：无需显式配置 gin.enabled=true 等
	// 启动报告自动打印，显示 Bean 数量、Starter 列表等
	app, err := boot.NewApplication(
		boot.WithAppName("enhanced-demo"),
		boot.WithVersion("2.0.0"),

		// P1: 使用模块化组合 API
		boot.WithModules(BackendModules),

		// P2: 使用插件系统
		boot.WithPlugin(&MonitoringPlugin{}),

		// 可选：禁用启动报告
		// boot.WithoutStartupReport(),
	)
	if err != nil {
		panic(err)
	}

	// 启动应用
	if err := app.Start(); err != nil {
		panic(err)
	}

	// 停止应用
	defer func() {
		if err := app.Stop(); err != nil {
			fmt.Printf("应用停止失败: %v\n", err)
		}
	}()

	fmt.Println("\n应用已启动，按 Ctrl+C 停止...")
	select {} // 阻塞
}
