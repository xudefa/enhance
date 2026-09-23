// Package main 演示 enhance 框架的父子容器功能
//
// 该示例展示：
// - 父子容器的创建
// - 子容器访问父容器的 Bean
// - 子容器覆盖父容器的 Bean
// - Bean 查找的优先级
package main

import (
	"fmt"

	"github.com/xudefa/enhance/core"
)

// Config 配置对象
type Config struct {
	Name string
}

// Logger 日志接口
type Logger struct {
	Prefix string
}

func (l *Logger) Log(msg string) {
	fmt.Printf("[%s] %s\n", l.Prefix, msg)
}

// ChildLogger 子容器的日志实现
type ChildLogger struct {
	Prefix string
}

func (l *ChildLogger) Log(msg string) {
	fmt.Printf("[Child-%s] %s\n", l.Prefix, msg)
}

func main() {
	fmt.Println("=== 父子容器示例 ===")
	fmt.Println()

	// 1. 创建父容器
	fmt.Println("1. 创建父容器")
	parent := core.NewContainer()

	core.Register[*Config](parent,
		core.WithFactory[*Config](func(c ...any) (any, error) {
			return &Config{Name: "ParentConfig"}, nil
		}),
	)

	core.Register[*Logger](parent,
		core.WithFactory[*Logger](func(c ...any) (any, error) {
			return &Logger{Prefix: "Parent"}, nil
		}),
	)

	if err := parent.Initialize(); err != nil {
		fmt.Printf("  父容器初始化失败: %v\n", err)
		return
	}
	fmt.Println("  父容器初始化完成")
	fmt.Println()

	// 2. 创建子容器
	fmt.Println("2. 创建子容器（带父容器）")
	child := core.NewContainer()
	child.(core.ContainerExt).SetParent(parent)

	// 子容器注册自己的 Bean
	core.Register[*ChildLogger](child,
		core.WithFactory[*ChildLogger](func(c ...any) (any, error) {
			return &ChildLogger{Prefix: "Child"}, nil
		}),
	)

	if err := child.Initialize(); err != nil {
		fmt.Printf("  子容器初始化失败: %v\n", err)
		return
	}
	fmt.Println("  子容器初始化完成")
	fmt.Println()

	// 3. 子容器访问父容器的 Bean
	fmt.Println("3. 子容器访问父容器的 Bean")
	config := core.MustGet[*Config](child, "")
	fmt.Printf("  从子容器获取 Config: Name=%s\n", config.Name)
	fmt.Println()

	// 4. 子容器覆盖父容器的 Bean
	fmt.Println("4. 子容器覆盖父容器的 Bean")
	// 注意：子容器已初始化，无法再注册 Bean
	// 这里演示的是查找优先级：子容器优先查找自己的 Bean
	childLogger := core.MustGet[*ChildLogger](child, "")
	childLogger.Log("来自子容器的日志")
	fmt.Println()

	// 5. 父容器无法访问子容器的 Bean
	fmt.Println("5. 父容器无法访问子容器的 Bean")
	if _, err := core.GetByName[*ChildLogger](parent, ""); err != nil {
		fmt.Printf("  从父容器获取 ChildLogger 失败（符合预期）: %v\n", err)
	}
	fmt.Println()

	// 6. 销毁容器
	fmt.Println("6. 销毁容器")
	child.Destroy()
	parent.Destroy()
	fmt.Println()

	fmt.Println("=== 父子容器示例完成 ===")
}
