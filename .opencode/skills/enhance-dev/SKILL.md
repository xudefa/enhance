---
name: enhance-dev
description: Use when developing, modifying, debugging, or reviewing the enhance Go framework codebase. Load this skill to get the full iteration workflow: doc navigation, hard rules, task SOPs, and verification gates. Triggers on any task in the enhance repository.
---

# enhance-dev — enhance 框架 AI 迭代工作流

> enhance 是一个 Go 企业级框架（零核心外部依赖、doc.go 门面、三层依赖方向）。
> 本 skill 封装完整迭代工作流。开始任何改动前先读本节。

## 0. 文档导航（按序加载）

1. `docs/AI_INDEX.md` — 入口总览
2. `docs/AI_NAVIGATION_MAP.md` — 逐包代码导航（定位文件/接口）
3. `AGENTS.md` — 硬性规范
4. `docs/AI_CHANGE_IMPACT.md` — 修改影响面（改代码前必读）
5. `docs/AI_TASK_HANDBOOK.md` — 任务 SOP

排查诡异行为：`docs/AI_PITFALLS.md`、`docs/ADR.md`
快速查规则/模板：`docs/AI_QUICK_REF.md`
依赖/质量/变更：`docs/DEPENDENCY_GRAPH.md`、`docs/QUALITY_SCORE.md`、`docs/API_CHANGELOG.md`

## 1. 硬性规范（违反即拒收）

- 核心包零外部依赖；第三方集成必须在 `starter/`（独立 go.mod）。
- `doc.go` 是门面：只放定义，不写实现；构造函数返回接口。
- 接口 `er` 后缀、≤5 方法；禁止 `I` 前缀。
- 依赖方向单向（Boot → Core → Infrastructure），禁止循环依赖。
- Go 惯用法：早期返回、函数式选项、组合、泛型优先于反射。
- 测试：表驱动 + `t.Parallel()`（含子测试），覆盖率 ≥ 80%。
- 文件：生产 ≤500 行，测试 ≤1000 行；函数 ≤80 行（不含注释和空行）。
- 算法嵌套深度 ≤17 层；时间+空间复杂度不能同时 ≥ O(n²)。
- 导出符号必须有中文 godoc。

## 2. 任务 SOP

### 新增包
1. 在模块根（如 core/）下建目录，`doc.go` 放定义，实现文件放代码。
2. 遵循 AGENTS.md §1.4 文件拆分规则。
3. 表驱动测试 + 子测试 `t.Parallel()`。

### 修改接口
1. 改前读 `docs/AI_CHANGE_IMPACT.md`；确认无破坏性变更。
2. 改后更新 `docs/API_CHANGELOG.md`。
3. 检查依赖方向：`make ai-deps-check`。

### 新增 auto-config / starter
1. 参考 `boot` 包既有 auto-config 与现有 starter 组织方式。
2. starter 互不依赖；示例只能引用库存放。

### 编写测试
1. 表驱动；`t.Parallel()`；覆盖正常+错误路径。
2. 参考 `docs/TESTING_GUIDE.md`。

## 3. 完成前验证（必做）

1. `go build ./...`
2. `go test ./...`（改动模块加 `-race`）
3. `go fmt ./...`
4. `make ai-verify`
5. `make quality-score`（≥ 80）
6. 更新相关文档 + CHANGELOG + API_CHANGELOG（如涉 API）
7. 变更后运行 `scripts/skill-sync.sh check`（若改了 AGENT_RULES.md 则先重新生成）

## 4. 平台规则文件

规则单一来源：`docs/AGENT_RULES.md`，用 `make skill-sync` 重新生成 4 平台文件，禁止手改 `.cursorrules` / `.windsurfrules` / `.codex` / `.github/copilot-instructions.md`。