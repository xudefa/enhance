// Package casbin 提供基于 Casbin 的细粒度访问控制机制。
//
// 本模块是 enhance 框架的第三方授权集成模块，基于 github.com/casbin/casbin/v2 实现，
// 提供 RBAC（基于角色的访问控制）、RESTful 授权等能力，
// 与 security 模块无缝集成。
//
// 官方文档：https://casbin.org/
//
// # 模块独立性
//
// 本模块采用独立模块设计（拥有独立的 go.mod），依赖隔离确保：
//   - 用户只使用 enhance 核心模块时，不会引入 casbin 依赖
//   - 用户显式引入本模块时，才会下载 casbin 及其间接依赖
//   - 避免依赖污染，保持用户项目的依赖树清晰
//
// # 架构设计
//
// 核心组件：
//   - DefaultCasbinEnforcer: 默认 Casbin 执行器实现（基于内存）
//   - CasbinAutoConfiguration: 自动配置类，根据配置文件启用 Casbin 授权
//
// 接口定义：
//   - CasbinEnforcer: 定义在 security 包中，提供权限检查接口
//   - CasbinVoter: 定义在 security 包中，实现 security.AccessDecisionVoter 接口
//
// 自动配置：
//   - 当 security.enabled=true 且 security.casbin.enabled=true 时自动生效
//   - CasbinVoter 自动注册到 AccessDecisionManager 中
//   - 如果容器中已有 CasbinEnforcer 实例则直接使用，否则自动创建 DefaultCasbinEnforcer
//
// # 核心功能
//
//   - RBAC 授权: 基于角色的访问控制（支持角色继承）
//   - RESTful 授权: 支持 HTTP 方法和路径模式匹配
//   - 动态策略: 支持运行时动态添加/移除策略
//   - 策略热加载: 支持重新加载策略文件
//   - 多模型支持: 支持 ACL、RBAC、ABAC 等多种授权模型
//   - 多加载方式: 支持从文件、字符串、数据库加载模型和策略
//
// # 快速开始
//
// 1. 在配置文件中启用 Casbin：
//
//	{
//	  "security": {
//	    "enabled": true,
//	    "casbin": {
//	      "enabled": true,
//	      "model-type": "file",
//	      "model-path": "config/casbin_model.conf",
//	      "policy-type": "file",
//	      "policy-path": "config/casbin_policy.csv"
//	    }
//	  }
//	}
//
// 2. 创建 Casbin 模型文件（casbin_model.conf）：
//
//	[request_definition]
//	r = sub, obj, act
//
//	[policy_definition]
//	p = sub, obj, act
//
//	[role_definition]
//	g = _, _
//
//	[policy_effect]
//	e = some(where (p.eft == allow))
//
//	[matchers]
//	m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act
//
// 3. 创建策略文件（casbin_policy.csv）：
//
//	p, admin, /api/users/*, GET
//	p, admin, /api/users/*, POST
//	p, user, /api/profile, GET
//	g, alice, admin
//	g, bob, user
//
// # 配置说明
//
//   - security.casbin.enabled: 是否启用 Casbin（默认 false）
//   - security.casbin.model-type: 模型加载方式，支持 file/string（默认 file）
//   - security.casbin.model-path: Casbin 模型文件路径（model-type=file 时使用）
//   - security.casbin.model-text: Casbin 模型文本内容（model-type=string 时使用）
//   - security.casbin.policy-type: 策略加载方式，支持 file/string/gorm（默认 file）
//   - security.casbin.policy-path: Casbin 策略文件路径（policy-type=file 时使用）
//   - security.casbin.policy-text: Casbin 策略文本内容（policy-type=string 时使用）
//
// # 与 security 模块集成
//
// Casbin 模块自动与 security 模块集成：
//   - CasbinVoter 实现 security.AccessDecisionVoter 接口
//   - 自动注册到 AccessDecisionManager 的投票者列表中
//   - 在 FilterSecurityInterceptor 执行访问决策时参与投票
//   - 与 JWT 认证模块兼容，可组合使用
//
// # 执行流程
//
//  1. JwtAuthenticationFilter 认证用户身份
//  2. FilterSecurityInterceptor 获取请求的访问属性
//  3. AccessDecisionManager 调用所有投票者（包括 CasbinVoter）
//  4. CasbinVoter 根据 Casbin 策略进行投票（ACCESS_GRANTED / ACCESS_DENIED / ACCESS_ABSTAIN）
//  5. AffirmativeBased 决策管理器根据投票结果决定是否允许访问
//
// # 策略语法
//
// Casbin 支持多种策略语法：
//   - p: 策略定义（policy），格式为 p = sub, obj, act
//   - g: 角色定义（group），格式为 g = user, role
//   - r: 请求定义（request），格式为 r = sub, obj, act
//
// 路径匹配函数：
//   - keyMatch: 精确匹配
//   - keyMatch2: 支持 RESTful 路径（如 /api/users/:id）
//   - keyMatch3: 支持通配符（如 /api/users/*）
//   - keyMatch4: 支持通配符和参数
//
// # 依赖说明
//
// 本模块依赖：
//   - github.com/casbin/casbin/v2: Casbin 核心库
//   - github.com/xudefa/enhance/security: 安全框架核心
//
// 用户项目引入本模块后，会自动引入上述依赖。
package casbin

// 注意：CasbinEnforcer 接口和 CasbinVoter 类型已移至 security 包中。
// 本包仅提供 DefaultCasbinEnforcer 实现和自动配置功能。
//
// 使用示例：
//
//	import "github.com/xudefa/enhance/security"
//
//	var enforcer security.CasbinEnforcer
//	voter := security.NewCasbinVoter(enforcer)
