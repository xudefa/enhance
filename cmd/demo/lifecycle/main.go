// Package main 演示 enhance 框架的 Bean 生命周期管理
//
// 该示例展示：
// - Bean 的初始化和销毁回调
// - 生命周期阶段监听器
// - 容器初始化和销毁流程
package main

import (
	"fmt"

	"github.com/xudefa/enhance/core"
)

// Database 模拟数据库连接
type Database struct {
	dsn       string
	connected bool
}

// Init 初始化回调，在容器初始化时调用
func (db *Database) Init() error {
	fmt.Printf("  [Database] 正在连接到 %s...\n", db.dsn)
	db.connected = true
	fmt.Println("  [Database] 连接成功")
	return nil
}

// Destroy 销毁回调，在容器销毁时调用
func (db *Database) Destroy() error {
	fmt.Println("  [Database] 正在关闭连接...")
	db.connected = false
	fmt.Println("  [Database] 连接已关闭")
	return nil
}

// UserService 用户服务，依赖 Database
type UserService struct {
	db *Database
}

// Init 初始化回调
func (svc *UserService) Init() error {
	if !svc.db.connected {
		return fmt.Errorf("database not connected")
	}
	fmt.Println("  [UserService] 初始化完成")
	return nil
}

// Destroy 销毁回调
func (svc *UserService) Destroy() error {
	fmt.Println("  [UserService] 清理资源")
	return nil
}

func main() {
	fmt.Println("=== Bean 生命周期管理示例 ===")
	fmt.Println()

	// 1. 创建容器
	fmt.Println("1. 创建容器")
	container := core.NewContainer()
	fmt.Println()

	// 2. 注册带生命周期回调的 Bean
	fmt.Println("2. 注册带生命周期回调的 Bean")

	core.Register[*Database](container,
		core.WithFactory[*Database](func(c ...any) (any, error) {
			return &Database{dsn: "localhost:5432"}, nil
		}),
		core.WithInit[*Database](func(bean any) error {
			return bean.(*Database).Init()
		}),
		core.WithDestroy[*Database](func(bean any) error {
			return bean.(*Database).Destroy()
		}),
	)

	core.Register[*UserService](container,
		core.WithFactory[*UserService](func(c ...any) (any, error) {
			db := core.MustGet[*Database](container, "")
			return &UserService{db: db}, nil
		}),
		core.WithInit[*UserService](func(bean any) error {
			return bean.(*UserService).Init()
		}),
		core.WithDestroy[*UserService](func(bean any) error {
			return bean.(*UserService).Destroy()
		}),
	)
	fmt.Println()

	// 3. 初始化容器（触发所有 Bean 的 Init 回调）
	fmt.Println("3. 初始化容器（触发 Init 回调）")
	if err := container.Initialize(); err != nil {
		fmt.Printf("  初始化失败: %v\n", err)
		return
	}
	fmt.Println()

	// 4. 使用 Bean
	fmt.Println("4. 使用 Bean")
	svc := core.MustGet[*UserService](container, "")
	fmt.Printf("  UserService 已就绪，数据库连接状态: %v\n", svc.db.connected)
	fmt.Println()

	// 5. 销毁容器（触发所有 Bean 的 Destroy 回调）
	fmt.Println("5. 销毁容器（触发 Destroy 回调）")
	if err := container.Destroy(); err != nil {
		fmt.Printf("  销毁失败: %v\n", err)
		return
	}
	fmt.Println()

	fmt.Println("=== 生命周期管理示例完成 ===")
}
