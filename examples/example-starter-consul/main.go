// Package main demonstrates the Consul starter usage.
//
// This example shows how to use the Consul starter to:
// 1. Auto-configure Consul client
// 2. Register services
// 3. Discover services
//
// Prerequisites:
// - Consul server running on localhost:8500
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/consul"
)

func main() {
	fmt.Println("=== Consul Starter Example ===")
	fmt.Println()

	app := newApp()
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Consul connection failed: %v\n", err)
		fmt.Println("This example requires a running Consul server.")
		fmt.Println("Please start Consul and try again.")
		return
	}

	client, err := getConsulClient(app)
	if err != nil {
		return
	}

	if err := demoRegisterServices(client); err != nil {
		return
	}
	if err := demoDiscoverServices(client); err != nil {
		return
	}
	if err := demoKVOperations(client); err != nil {
		return
	}
	if err := demoCleanup(client); err != nil {
		return
	}

	fmt.Println("\n=== Example completed successfully ===")
}

// newApp 创建应用实例，配置名称与激活的 Profile。
func newApp() *boot.Boot {
	app, err := boot.NewApplication(
		boot.WithAppName("consul-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// getConsulClient 从容器的 Bean 中获取 Consul 客户端。
func getConsulClient(app *boot.Boot) (*consulapi.Client, error) {
	client, err := core.GetByName[*consulapi.Client](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get Consul client: %v\n", err)
		return nil, fmt.Errorf("Failed to get Consul client: %w", err)
	}
	return client, nil
}

// demoRegisterServices 演示注册两个服务实例（Demo 1-2）。
func demoRegisterServices(client *consulapi.Client) error {
	// Demo 1: Register a service
	fmt.Println("--- Demo 1: Register Service ---")
	registration := &consulapi.AgentServiceRegistration{
		ID:      "my-service-1",
		Name:    "my-service",
		Port:    8080,
		Address: "127.0.0.1",
		Tags:    []string{"web", "api"},
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           "http://127.0.0.1:8080/health",
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
		Meta: map[string]string{
			"version": "1.0.0",
		},
	}

	if err := client.Agent().ServiceRegister(registration); err != nil {
		fmt.Printf("Failed to register service: %v\n", err)
		return fmt.Errorf("Failed to register service: %w", err)
	}
	fmt.Println("Service registered: my-service-1")

	// Demo 2: Register another service
	fmt.Println("\n--- Demo 2: Register Another Service ---")
	registration2 := &consulapi.AgentServiceRegistration{
		ID:      "my-service-2",
		Name:    "my-service",
		Port:    8081,
		Address: "127.0.0.1",
		Tags:    []string{"web", "api"},
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           "http://127.0.0.1:8081/health",
			Interval:                       "10s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
		},
	}

	if err := client.Agent().ServiceRegister(registration2); err != nil {
		fmt.Printf("Failed to register service: %v\n", err)
		return fmt.Errorf("Failed to register service: %w", err)
	}
	fmt.Println("Service registered: my-service-2")
	return nil
}

// demoDiscoverServices 演示服务发现（Demo 3-4）。
func demoDiscoverServices(client *consulapi.Client) error {
	// Demo 3: Discover services
	fmt.Println("\n--- Demo 3: Discover Services ---")
	services, _, err := client.Health().Service("my-service", "", true, nil)
	if err != nil {
		fmt.Printf("Failed to discover services: %v\n", err)
		return fmt.Errorf("Failed to discover services: %w", err)
	}
	fmt.Printf("Found %d healthy instances of 'my-service'\n", len(services))
	for _, svc := range services {
		fmt.Printf("  - ID: %s, Address: %s:%d\n", svc.Service.ID, svc.Service.Address, svc.Service.Port)
	}

	// Demo 4: Get service by ID
	fmt.Println("\n--- Demo 4: Get Service by ID ---")
	service, _, err := client.Health().Service("my-service-1", "", true, nil)
	if err != nil {
		fmt.Printf("Failed to get service: %v\n", err)
		return fmt.Errorf("Failed to get service: %w", err)
	}
	if len(service) > 0 {
		svc := service[0].Service
		fmt.Printf("Service: %s, Address: %s:%d\n", svc.ID, svc.Address, svc.Port)
	}
	return nil
}

// demoKVOperations 演示 KV 的写入、读取与列表（Demo 5-7）。
func demoKVOperations(client *consulapi.Client) error {
	// Demo 5: Set a key-value pair
	fmt.Println("\n--- Demo 5: Set Key-Value ---")
	kv := client.KV()
	p := &consulapi.KVPair{
		Key:   "config/my-service/debug",
		Value: []byte("true"),
	}
	_, err := kv.Put(p, nil)
	if err != nil {
		fmt.Printf("Failed to set key-value: %v\n", err)
		return fmt.Errorf("Failed to set key-value: %w", err)
	}
	fmt.Println("Key-value set: config/my-service/debug = true")

	// Demo 6: Get a key-value pair
	fmt.Println("\n--- Demo 6: Get Key-Value ---")
	pair, _, err := kv.Get("config/my-service/debug", nil)
	if err != nil {
		fmt.Printf("Failed to get key-value: %v\n", err)
		return fmt.Errorf("Failed to get key-value: %w", err)
	}
	if pair != nil {
		fmt.Printf("Key: %s, Value: %s\n", pair.Key, string(pair.Value))
	}

	// Demo 7: List keys
	fmt.Println("\n--- Demo 7: List Keys ---")
	keys, _, err := kv.Keys("config/", "", nil)
	if err != nil {
		fmt.Printf("Failed to list keys: %v\n", err)
		return fmt.Errorf("Failed to list keys: %w", err)
	}
	fmt.Printf("Found %d keys under 'config/'\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  - %s\n", key)
	}
	return nil
}

// demoCleanup 演示注销服务与删除 KV 的清理流程（Demo 8-9）。
func demoCleanup(client *consulapi.Client) error {
	// Demo 8: Deregister services
	fmt.Println("\n--- Demo 8: Deregister Services ---")
	if err := client.Agent().ServiceDeregister("my-service-1"); err != nil {
		fmt.Printf("Failed to deregister service: %v\n", err)
		return fmt.Errorf("Failed to deregister service: %w", err)
	}
	fmt.Println("Service deregistered: my-service-1")

	if err := client.Agent().ServiceDeregister("my-service-2"); err != nil {
		fmt.Printf("Failed to deregister service: %v\n", err)
		return fmt.Errorf("Failed to deregister service: %w", err)
	}
	fmt.Println("Service deregistered: my-service-2")

	// Demo 9: Delete key-value
	fmt.Println("\n--- Demo 9: Delete Key-Value ---")
	_, err := client.KV().Delete("config/my-service/debug", nil)
	if err != nil {
		fmt.Printf("Failed to delete key-value: %v\n", err)
		return fmt.Errorf("Failed to delete key-value: %w", err)
	}
	fmt.Println("Key-value deleted: config/my-service/debug")

	// Wait a bit for deregistration to complete
	time.Sleep(1 * time.Second)
	return nil
}
