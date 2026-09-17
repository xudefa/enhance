# AI 维护文档体系设计文档

> **日期**：2026-09-16
> **状态**：已批准
> **范围**：enhance 框架 AI 可读性与可维护性优化索引级文档体系

---

## 1. 背景与目标

enhance 是一个大型 Go 框架项目（526 个 Go 文件、30+ 顶层包）。项目已有完善的根文档（`AGENTS.md`、`ARCHITECTURE.md`、`CODING_STYLE.md`、`DEPENDENCIES.md` 等），但缺乏**面向 AI 的代码维护导航文档**。

AI（或新开发者）接手代码时虽然有 `AGENTS.md` 规范，但：
- 不知道每个包的源码文件各自负责什么、接口定义在哪个文件
- 不清楚修改某个包会影响到哪些下游包
- 没有"新增一个包 / 修改一个接口"的标准操作流程（SOP）
- 各包的设计决策与坑点散落在源码中，无系统记录

### 目标

在 `docs/` 下建立一组**导航索引级**的 AI 维护文档，让 AI 能：
1. **快速定位**：任意包的接口定义、关键实现、文件职责一目了然
2. **准确评估变更影响**：修改任意包前知晓其下游影响面
3. **按 SOP 执行任务**：新增包/接口/模块、加测试、验收都有标准步骤
4. **避开已知坑点**：各包的历史决策、隐蔽约定、易错点有系统记录

### 非目标

- 不做逐包深度实现讲解（那是各包 README 的职责）
- 不重写现有根文档（`AGENTS.md` 等仅做入口打通，不重构内容）
- 不进行大规模代码重构（仅在文档编写过程中顺带修复明确的 AI 可读性弱项）

---

## 2. 方案选择

对比三种方案后选定 **方案 A：分主题多文件**：

| 方案 | 结构 | 结论 |
|------|------|------|
| A. 分主题多文件 | 5 个各司其职的文档 + 总入口 | ✅ 推荐：文件少、职责单一、AI 按需加载 |
| B. 聚合式 | 1 个入口 + 1 个超大文件 | ❌ AI 加载浪费上下文、难更新 |
| C. 逐包目录 | `docs/packages/<pkg>.md` 每包一篇 | ❌ 30+ 文件、导航分散、维护成本高 |

选择理由：
- AI 上下文有限，文档应支持"按需加载"而非"全量加载"
- 文档文件少（5 个），便于人工与 AI 共同维护
- 每个文件聚焦一个主题，与现有根文档体系互补不重叠

---

## 3. 文档结构与内容

### 3.1 文件清单

| 文件 | 主题 | 核心读者 |
|------|------|----------|
| `docs/AI_INDEX.md` | 总入口 + 快速导航 | 新接入的 AI / 开发者 |
| `docs/AI_NAVIGATION_MAP.md` | 逐包代码导航地图 | 需要定位代码的 AI |
| `docs/AI_CHANGE_IMPACT.md` | 包间依赖 + 变更影响 | 修改前做影响分析的 AI |
| `docs/AI_TASK_HANDBOOK.md` | 常见任务的 SOP | 执行代码任务的 AI |
| `docs/AI_PITFALLS.md` | 各包设计决策与坑点 | 遇到隐蔽行为的 AI |

### 3.2 各文档详细内容

#### 3.2.1 `docs/AI_INDEX.md` — 总入口

- 一句话说明文档体系用途与阅读顺序
- 文档导航表（指向本组 5 份文档 + 现有根文档）
- **第一遍阅读路径**：AI 第一次接触代码库时按顺序读哪些文件
- **按任务类型读文档**（新增功能读哪篇、改接口读哪篇…）
- 关键约定速查（包命名、doc.go 规范、依赖方向、禁止事项摘要）
- 指向 `core/doc.go`、`boot/doc.go` 等核心包入口

#### 3.2.2 `docs/AI_NAVIGATION_MAP.md` — 逐包代码导航地图

对每个顶层包采用**统一模板**：

```markdown
### `package`（一句话职责）

- **核心接口**：`doc.go` 中的接口名（链接到 doc.go）
- **文件职责**：
  | 文件 | 职责 |
  |------|------|
  | `doc.go` | 接口与公共类型定义（门面） |
  | `xxx.go` | 实现说明 |
  | ... | ... |
- **关键实现文件**：实现核心逻辑的文件与函数
- **依赖**：引用的其他 enhance 包
- **注意事项**：该包独有的约定/限制
```

覆盖范围：`core`、`boot`、`context`、`condition`、`lifecycle`、`config`、`cache`、`web`、`event`、`schedule`、`security`、`metrics`、`actuator`、`log`、`async`、`retry`、`audit`、`i18n`、`spel`、`validation`、`exception`、`resilience`、`mq`、`tenant`、`tracing`、`openapi`、`testing`、`devtools`、`email`、`webtest`、`metadata` 等全部顶层包。

