// Package main 演示本地缓存的使用
//
// 该示例展示：
// - LRU 缓存
// - 分片缓存
// - 缓存过期策略
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xudefa/enhance/cache"
)

// User 用户模型
type User struct {
	ID    int
	Name  string
	Email string
}

func main() {
	ctx := context.Background()

	fmt.Println("=== 本地缓存示例 ===")
	fmt.Println()

	// 1. LRU 缓存示例
	fmt.Println("1. LRU 缓存示例:")
	lruCache := cache.NewLRUCache(100) // 最大 100 个条目

	// 设置缓存
	lruCache.Set(ctx, "user:1", &User{ID: 1, Name: "张三", Email: "zhangsan@example.com"}, 0)
	lruCache.Set(ctx, "user:2", &User{ID: 2, Name: "李四", Email: "lisi@example.com"}, 0)
	lruCache.Set(ctx, "user:3", &User{ID: 3, Name: "王五", Email: "wangwu@example.com"}, 0)

	// 获取缓存
	if val, err := lruCache.Get(ctx, "user:1"); err == nil {
		user := val.(*User)
		fmt.Printf("   获取 user:1 -> %s (%s)\n", user.Name, user.Email)
	}

	fmt.Printf("   缓存大小: %d\n", lruCache.Len())
	fmt.Println()

	// 2. 分片缓存示例
	fmt.Println("2. 分片缓存示例:")
	shardedCache := cache.NewShardedLRUCache(1000, 16) // 1000 条目，16 个分片

	// 设置带过期的缓存
	shardedCache.Set(ctx, "session:abc123", map[string]any{
		"user_id": 1,
		"role":    "admin",
	}, time.Minute)

	if val, err := shardedCache.Get(ctx, "session:abc123"); err == nil {
		session := val.(map[string]any)
		fmt.Printf("   获取 session -> user_id: %v, role: %v\n", session["user_id"], session["role"])
	}

	fmt.Printf("   缓存大小: %d\n", shardedCache.Len())
	fmt.Println()

	// 3. 缓存统计
	fmt.Println("3. 缓存统计:")
	fmt.Printf("   LRU 缓存条目数: %d\n", lruCache.Len())
	fmt.Printf("   分片缓存条目数: %d\n", shardedCache.Len())

	fmt.Println()
	fmt.Println("=== 示例完成 ===")
}
