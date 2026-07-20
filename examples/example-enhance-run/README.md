# Enhance GORM 示例

这个示例演示如何使用 `enhance.Run()` 简洁启动一个带有 GORM 数据库支持的 Web 服务。

## 功能特性

- ✅ 类似 Spring Boot 的一行启动 API
- ✅ GORM 模块自动连接 MySQL 数据库
- ✅ 自动表迁移
- ✅ RESTful API 控制器
- ✅ 依赖注入
- ✅ 优雅关闭

## 快速开始

### 1. 准备 MySQL 数据库

确保你有一个运行中的 MySQL 数据库，并创建一个数据库：

```sql
CREATE DATABASE enhance_demo;
```

### 2. 配置数据库连接

编辑 `config/application.json` 文件，启用 GORM 并配置数据库连接：

```json
{
  "gorm": {
    "enabled": true,
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "your_password",
    "database": "enhance_demo",
    "charset": "utf8mb4",
    "max-open-conns": 100,
    "max-idle-conns": 10,
    "conn-max-lifetime": 3600
  }
}
```

### 3. 运行示例

```bash
cd examples/example-enhance-run
go run .
```

### 4. 测试 API

服务启动后，你可以使用 curl 或 Postman 测试以下 API：

#### 创建用户

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "张三",
    "email": "zhangsan@example.com",
    "age": 25
  }'
```

#### 获取用户列表

```bash
curl http://localhost:8080/api/users
```

#### 获取单个用户

```bash
curl http://localhost:8080/api/users/1
```

#### 更新用户

```bash
curl -X PUT http://localhost:8080/api/users/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "李四",
    "email": "lisi@example.com",
    "age": 30
  }'
```

#### 删除用户

```bash
curl -X DELETE http://localhost:8080/api/users/1
```

## 代码结构

```
example-enhance-run/
├── main.go              # 主程序入口
├── gorm_module.go       # GORM 模块定义
├── models.go            # 数据模型
├── user_controller.go   # 用户控制器
├── autoconfig.go        # 自动配置
└── README.md            # 说明文档
```

## 核心代码说明

### 1. 最简启动方式

```go
func main() {
    enhance.Run()
}
```

### 2. 带 GORM 模块启动

```go
func main() {
    enhance.Run(
        boot.WithAppName("enhance-gorm-demo"),
        boot.WithVersion("1.0.0"),
        boot.WithModulesOption(NewGormModule()),
    )
}
```

### 3. 定义 GORM 模块

```go
func NewGormModule() *boot.ModuleBuilder {
    return boot.NewModule().
        Name("gorm").
        Bean(boot.Provide(NewGormDB)).
        Starter(&GormStarter{})
}
```

### 4. 自动配置

```go
func init() {
    boot.RegisterAutoConfig(&WebAutoConfig{})
    boot.RegisterAutoConfig(&DatabaseAutoConfig{})
}
```

## 自定义

你可以根据需要修改：

- **数据模型**：编辑 `models.go` 添加新的模型
- **控制器**：编辑 `user_controller.go` 或创建新的控制器
- **配置**：编辑 `config/application.json` 修改数据库连接等配置

## 注意事项

- 确保 MySQL 数据库已启动并且连接配置正确
- 首次运行时会自动创建 `users` 表
- 服务会监听 `config/application.json` 中配置的端口（默认 8080）