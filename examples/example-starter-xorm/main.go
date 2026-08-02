// Package main demonstrates the XORM starter usage.
//
// This example shows how to use the XORM starter to:
// 1. Auto-configure XORM engine
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
	"xorm.io/xorm"

	_ "github.com/xudefa/enhance/starter/xorm"
)

// User represents a user model
type User struct {
	Id        int64     `xorm:"int pk autoincr" json:"id"`
	Name      string    `xorm:"varchar(100) not null" json:"name"`
	Email     string    `xorm:"varchar(200) unique" json:"email"`
	Age       int       `xorm:"int" json:"age"`
	CreatedAt time.Time `xorm:"created" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated" json:"updated_at"`
	DeletedAt time.Time `xorm:"deleted" json:"-"`
}

func main() {
	fmt.Println("=== XORM Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("xorm-example"),
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

	// Get the XORM engine from container
	engine, err := core.GetByName[*xorm.Engine](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get XORM engine: %v\n", err)
		return
	}

	// Demo 1: Sync table structure
	fmt.Println("--- Demo 1: Sync Table Structure ---")
	if err := engine.Sync(new(User)); err != nil {
		fmt.Printf("Failed to sync table: %v\n", err)
		return
	}
	fmt.Println("Table 'user' synchronized")

	// Demo 2: Insert a user
	fmt.Println("\n--- Demo 2: Insert User ---")
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}
	_, err = engine.Insert(&user)
	if err != nil {
		fmt.Printf("Failed to insert user: %v\n", err)
		return
	}
	fmt.Printf("Inserted user: %+v\n", user)

	// Demo 3: Insert multiple users
	fmt.Println("\n--- Demo 3: Insert Multiple Users ---")
	users := []User{
		{Name: "Jane Doe", Email: "jane@example.com", Age: 25},
		{Name: "Bob Smith", Email: "bob@example.com", Age: 35},
		{Name: "Alice Johnson", Email: "alice@example.com", Age: 28},
	}
	_, err = engine.Insert(&users)
	if err != nil {
		fmt.Printf("Failed to insert users: %v\n", err)
		return
	}
	fmt.Printf("Inserted %d users\n", len(users))

	// Demo 4: Find a user
	fmt.Println("\n--- Demo 4: Find User ---")
	var foundUser User
	_, err = engine.Where("name = ?", "John Doe").Get(&foundUser)
	if err != nil {
		fmt.Printf("Failed to find user: %v\n", err)
		return
	}
	fmt.Printf("Found user: %+v\n", foundUser)

	// Demo 5: Find multiple users
	fmt.Println("\n--- Demo 5: Find Multiple Users ---")
	var foundUsers []User
	err = engine.Where("age > ?", 25).Find(&foundUsers)
	if err != nil {
		fmt.Printf("Failed to find users: %v\n", err)
		return
	}
	fmt.Printf("Found %d users with age > 25\n", len(foundUsers))
	for _, u := range foundUsers {
		fmt.Printf("  - %s (age: %d)\n", u.Name, u.Age)
	}

	// Demo 6: Update a user
	fmt.Println("\n--- Demo 6: Update User ---")
	_, err = engine.ID(foundUser.Id).Cols("age").Update(&User{Age: 31})
	if err != nil {
		fmt.Printf("Failed to update user: %v\n", err)
		return
	}
	fmt.Printf("Updated user age to 31\n")

	// Demo 7: Delete a user
	fmt.Println("\n--- Demo 7: Delete User ---")
	_, err = engine.Where("email = ?", "bob@example.com").Delete(&User{})
	if err != nil {
		fmt.Printf("Failed to delete user: %v\n", err)
		return
	}
	fmt.Println("Deleted user with email bob@example.com")

	// Demo 8: Count users
	fmt.Println("\n--- Demo 8: Count Users ---")
	count, err := engine.Count(&User{})
	if err != nil {
		fmt.Printf("Failed to count users: %v\n", err)
		return
	}
	fmt.Printf("Total users: %d\n", count)

	// Demo 9: Use transactions
	fmt.Println("\n--- Demo 9: Transaction ---")
	session := engine.NewSession()
	defer session.Close()

	if err := session.Begin(); err != nil {
		fmt.Printf("Failed to begin transaction: %v\n", err)
		return
	}

	// Create a user in transaction
	txUser := User{
		Name:  "Transaction User",
		Email: "transaction@example.com",
		Age:   40,
	}
	if _, err := session.Insert(&txUser); err != nil {
		session.Rollback()
		fmt.Printf("Failed to insert user in transaction: %v\n", err)
		return
	}

	// Commit transaction
	if err := session.Commit(); err != nil {
		fmt.Printf("Failed to commit transaction: %v\n", err)
		return
	}
	fmt.Println("Transaction committed successfully")

	// Demo 10: Raw query
	fmt.Println("\n--- Demo 10: Raw Query ---")
	var rawUsers []User
	err = engine.SQL("SELECT * FROM user WHERE age > ?", 30).Find(&rawUsers)
	if err != nil {
		fmt.Printf("Failed to execute raw query: %v\n", err)
		return
	}
	fmt.Printf("Found %d users with age > 30\n", len(rawUsers))

	// Demo 11: Use conditions
	fmt.Println("\n--- Demo 11: Use Conditions ---")
	var conditionUsers []User
	err = engine.Where("age > ?", 25).And("age < ?", 40).Find(&conditionUsers)
	if err != nil {
		fmt.Printf("Failed to find users with conditions: %v\n", err)
		return
	}
	fmt.Printf("Found %d users with 25 < age < 40\n", len(conditionUsers))

	// Demo 12: Use ordering
	fmt.Println("\n--- Demo 12: Use Ordering ---")
	var orderedUsers []User
	err = engine.Asc("age").Find(&orderedUsers)
	if err != nil {
		fmt.Printf("Failed to find users with ordering: %v\n", err)
		return
	}
	fmt.Println("Users ordered by age (ascending):")
	for _, u := range orderedUsers {
		fmt.Printf("  - %s (age: %d)\n", u.Name, u.Age)
	}

	fmt.Println("\n=== Example completed successfully ===")
}
