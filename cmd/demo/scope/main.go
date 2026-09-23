// Package main 演示 enhance 框架的作用域（Scope）功能
//
// 该示例展示：
// - Singleton 作用域（单例，默认）
// - Prototype 作用域（每次获取创建新实例）
// - 不同作用域的使用场景
package main

import (
	"fmt"

	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
)

// Connection 模拟数据库连接
type Connection struct {
	ID int
}

var connCounter = 0

func newConnection() *Connection {
	connCounter++
	return &Connection{ID: connCounter}
}

// RequestHandler 请求处理器（Prototype 作用域）
type RequestHandler struct {
	ID   int
	conn *Connection
}

var handlerCounter = 0

func main() {
	fmt.Println("=== Bean 作用域（Scope）示例 ===")
	fmt.Println()

	// 1. 创建容器
	fmt.Println("1. 创建容器")
	container := core.NewContainer()
	fmt.Println()

	// 2. 注册 Singleton Bean（默认作用域）
	fmt.Println("2. 注册 Singleton Bean（默认作用域）")
	core.Register[*Connection](container,
		core.WithFactory[*Connection](func(c ...any) (any, error) {
			return newConnection(), nil
		}),
	)
	fmt.Println("  注册 Connection（Singleton）")
	fmt.Println()

	// 3. 注册 Prototype Bean
	fmt.Println("3. 注册 Prototype Bean")
	core.Register[*RequestHandler](container,
		core.WithFactory[*RequestHandler](func(c ...any) (any, error) {
			handlerCounter++
			conn := core.MustGet[*Connection](container, "")
			return &RequestHandler{
				ID:   handlerCounter,
				conn: conn,
			}, nil
		}),
		core.WithScope[*RequestHandler](registry.Prototype),
	)
	fmt.Println("  注册 RequestHandler（Prototype）")
	fmt.Println()

	// 4. 初始化容器
	fmt.Println("4. 初始化容器")
	if err := container.Initialize(); err != nil {
		fmt.Printf("  初始化失败: %v\n", err)
		return
	}
	fmt.Println()

	// 5. 演示 Singleton 作用域
	fmt.Println("5. 演示 Singleton 作用域")
	conn1 := core.MustGet[*Connection](container, "")
	conn2 := core.MustGet[*Connection](container, "")
	fmt.Printf("  conn1.ID = %d\n", conn1.ID)
	fmt.Printf("  conn2.ID = %d\n", conn2.ID)
	fmt.Printf("  是否为同一实例: %v\n", conn1 == conn2)
	fmt.Println()

	// 6. 演示 Prototype 作用域
	fmt.Println("6. 演示 Prototype 作用域")
	handler1 := core.MustGet[*RequestHandler](container, "")
	handler2 := core.MustGet[*RequestHandler](container, "")
	fmt.Printf("  handler1.ID = %d, conn.ID = %d\n", handler1.ID, handler1.conn.ID)
	fmt.Printf("  handler2.ID = %d, conn.ID = %d\n", handler2.ID, handler2.conn.ID)
	fmt.Printf("  是否为同一实例: %v\n", handler1 == handler2)
	fmt.Printf("  但共享同一 Connection: %v\n", handler1.conn == handler2.conn)
	fmt.Println()

	// 7. 销毁容器
	fmt.Println("7. 销毁容器")
	container.Destroy()
	fmt.Println()

	fmt.Println("=== 作用域示例完成 ===")
}
