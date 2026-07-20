// Package casbingorm 提供 Casbin 与 GORM 的集成能力。
//
// 本模块是 enhance 框架的扩展模块，将 Casbin 权限管理与 GORM 数据库访问能力集成，
// 支持将 Casbin 策略持久化到数据库中。
//
// # 模块独立性
//
// 本模块采用独立模块设计（拥有独立的 go.mod），依赖隔离确保：
//   - 用户只使用 enhance 核心模块时，不会引入 casbin-gorm 依赖
//   - 用户显式引入本模块时，才会下载 casbin-gorm 及其间接依赖
//   - 避免依赖污染，保持用户项目的依赖树清晰
//
// # 架构设计
//
// 核心组件：
//   - CasbinGormAutoConfiguration: 自动配置类，根据配置文件创建和注册基于 GORM 的 CasbinEnforcer
//
// 自动配置：
//   - 当 security.enabled=true 且 security.casbin.enabled=true 且 security.casbin.policy-type=gorm 时自动生效
//   - 自动创建基于 GORM 的 Casbin 适配器并注册到 IoC 容器
//   - 支持数据库策略持久化
//
// # 快速开始
//
// 1. 在配置文件中启用 Casbin 和 GORM：
//
//	{
//	  "security": {
//	    "enabled": true,
//	    "casbin": {
//	      "enabled": true,
//	      "model-type": "file",
//	      "model-path": "config/casbin_model.conf",
//	      "policy-type": "gorm"
//	    }
//	  },
//	  "db": {
//	    "gorm": {
//	      "enabled": true,
//	      "host": "localhost",
//	      "port": 3306,
//	      "username": "root",
//	      "password": "password",
//	      "database": "mydb"
//	    }
//	  }
//	}
//
// 2. 在代码中使用：
//
//	type UserController struct {
//		enforcer casbin.CasbinEnforcer
//	}
//
//	func NewUserController(enforcer casbin.CasbinEnforcer) *UserController {
//		return &UserController{enforcer: enforcer}
//	}
//
// # 配置说明
//
//   - security.casbin.enabled: 是否启用 Casbin（默认 false）
//   - security.casbin.policy-type: 策略加载方式，设置为 gorm 时使用数据库存储
//   - security.casbin.model-type: 模型加载方式，支持 file/string（默认 file）
//   - security.casbin.model-path: Casbin 模型文件路径（model-type=file 时使用）
//   - security.casbin.model-text: Casbin 模型文本内容（model-type=string 时使用）
//   - security.casbin.table-name: 策略表名（默认 casbin_rule）
//   - security.casbin.database-prefix: 数据库前缀（默认空）
//   - security.casbin.auto-create-table: 是否自动创建策略表（默认 true）
//   - security.casbin.auto-load: 是否自动刷新策略（默认 false）
//   - security.casbin.auto-load-interval: 自动刷新间隔，单位分钟（默认 5）
//   - db.gorm.enabled: 是否启用 GORM（默认 false）
//
// # 依赖说明
//
// 本模块依赖：
//   - github.com/casbin/casbin/v2: Casbin 核心库
//   - github.com/casbin/gorm-adapter/v3: Casbin GORM 适配器
//   - github.com/xudefa/enhance/security: 安全框架核心
//
// 用户项目引入本模块后，会自动引入上述依赖。
package casbingorm

// ==================== 配置键常量 ====================

const (
	// Casbin GORM 配置
	CasbinGormEnabled         = "security.casbin.enabled"
	CasbinGormPolicyType      = "security.casbin.policy-type"
	CasbinGormAutoCreateTable = "security.casbin.auto-create-table"
	CasbinGormTableName       = "security.casbin.table-name"
	CasbinGormDatabasePrefix  = "security.casbin.database-prefix"

	// 日志字段常量
	LogFieldPolicyType = "policy-type"
)

// ==================== 默认值常量 ====================

const (
	// Casbin GORM 默认值
	DefaultCasbinGormPolicyType      = "gorm"
	DefaultCasbinGormAutoCreateTable = true
	DefaultCasbinGormTableName       = "casbin_rule"
	DefaultCasbinGormDatabasePrefix  = ""

	// 条件值常量
	ConditionTrue = "true"
)
