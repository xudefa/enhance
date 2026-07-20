// Package elasticsearch 提供 Elasticsearch 搜索引擎自动配置。
//
// Elasticsearch 是分布式搜索和分析引擎。
//
// 功能特性：
//   - 自动配置 Elasticsearch 连接
//   - 支持多节点集群
//   - 支持认证
//   - 连接池管理
//
// 配置示例：
//
//	{
//	  "elasticsearch": {
//	    "enabled": true,
//	    "addresses": ["localhost:9200"],
//	    "username": "elastic",
//	    "password": "changeme"
//	  }
//	}
//
// 使用示例：
//
//	client := core.MustGetBean[*elasticsearch.Client](app.Container())
//	res, _ := client.Search(client.Search.WithIndex("my-index"))
package elasticsearch
