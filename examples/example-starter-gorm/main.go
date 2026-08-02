// Package main demonstrates the GORM starter usage.
//
// This example shows how to use the GORM starter to:
// 1. Auto-configure GORM database connection
// 2. Define models
// 3. Perform CRUD operations
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
	"time"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"
	"gorm.io/gorm"

	_ "github.com/xudefa/enhance/starter/gorm"
)

// User represents a user model
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:100;not null"`
	Email     string         `json:"email" gorm:"size:200;uniqueIndex"`
	Age       int            `json:"age"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func main() {
	fmt.Println("=== GORM Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("gorm-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Database connection failed: %v\n", err)
		fmt.Println("This example requires a running MySQL server.")
		fmt.Println("Please start MySQL and try again.")
		return
	}

	// Get the GORM DB from container
	db, err := core.GetByName[*gorm.DB](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get GORM DB: %v\n", err)
		return
	}

	// Demo 1: Auto-migrate
	fmt.Println("--- Demo 1: Auto-Migrate ---")
	if err := db.AutoMigrate(&User{}); err != nil {
		fmt.Printf("Failed to migrate: %v\n", err)
		return
	}
	fmt.Println("Database migrated successfully")

	// Demo 2: Create a user
	fmt.Println("\n--- Demo 2: Create User ---")
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	if err := db.Create(&user).Error; err != nil {
		fmt.Printf("Failed to create user: %v\n", err)
		return
	}
	fmt.Printf("Created user: %+v\n", user)

	// Demo 3: Create multiple users
	fmt.Println("\n--- Demo 3: Create Multiple Users ---")
	users := []User{
		{Name: "Jane Doe", Email: "jane@example.com", Age: 25},
		{Name: "Bob Smith", Email: "bob@example.com", Age: 35},
		{Name: "Alice Johnson", Email: "alice@example.com", Age: 28},
	}
	if err := db.Create(&users).Error; err != nil {
		fmt.Printf("Failed to create users: %v\n", err)
		return
	}
	fmt.Printf("Created %d users\n", len(users))

	// Demo 4: Find a user
	fmt.Println("\n--- Demo 4: Find User ---")
	var foundUser User
	if err := db.Where("name = ?", "John Doe").First(&foundUser).Error; err != nil {
		fmt.Printf("Failed to find user: %v\n", err)
		return
	}
	fmt.Printf("Found user: %+v\n", foundUser)

	// Demo 5: Find multiple users
	fmt.Println("\n--- Demo 5: Find Multiple Users ---")
	var foundUsers []User
	if err := db.Where("age > ?", 25).Find(&foundUsers).Error; err != nil {
		fmt.Printf("Failed to find users: %v\n", err)
		return
	}
	fmt.Printf("Found %d users with age > 25\n", len(foundUsers))
	for _, u := range foundUsers {
		fmt.Printf("  - %s (age: %d)\n", u.Name, u.Age)
	}

	// Demo 6: Update a user
	fmt.Println("\n--- Demo 6: Update User ---")
	if err := db.Model(&foundUser).Update("age", 31).Error; err != nil {
		fmt.Printf("Failed to update user: %v\n", err)
		return
	}
	fmt.Printf("Updated user age to %d\n", foundUser.Age)

	// Demo 7: Save a user
	fmt.Println("\n--- Demo 7: Save User ---")
	foundUser.Name = "John Smith"
	if err := db.Save(&foundUser).Error; err != nil {
		fmt.Printf("Failed to save user: %v\n", err)
		return
	}
	fmt.Printf("Saved user: %+v\n", foundUser)

	// Demo 8: Delete a user
	fmt.Println("\n--- Demo 8: Delete User ---")
	if err := db.Delete(&User{}, "email = ?", "bob@example.com").Error; err != nil {
		fmt.Printf("Failed to delete user: %v\n", err)
		return
	}
	fmt.Println("Deleted user with email bob@example.com")

	// Demo 9: Count users
	fmt.Println("\n--- Demo 9: Count Users ---")
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		fmt.Printf("Failed to count users: %v\n", err)
		return
	}
	fmt.Printf("Total users: %d\n", count)

	// Demo 10: Use transactions
	fmt.Println("\n--- Demo 10: Transaction ---")
	tx := db.Begin()
	if tx.Error != nil {
		fmt.Printf("Failed to begin transaction: %v\n", tx.Error)
		return
	}

	// Create a user in transaction
	txUser := User{
		Name:  "Transaction User",
		Email: "transaction@example.com",
		Age:   40,
	}
	if err := tx.Create(&txUser).Error; err != nil {
		tx.Rollback()
		fmt.Printf("Failed to create user in transaction: %v\n", err)
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		fmt.Printf("Failed to commit transaction: %v\n", err)
		return
	}
	fmt.Println("Transaction committed successfully")

	// Demo 11: Raw query
	fmt.Println("\n--- Demo 11: Raw Query ---")
	var rawUsers []User
	if err := db.Raw("SELECT * FROM users WHERE age > ?", 30).Scan(&rawUsers).Error; err != nil {
		fmt.Printf("Failed to execute raw query: %v\n", err)
		return
	}
	fmt.Printf("Found %d users with age > 30\n", len(rawUsers))

	// Demo 12: Scopes
	fmt.Println("\n--- Demo 12: Scopes ---")
	// Define a scope
	adultScope := func(db *gorm.DB) *gorm.DB {
		return db.Where("age >= ?", 18)
	}

	var adults []User
	if err := db.Scopes(adultScope).Find(&adults).Error; err != nil {
		fmt.Printf("Failed to find adults: %v\n", err)
		return
	}
	fmt.Printf("Found %d adults\n", len(adults))

	fmt.Println("\n=== Example completed successfully ===")
}
