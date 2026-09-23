// Package main 演示 enhance 框架的自动配置机制
//
// 该示例展示：
// - 自动配置类的注册
// - 条件化自动配置
// - 自动配置的依赖排序
// - 应用启动流程
package main

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/core"
)

// 自动配置优先级常量
const (
	dataLayerPriority     = -2000 // 数据层优先级
	businessLayerPriority = 1000  // 业务层优先级
)

// DatabaseAutoConfiguration 数据库自动配置
type DatabaseAutoConfiguration struct{}

func (c *DatabaseAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	fmt.Println("  [DatabaseAutoConfiguration] 正在配置数据库...")

	// 注册数据库 Bean
	ctx.Register(
		reflect.TypeOf((*Database)(nil)).Elem(),
		core.WithFactory[Database](func(c ...any) (any, error) {
			return &Database{DSN: "localhost:5432"}, nil
		}),
	)

	fmt.Println("  [DatabaseAutoConfiguration] 数据库配置完成")
	return nil
}

func (c *DatabaseAutoConfiguration) Order() int {
	return dataLayerPriority
}

func init() {
	boot.RegisterAutoConfig(
		&DatabaseAutoConfiguration{},
		condition.OnProperty("database.enabled", "true"),
	)
}

// UserServiceAutoConfiguration 用户服务自动配置
type UserServiceAutoConfiguration struct{}

func (c *UserServiceAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	fmt.Println("  [UserServiceAutoConfiguration] 正在配置用户服务...")

	// 依赖数据库 Bean
	ctx.Register(
		reflect.TypeOf((*UserService)(nil)).Elem(),
		core.WithFactory[UserService](func(c ...any) (any, error) {
			container := c[0].(core.Container)
			db, err := core.GetByName[*Database](container, "")
			if err != nil {
				return nil, err
			}
			return &UserService{DB: db}, nil
		}),
	)

	fmt.Println("  [UserServiceAutoConfiguration] 用户服务配置完成")
	return nil
}

func (c *UserServiceAutoConfiguration) Order() int {
	return businessLayerPriority
}

func (c *UserServiceAutoConfiguration) DependsOn() []string {
	return []string{"DatabaseAutoConfiguration"}
}

func init() {
	boot.RegisterAutoConfig(&UserServiceAutoConfiguration{})
}

// Database 数据库连接
type Database struct {
	DSN string
}

// UserService 用户服务
type UserService struct {
	DB *Database
}

func main() {
	fmt.Println("=== 自动配置示例 ===")
	fmt.Println()

	// 设置配置属性
	os.Setenv("ENHANCE_DATABASE_ENABLED", "true")

	// 创建应用
	fmt.Println("1. 创建应用")
	app, err := boot.NewApplication(
		boot.WithAppName("autoconfig-demo"),
		boot.WithVersion("1.0.0"),
		boot.WithProperty("database.enabled", "true"),
	)
	if err != nil {
		fmt.Printf("  创建应用失败: %v\n", err)
		return
	}
	fmt.Println()

	// 启动应用
	fmt.Println("2. 启动应用（执行自动配置）")
	if err := app.Start(); err != nil {
		fmt.Printf("  启动失败: %v\n", err)
		return
	}
	fmt.Println()

	// 使用 Bean
	fmt.Println("3. 使用自动配置的 Bean")
	container := app.Container()
	userSvc, err := core.GetByName[*UserService](container, "")
	if err != nil {
		fmt.Printf("  获取 UserService 失败: %v\n", err)
		return
	}
	fmt.Printf("  UserService 已就绪，数据库 DSN: %s\n", userSvc.DB.DSN)
	fmt.Println()

	// 停止应用
	fmt.Println("4. 停止应用")
	if err := app.Stop(); err != nil {
		slog.Error("停止应用失败", "error", err)
	}
	fmt.Println()

	fmt.Println("=== 自动配置示例完成 ===")
}
