# enhance 框架 AI 维护导航总入口

> 本文档是 AI/开发者维护 enhance 代码库的**第一入口**。所有变更任务开始前请先阅读本节，按需加载导航文档。

## 1. 这是什么

enhance 是一个 Go 企业级框架（526+ Go 文件、30+ 顶层包）。为降低 AI 维护成本，docs/ 下提供一组**导航索引级**文档，AI 可按需加载。

## 2. 文档导航

| 文档 | 用途 | 什么时候读 |
|------|------|-----------|
| [AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md) | 逐包代码导航：文件职责、核心接口、关键实现 | 需要定位某个包/文件/接口时 |
| [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md) | 包间依赖、修改影响面、回归范围 | 动手改代码**之前** |
| [AI_TASK_HANDBOOK.md](AI_TASK_HANDBOOK.md) | 新增包/接口/模块/测试/验收的 SOP | 执行具体开发任务时 |
| [AI_PITFALLS.md](AI_PITFALLS.md) | 各包设计决策、坑点、隐蔽约定 | 遇到诡异行为/历史决策时 |
| [ADR.md](ADR.md) | 架构决策记录归档（现有） | 想了解"为什么这样设计"时 |
| [TESTING_GUIDE.md](TESTING_GUIDE.md) | 测试规范和最佳实践 | 编写测试时 |
| [DEPENDENCY_GRAPH.md](DEPENDENCY_GRAPH.md) | 包间依赖关系可视化 | 理解模块依赖时 |
| [API_CHANGELOG.md](API_CHANGELOG.md) | API 变更记录 | 追踪接口演进时 |
| [QUALITY_SCORE.md](QUALITY_SCORE.md) | 代码质量评分体系 | 评估代码质量时 |
| [AI_QUICK_REF.md](AI_QUICK_REF.md) | AI 快速参考卡（代码模板+规则速查+命令速查） | 需要快速查规则/模板时 |

现有根文档（补充，不重复内容）：
[AGENTS.md](../AGENTS.md)（全局规范） · [ARCHITECTURE.md](../ARCHITECTURE.md)（架构设计） · [DEPENDENCIES.md](../DEPENDENCIES.md)（模块依赖） · [CODING_STYLE.md](../CODING_STYLE.md)（代码风格） · [CONTRIBUTING.md](../CONTRIBUTING.md)（贡献指南） · [CHANGELOG.md](../CHANGELOG.md)（变更日志）

## 3. 首次接入：按顺序完成

1. 读本文件（第 1-2 节）
2. 读 [AI_NAVIGATION_MAP.md](AI_NAVIGATION_MAP.md) 了解包结构
3. 读 [AGENTS.md](../AGENTS.md) 了解硬性规范（零外部依赖、doc.go 门面、依赖方向）
4. 读 [AI_CHANGE_IMPACT.md](AI_CHANGE_IMPACT.md) 了解你准备修改的包的影响面
5. 读 [AI_TASK_HANDBOOK.md](AI_TASK_HANDBOOK.md) SOP 6 验收清单，再开始动手

## 4. 按任务类型读文档

| 任务 | 必读 |
|------|------|
| 新增一个包 | AI_TASK_HANDBOOK §新增包、AI_NAVIGATION_MAP、TESTING_GUIDE |
| 修改现有包的接口 | AI_CHANGE_IMPACT、AI_TASK_HANDBOOK §修改接口、API_CHANGELOG |
| 新增 auto-config/starter | AI_TASK_HANDBOOK §新增自动配置、AI_PITFALLS |
| 新增测试 | AI_TASK_HANDBOOK §测试、AGENTS.md §2.3、TESTING_GUIDE |
| 排查诡异行为 | AI_PITFALLS、ADR.md |
| 理解模块依赖 | DEPENDENCY_GRAPH、DEPENDENCIES.md |
| 评估代码质量 | QUALITY_SCORE |
| 追踪 API 变更 | API_CHANGELOG |

## 5. 硬性规范速查（违反会被拒收）

### 5.1 必须遵守
- ✅ 核心框架包（core/boot/...）**零外部依赖**，只 import 标准库 + enhance 内部包
- ✅ 包间只能通过接口交互，实现隐藏；`doc.go` 是门面（只放定义）
- ✅ 构造函数返回接口；接口命名 `er` 后缀，禁止 `I` 前缀；接口 ≤5 方法
- ✅ 依赖方向单向（Boot → Core → Infrastructure），禁止循环依赖
- ✅ Go 惯用法：早期返回、函数式选项、组合优于继承、泛型优先于反射
- ✅ 测试用表驱动，测试函数与子测试 `t.Parallel()`

### 5.2 绝对禁止
- ❌ 核心包引入第三方依赖（第三方集成必须放 `starter/`）
- ❌ 相对导入、忽略错误、裸 goroutine、直接用 `==` 比较错误
- ❌ 在 doc.go 写实现逻辑
- ❌ 大改现有接口而不做影响分析与兼容处理

## 6. 验收标准（任务完成前自检）

1. `go build ./...` 通过
2. 相关包 `go test` 通过
3. 导出符号均有 godoc 注释
4. 文档引用的文件名/接口名与源码一致（用 grep 抽查）
5. 运行 `make ai-verify` 通过
6. 运行 `make quality-score` 评分 ≥ 80