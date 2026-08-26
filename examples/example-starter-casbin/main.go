// Package main demonstrates the Casbin starter usage.
//
// This example shows how to use the Casbin starter to:
// 1. Auto-configure Casbin enforcer
// 2. Define policies
// 3. Check permissions
//
// Run:
//
//	go run main.go
package main

import (
	"context"
	"fmt"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/starter/casbin"

	_ "github.com/xudefa/enhance/starter/casbin"
)

func main() {
	ctx := context.Background()
	fmt.Println("=== Casbin Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("casbin-example"),
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

	// Get the Casbin enforcer from container
	enforcer, err := core.GetByName[*casbin.DefaultCasbinEnforcer](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get Casbin enforcer: %v\n", err)
		return
	}

	// Demo 1: Check permissions
	fmt.Println("--- Demo 1: Check Permissions ---")
	// Alice can read data1
	allowed, err := enforcer.Enforce(ctx, "alice", "data1", "read")
	if err != nil {
		fmt.Printf("Failed to check permission: %v\n", err)
		return
	}
	fmt.Printf("alice can read data1: %v\n", allowed)

	// Alice can write data1 (as admin)
	allowed, err = enforcer.Enforce(ctx, "alice", "data1", "write")
	if err != nil {
		fmt.Printf("Failed to check permission: %v\n", err)
		return
	}
	fmt.Printf("alice can write data1 (as admin): %v\n", allowed)

	// Alice cannot read data2
	allowed, err = enforcer.Enforce(ctx, "alice", "data2", "read")
	if err != nil {
		fmt.Printf("Failed to check permission: %v\n", err)
		return
	}
	fmt.Printf("alice can read data2: %v\n", allowed)

	// Demo 2: Check Bob's permissions
	fmt.Println("\n--- Demo 2: Bob's Permissions ---")
	// Bob can write data2
	allowed, err = enforcer.Enforce(ctx, "bob", "data2", "write")
	if err != nil {
		fmt.Printf("Failed to check permission: %v\n", err)
		return
	}
	fmt.Printf("bob can write data2: %v\n", allowed)

	// Bob cannot read data1
	allowed, err = enforcer.Enforce(ctx, "bob", "data1", "read")
	if err != nil {
		fmt.Printf("Failed to check permission: %v\n", err)
		return
	}
	fmt.Printf("bob can read data1: %v\n", allowed)

	// Demo 3: Check admin permissions
	fmt.Println("\n--- Demo 3: Admin Permissions ---")
	// Admin can read and write both data1 and data2
	adminTests := []struct {
		sub  string
		obj  string
		act  string
		desc string
	}{
		{"admin", "data1", "read", "admin can read data1"},
		{"admin", "data1", "write", "admin can write data1"},
		{"admin", "data2", "read", "admin can read data2"},
		{"admin", "data2", "write", "admin can write data2"},
	}

	for _, test := range adminTests {
		allowed, err = enforcer.Enforce(ctx, test.sub, test.obj, test.act)
		if err != nil {
			fmt.Printf("Failed to check permission: %v\n", err)
			continue
		}
		fmt.Printf("%s: %v\n", test.desc, allowed)
	}

	// Demo 4: Add a new policy
	fmt.Println("\n--- Demo 4: Add New Policy ---")
	err = enforcer.AddPolicy(ctx, "charlie", "data3", "read")
	if err != nil {
		fmt.Printf("Failed to add policy: %v\n", err)
		return
	}
	fmt.Println("Added policy: charlie can read data3")

	// Check the new policy
	allowed, err = enforcer.Enforce(ctx, "charlie", "data3", "read")
	if err != nil {
		fmt.Printf("Failed to check permission: %v\n", err)
		return
	}
	fmt.Printf("charlie can read data3: %v\n", allowed)

	// Demo 5: Remove a policy
	fmt.Println("\n--- Demo 5: Remove Policy ---")
	err = enforcer.RemovePolicy(ctx, "charlie", "data3", "read")
	if err != nil {
		fmt.Printf("Failed to remove policy: %v\n", err)
		return
	}
	fmt.Println("Removed policy: charlie can read data3")

	// Check the removed policy
	allowed, err = enforcer.Enforce(ctx, "charlie", "data3", "read")
	if err != nil {
		fmt.Printf("Failed to check permission: %v\n", err)
		return
	}
	fmt.Printf("charlie can read data3 after removal: %v\n", allowed)

	// Demo 6: Get all policies
	fmt.Println("\n--- Demo 6: Get All Policies ---")
	policies, _ := enforcer.GetPolicy(ctx)
	fmt.Printf("Total policies: %d\n", len(policies))
	for _, p := range policies {
		fmt.Printf("  - %v\n", p)
	}

	// Demo 7: Get roles for a user
	fmt.Println("\n--- Demo 7: Get Roles ---")
	roles, err := enforcer.GetRolesForUser(ctx, "alice")
	if err != nil {
		fmt.Printf("Failed to get roles: %v\n", err)
		return
	}
	fmt.Printf("Roles for alice: %v\n", roles)

	// Demo 8: Get users for a role
	fmt.Println("\n--- Demo 8: Get Users for Role ---")
	users, err := enforcer.GetUsersForRole(ctx, "admin")
	if err != nil {
		fmt.Printf("Failed to get users: %v\n", err)
		return
	}
	fmt.Printf("Users with admin role: %v\n", users)

	// Demo 9: Add a role
	fmt.Println("\n--- Demo 9: Add Role ---")
	_, err = enforcer.AddRoleForUser(ctx, "bob", "admin")
	if err != nil {
		fmt.Printf("Failed to add role: %v\n", err)
		return
	}
	fmt.Println("Added role: bob is now admin")

	// Check Bob's new permissions
	allowed, err = enforcer.Enforce(ctx, "bob", "data1", "read")
	if err != nil {
		fmt.Printf("Failed to check permission: %v\n", err)
		return
	}
	fmt.Printf("bob can read data1 (as admin): %v\n", allowed)

	// Demo 10: Remove a role
	fmt.Println("\n--- Demo 10: Remove Role ---")
	_, err = enforcer.DeleteRoleForUser(ctx, "bob", "admin")
	if err != nil {
		fmt.Printf("Failed to remove role: %v\n", err)
		return
	}
	fmt.Println("Removed role: bob is no longer admin")

	// Check Bob's permissions after removing role
	allowed, err = enforcer.Enforce(ctx, "bob", "data1", "read")
	if err != nil {
		fmt.Printf("Failed to check permission: %v\n", err)
		return
	}
	fmt.Printf("bob can read data1 after role removal: %v\n", allowed)

	fmt.Println("\n=== Example completed successfully ===")
}
