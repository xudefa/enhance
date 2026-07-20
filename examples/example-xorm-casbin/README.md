# XORM + Casbin-XORM 集成示例

本示例演示如何将 XORM 数据库 ORM 与 Casbin 权限管理集成，实现基于数据库的 RBAC 访问控制。

## 功能特性

- **XORM 数据库集成**: 自动配置数据库连接、连接池管理
- **Casbin-XORM 集成**: 将 Casbin 策略持久化到 MySQL 数据库
- **JWT 认证**: 基于 Token 的用户认证
- **安全过滤器链**: 自动化的 URL 访问控制
- **策略管理 API**: 运行时动态添加/删除权限策略

## 快速开始

### 1. 准备数据库

```sql
CREATE DATABASE demo DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE USER 'scott'@'localhost' IDENTIFIED BY '123456';
GRANT ALL PRIVILEGES ON demo.* TO 'scott'@'localhost';
FLUSH PRIVILEGES;
```

### 2. 修改配置

编辑 `config/application.json`，根据你的数据库配置修改：

```json
{
  "db": {
    "xorm": {
      "host": "127.0.0.1",
      "port": 3306,
      "username": "scott",
      "password": "123456",
      "database": "demo"
    }
  }
}
```

### 3. 运行示例

```bash
cd examples/example-xorm-casbin
go mod tidy
go run .
```

### 4. 测试 API

#### 登录获取 Token

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"test"}'
```

#### 访问受保护的资源

```bash
# 使用返回的 token
TOKEN="your-token-here"

# 访问用户资料（需要认证）
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer $TOKEN"

# 管理员接口（需要 admin 角色）
curl http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN"
```

#### 管理 Casbin 策略

```bash
# 获取所有策略
curl http://localhost:8080/api/casbin/policies \
  -H "Authorization: Bearer $TOKEN"

# 添加策略
curl -X POST http://localhost:8080/api/casbin/policy \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"subject":"user","object":"/api/data","action":"GET"}'

# 删除策略
curl -X DELETE http://localhost:8080/api/casbin/policy \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"subject":"user","object":"/api/data","action":"GET"}'
```

## 配置说明

### XORM 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| db.xorm.enabled | 是否启用 XORM | false |
| db.xorm.type | 数据库类型 | mysql |
| db.xorm.host | 数据库主机 | localhost |
| db.xorm.port | 数据库端口 | 3306 |
| db.xorm.username | 用户名 | scott |
| db.xorm.password | 密码 | 123456 |
| db.xorm.database | 数据库名 | demo |
| db.xorm.show-sql | 显示 SQL 日志 | false |

### Casbin 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| security.casbin.enabled | 是否启用 Casbin | false |
| security.casbin.policy-type | 策略存储类型 | xorm |
| security.casbin.model-path | 模型文件路径 | config/casbin_model.conf |
| security.casbin.table-name | 策略表名 | casbin_rule |
| security.casbin.auto-create-table | 自动创建表 | true |
| security.casbin.auto-load | 自动刷新策略 | false |

## 架构说明

### 自动配置执行顺序

1. **XORM 配置** (Order: -2000): 创建数据库连接
2. **Casbin-XORM 配置** (Order: -1300): 创建基于 XORM 的 Casbin 适配器
3. **Casbin 基础配置** (Order: -1200): 加载模型和策略
4. **JWT 认证** (Order: -1500): 配置 JWT Token 验证
5. **Security 配置** (Order: -100): 配置安全过滤器链

### 依赖关系

```
XORM AutoConfiguration (-2000)
    ↓
Casbin-XORM AutoConfiguration (-1300)
    ↓
Casbin AutoConfiguration (-1200)
    ↓
Security AutoConfiguration (-100)
```

## 测试

运行单元测试：

```bash
# 测试 xorm starter
cd starter/xorm
go test -v ./...

# 测试 casbin-xorm starter
cd starter/casbin-xorm
go test -v ./...
```