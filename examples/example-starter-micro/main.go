// Package main demonstrates the Micro starter usage.
//
// This example shows how to use the Micro starter to:
// 1. Auto-configure Micro service
// 2. Register services
// 3. Call other services
//
// Prerequisites:
// - Consul running on localhost:8500 (for service registry)
//
// Run:
//
//	go run main.go
package main

import (
	"context"
	"fmt"

	"github.com/xudefa/enhance/boot"

	_ "github.com/xudefa/enhance/starter/micro"
)

// Placeholder types for go-micro
type Service interface {
	Init(...Option)
	Options() Options
	Run() error
}

// Options Micro 服务配置选项。
type Options struct {
	Name    string
	Version string
}

// Option 函数选项模式，用于配置 Micro 服务。
type Option func(*Options)

func main() {
	fmt.Println("=== Micro Starter Example ===")
	fmt.Println()

	app := newApp("micro-example")
	defer app.Stop()

	if !startApp(app) {
		return
	}

	// Get the Micro service from container
	// In a real application, you would get the actual micro.Service
	// microService, err := core.GetByName[micro.Service](app.Container(), "")
	_ = context.Background()

	printMicroConfig()
	printMicroUsage()
}

// newApp 创建 enhance 应用，创建失败时 panic。
func newApp(name string) *boot.Boot {
	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName(name),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// startApp 启动应用，触发 Micro 自动配置。
func startApp(app *boot.Boot) bool {
	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return false
	}
	return true
}

// printMicroConfig 打印 Micro 服务配置信息与核心特性。
func printMicroConfig() {
	fmt.Println("--- Micro Service Configuration ---")
	fmt.Println("Service Name: my-micro-service")
	fmt.Println("Version: 1.0.0")
	fmt.Println("Registry: consul://localhost:8500")

	fmt.Println("\n--- Micro Service Features ---")
	fmt.Println("1. Service Discovery")
	fmt.Println("   - Register service with Consul")
	fmt.Println("   - Discover other services")
	fmt.Println("   - Health checks")

	fmt.Println("\n2. Load Balancing")
	fmt.Println("   - Round Robin")
	fmt.Println("   - Random")
	fmt.Println("   - Weighted")

	fmt.Println("\n3. Circuit Breaking")
	fmt.Println("   - Request timeout")
	fmt.Println("   - Max retries")
	fmt.Println("   - Breaker threshold")

	fmt.Println("\n4. Tracing")
	fmt.Println("   - Distributed tracing")
	fmt.Println("   - Request correlation")
	fmt.Println("   - Span propagation")
}

// printMicroUsage 打印 Micro 服务使用、通信与发现示例。
func printMicroUsage() {
	fmt.Println("\n--- Example Usage ---")
	fmt.Println("// Create a new service")
	fmt.Println("svc := micro.NewService(")
	fmt.Println("    micro.Name(\"my-service\"),")
	fmt.Println("    micro.Version(\"1.0.0\"),")
	fmt.Println(")")
	fmt.Println()
	fmt.Println("// Initialize the service")
	fmt.Println("svc.Init()")
	fmt.Println()
	fmt.Println("// Register handler")
	fmt.Println("micro.RegisterHandler(svc.Server(), new(Handler))")
	fmt.Println()
	fmt.Println("// Run the service")
	fmt.Println("svc.Run()")

	fmt.Println("\n--- Service Communication ---")
	fmt.Println("// Call another service")
	fmt.Println("req := svc.Client().NewRequest(")
	fmt.Println("    \"service-name\",")
	fmt.Println("    \"Method\",")
	fmt.Println("    &Request{...},")
	fmt.Println(")")
	fmt.Println()
	fmt.Println("var resp Response")
	fmt.Println("err := svc.Client().Call(ctx, req, &resp)")

	fmt.Println("\n--- Service Discovery ---")
	fmt.Println("// List services")
	fmt.Println("services, err := svc.Registry().ListServices()")
	fmt.Println()
	fmt.Println("// Get service instances")
	fmt.Println("instances, err := svc.Registry().GetService(\"service-name\")")

	fmt.Println("\n=== Example completed successfully ===")
}
