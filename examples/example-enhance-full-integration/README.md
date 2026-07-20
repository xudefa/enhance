## 项目介绍

这是一个基于 Enhance 的全功能示例项目，展示了 Enhance 的所有核心功能。

### 功能特性

- **认证与授权**：JWT Token 认证 + Casbin RBAC 权限控制
- **用户管理**：用户 CRUD 操作（创建、查询、更新、删除）
- **数据库操作**：基于 GORM 的 ORM 数据库操作
- **日志记录**：基于 Zerolog 的高性能结构化日志
- **策略管理**：Casbin 策略存储到数据库，支持动态增删查改
- **安全过滤器链**：可配置的 URL 级别访问控制

### 项目结构

```
example-enhance-full-integration/
├── config/
│   ├── application.json      # 应用配置文件
│   ├── casbin_model.conf     # Casbin 权限模型
│   └── casbin_policy.csv     # Casbin 初始策略
├── main.go                   # 主程序入口
├── user_controller.go        # 用户管理控制器
├── casbin_controller.go      # Casbin 策略管理控制器
├── autoconfig.go             # 自动配置
├── go.mod                    # Go 模块依赖
└── README.md                 # 项目文档
```

### 项目运行

1. 确保 MySQL 数据库服务已启动，数据库名称为 `demo`
2. 确保数据库配置正确（`config/application.json` 中的 `db.gorm` 配置）
3. 运行项目：

```bash
go run .
```

4. 服务监听端口：`8080`

### 项目配置

| 配置项 | 位置 | 说明 |
|--------|------|------|
| 数据库连接 | `db.gorm` | MySQL 连接配置 |
| 日志记录 | `log.zerolog` | Zerolog 日志配置 |
| JWT 认证 | `security.jwt` | JWT Secret 和过期时间 |
| Casbin 权限 | `security.casbin` | 策略存储方式（gorm） |
| 安全规则 | `security.rules` | URL 访问控制规则 |

### Casbin GORM 集成

本项目使用 `starter/casbin-gorm` 模块将 Casbin 策略存储到数据库表中，实现策略的持久化管理。

#### 配置说明

```json
{
  "security": {
    "casbin": {
      "enabled": true,
      "model-type": "file",
      "model-path": "config/casbin_model.conf",
      "policy-type": "gorm",
      "table-name": "casbin_rule",
      "auto-create-table": true,
      "auto-load": true,
      "auto-load-interval": 5
    }
  }
}
```

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `model-type` | 模型加载方式（file/string） | file |
| `model-path` | 模型文件路径（model-type=file 时） | - |
| `policy-type` | 策略存储方式，设置为 `gorm` 时使用数据库 | file |
| `table-name` | 策略表名 | casbin_rule |
| `auto-create-table` | 是否自动创建策略表 | true |
| `auto-load` | 是否自动刷新策略 | false |
| `auto-load-interval` | 自动刷新间隔（分钟） | 5 |

#### 数据库表结构

策略存储在 `casbin_rule` 表中，结构如下：

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| ptype | string | 策略类型（p/g） |
| v0 | string | 主体（角色/用户） |
| v1 | string | 对象（资源） |
| v2 | string | 动作（操作） |
| v3-v5 | string | 扩展字段 |

---

## 接口列表

### 认证接口

| 接口 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/login` | POST | 用户登录，返回 JWT Token | 公开 |

### 用户接口

| 接口 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/users` | GET | 获取用户列表 | 认证 |
| `/api/users/{id}` | GET | 获取单个用户 | 认证 |
| `/api/users` | POST | 创建用户 | 认证 |
| `/api/users/{id}` | PUT | 更新用户 | 认证 |
| `/api/users/{id}` | DELETE | 删除用户 | 认证 |

### 安全保护接口

| 接口 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/profile` | GET | 获取当前用户资料 | 认证 |
| `/api/admin/users` | GET | 管理员用户列表 | hasRole('admin') |

### Casbin 策略管理接口

| 接口 | 方法 | 说明 | 权限 |
|------|------|------|------|
| `/api/casbin/policies` | GET | 获取所有策略 | 认证 |
| `/api/casbin/policy` | POST | 添加策略 | 认证 |
| `/api/casbin/policy` | DELETE | 移除策略 | 认证 |
| `/api/casbin/check` | POST | 检查权限 | 认证 |

---

## 完整测试脚本

以下脚本涵盖了所有接口的测试，可以直接复制执行。

### 1. 基础认证测试

```bash
# === 1. alice 登录（获取 admin 角色）===
echo "=== 1. alice 登录 ==="
ALICE_TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"test"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "alice Token: ${ALICE_TOKEN:0:30}..."

