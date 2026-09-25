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

	app := newApp()
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Warning: Elasticsearch connection failed: %v\n", err)
		fmt.Println("This example requires a running Elasticsearch server.")
		fmt.Println("Please start Elasticsearch and try again.")
		return
	}

	es, err := getESClient(app)
	if err != nil {
		return
	}

	if err := demoClusterHealth(es); err != nil {
		return
	}
	if err := demoIndexing(es); err != nil {
		return
	}
	if err := demoIndexMany(es); err != nil {
		return
	}
	if err := demoSearch(es); err != nil {
		return
	}
	if err := demoDocumentOps(es); err != nil {
		return
	}
	if err := demoDeleteIndex(es); err != nil {
		return
	}

	fmt.Println("\n=== Example completed successfully ===")
}

// newApp 创建应用实例，配置名称与激活的 Profile。
func newApp() *boot.Boot {
	app, err := boot.NewApplication(
		boot.WithAppName("elasticsearch-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// getESClient 从容器的 Bean 中获取 Elasticsearch 客户端。
func getESClient(app *boot.Boot) (*elasticsearch.Client, error) {
	es, err := core.GetByName[*elasticsearch.Client](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get Elasticsearch client: %v\n", err)
		return nil, fmt.Errorf("Failed to get Elasticsearch client: %w", err)
	}
	return es, nil
}

// demoClusterHealth 演示查看集群健康状态（Demo 1）。
func demoClusterHealth(es *elasticsearch.Client) error {
	// Demo 1: Check cluster health
	fmt.Println("--- Demo 1: Cluster Health ---")
	res, err := es.Cluster.Health()
	if err != nil {
		fmt.Printf("Failed to get cluster health: %v\n", err)
		return fmt.Errorf("Failed to get cluster health: %w", err)
	}
	defer res.Body.Close()
	fmt.Printf("Cluster status: %s\n", res.Status())
	return nil
}

// demoIndexing 演示创建索引与写入单条文档（Demo 2-3）。
func demoIndexing(es *elasticsearch.Client) error {
	// Demo 2: Create an index
	fmt.Println("\n--- Demo 2: Create Index ---")
	createIndexRes, err := es.Indices.Create("users")
	if err != nil {
		fmt.Printf("Failed to create index: %v\n", err)
		return fmt.Errorf("Failed to create index: %w", err)
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
		return fmt.Errorf("Failed to index document: %w", err)
	}
	defer indexRes.Body.Close()
	fmt.Println("Document indexed")
	return nil
}

// demoIndexMany 演示批量写入多份文档（Demo 4）。
func demoIndexMany(es *elasticsearch.Client) error {
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
	return nil
}

// demoSearch 演示按条件搜索文档（Demo 5）。
func demoSearch(es *elasticsearch.Client) error {
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
		return fmt.Errorf("Failed to search documents: %w", err)
	}
	defer searchRes.Body.Close()
	fmt.Println("Search completed")
	return nil
}

// demoDocumentOps 演示按 ID 获取与删除文档（Demo 6-7）。
func demoDocumentOps(es *elasticsearch.Client) error {
	// Demo 6: Get a document by ID
	fmt.Println("\n--- Demo 6: Get Document ---")
	getRes, err := es.Get("users", "1")
	if err != nil {
		fmt.Printf("Failed to get document: %v\n", err)
		return fmt.Errorf("Failed to get document: %w", err)
	}
	defer getRes.Body.Close()
	fmt.Println("Document retrieved")

	// Demo 7: Delete a document
	fmt.Println("\n--- Demo 7: Delete Document ---")
	deleteRes, err := es.Delete("users", "1")
	if err != nil {
		fmt.Printf("Failed to delete document: %v\n", err)
		return fmt.Errorf("Failed to delete document: %w", err)
	}
	defer deleteRes.Body.Close()
	fmt.Println("Document deleted")
	return nil
}

// demoDeleteIndex 演示删除索引并执行收尾清理（Demo 8）。
func demoDeleteIndex(es *elasticsearch.Client) error {
	// Demo 8: Delete an index
	fmt.Println("\n--- Demo 8: Delete Index ---")
	deleteIndexRes, err := es.Indices.Delete([]string{"users"})
	if err != nil {
		fmt.Printf("Failed to delete index: %v\n", err)
		return fmt.Errorf("Failed to delete index: %w", err)
	}
	defer deleteIndexRes.Body.Close()
	fmt.Println("Index 'users' deleted")

	// Cleanup
	_, _ = es.Indices.Delete([]string{"users"})
	return nil
}
