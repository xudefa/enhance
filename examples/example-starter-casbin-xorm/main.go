// Package main demonstrates the Casbin-XORM starter usage.
//
// This example shows how to use the Casbin-XORM starter to:
// 1. Auto-configure Casbin with XORM adapter
// 2. Store policies in database
// 3. Check permissions
//
// Prerequisites:
// - MySQL server running on localhost:3306
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/casbin-xorm"
	_ "github.com/xudefa/enhance/starter/xorm"
)

func main() {
	fmt.Println("=== Casbin-XORM Starter Example ===")
	fmt.Println()

	app, err := boot.NewApplication(
		boot.WithAppName("casbin-xorm-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Database connection failed: %v\n", err)
		fmt.Println("This example requires a running MySQL server.")
		return
	}

	enforcer, err := core.GetByName[*casbin.Enforcer](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get Casbin enforcer: %v\n", err)
		return
	}

	fmt.Println("--- Check Permissions ---")
	allowed, _ := enforcer.Enforce("alice", "data1", "read")
	fmt.Printf("alice can read data1: %v\n", allowed)

	fmt.Println("\n--- Add Policy ---")
	enforcer.AddPolicy("charlie", "data3", "read")
	allowed, _ = enforcer.Enforce("charlie", "data3", "read")
	fmt.Printf("charlie can read data3: %v\n", allowed)

	fmt.Println("\n--- Add Role ---")
	enforcer.AddRoleForUser("bob", "admin")
	allowed, _ = enforcer.Enforce("bob", "data1", "read")
	fmt.Printf("bob can read data1 (as admin): %v\n", allowed)

	fmt.Println("\n--- All Policies ---")
	policies, _ := enforcer.GetPolicy()
	for _, p := range policies {
		fmt.Printf("  %v\n", p)
	}

	fmt.Println("\n=== Example completed successfully ===")
}