# === 2. bob 登录（获取 user 角色）===
echo ""
echo "=== 2. bob 登录 ==="
BOB_TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"test"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "bob Token: ${BOB_TOKEN:0:30}..."
```

### 2. 安全保护接口测试

```bash
# === 3. alice 查看个人资料（需要认证）===
echo ""
echo "=== 3. alice 查看个人资料 ==="
curl -s http://localhost:8080/api/profile \
  -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

# === 4. alice 访问管理员接口（admin 角色，应该成功）===q
echo ""
echo "=== 4. alice 访问管理员接口 ==="
curl -s http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

# === 5. bob 访问管理员接口（user 角色，应该返回 403）===
echo ""
echo "=== 5. bob 访问管理员接口 (应该 403) ==="
curl -s -o /dev/null -w "HTTP Status: %{http_code}" http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $BOB_TOKEN"
echo ""

# === 6. 未登录访问受保护接口（应该返回 401）===
echo ""
echo "=== 6. 未登录访问 /api/profile (应该 401) ==="
curl -s -o /dev/null -w "HTTP Status: %{http_code}" http://localhost:8080/api/profile
echo ""
```

### 3. 用户管理接口测试

```bash
# === 7. 获取用户列表 ===
echo "=== 7. 获取用户列表 ==="
curl -s http://localhost:8080/api/users \
  -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

# === 8. 创建用户 ===
echo ""
echo "=== 8. 创建用户 ==="
curl -s -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"张三","email":"zhangsan@example.com","age":30}' | python3 -m json.tool

# === 9. 获取刚创建的用户（假设 ID 为 1，根据实际情况调整）===
echo ""
echo "=== 9. 获取用户详情 ==="
curl -s http://localhost:8080/api/users/1 \
  -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

# === 10. 更新用户 ===
echo ""
echo "=== 10. 更新用户 ==="
curl -s -X PUT http://localhost:8080/api/users/1 \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"张三更新","email":"zhangsan_new@example.com","age":31}' | python3 -m json.tool

# === 11. 删除用户 ===
echo ""
echo "=== 11. 删除用户 ==="
curl -s -X DELETE http://localhost:8080/api/users/1 \
  -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool
```

### 4. Casbin 策略管理接口测试

```bash
# === 12. 查看当前所有策略 ===
echo "=== 12. 查看 Casbin 策略列表 ==="
curl -s http://localhost:8080/api/casbin/policies \
  -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

# === 13. 添加策略（允许 admin 角色访问 /api/admin/users）===
echo ""
echo "=== 13. 添加策略 ==="
curl -s -X POST http://localhost:8080/api/casbin/policy \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subject":"admin","object":"/api/admin/users","action":"GET"}' | python3 -m json.tool

# === 14. 再次查看策略列表（应该包含新添加的策略）===
echo ""
echo "=== 14. 再次查看策略列表 ==="
curl -s http://localhost:8080/api/casbin/policies \
  -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

# === 15. 权限检查 ===
echo ""
echo "=== 15. 权限检查 ==="
curl -s -X POST http://localhost:8080/api/casbin/check \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subject":"admin","object":"/api/admin/users","action":"GET"}' | python3 -m json.tool

# === 16. 删除策略 ===
echo ""
echo "=== 16. 删除策略 ==="
curl -s -X DELETE http://localhost:8080/api/casbin/policy \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subject":"admin","object":"/api/admin/users","action":"GET"}' | python3 -m json.tool

# === 17. 最终策略列表（应该为空）===
echo ""
echo "=== 17. 最终策略列表 ==="
curl -s http://localhost:8080/api/casbin/policies \
  -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool
```

### 5. 一键完整测试脚本

将以上所有步骤合并为一个脚本，方便快速验证：

```bash
#!/bin/bash
# test_all.sh - 完整测试所有接口

BASE_URL="http://localhost:8080"

echo "=========================================="
echo "  Enhance 全功能示例项目 - 接口测试"
echo "=========================================="

