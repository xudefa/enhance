// Package main demonstrates the Cobra starter usage.
//
// This example shows how to use the Cobra starter to:
// 1. Auto-configure Cobra CLI framework
// 2. Add commands and flags
// 3. Build a CLI application
//
// Run:
//
//	go run main.go --help
//	go run main.go greet --name World
//	go run main.go version
package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/cobra"
)

func main() {
	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("cobra-example"),
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

	// Get the root command from container
	rootCmd, err := core.GetByName[*cobra.Command](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get root command: %v\n", err)
		return
	}

	// Add a "greet" command
	greetCmd := &cobra.Command{
		Use:   "greet",
		Short: "Greet someone",
		Long:  "Print a greeting message to someone",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				name = "World"
			}
			fmt.Printf("Hello, %s!\n", name)
			return nil
		},
	}

	greetCmd.Flags().StringP("name", "n", "", "Name to greet")
	rootCmd.AddCommand(greetCmd)

	// Add a "version" command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print the version information of the application",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Cobra Example v1.0.0\n")
			fmt.Printf("Built with enhance framework\n")
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Add a "config" command with subcommands
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "Manage application configuration",
	}

	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Current configuration:")
			fmt.Println("  app.name: cobra-example")
			fmt.Println("  app.version: 1.0.0")
			fmt.Println("  server.host: 0.0.0.0")
			fmt.Println("  server.port: 8080")
		},
	}

	configSetCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key := args[0]
			value := args[1]
			fmt.Printf("Setting %s = %s\n", key, value)
		},
	}

	configCmd.AddCommand(configShowCmd, configSetCmd)
	rootCmd.AddCommand(configCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
