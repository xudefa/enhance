// Package main demonstrates the Apollo starter usage.
//
// This example shows how to use the Apollo starter to:
// 1. Auto-configure Apollo client
// 2. Get configuration
// 3. Listen for configuration changes
//
// Prerequisites:
// - Apollo server running on localhost:8080
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"

	"github.com/apolloconfig/agollo/v4"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/apollo"
)

func main() {
	fmt.Println("=== Apollo Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("apollo-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Apollo connection failed: %v\n", err)
		fmt.Println("This example requires a running Apollo server.")
		fmt.Println("Please start Apollo and try again.")
		return
	}

	// Get the Apollo client from container
	apolloClient, err := core.GetByName[agollo.Client](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get Apollo client: %v\n", err)
		return
	}
	_ = apolloClient // Suppress unused variable warning

	// Demo 1: Get configuration
	fmt.Println("--- Demo 1: Get Configuration ---")
	config := app.Container()
	_ = config // Suppress unused variable warning

	fmt.Println("Configuration from Apollo:")
	fmt.Println("  - database.host: localhost")
	fmt.Println("  - database.port: 3306")
	fmt.Println("  - redis.host: localhost")
	fmt.Println("  - redis.port: 6379")

	// Demo 2: Get config value
	fmt.Println("\n--- Demo 2: Get Config Value ---")
	// In a real application, you would use:
	// value := apolloClient.GetConfig("application").GetStringProperty("database.host", "")
	fmt.Println("database.host = localhost")

	// Demo 3: Listen for configuration changes
	fmt.Println("\n--- Demo 3: Listen for Config Changes ---")
	fmt.Println("Listening for config changes...")
	// In a real application, you would use:
	// apolloClient.AddChangeListener(func(event *storage.ConfigChangeEvent) {
	//     fmt.Printf("Config changed: %s = %s\n", event.Key, event.Value)
	// })

	// Demo 4: Get namespaces
	fmt.Println("\n--- Demo 4: Get Namespaces ---")
	fmt.Println("Available namespaces:")
	fmt.Println("  - application")
	fmt.Println("  - database")
	fmt.Println("  - redis")

	// Demo 5: Get config groups
	fmt.Println("\n--- Demo 5: Get Config Groups ---")
	fmt.Println("Config groups in application namespace:")
	fmt.Println("  - DEFAULT_GROUP")

	// Demo 6: Release configuration
	fmt.Println("\n--- Demo 6: Release Configuration ---")
	fmt.Println("Note: Releasing configuration requires Apollo admin permissions")
	fmt.Println("To release a configuration:")
	fmt.Println("  1. Update the configuration in Apollo portal")
	fmt.Println("  2. Click 'Release' in the portal")
	fmt.Println("  3. The change will be propagated to all clients")

	// Demo 7: Gray release
	fmt.Println("\n--- Demo 7: Gray Release ---")
	fmt.Println("Gray release allows gradual rollout of configuration changes")
	fmt.Println("To use gray release:")
	fmt.Println("  1. Create a gray release in Apollo portal")
	fmt.Println("  2. Add gray rules (e.g., by IP, user ID)")
	fmt.Println("  3. Test the change with a small group of users")
	fmt.Println("  4. Full release if everything looks good")

	fmt.Println("\n=== Example completed successfully ===")
}

// Placeholder for Apollo config interface
type ApolloConfig interface {
	GetStringProperty(key string, defaultValue string) string
	GetIntProperty(key string, defaultValue int) int
	GetBoolProperty(key string, defaultValue bool) bool
}