# 1. 登录
echo -e "\n=== 1. alice 登录 ==="
ALICE_TOKEN=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"test"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "Token: ${ALICE_TOKEN:0:30}..."

echo -e "\n=== 2. bob 登录 ==="
BOB_TOKEN=$(curl -s -X POST $BASE_URL/login \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"test"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
echo "Token: ${BOB_TOKEN:0:30}..."

# 2. 基础接口
echo -e "\n=== 4. alice 查看个人资料 ==="
curl -s $BASE_URL/api/profile -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

echo -e "\n=== 5. alice 访问管理员接口 ==="
curl -s $BASE_URL/api/admin/users -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

echo -e "\n=== 6. bob 访问管理员接口 (应该 403) ==="
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" $BASE_URL/api/admin/users -H "Authorization: Bearer $BOB_TOKEN"

echo -e "\n=== 7. 未登录访问 (应该 401) ==="
curl -s -o /dev/null -w "HTTP Status: %{http_code}\n" $BASE_URL/api/profile

# 3. 用户管理
echo -e "\n=== 8. 获取用户列表 ==="
curl -s $BASE_URL/api/users -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

echo -e "\n=== 9. 创建用户 ==="
curl -s -X POST $BASE_URL/api/users \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试用户","email":"test@example.com","age":25}' | python3 -m json.tool

echo -e "\n=== 10. 获取用户详情 ==="
curl -s $BASE_URL/api/users/1 -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

echo -e "\n=== 11. 更新用户 ==="
curl -s -X PUT $BASE_URL/api/users/1 \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"更新用户","email":"updated@example.com","age":26}' | python3 -m json.tool

echo -e "\n=== 12. 删除用户 ==="
curl -s -X DELETE $BASE_URL/api/users/1 \
  -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

# 4. Casbin 策略管理
echo -e "\n=== 13. 查看策略列表 ==="
curl -s $BASE_URL/api/casbin/policies -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

echo -e "\n=== 14. 添加策略 ==="
curl -s -X POST $BASE_URL/api/casbin/policy \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subject":"admin","object":"/api/admin/users","action":"GET"}' | python3 -m json.tool

echo -e "\n=== 15. 再次查看策略列表 ==="
curl -s $BASE_URL/api/casbin/policies -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

echo -e "\n=== 16. 权限检查 ==="
curl -s -X POST $BASE_URL/api/casbin/check \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subject":"admin","object":"/api/admin/users","action":"GET"}' | python3 -m json.tool

echo -e "\n=== 17. 删除策略 ==="
curl -s -X DELETE $BASE_URL/api/casbin/policy \
  -H "Authorization: Bearer $ALICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"subject":"admin","object":"/api/admin/users","action":"GET"}' | python3 -m json.tool

echo -e "\n=== 18. 最终策略列表 ==="
curl -s $BASE_URL/api/casbin/policies -H "Authorization: Bearer $ALICE_TOKEN" | python3 -m json.tool

echo -e "\n=========================================="
echo "  测试完成！"
echo "=========================================="
```

---

## 预期测试结果

| 测试项 | 预期结果 | 说明 |
|--------|----------|------|
| alice 登录 | 200，返回 Token | alice 拥有 admin 角色 |
| bob 登录 | 200，返回 Token | bob 拥有 user 角色 |
| 健康检查 | 200，返回数据库状态 | 公开接口，无需认证 |
| alice 查看个人资料 | 200，返回用户信息 | 需要认证 |
| alice 访问管理员接口 | 200，返回用户列表 | admin 角色有权限 |
| bob 访问管理员接口 | 403 Forbidden | user 角色无权限 |
| 未登录访问 | 401 Unauthorized | 需要认证 |
| 获取用户列表 | 200，返回用户数据 | 需要认证 |
| 创建用户 | 201，返回创建的用户 | 需要认证 |
| 获取用户详情 | 200，返回用户数据 | 需要认证 |
| 更新用户 | 200，返回更新后的用户 | 需要认证 |
| 删除用户 | 200，返回成功消息 | 需要认证 |
| 查看策略列表 | 200，返回策略数据 | 需要认证 |
| 添加策略 | 200，返回成功消息 | 需要认证 |
| 权限检查 | 200，返回 allowed=true | 需要认证 |
| 删除策略 | 200，返回成功消息 | 需要认证 |

---

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |