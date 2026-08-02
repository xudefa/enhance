// Package main demonstrates the Viper starter usage.
//
// This example shows how to use the Viper starter to:
// 1. Load configuration from YAML file
// 2. Access configuration values
// 3. Watch for configuration changes
//
// Configuration file: config/application.yaml
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/viper"
)

func main() {
	fmt.Println("=== Viper Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("viper-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	// Get the Viper instance from container
	v, err := core.GetByName[*viper.Viper](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get viper: %v\n", err)
		return
	}

	// Demonstrate configuration access
	fmt.Println("--- Configuration Access ---")

	// Read app configuration
	appName := v.GetString("app.name")
	appVersion := v.GetString("app.version")
	fmt.Printf("App Name: %s\n", appName)
	fmt.Printf("App Version: %s\n", appVersion)

	// Read server configuration
	host := v.GetString("server.host")
	port := v.GetInt("server.port")
	fmt.Printf("Server: %s:%d\n", host, port)

	// Read database configuration
	dbHost := v.GetString("database.host")
	dbPort := v.GetInt("database.port")
	dbName := v.GetString("database.name")
	fmt.Printf("Database: %s:%d/%s\n", dbHost, dbPort, dbName)

	// Read logging configuration
	logLevel := v.GetString("logging.level")
	logFormat := v.GetString("logging.format")
	fmt.Printf("Logging: level=%s, format=%s\n", logLevel, logFormat)

	// Demonstrate default values
	fmt.Println("\n--- Default Values ---")
	nonExistent := v.GetString("non.existent.key")
	fmt.Printf("Non-existent key (with default): %q\n", nonExistent)

	// Demonstrate configuration map
	fmt.Println("\n--- Configuration Map ---")
	allSettings := v.AllSettings()
	for key, value := range allSettings {
		fmt.Printf("%s: %v\n", key, value)
	}

	fmt.Println("\n=== Example completed successfully ===")
}
