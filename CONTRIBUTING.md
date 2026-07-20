# 贡献指南

感谢你对 enhance 项目的关注！本文档将帮助你了解如何参与项目开发，从环境搭建到提交第一个 Pull Request。

> **相关文档**：[开发规范（AI 必读）](AGENTS.md) | [代码风格指南](CODING_STYLE.md) | [架构设计](ARCHITECTURE.md)

---

## 目录

- [快速开始](#快速开始)
- [开发流程](#开发流程)
- [代码规范](#代码规范)
- [测试要求](#测试要求)
- [提交 Pull Request](#提交-pull-request)
- [架构设计原则](#架构设计原则)
- [问题反馈](#问题反馈)
- [行为准则](#行为准则)

---

## 快速开始

### 前提条件

| 工具 | 版本要求 | 说明 |
|------|---------|------|
| Go | 1.21+ | Go 编程语言 |
| Git | 2.0+ | 版本控制 |
| 编辑器 | 任意 | 推荐 VS Code 或 GoLand |

### 1. 克隆仓库

```bash
git clone https://github.com/xudefa/enhance.git
cd enhance
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 运行测试

```bash
# 运行所有测试
go test ./...

# 运行测试并检测数据竞争
go test -race ./...

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 4. 代码格式化

```bash
go fmt ./...
go mod tidy
```

---

## 开发流程

### 分支策略

| 分支 | 用途 | 命名规范 |
|------|------|----------|
| `master` | 主分支，保持稳定 | — |
| `feature/*` | 功能开发 | `feature/ioc-container` |
| `fix/*` | Bug 修复 | `fix/aop-pointcut-match` |
| `docs/*` | 文档更新 | `docs/readme-update` |
| `release/*` | 发布准备 | `release/v1.0.0` |

### 提交规范

提交信息遵循 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <subject>

[optional body]

[optional footer(s)]
```

#### 常用 type

| 类型 | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(core): implement singleton scope` |
| `fix` | Bug 修复 | `fix(aop): resolve pointcut matching issue` |
| `docs` | 文档更新 | `docs(readme): add installation instructions` |
| `refactor` | 代码重构 | `refactor(container): simplify registration` |
| `test` | 测试相关 | `test(core): add container lifecycle tests` |
| `chore` | 构建/工具 | `chore(ci): add golangci-lint config` |
| `perf` | 性能优化 | `perf(container): cache reflection results` |

---

## 代码规范

详细的代码规范请参阅 [代码风格指南](CODING_STYLE.md)。以下是核心规则摘要：

### 命名规范

| 类型 | 规则 | 示例 |
|------|------|------|
| 包名 | 小写，无下划线 | `container`, `userservice` |
| 导出标识符 | 大写驼峰 | `UserID`, `GetUser` |
| 非导出标识符 | 小写驼峰 | `userID`, `getUser` |
| 错误变量 | `Err` 前缀 | `ErrNotFound` |
| 接口 | `er` 后缀或功能描述 | `Reader`, `Logger` |

### 注释规范

- 使用中文注释，技术术语保留英文
- 导出函数/类型必须有 godoc 注释
- 注释应说明"为什么"而非"做什么"

### 错误处理

```go
// 定义哨兵错误
var ErrNotFound = errors.New("not found")

// 包装错误，保留错误链
if err := find(id); err != nil {
    return fmt.Errorf("lookup failed: %w", err)
}

// 使用 errors.Is/As 判断错误
if errors.Is(err, ErrNotFound) {
    // 处理特定错误
}
```

### 函数式选项模式

```go
func NewContainer(opts ...ContainerOption) Container {
    cfg := defaultContainerConfig()
    for _, opt := range opts {
        opt(cfg)
    }
    // ...
}

// 使用
container := NewContainer(
    WithSingleton(),
    WithLazyInit(),
)
```

---

## 测试要求

### 测试命名

```
Test功能_条件_期望行为
```

示例：

```go
func TestContainer_RegisterAndGet_Success(t *testing.T)
func TestContainer_GetT_NotFound_ReturnsError(t *testing.T)
```

### 表驱动测试

```go
func TestCalculateDiscount(t *testing.T) {
    t.Parallel()
    
    tests := []struct {
        name        string
        basePrice   float64
        quantity    int
        expected    float64
        expectError bool
    }{
        {"normal", 100.0, 10, 95.0, false},
        {"negative price", -1.0, 10, 0, true},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // 测试逻辑
        })
    }
}
```

### 覆盖率目标

| 模块 | 目标覆盖率 |
|------|-----------|
| 核心框架 | 90%+ |
| 工具函数 | 95%+ |
| 集成测试 | 覆盖主要流程 |

### 运行特定测试

```bash
# 运行特定包的测试
go test ./core/... -v

# 运行特定测试函数
go test ./core/... -run TestContainer_Register -v

# 并发运行测试
go test -parallel=4 ./...
```

---

## 提交 Pull Request

### 步骤

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 提交更改：`git commit -m 'feat: add my feature'`
4. 推送分支：`git push origin feature/my-feature`
5. 在 GitHub 上创建 Pull Request

### PR 检查清单

提交 PR 前，请确认：

- [ ] 代码通过所有测试（`go test ./...`）
- [ ] 无数据竞争（`go test -race ./...`）
- [ ] 代码已格式化（`go fmt ./...`）
- [ ] 新增功能包含测试
- [ ] 更新相关文档（如适用）
- [ ] 提交信息遵循规范

### PR 模板

```markdown
## 描述
简要描述此 PR 的目的和变更内容。

## 相关 Issue
Closes #123

## 变更类型
- [ ] 新功能 (feat)
- [ ] Bug 修复 (fix)
- [ ] 文档更新 (docs)
- [ ] 代码重构 (refactor)
- [ ] 测试相关 (test)
- [ ] 其他 (chore)

## 测试
- [ ] 已添加/更新单元测试
- [ ] 已通过所有测试

## 备注
其他需要说明的信息。
```

---

## 架构设计原则

### 零外部依赖

核心框架仅使用 Go 标准库，不引入任何第三方依赖。第三方集成放在 `starter/` 目录。

### 接口优先

优先定义小接口，再提供默认实现。这使得用户可以轻松替换实现。

```go
// 定义接口
type Cache interface {
    Get(key string) (any, bool)
    Set(key string, value any, ttl time.Duration) error
}

// 提供默认实现
type MemoryCache struct { ... }
```

### Go 语言优先

参考 Spring 的设计哲学，但遵循 Go 的惯用法：

- 使用组合而非继承
- 使用接口抽象
- 使用泛型提供类型安全
- 使用 context 传递请求上下文

---

## 问题反馈

### Issue 类型

| 类型 | 标签 | 说明 |
|------|------|------|
| Bug 报告 | `bug` | 报告程序错误 |
| 功能请求 | `enhancement` | 提议新功能 |
| 问题咨询 | `question` | 询问使用问题 |
| 文档改进 | `documentation` | 建议文档更新 |
| 性能问题 | `performance` | 报告性能问题 |

### Bug 报告模板

```markdown
## 描述
简要描述 bug 现象。

## 复现步骤
1. ...
2. ...
3. ...

## 期望行为
描述期望的行为。

## 实际行为
描述实际的行为。

## 环境信息
- Go 版本：
- enhance 版本：
- 操作系统：
```

---

## 行为准则

- 尊重所有贡献者
- 接受建设性批评
- 关注问题而非个人
- 欢迎新贡献者
- 保持耐心和友善

---

## 参考文档

- [README.md](README.md) — 项目介绍和快速开始
- [AGENTS.md](AGENTS.md) — AI 智能体开发规范
- [CODING_STYLE.md](CODING_STYLE.md) — 代码风格指南
- [ARCHITECTURE.md](ARCHITECTURE.md) — 架构设计文档

感谢你的贡献！