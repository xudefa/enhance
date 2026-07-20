# Elasticsearch Starter

Elasticsearch 搜索引擎自动配置模块，提供全文搜索和数据分析支持。

## 功能特性

- ✅ 自动配置 Elasticsearch 客户端
- ✅ 支持多节点集群
- ✅ 支持认证
- ✅ 连接池管理
- ✅ 便捷的索引和搜索方法

## 快速开始

### 1. 引入依赖

```go
import (
    "github.com/xudefa/enhance/starter/elasticsearch"
)
```

### 2. 配置文件

在 `application.json` 中添加 Elasticsearch 配置：

```json
{
  "elasticsearch": {
    "enabled": true,
    "addresses": ["localhost:9200"],
    "username": "",
    "password": "",
    "max-idle-conns": 10,
    "timeout": 30
  }
}
```

### 3. 使用示例

```go
package main

import (
    "bytes"
    "context"
    "encoding/json"
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/core"
    "github.com/elastic/go-elasticsearch/v8"
)

type Document struct {
    Title   string `json:"title"`
    Content string `json:"content"`
    Author  string `json:"author"`
}

func main() {
    app, _ := boot.NewApplication(
        boot.WithAppName("elasticsearch-demo"),
    )
    defer app.Stop()
    
    // 获取 Elasticsearch Client 实例
    client := core.MustGetBean[*elasticsearch.Client](app.Container())
    
    ctx := context.Background()
    
    // 创建索引
    doc := Document{
        Title:   "Elasticsearch 入门",
        Content: "Elasticsearch 是一个强大的搜索引擎",
        Author:  "John",
    }
    
    data, _ := json.Marshal(doc)
    res, err := client.Index("articles", bytes.NewReader(data))
    if err != nil {
        // 处理错误
    }
    defer res.Body.Close()
    
    // 搜索文档
    query := map[string]interface{}{
        "query": map[string]interface{}{
            "match": map[string]interface{}{
                "title": "Elasticsearch",
            },
        },
    }
    
    buf := new(bytes.Buffer)
    json.NewEncoder(buf).Encode(query)
    
    res, err = client.Search(
        client.Search.WithIndex("articles"),
        client.Search.WithBody(buf),
    )
    if err != nil {
        // 处理错误
    }
    defer res.Body.Close()
}
```

## 配置说明

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `elasticsearch.enabled` | bool | false | 是否启用 Elasticsearch |
| `elasticsearch.addresses` | []string | [localhost:9200] | 节点地址列表 |
| `elasticsearch.username` | string | "" | 用户名 |
| `elasticsearch.password` | string | "" | 密码 |
| `elasticsearch.max-idle-conns` | int | 10 | 最大空闲连接数 |
| `elasticsearch.timeout` | int | 30 | 请求超时（秒） |

## 高级用法

### 批量操作

```go
client := core.MustGetBean[*elasticsearch.Client](app.Container())

// 批量索引文档
var buf bytes.Buffer
for _, doc := range documents {
    meta := []byte(fmt.Sprintf(`{"index":{"_index":"articles"}}`))
    data, _ := json.Marshal(doc)
    
    buf.Write(meta)
    buf.WriteByte('\n')
    buf.Write(data)
    buf.WriteByte('\n')
}

res, err := client.Bulk(&buf, client.Bulk.WithIndex("articles"))
```

### 复杂查询

```go
query := map[string]interface{}{
    "query": map[string]interface{}{
        "bool": map[string]interface{}{
            "must": []map[string]interface{}{
                {"match": map[string]interface{}{"title": "Elasticsearch"}},
                {"range": map[string]interface{}{
                    "created_at": map[string]interface{}{
                        "gte": "2024-01-01",
                    },
                }},
            },
            "filter": []map[string]interface{}{
                {"term": map[string]interface{}{"status": "published"}},
            },
        },
    },
    "sort": []map[string]interface{}{
        {"created_at": map[string]interface{}{"order": "desc"}},
    },
    "from": 0,
    "size": 10,
}
```

## 启动顺序

- **优先级**: `OrderPriorityDataLayer` (-2000)
- **触发条件**: `elasticsearch.enabled=true`

## 依赖

- `github.com/elastic/go-elasticsearch/v8`