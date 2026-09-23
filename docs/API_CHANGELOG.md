# API 变更日志

本文档记录 enhance 项目的 API 变更，便于追踪接口演进。

## 1. 变更类型

### 1.1 新增 (Added)
- 新增接口
- 新增方法
- 新增类型
- 新增函数

### 1.2 修改 (Changed)
- 修改接口签名
- 修改方法行为
- 修改类型定义

### 1.3 废弃 (Deprecated)
- 废弃接口
- 废弃方法
- 废弃类型

### 1.4 移除 (Removed)
- 移除接口
- 移除方法
- 移除类型

## 2. 变更记录

### 2026-09-16

#### Added
- `core.Container` 接口新增 `MustGetBean[T]` 泛型方法
- `boot.Application` 接口新增 `WithProfiles` 方法
- `event.EventBus` 接口新增 `PublishAsync` 方法

#### Changed
- `web.Context` 接口方法签名更新，增加 `context.Context` 参数

#### Deprecated
- `boot.BootErrorStruct` 类型别名，建议使用 `BootError` 接口

### 2026-09-01

#### Added
- `core.Container` 接口
- `boot.Application` 接口
- `event.EventBus` 接口
- `web.Router` 接口
- `security.HttpSecurity` 接口
- `cache.Cache` 接口
- `schedule.Scheduler` 接口
- `log.Logger` 接口
- `metrics.MeterRegistry` 接口
- `actuator.Endpoint` 接口
- `tracing.Tracer` 接口

## 3. 兼容性策略

### 3.1 向后兼容
- 新增方法不破坏现有实现
- 新增类型不破坏现有代码
- 废弃方法保留至少一个版本

### 3.2 破坏性变更
- 修改接口签名需要版本升级
- 移除方法需要版本升级
- 修改行为需要文档说明

### 3.3 迁移指南
- 废弃方法提供替代方案
- 破坏性变更提供迁移脚本
- 重大变更提供详细文档

## 4. 版本语义

### 4.1 主版本号 (Major)
- 破坏性 API 变更
- 架构重大调整

### 4.2 次版本号 (Minor)
- 新增功能
- 新增接口
- 向后兼容

### 4.3 修订号 (Patch)
- Bug 修复
- 文档更新
- 性能优化

## 5. 使用方法

### 5.1 查看当前 API
```bash
# 查看包的导出接口
go doc ./core/...
go doc ./boot/...
```

### 5.2 检查 API 兼容性
```bash
# 使用 go vet 检查
go vet ./...

# 使用静态分析工具
golangci-lint run
```

### 5.3 生成 API 文档
```bash
# 使用 go doc 生成文档
go doc -all ./core/ > docs/api/core.txt
go doc -all ./boot/ > docs/api/boot.txt
```

## 6. 最佳实践

### 6.1 接口设计
- 保持接口小而精（1-5 个方法）
- 使用组合形成大接口
- 避免频繁修改接口

### 6.2 向后兼容
- 新增方法不破坏现有实现
- 废弃方法保留至少一个版本
- 提供迁移指南

### 6.3 文档更新
- 及时更新 API 文档
- 记录变更原因
- 提供使用示例
