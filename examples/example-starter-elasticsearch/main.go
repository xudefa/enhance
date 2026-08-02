// Package main demonstrates the Elasticsearch starter usage.
//
// This example shows how to use the Elasticsearch starter to:
// 1. Auto-configure Elasticsearch client
// 2. Index documents
// 3. Search documents
//
// Prerequisites:
// - Elasticsearch server running on localhost:9200
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/elasticsearch"
)

func main() {
	fmt.Println("=== Elasticsearch Starter Example ===")
	fmt.Println()

	// Create application with boot
	app, err := boot.NewApplication(
		boot.WithAppName("elasticsearch-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	defer app.Stop()

	// Start the application (triggers auto-configuration)
	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Elasticsearch connection failed: %v\n", err)
		fmt.Println("This example requires a running Elasticsearch server.")
		fmt.Println("Please start Elasticsearch and try again.")
		return
	}

	// Get the Elasticsearch client from container
	es, err := core.GetByName[*elasticsearch.Client](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get Elasticsearch client: %v\n", err)
		return
	}

	// Demo 1: Check cluster health
	fmt.Println("--- Demo 1: Cluster Health ---")
	res, err := es.Cluster.Health()
	if err != nil {
		fmt.Printf("Failed to get cluster health: %v\n", err)
		return
	}
	defer res.Body.Close()
	fmt.Printf("Cluster status: %s\n", res.Status)

	// Demo 2: Create an index
	fmt.Println("\n--- Demo 2: Create Index ---")
	createIndexRes, err := es.Indices.Create("users")
	if err != nil {
		fmt.Printf("Failed to create index: %v\n", err)
		return
	}
	defer createIndexRes.Body.Close()
	fmt.Println("Index 'users' created")

	// Demo 3: Index a document
	fmt.Println("\n--- Demo 3: Index Document ---")
	doc := `{
		"name": "John Doe",
		"email": "john@example.com",
		"age": 30,
		"city": "New York"
	}`
	indexRes, err := es.Index(
		"users",
		strings.NewReader(doc),
	)
	if err != nil {
		fmt.Printf("Failed to index document: %v\n", err)
		return
	}
	defer indexRes.Body.Close()
	fmt.Println("Document indexed")

	// Demo 4: Index multiple documents
	fmt.Println("\n--- Demo 4: Index Multiple Documents ---")
	documents := []string{
		`{"name": "Jane Doe", "email": "jane@example.com", "age": 25, "city": "Los Angeles"}`,
		`{"name": "Bob Smith", "email": "bob@example.com", "age": 35, "city": "Chicago"}`,
		`{"name": "Alice Johnson", "email": "alice@example.com", "age": 28, "city": "New York"}`,
	}

	for i, doc := range documents {
		res, err := es.Index("users", strings.NewReader(doc))
		if err != nil {
			fmt.Printf("Failed to index document %d: %v\n", i+1, err)
			continue
		}
		defer res.Body.Close()
		fmt.Printf("Document %d indexed\n", i+1)
	}

	// Demo 5: Search documents
	fmt.Println("\n--- Demo 5: Search Documents ---")
	searchBody := `{
		"query": {
			"match": {
				"city": "New York"
			}
		}
	}`
	searchRes, err := es.Search(
		es.Search.WithIndex("users"),
		es.Search.WithBody(strings.NewReader(searchBody)),
	)
	if err != nil {
		fmt.Printf("Failed to search documents: %v\n", err)
		return
	}
	defer searchRes.Body.Close()
	fmt.Println("Search completed")

	// Demo 6: Get a document by ID
	fmt.Println("\n--- Demo 6: Get Document ---")
	getRes, err := es.Get("users", "1")
	if err != nil {
		fmt.Printf("Failed to get document: %v\n", err)
		return
	}
	defer getRes.Body.Close()
	fmt.Println("Document retrieved")

	// Demo 7: Delete a document
	fmt.Println("\n--- Demo 7: Delete Document ---")
	deleteRes, err := es.Delete("users", "1")
	if err != nil {
		fmt.Printf("Failed to delete document: %v\n", err)
		return
	}
	defer deleteRes.Body.Close()
	fmt.Println("Document deleted")

	// Demo 8: Delete an index
	fmt.Println("\n--- Demo 8: Delete Index ---")
	deleteIndexRes, err := es.Indices.Delete([]string{"users"})
	if err != nil {
		fmt.Printf("Failed to delete index: %v\n", err)
		return
	}
	defer deleteIndexRes.Body.Close()
	fmt.Println("Index 'users' deleted")

	// Cleanup
	_, _ = es.Indices.Delete([]string{"users"})

	fmt.Println("\n=== Example completed successfully ===")
}
