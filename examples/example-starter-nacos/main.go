// Package main demonstrates the Nacos starter usage.
//
// This example shows how to use the Nacos starter to:
// 1. Auto-configure Nacos client
// 2. Get configuration
// 3. Listen for configuration changes
//
// Prerequisites:
// - Nacos server running on localhost:8848
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"

	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/nacos"
)

func main() {
	fmt.Println("=== Nacos Starter Example ===")
	fmt.Println()

	app := newApp()
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Nacos connection failed: %v\n", err)
		fmt.Println("This example requires a running Nacos server.")
		fmt.Println("Please start Nacos and try again.")
		return
	}

	configClient, err := getConfigClient(app)
	if err != nil {
		return
	}

	if err := demoPublishConfig(configClient); err != nil {
		return
	}
	if err := demoGetConfig(configClient); err != nil {
		return
	}
	if err := demoListenConfig(configClient); err != nil {
		return
	}
	if err := demoDeleteConfig(configClient); err != nil {
		return
	}
	if err := demoListConfigs(); err != nil {
		return
	}

	fmt.Println("\n=== Example completed successfully ===")
}

// newApp 创建应用实例，配置名称与激活的 Profile。
func newApp() *boot.Boot {
	app, err := boot.NewApplication(
		boot.WithAppName("nacos-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// getConfigClient 从容器的 Bean 中获取 Nacos 配置客户端。
func getConfigClient(app *boot.Boot) (config_client.IConfigClient, error) {
	configClient, err := core.GetByName[config_client.IConfigClient](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get Nacos config client: %v\n", err)
		return nil, fmt.Errorf("Failed to get Nacos config client: %w", err)
	}
	return configClient, nil
}

// demoPublishConfig 演示发布一份配置（Demo 1）。
func demoPublishConfig(configClient config_client.IConfigClient) error {
	// Demo 1: Publish a configuration
	fmt.Println("--- Demo 1: Publish Configuration ---")
	content := `{
		"database": {
			"host": "localhost",
			"port": 3306,
			"name": "mydb"
		},
		"redis": {
			"host": "localhost",
			"port": 6379
		}
	}`

	success, err := configClient.PublishConfig(vo.ConfigParam{
		DataId:  "application.json",
		Group:   "DEFAULT_GROUP",
		Content: content,
		Type:    "json",
	})
	if err != nil {
		fmt.Printf("Failed to publish config: %v\n", err)
		return fmt.Errorf("Failed to publish config: %w", err)
	}
	fmt.Printf("Config published: %v\n", success)
	return nil
}

// demoGetConfig 演示读取一份配置（Demo 2）。
func demoGetConfig(configClient config_client.IConfigClient) error {
	// Demo 2: Get a configuration
	fmt.Println("\n--- Demo 2: Get Configuration ---")
	config, err := configClient.GetConfig(vo.ConfigParam{
		DataId: "application.json",
		Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		fmt.Printf("Failed to get config: %v\n", err)
		return fmt.Errorf("Failed to get config: %w", err)
	}
	fmt.Printf("Config content:\n%s\n", config)
	return nil
}

// demoListenConfig 演示监听配置变更（Demo 3）。
func demoListenConfig(configClient config_client.IConfigClient) error {
	// Demo 3: Listen for configuration changes
	fmt.Println("\n--- Demo 3: Listen for Config Changes ---")
	err := configClient.ListenConfig(vo.ConfigParam{
		DataId: "application.json",
		Group:  "DEFAULT_GROUP",
		OnChange: func(namespace, group, dataId, data string) {
			fmt.Printf("Config changed: namespace=%s, group=%s, dataId=%s\n", namespace, group, dataId)
			fmt.Printf("New content:\n%s\n", data)
		},
	})
	if err != nil {
		fmt.Printf("Failed to listen config: %v\n", err)
		return fmt.Errorf("Failed to listen config: %w", err)
	}
	fmt.Println("Listening for config changes...")
	return nil
}

// demoDeleteConfig 演示删除一份配置（Demo 4）。
func demoDeleteConfig(configClient config_client.IConfigClient) error {
	// Demo 4: Delete a configuration
	fmt.Println("\n--- Demo 4: Delete Configuration ---")
	success, err := configClient.DeleteConfig(vo.ConfigParam{
		DataId: "application.json",
		Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		fmt.Printf("Failed to delete config: %v\n", err)
		return fmt.Errorf("Failed to delete config: %w", err)
	}
	fmt.Printf("Config deleted: %v\n", success)
	return nil
}

// demoListConfigs 演示展示可用配置列表（Demo 5）。
func demoListConfigs() error {
	// Demo 5: List configurations
	fmt.Println("\n--- Demo 5: List Configurations ---")
	fmt.Println("Note: Listing configurations requires admin permissions")
	fmt.Println("Available configs in DEFAULT_GROUP:")
	fmt.Println("  - application.json (if not deleted)")
	return nil
}
