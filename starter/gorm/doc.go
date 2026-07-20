// Package gorm 提供基于 GORM 的数据库访问能力。
//
// 本模块是 enhance 框架的数据库集成模块，基于 gorm.io/gorm 实现，
// 提供 ORM 数据库操作、连接池管理、自动迁移等能力。
//
// 官方文档：https://gorm.io/
//
// # 模块独立性
//
// 本模块采用独立模块设计（拥有独立的 go.mod），依赖隔离确保：
//   - 用户只使用 enhance 核心模块时，不会引入 GORM 依赖
//   - 用户显式引入本模块时，才会下载 GORM 及其间接依赖
//   - 避免依赖污染，保持用户项目的依赖树清晰
//
// # 架构设计
//
// 核心组件：
//   - GormAutoConfiguration: 自动配置类，根据配置文件创建和注册 GORM 实例
//
// 自动配置：
//   - 当 gorm.enabled=true 时自动生效
//   - 自动创建数据库连接并注册 *gorm.DB 到 IoC 容器
//   - 支持连接池配置（最大连接数、空闲连接数、连接生命周期）
//
// # 快速开始
//
// 1. 在配置文件中启用 GORM：
//
//	{
//	  "gorm": {
//	    "enabled": true,
//	    "host": "localhost",
//	    "port": 3306,
//	    "username": "root",
//	    "password": "password",
//	    "database": "mydb",
//	    "charset": "utf8mb4",
//	    "max-open-conns": 100,
//	    "max-idle-conns": 10,
//	    "conn-max-lifetime": 3600
//	  }
//	}
//
// 2. 在代码中使用：
//
//	type UserRepository struct {
//		db *gorm.DB
//	}
//
//	func NewUserRepository(db *gorm.DB) *UserRepository {
//		return &UserRepository{db: db}
//	}
//
// # 配置说明
//
//   - gorm.enabled: 是否启用 GORM（默认 false）
//   - gorm.host: 数据库主机地址
//   - gorm.port: 数据库端口
//   - gorm.username: 数据库用户名
//   - gorm.password: 数据库密码
//   - gorm.database: 数据库名称
//   - gorm.charset: 字符集（默认 utf8mb4）
//   - gorm.max-open-conns: 最大打开连接数（默认 100）
//   - gorm.max-idle-conns: 最大空闲连接数（默认 10）
//   - gorm.conn-max-lifetime: 连接最大生命周期，单位秒（默认 3600）
//
// # 依赖说明
//
// 本模块依赖：
//   - gorm.io/gorm: GORM 核心库
//   - gorm.io/driver/mysql: MySQL 驱动
//
// 用户项目引入本模块后，会自动引入上述依赖。
package gorm

// ==================== 配置键常量 ====================

const (
	// GORM 配置
	GORMEnabled         = "db.gorm.enabled"
	GORMHost            = "db.gorm.host"
	GORMPort            = "db.gorm.port"
	GORMUsername        = "db.gorm.username"
	GORMPassword        = "db.gorm.password"
	GORMDatabase        = "db.gorm.database"
	GORMCharset         = "db.gorm.charset"
	GORMMaxOpenConns    = "db.gorm.max-open-conns"
	GORMMaxIdleConns    = "db.gorm.max-idle-conns"
	GORMConnMaxLifetime = "db.gorm.conn-max-lifetime"

	// 日志字段常量
	LogFieldHost     = "host"
	LogFieldPort     = "port"
	LogFieldDatabase = "database"
)

// ==================== 默认值常量 ====================

const (
	// GORM 默认值
	DefaultGORMDriver          = "mysql"
	DefaultGORMHost            = "localhost"
	DefaultGORMPort            = 3306
	DefaultGORMUsername        = "scott"
	DefaultGORMPassword        = "123456"
	DefaultGORMDatabase        = "demo"
	DefaultGORMCharset         = "utf8mb4"
	DefaultGORMMaxOpenConns    = 100
	DefaultGORMMaxIdleConns    = 10
	DefaultGORMConnMaxLifetime = 3600

	// 条件值常量
	ConditionTrue = "true"
)
