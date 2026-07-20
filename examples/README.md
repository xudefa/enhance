# examples — 集成示例和测试

> **说明**: 完整的集成示例和测试用例，展示 enhance 框架与第三方库的集成

## 概述

本目录包含 enhance 框架与第三方库集成的完整示例和集成测试，帮助开发者快速上手和理解框架的使用方式。

### 示例列表

| 示例 | 说明 |
|------|------|
| **example-enhance-full-integration** | 完整集成示例：zerolog + jwt + casbin + security + gorm |
| **example-enhance-jwt-casbin** | JWT 认证 + Casbin 授权示例 |
| **example-enhance-kafka-otel** | Kafka + OpenTelemetry 集成示例 |
| **example-enhance-run** | 数据库集成示例（GORM + MySQL） |
| **example-enhance-utilities** | 工具类集成示例 |
| **example-xorm-casbin** | XORM + Casbin 集成示例 |
| **example-gin-tracing** | Gin + OpenTelemetry 链路追踪 |
| **example-echo-tracing** | Echo + OpenTelemetry 链路追踪 |
| **example-fiber-tracing** | Fiber + OpenTelemetry 链路追踪 |
| **example-chi-tracing** | Chi + OpenTelemetry 链路追踪 |

## 运行示例

### 1. 完整集成示例

```bash
cd example-enhance-full-integration
go run main.go

# 访问 API 端点:
# POST   http://localhost:8080/login
# GET    http://localhost:8080/api/profile
# GET    http://localhost:8080/api/admin/users
```

### 2. JWT + Casbin 示例

```bash
cd example-enhance-jwt-casbin
go run main.go
```

### 3. Kafka + OpenTelemetry 示例

```bash
cd example-enhance-kafka-otel
go run main.go
```

### 4. 数据库集成示例

```bash
cd example-enhance-run
go run main.go
```

### 5. XORM + Casbin 示例

```bash
cd example-xorm-casbin
go run main.go
```

## 运行测试

### 运行所有测试

```bash
cd examples
go test ./...
```

### 运行特定测试

```bash
# 运行所有示例的测试
cd examples
go test -v ./...

# 运行特定示例的测试
go test -v ./example-enhance-full-integration/...
go test -v ./example-gin-tracing/...

# 运行基准测试
go test -bench=. -benchmem ./example-enhance-full-integration/...
```

### 运行短测试

```bash
go test -short ./...
```

## 添加新示例

1. 在 `examples/` 下创建新目录
2. 创建 `main.go` 主程序
3. 创建 `*_test.go` 集成测试（如有需要）
4. 创建 `application.json` 配置文件（如需要）
5. 创建 `go.mod` 文件（如有第三方依赖）
6. 更新本 README

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `SKIP_REDIS_TESTS` | 跳过 Redis 测试 | 未设置 |
| `SKIP_MYSQL_TESTS` | 跳过 MySQL 测试 | 未设置 |
| `SKIP_KAFKA_TESTS` | 跳过 Kafka 测试 | 未设置 |
| `SKIP_RABBITMQ_TESTS` | 跳过 RabbitMQ 测试 | 未设置 |
| `SKIP_ES_TESTS` | 跳过 Elasticsearch 测试 | 未设置 |
| `SKIP_MINIO_TESTS` | 跳过 MinIO 测试 | 未设置 |

## 外部服务要求

运行完整测试需要以下外部服务：

| 服务 | 默认地址 | 用途 |
|------|----------|------|
| Redis | localhost:6379 | 缓存测试 |
| MySQL | localhost:3306 | 数据库测试 |
| Kafka | localhost:9092 | 消息队列测试 |
| RabbitMQ | localhost:5672 | 消息队列测试 |
| Elasticsearch | localhost:9200 | 搜索引擎测试 |
| MinIO | localhost:9000 | 对象存储测试 |
| Nacos | localhost:8848 | 配置中心测试 |
| Consul | localhost:8500 | 服务注册测试 |
| etcd | localhost:2379 | 配置中心测试 |

可以使用 Docker Compose 快速启动所有服务：

```bash
docker-compose up -d
```

## 故障排除

### 测试失败

1. 检查外部服务是否运行
2. 设置 `SKIP_XXX_TESTS` 环境变量跳过特定测试
3. 使用 `-short` 标志运行短测试

### 连接被拒绝

1. 检查服务地址和端口
2. 检查防火墙设置
3. 使用 `docker ps` 确认容器运行

### 测试超时

1. 增加超时时间
2. 检查网络连接
3. 确认服务响应正常