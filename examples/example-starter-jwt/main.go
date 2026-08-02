// Package main demonstrates the JWT starter usage.
//
// This example shows how to use the JWT starter to:
// 1. Auto-configure JWT token provider
// 2. Generate and validate JWT tokens
// 3. Parse and refresh tokens
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
	"github.com/xudefa/enhance/starter/jwt"
)

func main() {
	fmt.Println("=== JWT Starter Example ===")
	fmt.Println()

	app, err := boot.NewApplication(
		boot.WithAppName("jwt-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	tokenProvider, err := core.GetByName[*jwt.DefaultTokenProvider](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get token provider: %v\n", err)
		return
	}

	ctx := context.Background()

	// Demo 1: Generate a token
	fmt.Println("--- Demo 1: Generate Token ---")
	accessToken, err := tokenProvider.GenerateToken(ctx, "user-123", []string{"admin", "user"})
	if err != nil {
		fmt.Printf("Failed to generate token: %v\n", err)
		return
	}
	fmt.Printf("Access Token generated: %s...\n", accessToken[:min(50, len(accessToken))])

	// Demo 2: Parse token to get claims
	fmt.Println("\n--- Demo 2: Parse Token ---")
	claims, err := tokenProvider.ParseToken(ctx, accessToken)
	if err != nil {
		fmt.Printf("Failed to parse token: %v\n", err)
		return
	}
	fmt.Printf("Token Claims:\n")
	fmt.Printf("  Subject: %s\n", claims.Subject)
	fmt.Printf("  Authorities: %v\n", claims.Authorities)

	// Demo 3: Validate token
	fmt.Println("\n--- Demo 3: Validate Token ---")
	err = tokenProvider.ValidateToken(ctx, accessToken)
	if err != nil {
		fmt.Printf("Token validation failed: %v\n", err)
		return
	}
	fmt.Println("Token is valid!")

	// Demo 4: Refresh token
	fmt.Println("\n--- Demo 4: Refresh Token ---")
	newToken, err := tokenProvider.RefreshToken(ctx, accessToken)
	if err != nil {
		fmt.Printf("Failed to refresh token: %v\n", err)
		return
	}
	fmt.Printf("New Token generated: %s...\n", newToken[:min(50, len(newToken))])

	fmt.Println("\n=== Example completed successfully ===")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
