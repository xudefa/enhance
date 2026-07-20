package boot_test

import (
	"fmt"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
)

// ExampleOrderPriority 演示如何使用 OrderPriority 枚举设置执行顺序
func ExampleOrderPriority() {
	// 基础设施层：日志组件（最先执行）
	boot.RegisterAutoConfigWith(&ExampleLoggerConfig{},
		boot.WithConditions(
			condition.OnProperty("log.enabled", "true"),
		),
		boot.WithOrder(int(boot.OrderPriorityInfrastructure)),
	)

	// 数据层：数据库组件
	boot.RegisterAutoConfigWith(&ExampleDatabaseConfig{},
		boot.WithConditions(
			condition.OnProperty("db.enabled", "true"),
		),
		boot.WithOrder(int(boot.OrderPriorityDataLayer)),
	)

	// 认证层：JWT 组件
	boot.RegisterAutoConfigWith(&ExampleJwtConfig{},
		boot.WithConditions(
			condition.OnProperty("security.jwt.enabled", "true"),
		),
		boot.WithOrder(int(boot.OrderPriorityAuthentication)),
	)

	// 授权层-GORM：Casbin GORM 适配器（依赖数据库）
	boot.RegisterAutoConfigWith(&ExampleCasbinGormConfig{},
		boot.WithConditions(
			condition.OnProperty("security.casbin.enabled", "true"),
			condition.OnProperty("security.casbin.policy-type", "gorm"),
		),
		boot.WithOrder(int(boot.OrderPriorityAuthorizationGorm)),
	)

	// 授权层：Casbin 基础配置（检测容器中是否有 Enforcer）
	boot.RegisterAutoConfigWith(&ExampleCasbinConfig{},
		boot.WithConditions(
			condition.OnProperty("security.casbin.enabled", "true"),
		),
		boot.WithOrder(int(boot.OrderPriorityAuthorization)),
	)

	// 安全核心层：安全过滤器链
	boot.RegisterAutoConfigWith(&ExampleSecurityConfig{},
		boot.WithConditions(
			condition.OnProperty("security.enabled", "true"),
		),
		boot.WithOrder(int(boot.OrderPrioritySecurityCore)),
	)

	// Web 层：HTTP 服务器
	boot.RegisterAutoConfigWith(&ExampleWebConfig{},
		boot.WithConditions(
			condition.OnProperty("web.enabled", "true"),
		),
		boot.WithOrder(int(boot.OrderPriorityWebLayer)),
	)

	// 业务层：定时任务
	boot.RegisterAutoConfigWith(&ExampleScheduleConfig{},
		boot.WithConditions(
			condition.OnProperty("schedule.enabled", "true"),
		),
		boot.WithOrder(int(boot.OrderPriorityBusinessLayer)),
	)

	// 监控层：Actuator
	boot.RegisterAutoConfigWith(&ExampleActuatorConfig{},
		boot.WithConditions(
			condition.OnProperty("actuator.enabled", "true"),
		),
		boot.WithOrder(int(boot.OrderPriorityMonitoringLayer)),
	)

	fmt.Println("执行顺序: Infrastructure → DataLayer → Authentication → AuthorizationGorm → Authorization → SecurityCore → WebLayer → BusinessLayer → MonitoringLayer")
	// Output: 执行顺序: Infrastructure → DataLayer → Authentication → AuthorizationGorm → Authorization → SecurityCore → WebLayer → BusinessLayer → MonitoringLayer
}

// 示例自动配置类（仅用于演示）

type ExampleLoggerConfig struct{}

func (c *ExampleLoggerConfig) Configure(ctx boot.ApplicationContext) error {
	return nil
}

type ExampleDatabaseConfig struct{}

func (c *ExampleDatabaseConfig) Configure(ctx boot.ApplicationContext) error {
	return nil
}

type ExampleJwtConfig struct{}

func (c *ExampleJwtConfig) Configure(ctx boot.ApplicationContext) error {
	return nil
}

type ExampleCasbinGormConfig struct{}

func (c *ExampleCasbinGormConfig) Configure(ctx boot.ApplicationContext) error {
	return nil
}

type ExampleCasbinConfig struct{}

func (c *ExampleCasbinConfig) Configure(ctx boot.ApplicationContext) error {
	return nil
}

type ExampleSecurityConfig struct{}

func (c *ExampleSecurityConfig) Configure(ctx boot.ApplicationContext) error {
	return nil
}

type ExampleWebConfig struct{}

func (c *ExampleWebConfig) Configure(ctx boot.ApplicationContext) error {
	return nil
}

type ExampleScheduleConfig struct{}

func (c *ExampleScheduleConfig) Configure(ctx boot.ApplicationContext) error {
	return nil
}

type ExampleActuatorConfig struct{}

func (c *ExampleActuatorConfig) Configure(ctx boot.ApplicationContext) error {
	return nil
}
