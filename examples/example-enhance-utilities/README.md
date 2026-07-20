# Enhance Utilities 示例

本示例演示如何使用 enhance 框架的实用工具模块：Validator、RateLimiter、Cron 和 Asynq。

## 功能特性

- ✅ 数据验证（Validator）
- ✅ 请求限流（RateLimiter）
- ✅ 定时任务（Cron）
- ✅ 异步任务队列（Asynq）

## 快速开始

### 1. 安装依赖

```bash
cd examples/example-enhance-utilities
go mod tidy
```

### 2. 运行示例

```bash
go run main.go
```

### 3. 测试 API

```bash
# 创建用户（带验证）
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "email": "zhangsan@example.com",
    "age": 30,
    "phone": "13800138000"
  }'

# 健康检查
curl http://localhost:8080/api/health

# 发送邮件任务
curl -X POST http://localhost:8080/api/tasks/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Welcome",
    "body": "Welcome to our platform!"
  }'
```

## 配置说明

配置文件位于 `config/application.json`：

```json
{
  "validator": {
    "enabled": true,
    "enable-custom-validators": true
  },
  "ratelimiter": {
    "enabled": true,
    "rate": 100.0,
    "burst": 200
  },
  "cron": {
    "enabled": true,
    "with-logger": true
  },
  "asynq": {
    "enabled": true,
    "host": "localhost",
    "port": 6379
  }
}
```

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/users | 创建用户（带验证） |
| GET | /api/health | 健康检查 |
| POST | /api/tasks/email | 创建邮件发送任务 |

## 验证规则

| 字段 | 规则 | 说明 |
|------|------|------|
| name | required, min=2, max=50 | 必填，长度 2-50 |
| email | required, email | 必填，邮箱格式 |
| age | required, min=1, max=130 | 必填，年龄 1-130 |
| phone | required, phone | 必填，手机号格式 |