> 文件—职责对照需基于实际包内容整理，不臆造。若某个真实文件未被原始文档记录，以源码为准。

#### 3.2.3 `docs/AI_CHANGE_IMPACT.md` — 包间依赖与变更影响

- **依赖方向总览**：Boot → Core → Infrastructure（与 `DEPENDENCIES.md` 对齐）
- **包依赖矩阵**：`A 包依赖 → B/C/D`（对应"改 A 前的编译前提"）
- **反向影响表**：`修改 X 包 → 会影响 Y/Z 包`（对应"改 X 后的回归测试范围"）
- **执行顺序表**：各模块的启动/执行优先级（引用 `DEPENDENCIES.md` 与 `boot/doc.go` 中 `OrderPriority`）
- **验证清单**：修改某包后的必做检查项（编译、单测、race、受影响包测试）

#### 3.2.4 `docs/AI_TASK_HANDBOOK.md` — 常见任务 SOP

每个任务按**标准可执行步骤**编写，含命令与检查点：

| 任务 | SOP 关键步骤 |
|------|-------------|
| 新增一个包 | 目录结构、doc.go 门面规范、README、测试、导入检查 |
| 新增接口/类型 | 放 doc.go、er 后缀、godoc 注释、≤5 方法、构造函数返回接口 |
| 修改现有接口/类型 | 影响面排查、类型别名兼容、调用方更新、测试更新 |
| 新增 auto-config/starter | boot.RegisterAutoConfig、条件装配、OrderPriority、循环依赖检查 |
| 新增测试用例 | 表驱动、t.Parallel、覆盖正/反路径、无全局状态竞争 |
| 新增第三方示例 | examples/ 独立 go.mod、starter 集成规则 |
| 任务验收 | go build / go test / go test -race / go vet / gofmt |

#### 3.2.5 `docs/AI_PITFALLS.md` — 设计决策与坑点

- **全局坑点**：禁止事项清单（来自 AGENTS.md §6）
- **每包的特殊决策与坑点**：从源码注释、ADR.md、doc.go 中的设计说明提炼
  - 例如 `core`：Go 接口方法不能带类型参数，容器内部用 reflect.Type、用户层用泛型 API
  - 例如 `boot`：BootError 从结构体改为接口，保留 `BootErrorStruct` 类型别名做兼容
  - 例如 `event`：用字符串类型而非反射类型做事件路由
  - 实际内容以源码与 ADR 提炼为准，标注入 `ADR-XXX` 出处
- **易错模式**：AI 最容易犯的错（如把接口定义写进实现文件、构造函数返回具体类型）

### 3.3 与现有文档的关系

| 现有文档 | 本体系的定位 | 关系 |
|----------|-------------|------|
| `AGENTS.md` | 全局规范 | 入口打通：在 §0 或 §1 增加指向 AI_INDEX.md 的指引 |
| `ARCHITECTURE.md` | 架构设计 | AI_NAVIGATION_MAP 的"为什么"；本体系聚焦"文件在哪" |
| `DEPENDENCIES.md` | 模块依赖 | AI_CHANGE_IMPACT 的数据来源，保持一致不重复造表 |
| `CODING_STYLE.md` | 代码风格 | AI_TASK_HANDBOOK 的细则，SOP 引用而非复制 |

---

## 4. 入口打通

在 `AGENTS.md` 顶部"重要"引用区新增一行：

```markdown
> **AI 维护导航**：首次接触代码库的 AI 请先阅读 [docs/AI_INDEX.md](docs/AI_INDEX.md)，了解代码导航、变更影响与任务 SOP。
```

确保新接入的 AI 会主动加载文档体系。

---

## 5. 顺手修复原则

写文档过程中如发现明确的 AI 可读性弱项，做**最小改动**修复：

- 允许：导出类型/函数缺少 godoc 注释时补注释；明显命名不规范（非接口滥用 I 前缀等）；错误的链接/文档描述
- 不允许：重构实现逻辑、大范围重命名、修改行为
- 每次顺手修复必须在对应文档的"注意事项"中记录，并跑 `go build ./...` + `go test` 验证

---

## 6. 验收标准

1. `docs/` 下新增 5 个文档文件，均有实际内容且互相索引
2. `AGENTS.md` 已加入 AI_INDEX.md 入口指引
3. 每个顶层包在 `AI_NAVIGATION_MAP.md` 中有条目，文件—职责与实际源码一致
4. `AI_CHANGE_IMPACT` 的依赖信息与 `DEPENDENCIES.md`、实际 import 一致
5. `AI_TASK_HANDBOOK` 的每个 SOP 可被 AI 直接执行（含命令）
6. `AI_PITFALLS` 的每条坑点能对应到实际源码/ADR
7. 代码改动（若有）通过 `go build ./...` 与相关包 `go test` 验证
8. 全部文档使用中文，与其他现有文档语言一致