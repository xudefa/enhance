// Package xorm 提供基于 XORM 的数据库访问能力。
//
// 本模块是 enhance 框架的数据库集成模块，基于 xorm.io/xorm 实现，
// 提供 ORM 数据库操作、连接池管理、自动迁移等能力。
//
// 官方文档：https://xorm.io/
//
// # 模块独立性
//
// 本模块采用独立模块设计（拥有独立的 go.mod），依赖隔离确保：
//   - 用户只使用 enhance 核心模块时，不会引入 XORM 依赖
//   - 用户显式引入本模块时，才会下载 XORM 及其间接依赖
//   - 避免依赖污染，保持用户项目的依赖树清晰
//
// # 架构设计
//
// 核心组件：
//   - XormAutoConfiguration: 自动配置类，根据配置文件创建和注册 XORM 实例
//
// 自动配置：
//   - 当 xorm.enabled=true 时自动生效
//   - 自动创建数据库连接并注册 *xorm.Engine 到 IoC 容器
//   - 支持连接池配置（最大连接数、空闲连接数、连接生命周期）
//
// # 快速开始
//
// 1. 在配置文件中启用 XORM：
//
//	{
//	  "db": {
//	    "xorm": {
//	      "enabled": true,
//	      "type": "mysql",
//	      "host": "localhost",
//	      "port": 3306,
//	      "username": "root",
//	      "password": "password",
//	      "database": "mydb",
//	      "charset": "utf8mb4",
//	      "max-open-conns": 100,
//	      "max-idle-conns": 10,
//	      "show-sql": false
//	    }
//	  }
//	}
//
// 2. 在代码中使用：
//
//	type UserRepository struct {
//		engine *xorm.Engine
//	}
//
//	func NewUserRepository(engine *xorm.Engine) *UserRepository {
//		return &UserRepository{engine: engine}
//	}
//
// # 配置说明
//
//   - db.xorm.enabled: 是否启用 XORM（默认 false）
//   - db.xorm.type: 数据库类型，支持 mysql/postgres/sqlite（默认 mysql）
//   - db.xorm.host: 数据库主机地址
//   - db.xorm.port: 数据库端口
//   - db.xorm.username: 数据库用户名
//   - db.xorm.password: 数据库密码
//   - db.xorm.database: 数据库名称
//   - db.xorm.charset: 字符集（默认 utf8mb4）
//   - db.xorm.max-open-conns: 最大打开连接数（默认 100）
//   - db.xorm.max-idle-conns: 最大空闲连接数（默认 10）
//   - db.xorm.show-sql: 是否显示 SQL 日志（默认 false）
//
// # 依赖说明
//
// 本模块依赖：
//   - xorm.io/xorm: XORM 核心库
//   - github.com/go-sql-driver/mysql: MySQL 驱动
//
// 用户项目引入本模块后，会自动引入上述依赖。
package xorm

// ==================== 配置键常量 ====================

const (
	// XORM 配置
	XORMEnabled         = "db.xorm.enabled"
	XORMType            = "db.xorm.type"
	XORMHost            = "db.xorm.host"
	XORMPort            = "db.xorm.port"
	XORMUsername        = "db.xorm.username"
	XORMPassword        = "db.xorm.password"
	XORMDatabase        = "db.xorm.database"
	XORMCharset         = "db.xorm.charset"
	XORMMaxOpenConns    = "db.xorm.max-open-conns"
	XORMMaxIdleConns    = "db.xorm.max-idle-conns"
	XORMConnMaxLifetime = "db.xorm.conn-max-lifetime"
	XORMShowSQL         = "db.xorm.show-sql"

	// 日志字段常量
	LogFieldHost     = "host"
	LogFieldPort     = "port"
	LogFieldDatabase = "database"
	LogFieldType     = "type"
)

// ==================== 默认值常量 ====================

const (
	// XORM 默认值
	DefaultXORMType            = "mysql"
	DefaultXORMHost            = "localhost"
	DefaultXORMPort            = 3306
	DefaultXORMUsername        = "scott"
	DefaultXORMPassword        = "123456"
	DefaultXORMDatabase        = "demo"
	DefaultXORMCharset         = "utf8mb4"
	DefaultXORMMaxOpenConns    = 100
	DefaultXORMMaxIdleConns    = 10
	DefaultXORMConnMaxLifetime = 3600
	DefaultXORMShowSQL         = false

	// 条件值常量
	ConditionTrue = "true"
)
