// Package main demonstrates the MongoDB starter usage.
//
// This example shows how to use the MongoDB starter to:
// 1. Auto-configure MongoDB connection
// 2. Perform CRUD operations
// 3. Use collections and documents
//
// Prerequisites:
// - MongoDB server running on localhost:27017
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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	_ "github.com/xudefa/enhance/starter/mongodb"
)

// User represents a user document
type User struct {
	ID    string `bson:"_id,omitempty" json:"id"`
	Name  string `bson:"name" json:"name"`
	Email string `bson:"email" json:"email"`
	Age   int    `bson:"age" json:"age"`
}

func main() {
	fmt.Println("=== MongoDB Starter Example ===")
	fmt.Println()

	app := newApp()
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Warning: MongoDB connection failed: %v\n", err)
		fmt.Println("This example requires a running MongoDB server.")
		fmt.Println("Please start MongoDB and try again.")
		return
	}

	client, err := getClient(app)
	if err != nil {
		return
	}
	ctx := context.Background()
	collection := client.Database("enhance").Collection("users")

	if err := demoInsertDocuments(ctx, collection); err != nil {
		return
	}
	if err := demoQueryDocuments(ctx, collection); err != nil {
		return
	}
	if err := demoUpdateDeleteCount(ctx, collection); err != nil {
		return
	}
	if err := demoCreateIndex(ctx, collection); err != nil {
		return
	}

	fmt.Println("\n=== Example completed successfully ===")
}

// newApp 创建应用实例，配置名称与激活的 Profile。
func newApp() *boot.Boot {
	app, err := boot.NewApplication(
		boot.WithAppName("mongodb-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// getClient 从容器的 Bean 中获取 MongoDB 客户端。
func getClient(app *boot.Boot) (*mongo.Client, error) {
	client, err := core.GetByName[*mongo.Client](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get MongoDB client: %v\n", err)
		return nil, fmt.Errorf("Failed to get MongoDB client: %w", err)
	}
	return client, nil
}

// demoInsertDocuments 演示插入单条与多条文档（Demo 1-2）。
func demoInsertDocuments(ctx context.Context, collection *mongo.Collection) error {
	// Demo 1: Insert a document
	fmt.Println("--- Demo 1: Insert Document ---")
	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	insertResult, err := collection.InsertOne(ctx, user)
	if err != nil {
		fmt.Printf("Failed to insert document: %v\n", err)
		return fmt.Errorf("Failed to insert document: %w", err)
	}
	fmt.Printf("Inserted document with ID: %v\n", insertResult.InsertedID)

	// Demo 2: Insert multiple documents
	fmt.Println("\n--- Demo 2: Insert Multiple Documents ---")
	users := []interface{}{
		User{Name: "Jane Doe", Email: "jane@example.com", Age: 25},
		User{Name: "Bob Smith", Email: "bob@example.com", Age: 35},
		User{Name: "Alice Johnson", Email: "alice@example.com", Age: 28},
	}

	insertManyResult, err := collection.InsertMany(ctx, users)
	if err != nil {
		fmt.Printf("Failed to insert documents: %v\n", err)
		return fmt.Errorf("Failed to insert documents: %w", err)
	}
	fmt.Printf("Inserted %d documents with IDs: %v\n", len(insertManyResult.InsertedIDs), insertManyResult.InsertedIDs)
	return nil
}

// demoQueryDocuments 演示查找单个与多个文档（Demo 3-4）。
func demoQueryDocuments(ctx context.Context, collection *mongo.Collection) error {
	// Demo 3: Find a document
	fmt.Println("\n--- Demo 3: Find Document ---")
	var foundUser User
	err := collection.FindOne(ctx, bson.M{"name": "John Doe"}).Decode(&foundUser)
	if err != nil {
		fmt.Printf("Failed to find document: %v\n", err)
		return fmt.Errorf("Failed to find document: %w", err)
	}
	fmt.Printf("Found user: %+v\n", foundUser)

	// Demo 4: Find multiple documents
	fmt.Println("\n--- Demo 4: Find Multiple Documents ---")
	cursor, err := collection.Find(ctx, bson.M{"age": bson.M{"$gte": 25}})
	if err != nil {
		fmt.Printf("Failed to find documents: %v\n", err)
		return fmt.Errorf("Failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var foundUsers []User
	if err := cursor.All(ctx, &foundUsers); err != nil {
		fmt.Printf("Failed to decode documents: %v\n", err)
		return fmt.Errorf("Failed to decode documents: %w", err)
	}
	fmt.Printf("Found %d users with age >= 25\n", len(foundUsers))
	for _, u := range foundUsers {
		fmt.Printf("  - %s (age: %d)\n", u.Name, u.Age)
	}
	return nil
}

// demoUpdateDeleteCount 演示更新、删除与统计文档（Demo 5-7）。
func demoUpdateDeleteCount(ctx context.Context, collection *mongo.Collection) error {
	// Demo 5: Update a document
	fmt.Println("\n--- Demo 5: Update Document ---")
	updateResult, err := collection.UpdateOne(
		ctx,
		bson.M{"name": "John Doe"},
		bson.M{"$set": bson.M{"age": 31}},
	)
	if err != nil {
		fmt.Printf("Failed to update document: %v\n", err)
		return fmt.Errorf("Failed to update document: %w", err)
	}
	fmt.Printf("Updated %d document(s)\n", updateResult.MatchedCount)

	// Demo 6: Delete a document
	fmt.Println("\n--- Demo 6: Delete Document ---")
	deleteResult, err := collection.DeleteOne(ctx, bson.M{"name": "Bob Smith"})
	if err != nil {
		fmt.Printf("Failed to delete document: %v\n", err)
		return fmt.Errorf("Failed to delete document: %w", err)
	}
	fmt.Printf("Deleted %d document(s)\n", deleteResult.DeletedCount)

	// Demo 7: Count documents
	fmt.Println("\n--- Demo 7: Count Documents ---")
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		fmt.Printf("Failed to count documents: %v\n", err)
		return fmt.Errorf("Failed to count documents: %w", err)
	}
	fmt.Printf("Total documents in collection: %d\n", count)
	return nil
}

// demoCreateIndex 演示为字段创建唯一索引（Demo 8）。
func demoCreateIndex(ctx context.Context, collection *mongo.Collection) error {
	// Demo 8: Create index
	fmt.Println("\n--- Demo 8: Create Index ---")
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	indexName, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		fmt.Printf("Failed to create index: %v\n", err)
		return fmt.Errorf("Failed to create index: %w", err)
	}
	fmt.Printf("Created index: %s\n", indexName)
	return nil
}
