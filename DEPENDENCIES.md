# enhance 模块依赖关系

> **文档版本**：v0.0 | **最后更新**：2026-07-29

本文档描述 enhance 框架各模块之间的依赖关系和执行顺序。

---

## 依赖总览

### 依赖方向

```
Boot Layer (启动层)
    ↓
Core Layer (核心层)
    ↓
Infrastructure Layer (基础设施层)
```

**核心规则**：
- 依赖方向只能从上到下，禁止循环依赖
- 核心包（`core/`）零外部依赖，仅使用 Go 标准库
- 各包通过接口交互，实现细节隐藏在包内部

### 模块优先级

| 层级 | 优先级 | 说明 |
|------|--------|------|
| 基础设施层 | -3000 | 日志、配置、条件判断 |
| 核心层 | -2500 | IoC 容器、AOP、生命周期 |
| 数据层 | -2000 | 数据库、缓存 |
| 认证层 | -1500 | JWT 认证 |
| 授权层 | -1200 | Casbin 授权 |
| 安全核心层 | -100 | 安全框架 |
| Web 层 | 0 | HTTP 服务器、MVC |
| 业务层 | 1000 | 事件、调度、异步 |
| 监控层 | 2000 | 指标、运维端点 |

---

## 核心模块依赖

### 1. 基础设施层 (-3000)

| 模块 | 说明 | 依赖 |
|------|------|------|
| `boot/` | 启动器、自动配置注册表 | 无 |
| `condition/` | 条件判断（OnProperty、OnBean） | `boot` |
| `config/` | 配置管理 | `boot`, `condition` |
| `log/` | 日志抽象 | 无 |

**执行顺序**：`boot → condition → config → log`

### 2. 核心层 (-2500)

| 模块 | 说明 | 依赖 |
|------|------|------|
| `core/` | IoC 容器、依赖注入 | 无（仅标准库） |
| `aop/` | 面向切面编程 | `core` |
| `lifecycle/` | 生命周期管理 | `boot` |

**执行顺序**：`core → aop → lifecycle`

### 3. 数据层 (-2000)

| 模块 | 说明 | 依赖 |
|------|------|------|
| `cache/` | 缓存抽象（LRU） | `core` |
| `starter/gorm` | GORM 数据库集成 | `boot`, `condition`, `log`, `config` |

### 4. 安全层 (-1500 ~ -100)

| 模块 | 优先级 | 说明 | 依赖 |
|------|--------|------|------|
| `starter/jwt` | -1500 | JWT 认证 | `boot`, `condition`, `security`, `log` |
| `starter/casbin` | -1200 | Casbin 授权 | `boot`, `condition`, `security`, `log` |
| `security/` | -100 | 安全框架 | `boot`, `condition`, `core`, `log` |

### 5. Web 层 (0)

| 模块 | 说明 | 依赖 |
|------|------|------|
| `web/` | HTTP 服务器、路由器 | `boot`, `log` |
| `web/mvc` | MVC 框架 | `web`, `security` |

### 6. 业务层 (1000)

| 模块 | 说明 | 依赖 |
|------|------|------|
| `event/` | 事件驱动 | `core`, `async` |
| `async/` | 异步执行 | `core` |
| `schedule/` | 定时任务 | `boot`, `condition`, `log` |
| `resilience/` | 弹性与容错 | `core`, `log` |

### 7. 监控层 (2000)

| 模块 | 说明 | 依赖 |
|------|------|------|
| `metrics/` | 指标收集 | `boot`, `condition` |
| `actuator/` | 运维端点 | `boot`, `condition`, `metrics` |

### 8. 工具层（无固定优先级）

| 模块 | 说明 | 依赖 |
|------|------|------|
| `i18n/` | 国际化 | `core`, `config` |
| `validation/` | 数据验证 | `core`, `web` |
| `spel/` | 表达式语言 | `core` |
| `context/` | 应用上下文 | `boot`, `core`, `config` |

---

## 第三方插件集成

### 插件优先级选择

