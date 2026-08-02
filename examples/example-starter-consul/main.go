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

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("consul-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Consul connection failed: %v\n", err)
		fmt.Println("This example requires a running Consul server.")
		fmt.Println("Please start Consul and try again.")
		return
	}

	// Get the Consul client from container
	consulClient, err := core.GetByName[*consulapi.Client](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get Consul client: %v\n", err)
		return
	}

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

	if err := consulClient.Agent().ServiceRegister(registration); err != nil {
		fmt.Printf("Failed to register service: %v\n", err)
		return
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

	if err := consulClient.Agent().ServiceRegister(registration2); err != nil {
		fmt.Printf("Failed to register service: %v\n", err)
		return
	}
	fmt.Println("Service registered: my-service-2")

	// Demo 3: Discover services
	fmt.Println("\n--- Demo 3: Discover Services ---")
	services, _, err := consulClient.Health().Service("my-service", "", true, nil)
	if err != nil {
		fmt.Printf("Failed to discover services: %v\n", err)
		return
	}
	fmt.Printf("Found %d healthy instances of 'my-service'\n", len(services))
	for _, svc := range services {
		fmt.Printf("  - ID: %s, Address: %s:%d\n", svc.Service.ID, svc.Service.Address, svc.Service.Port)
	}

	// Demo 4: Get service by ID
	fmt.Println("\n--- Demo 4: Get Service by ID ---")
	service, _, err := consulClient.Health().Service("my-service-1", "", true, nil)
	if err != nil {
		fmt.Printf("Failed to get service: %v\n", err)
		return
	}
	if len(service) > 0 {
		svc := service[0].Service
		fmt.Printf("Service: %s, Address: %s:%d\n", svc.ID, svc.Address, svc.Port)
	}

	// Demo 5: Set a key-value pair
	fmt.Println("\n--- Demo 5: Set Key-Value ---")
	kv := consulClient.KV()
	p := &consulapi.KVPair{
		Key:   "config/my-service/debug",
		Value: []byte("true"),
	}
	_, err = kv.Put(p, nil)
	if err != nil {
		fmt.Printf("Failed to set key-value: %v\n", err)
		return
	}
	fmt.Println("Key-value set: config/my-service/debug = true")

	// Demo 6: Get a key-value pair
	fmt.Println("\n--- Demo 6: Get Key-Value ---")
	pair, _, err := kv.Get("config/my-service/debug", nil)
	if err != nil {
		fmt.Printf("Failed to get key-value: %v\n", err)
		return
	}
	if pair != nil {
		fmt.Printf("Key: %s, Value: %s\n", pair.Key, string(pair.Value))
	}

	// Demo 7: List keys
	fmt.Println("\n--- Demo 7: List Keys ---")
	keys, _, err := kv.Keys("config/", "", nil)
	if err != nil {
		fmt.Printf("Failed to list keys: %v\n", err)
		return
	}
	fmt.Printf("Found %d keys under 'config/'\n", len(keys))
	for _, key := range keys {
		fmt.Printf("  - %s\n", key)
	}

	// Demo 8: Deregister services
	fmt.Println("\n--- Demo 8: Deregister Services ---")
	if err := consulClient.Agent().ServiceDeregister("my-service-1"); err != nil {
		fmt.Printf("Failed to deregister service: %v\n", err)
		return
	}
	fmt.Println("Service deregistered: my-service-1")

	if err := consulClient.Agent().ServiceDeregister("my-service-2"); err != nil {
		fmt.Printf("Failed to deregister service: %v\n", err)
		return
	}
	fmt.Println("Service deregistered: my-service-2")

	// Demo 9: Delete key-value
	fmt.Println("\n--- Demo 9: Delete Key-Value ---")
	_, err = kv.Delete("config/my-service/debug", nil)
	if err != nil {
		fmt.Printf("Failed to delete key-value: %v\n", err)
		return
	}
	fmt.Println("Key-value deleted: config/my-service/debug")

	// Wait a bit for deregistration to complete
	time.Sleep(1 * time.Second)

	fmt.Println("\n=== Example completed successfully ===")
}