| 插件类型 | 推荐优先级 | 说明 |
|---------|-----------|------|
| 日志框架 | -3000 | 必须最先初始化 |
| 配置中心 | -3000 | 提供配置管理能力 |
| 数据库驱动 | -2000 | 提供数据访问能力 |
| 缓存驱动 | -2000 | 提供缓存能力 |
| 认证框架 | -1500 | 提供认证能力 |
| 授权框架 | -1200 | 提供授权能力 |
| 消息队列 | 1000 | 提供消息传递能力 |
| 定时任务 | 1000 | 提供调度能力 |
| 链路追踪 | 2000 | 提供追踪能力 |
| 指标收集 | 2000 | 提供监控能力 |

### 插件开发示例

```go
package redis

import (
    "github.com/xudefa/enhance/boot"
    "github.com/xudefa/enhance/condition"
)

type RedisAutoConfiguration struct{}

func (r *RedisAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
    // 初始化 Redis 连接
    return nil
}

func init() {
    boot.RegisterAutoConfigWith(&RedisAutoConfiguration{},
        boot.WithConditions(
            condition.OnProperty("redis.enabled", "true"),
        ),
        boot.WithOrder(-2000), // 数据层优先级
    )
}
```

---

## 依赖检查清单

开发新模块时，请确认以下依赖关系：

- [ ] 是否依赖基础设施层（log、config、boot）？
- [ ] 是否依赖核心层（core、aop）？
- [ ] 是否依赖数据层（gorm、cache）？
- [ ] 是否依赖安全层（security、jwt、casbin）？
- [ ] 优先级设置是否正确？
- [ ] 是否存在循环依赖？
- [ ] 是否依赖 Web 层（web、mvc）？
- [ ] 是否依赖业务层（schedule、event、async）？
- [ ] 是否依赖监控层（metrics、actuator）？
- [ ] 是否设置了正确的 OrderPriority？
- [ ] 是否有循环依赖？
- [ ] 是否通过接口交互，而非直接依赖实现？
- [ ] 是否使用了 condition 条件装配？
- [ ] 是否提供了完整的依赖说明文档？

---

## 常见问题

### Q1: 为什么 casbin-gorm 的优先级比 casbin 高？

A: `casbin-gorm` 依赖 GORM 提供数据库连接，并在 `casbin` 基础配置之前执行，注册 GORM 版本的 `CasbinEnforcer`。`casbin` 基础配置会检测容器中是否已有 Enforcer，如果有则直接使用，否则创建默认的 `DefaultCasbinEnforcer`。

### Q2: 如何确定新模块的优先级？

A: 参考本文档的"第三方插件集成指南"表格，根据模块的职责选择对应的优先级。如果模块依赖多个层，选择依赖层中优先级最高的（值最小的）。

### Q3: 可以自定义优先级吗？

A: 可以，但建议使用 `OrderPriority` 枚举常量，而非直接使用魔法数字。如果需要在同一层级内调整顺序，可以在枚举值基础上加减（如 `OrderPriorityDataLayer - 100`）。

### Q4: 核心层（Core Layer）为什么是 -2500？

A: 核心层包含 IoC 容器、AOP 和生命周期管理，这些模块依赖基础设施层（日志、配置），但必须在数据层之前初始化，为上层提供依赖注入能力。因此设置在基础设施层（-3000）和数据层（-2000）之间。

### Q5: 工具层模块为什么没有固定优先级？

A: 工具层模块（如 i18n、exception、validation 等）是按需使用的工具类库，不参与自动配置的执行顺序。它们通过 IoC 容器注册为 Bean，在应用启动后的任何阶段都可以使用。

### Q6: 如何避免循环依赖？

A: 遵循以下原则：
1. 低优先级模块不能依赖高优先级模块
2. 使用接口而非具体实现进行模块间交互
3. 通过 IoC 容器解耦依赖关系
4. 使用 `condition.OnMissingBean` 避免重复注册

### Q7: 模块之间如何通信？

A: 推荐使用以下方式：
1. **IoC 容器**：通过依赖注入获取其他模块的 Bean
2. **事件总线**：通过 `event` 模块发布/订阅事件
3. **配置中心**：通过 `config` 模块共享配置
4. **日志系统**：通过 `log` 模块记录日志