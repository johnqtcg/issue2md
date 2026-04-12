---
title: 规范驱动开发（SDD）最佳实践
owner: john
status: active
last_updated: 2026-03-31
applicable_versions: spec-kit v0.x (2025-09 至今), Claude Code 2025+, Codex CLI 5.x, Cursor 2.x
---

# 规范驱动开发（SDD）最佳实践

<!-- Funnel Layer 1: Executive Summary — 所有角色阅读 -->

## Executive Summary

规范驱动开发（Specification-Driven Development, SDD）将规范而非代码作为软件的第一公民。核心主张：**先写规范、再让 AI 生成代码**，而非用模糊 prompt 反复试错。

GitHub 于 2025 年 9 月开源 [spec-kit](https://github.com/github/spec-kit) 工具包，提供 Constitution → Specify → Plan → Tasks → Implement 五阶段工作流，支持 Claude Code、Codex、Cursor 等主流 AI Agent。其中 **Constitution（项目宪法）** 是关键创新——通过不可变法案约束 AI 行为，确保生成代码始终符合架构原则和质量标准。

**对各角色的意义**：

| 角色 | SDD 带来的改变 |
|------|--------------|
| **Leader** | 规范版本化、可 Review，技术决策有据可查，团队知识不再依赖个人 |
| **PM** | 需求文档即开发输入，变更直接反映在实现中，意图与代码不再脱节 |
| **后端开发** | AI 在明确约束下生成代码，减少返工；宪法确保架构一致性 |
| **QA** | 规范中的验收标准直接驱动测试，覆盖率可追溯到需求 |

**关键判断**：SDD 适合中大型功能开发、新项目启动、需要合规性和可追溯性的场景；不适合快速原型和小 Bug 修复。本项目 issue2md 是 SDD + TDD 的直接产物，见第 7 章案例。

> *以下内容按深度递进：第 1-2 章为概念概览，第 3-5 章为核心机制，第 6-7 章为实操指南和案例，第 8-9 章为边界讨论和行动建议。*

---

## 目录

**阅读指引**：本文包含四类内容，用标签区分。可按需跳读。

| 标签 | 含义 | 建议读者 |
|------|------|---------|
| `概念` | 原理解释，回答"是什么、为什么" | 所有人 |
| `机制` | 设计细节，回答"如何约束 AI" | 开发、QA |
| `操作` | 具体步骤，回答"怎么做" | 动手实践者 |
| `案例` | 项目经验，回答"实际效果如何" | 评估可行性者 |

- [Executive Summary](#executive-summary)
- [1. SDD 是什么：从 Vibe Coding 到规范驱动](#1-sdd-是什么从-vibe-coding-到规范驱动) `概念`
- [2. 为什么是现在：三大趋势的交汇](#2-为什么是现在三大趋势的交汇) `概念`
- [3. spec-kit：GitHub 的 SDD 开源工具包](#3-spec-kitgithub-的-sdd-开源工具包) `概念` `操作`
- [4. 宪法基石：约束 AI Agent 行为的九条法案](#4-宪法基石约束-ai-agent-行为的九条法案) `机制`
- [5. 模板如何驯化 LLM：七种约束机制](#5-模板如何驯化-llm七种约束机制) `机制`
- [6. 工具实操：Claude Code / Codex / Cursor 的 SDD 落地](#6-工具实操claude-code--codex--cursor-的-sdd-落地) `操作`
- [7. 实战案例：issue2md 项目的 SDD 实践](#7-实战案例issue2md-项目的-sdd-实践) `案例`（含 7.4：九条法案 → 项目宪法映射）
- [8. SDD 的边界与批评](#8-sdd-的边界与批评) `概念`
- [9. 总结与行动建议](#9-总结与行动建议) `操作`
- [附录 A：术语表](#附录-a术语表)
- [附录 B：参考资料](#附录-b参考资料)

---

<!-- Funnel Layer 2: Overview — Leader + 高级开发阅读 -->

## 1. SDD 是什么：从 Vibe Coding 到规范驱动

### 1.1 问题：Vibe Coding 的代价

2025 年 2 月，Andrej Karpathy 提出了"Vibe Coding"一词，描述了一种用 AI 写代码的即兴模式——给出一个模糊 prompt，拿到一段"看起来对"的代码，编译通过就继续，不行就再 prompt 一轮。这个模式在原型验证时效率极高，但用于生产级系统时问题显著：

- **模糊性放大**：AI 模型擅长模式补全，但不会读心术。一个"给我加个登录功能"的 prompt，会迫使模型对上千个未说明的需求做猜测 [1]。
- **意图丢失**：当对话上下文增长或 session 切换后，AI 会忘记之前的设计决策，给出前后矛盾的实现 [5]。
- **架构失控**：缺乏全局约束，多次迭代后代码架构变成"自动化生成的遗留系统" [6]。

多个独立来源引述 2025 年 Stack Overflow 开发者调查的数据：84% 的开发者使用或计划使用 AI 工具，但只有 22% 持正面评价，46% 认为准确性是首要问题 [2]（注：该数据来自 orchestrator.dev 对 SO 调查的二次引用，非本文直接核实）。

### 1.2 解法：SDD 的权力反转

Spec-Driven Development（SDD）颠倒了代码与规范的权力关系 [3]：

| 维度 | 传统开发 | Vibe Coding | SDD |
|------|---------|-------------|-----|
| 信息流 | 需求 → 设计 → 手动编码 → 测试 | Prompt → AI 生成 → 修 → 再 prompt | 需求 → 详细规范 → AI 生成 → 验证 |
| 真相来源 | 代码 | 对话上下文 | 规范 |
| 变更成本 | 改代码 + 补文档 | 重新 prompt | 改规范 → 重新生成 |
| 可追溯性 | 弱 | 无 | 强（规范 → 计划 → 任务 → 代码） |

spec-kit 核心文档对此的表述相当激进 [3]：

> *"Specifications don't serve code — code serves specifications. The PRD isn't a guide for implementation; it's the source that generates implementation."*

这是 SDD 倡导者的理想愿景。实践中，Martin Fowler 团队观察到这种"权力反转"尚未完全实现——多数工具仍处于早期阶段 [7]。他们将 SDD 分为三个成熟度层级：

1. **Spec-first**：写好规范再用 AI 生成代码，代码仍需人工维护
2. **Spec-anchored**：规范和代码共存且同步演化
3. **Spec-as-source**：规范是唯一制品，人不直接接触代码

大多数实践目前处于 Spec-first 阶段，spec-kit 正向 Spec-anchored 演进。

---

## 2. 为什么是现在：三大趋势的交汇

spec-kit 核心文档明确指出，三个趋势的交汇使 SDD 从理论变为现实 [3]：

**趋势一：AI 能力跨过自然语言 → 代码的可靠性门槛**

AI 模型已经能够理解复杂的自然语言规范并生成可工作的代码。这不是替代开发者，而是将"从规范到实现的机械翻译"自动化 [3]。

**趋势二：软件复杂性指数级增长**

现代系统集成几十个服务、框架和依赖，通过手动流程保持所有组件与原始意图的一致性越来越困难。SDD 通过规范驱动的生成提供系统性对齐 [3]。

**趋势三：变更速度持续加快**

需求变更已不是意外，而是常态。传统开发将需求变更视为破坏性事件，每次变更需要手动传播到文档、设计和代码中。SDD 将需求变更转化为正常工作流——修改规范中的核心需求，受影响的实现计划自动标记更新 [3]。

---

<!-- Funnel Layer 3: Technical Detail — 开发和 QA 阅读 -->

## 3. spec-kit：GitHub 的 SDD 开源工具包

### 3.1 项目概况

[spec-kit](https://github.com/github/spec-kit) 是 GitHub 于 2025 年 9 月开源的 SDD 工具包，MIT 协议。截至目前，GitHub 页面显示 83,900+ Stars；README 列出 25+ 支持的 AI Agent（Claude Code、GitHub Copilot、Cursor、Gemini CLI、Codex CLI 等）和 40+ 社区扩展 [4][8]。

核心工具是 `specify` CLI（Python 实现），将自然语言描述转换为结构化、版本控制的 Markdown 规范文件。

### 3.2 五阶段工作流

spec-kit 将开发过程结构化为五个有序阶段，每个阶段控制下一个——未完成当前阶段，不能进入下一阶段 [1][3][4]：

```
┌─────────────┐    ┌──────────┐    ┌────────┐    ┌─────────┐    ┌───────────┐
│ Constitution │───▶│ Specify  │───▶│  Plan  │───▶│  Tasks  │───▶│ Implement │
│  项目宪法    │    │ 定义需求  │    │ 技术方案│    │ 任务拆解│    │  执行实现  │
└─────────────┘    └──────────┘    └────────┘    └─────────┘    └───────────┘
     ▲                  │              │              │               │
     │                  ▼              ▼              ▼               ▼
     │             人工审查 ✓     人工审查 ✓     人工审查 ✓      人工审查 ✓
     └──────────────── 反馈循环：生产指标、事故、运维经验 ◀──────────┘
```

| 阶段 | 命令 | 关注点 | 输出制品 |
|------|------|--------|---------|
| Constitution | `/speckit-constitution` | 项目不可变原则和编码标准 | `.specify/memory/constitution.md` |
| Specify | `/speckit-specify` | 用户旅程、业务需求、成功标准（不涉及技术栈） | `.specify/specs/<feature>/spec.md` |
| Plan | `/speckit-plan` | 技术选型、架构约束、数据模型、API 契约 | `plan.md`, `data-model.md`, `contracts/`, `research.md` |
| Tasks | `/speckit-tasks` | 将 Plan 拆解为可独立执行和测试的任务单元 | `tasks.md` |
| Implement | `/speckit-implement` | AI Agent 按任务列表逐个实现，人工审查增量变更 | 代码文件 |

**关键设计**：每个阶段都有 **人工审查检查点**——AI 生成制品，人验证正确性。这种门禁机制的有效性取决于审查者是否真正仔细阅读了制品（参见 8.2 节对于 Markdown 审查疲劳地讨论）。

### 3.3 安装与快速开始

**方案 A：uvx 临时运行（推荐，无需全局安装）**

```bash
# 第一步：安装 uv（如已安装可跳过）
curl -LsSf https://astral.sh/uv/install.sh | sh
source $HOME/.local/bin/env

# 第二步：在项目目录中直接初始化（uvx 会临时下载 specify-cli）
cd my-project
uvx --from git+https://github.com/github/spec-kit.git specify init --here --force --ai claude
```

**方案 B：全局安装后使用**

```bash
# 安装 specify-cli
uv tool install git+https://github.com/github/spec-kit.git

# 之后在任意项目中使用
cd my-project
specify init --here --force --ai claude      # Claude Code
specify init --here --force --ai cursor-agent    # Cursor
specify init --here --force --ai codex --ai-skills  # Codex CLI
```

> `--force` 参数：在**已有文件的项目**中初始化时必须加，跳过"目录非空"的交互确认；空目录可省略。

初始化后目录结构：

```
your-project/
├── .specify/                     # 规格文件存放目录
│   ├── constitution.md           # 项目宪法（由 /speckit-constitution 生成）
│   ├── specs/                    # 功能规格文档
│   └── plans/                    # 实现计划文档
└── .claude/
    └── skills/                   # spec-kit slash 命令
        ├── speckit-constitution/
        ├── speckit-specify/
        ├── speckit-clarify/
        ├── speckit-plan/
        ├── speckit-tasks/
        ├── speckit-implement/
        ├── speckit-analyze/
        ├── speckit-checklist/
        └── speckit-taskstoissues/
```

初始化完成后需**重启 Claude Code**，它会重新扫描 `.claude/skills/` 并加载这些命令。

### 3.4 社区生态

spec-kit 拥有丰富的社区扩展生态，按功能分类 [4]：

| 类别 | 代表扩展 | 用途                  |
|------|---------|---------------------|
| 流程编排 | Fleet Orchestrator, MAQA | 多 Agent 并行实现、人工审查门禁 |
| 质量保障 | Verify, Review, Cleanup | 实现后代码审查、规范符合性验证     |
| 文档同步 | Reconcile, Spec Sync | 检测规范与实现的漂移并修复       |
| 外部集成 | Jira, Azure DevOps, GitHub Projects | 将规范和任务同步到项目管理平台     |
| 项目健康 | Doctor, Status | 诊断项目结构、查看工作流进度      |

---

## 4. 宪法基石：约束 AI Agent 行为的九条法案

spec-kit 区别于其他 SDD 工具的核心设计是 **Constitution（项目宪法）**——不是普通的编码规范文档，而是一组显式的、版本化的约束规则，目标是让 AI Agent 在生成代码时遵守预定义的架构原则 [3]。

宪法的设计意图：**通过不可变原则来提升 AI 生成代码的一致性和质量**。其假设是：如果原则足够明确且嵌入到模板和提示中，AI 的行为就更可预测。但正如第 8.3 节将讨论的，这个假设并非总能成立。

spec-kit 的 `spec-driven.md` 文档将宪法分为九条核心法案 [3]：

### Article I：Library-First Principle（库优先原则）

> *"Every feature in Specify MUST begin its existence as a standalone library."*

**约束**：每个功能必须作为独立库开始，不允许直接在应用代码中实现。

**目的**：强制模块化设计。当 AI 生成实现方案时，必须将功能组织为具有清晰边界和最小依赖的可复用库组件，而非单体应用。

### Article II：CLI Interface Mandate（CLI 接口要求）

> *"All CLI interfaces MUST accept text as input and produce text as output."*

**约束**：每个库必须通过 CLI 暴露功能，支持 stdin/stdout 文本 I/O 和 JSON 结构化数据交换。

**目的**：强制可观测性和可测试性。AI 不能将功能隐藏在不透明的类内部——一切必须可通过文本接口验证。

### Article III：Test-First Imperative（测试先行要求，不可协商）

> *"This is NON-NEGOTIABLE: All implementation MUST follow strict Test-Driven Development."*

**约束**：
1. 先写单元测试
2. 测试经用户审查批准
3. 确认测试失败（Red 阶段）
4. 然后才能写实现代码

**目的**：彻底反转传统 AI 代码生成模式。不是先生成代码再期望它能工作，而是先生成全面的测试来定义行为，审批通过后再生成实现。

### Article IV - VI：（项目可根据需要定制）

这些条款通常涵盖安全性要求、代码审查流程、文档标准等，由各项目根据自身需求定义。

### Article VII：Simplicity（简洁性）

> *"Maximum 3 projects for initial implementation. Additional projects require documented justification."*

**约束**：初始实现最多 3 个项目/模块；额外项目需要书面理由。

**目的**：当 AI 自然倾向于创建复杂抽象时，这条法案迫使它为每一层复杂性提供正当理由。

### Article VIII：Anti-Abstraction（反过度抽象）

> *"Use framework features directly rather than wrapping them."*

**约束**：直接使用框架功能，不要包装；使用单一模型表示。

**目的**：与 Article VII 配对，共同抵制过度工程化。实现方案模板的 "Phase -1 Gates" 直接执行这些原则。

### Article IX：Integration-First Testing（集成优先测试）

> *"Tests MUST use realistic environments: prefer real databases over mocks, use actual service instances over stubs."*

**约束**：测试必须使用真实环境——优先使用真实数据库而非 mock，使用实际服务实例而非 stub，契约测试必须在实现之前编写。

**目的**：确保生成的代码在实践中可工作，而不仅仅是理论上可行。

### 宪法的力量

宪法的力量在于其 **不可变性** [3]。虽然实现细节可以演化，但核心原则保持恒定。这提供了：

1. **跨时间一致性**：今天生成的代码与明年遵循相同原则
2. **跨模型一致性**：不同 AI 模型产出架构兼容的代码
3. **架构完整性**：每个功能强化而非破坏系统设计
4. **质量保证**：测试先行、库优先、简洁性原则确保可维护的代码

宪法的修改有严格流程——需要明确的变更理由、维护者审批、向后兼容性评估 [3]。

---

## 5. 模板如何驯化 LLM：七种约束机制

spec-kit 的另一个核心设计是 **模板系统**。模板不仅仅是节省打字的快捷方式，而是精心设计的"LLM 行为约束器" [3]：

### 约束 1：阻止过早引入技术细节

```text
规范模板明确指示：
✅ Focus on WHAT users need and WHY
❌ Avoid HOW to implement (no tech stack, APIs, code structure)
```

这迫使 LLM 在规范阶段保持正确的抽象层级。当 LLM 自然倾向于跳到"用 React + Redux 实现"时，模板将其拉回到"用户需要实时数据更新" [3]。

### 约束 2：强制标记不确定性

```text
当从 prompt 创建规范时：
1. Mark all ambiguities: [NEEDS CLARIFICATION: specific question]
2. Don't guess: If the prompt doesn't specify something, mark it
```

阻止 LLM 常见的"做出看似合理但可能错误的假设"行为 [3]。

### 约束 3：通过 Checklist 强制结构化思考

```markdown
### Requirement Completeness
- [ ] No [NEEDS CLARIFICATION] markers remain
- [ ] Requirements are testable and unambiguous
- [ ] Success criteria are measurable
```

Checklist 就像规范的单元测试，迫使 LLM 系统性地自我审查输出 [3]。

### 约束 4：宪法合规门禁

```markdown
### Phase -1: Pre-Implementation Gates
#### Simplicity Gate (Article VII)
- [ ] Using ≤3 projects?
- [ ] No future-proofing?
#### Anti-Abstraction Gate (Article VIII)
- [ ] Using framework directly?
- [ ] Single model representation?
```

AI 在通过这些门禁之前不能继续，否则必须在"Complexity Tracking"部分记录正当理由 [3]。

### 约束 5：层级化信息管理

将详细的代码示例、算法和技术规格分离到 `implementation-details/` 子目录，主文档保持高层可读性 [3]。

### 约束 6：测试先行思维

文件创建顺序被强制为：先创建 `contracts/`（API 契约）→ 测试文件 → 源文件。确保 LLM 在实现之前先思考可测试性和契约 [3]。

### 约束 7：阻止投机性功能

```text
- [ ] No speculative or "might need" features
- [ ] All phases have clear prerequisites and deliverables
```

阻止 LLM 添加"可能需要"的功能，每个功能必须追溯到有明确验收标准的用户故事 [3]。

spec-kit 文档将这七种约束的复合效果描述为"将 LLM 从创意写作者转变为纪律规范工程师" [3]。这是其设计目标。实际效果取决于模型能力、上下文窗口大小和规范的具体质量——第 8 章会讨论已知的局限性。

---

<!-- Funnel Layer 4: Practical Guide — 动手实践时参考 -->

## 6. 工具实操：Claude Code / Codex / Cursor 的 SDD 落地

### 6.1 Claude Code 的 SDD 工作流

Claude Code 通过 `CLAUDE.md` 和 `.claude/` 目录原生支持 SDD 模式 [9][10]。

**Step 1：确认 claude CLI 在 PATH 中**

spec-kit 初始化时会检测 `claude` 命令是否存在。Claude Code 通过 npm 安装，路径通常类似 `~/.nvm/versions/node/v22.x.x/bin/claude`：

```bash
which claude   # 有输出说明已在 PATH 中
```

如果没有输出：

```bash
export PATH="$HOME/.nvm/versions/node/v22.19.0/bin:$PATH"
```

**Step 2：安装 spec-kit 并初始化项目**

推荐使用 `uvx` 临时运行（无需全局安装）：

```bash
# 安装 uv（如已安装可跳过）
curl -LsSf https://astral.sh/uv/install.sh | sh
source $HOME/.local/bin/env

# 在项目目录初始化
cd my-project
uvx --from git+https://github.com/github/spec-kit.git specify init --here --force --ai claude
```

或全局安装后使用：

```bash
uv tool install git+https://github.com/github/spec-kit.git
specify init --here --force --ai claude
```

> `--force` 在已有文件的项目中必须加，空目录可省略。

**初始化完成后，重启 Claude Code**（它需要重新扫描 `.claude/skills/` 目录来加载命令）。

**Step 3：五阶段执行（核心命令）**

以下为在 Claude Code 对话中输入的斜杠命令（不是终端命令）：

| 步骤 | 命令 | 作用 |
|------|------|------|
| 1 | `/speckit-constitution` | 定义项目技术原则、架构决策、编码规范（一次性，长期维护） |
| 2 | `/speckit-specify` | 用自然语言描述需求，生成结构化功能规格文档 |
| 3 | `/speckit-plan` | 基于规格生成详细技术实现方案 |
| 4 | `/speckit-tasks` | 将实现方案拆解为可执行的任务列表 |
| 5 | `/speckit-implement` | 按规格和计划执行代码实现 |

**可选增强命令**：

| 命令 | 何时使用 |
|------|---------|
| `/speckit-clarify` | 需求模糊时，在 `/speckit-plan` 前执行，澄清歧义 |
| `/speckit-analyze` | 在 `/speckit-tasks` 后执行，检查 constitution/spec/plan 三者一致性 |
| `/speckit-checklist` | 在 `/speckit-plan` 后执行，生成质量验证清单 |
| `/speckit-taskstoissues` | 将任务列表转换为 GitHub Issues，用于团队协作 |

**标准流程**（以 issue2md 为例）：

```text
/speckit-constitution 专注代码质量、测试标准和性能要求

/speckit-specify 一个 CLI 工具，将 GitHub Issue/PR/Discussion URL
转换为可归档的 Markdown 文档，支持 AI 摘要

/speckit-clarify   ← 可选：有歧义时澄清

/speckit-plan 使用 Go 标准库 net/http，go-github 做 REST 客户端，
GraphQL v4 获取 Discussion，最小外部依赖

/speckit-tasks

/speckit-implement
```

**Claude Code 的关键配置**：

| 文件 | 作用 | 优先级 |
|------|------|--------|
| `CLAUDE.md` | 项目级持久化指令（每次 session 自动加载） | 高 |
| `.claude/rules/` | 目录级或主题级规则 | 中 |
| `settings.json` | 客户端级强制规则（权限、Hook） | 最高（100% 执行） |

**高级技巧：使用 Hook 实现自动化门禁**

在 `.claude/settings.json` 中配置 PostToolUse hook，每次 Agent 修改文件后自动运行测试和 lint：

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [{ "type": "command", "command": "make test && make lint" }]
      }
    ]
  }
}
```

每次 Agent 完成写入操作后自动运行测试和 lint，确保质量门禁 [10][11]。

### 6.2 Codex CLI 的 SDD 工作流

OpenAI Codex CLI 通过 Skills 系统支持 SDD [4]：

终端初始化：

```bash
specify init --here --force --ai codex --ai-skills
```

> `--ai-skills` 是 Codex 专属参数，将 spec-kit 命令安装为 Agent Skills（Prompt.MD 格式），而非斜杠命令文件。

然后在 Codex 对话中使用 `$speckit-*` 前缀调用（Codex 的语法，不同于 Claude Code 的 `/`）：

```text
$speckit-constitution
$speckit-specify
$speckit-plan
$speckit-tasks
$speckit-implement
```

Codex 的 Skills 安装到 `.agents/skills/` 目录，作为可复用的工作流定义。

### 6.3 Cursor 的 SDD 工作流

Cursor 通过 Rules 和 Plan Mode 原生支持规范驱动模式 [4][11]：

**Step 1：创建项目规则**

```bash
specify init --here --force --ai cursor-agent
```

创建 `.cursor/rules/` 目录，包含 spec-kit 工作流规则。

**Step 2：使用 Plan Mode**

在 Agent 输入框中按 `Shift+Tab` 切换到 Plan Mode。Agent 会：

1. 搜索代码库，找到相关文件
2. 提出澄清问题
3. 创建详细的实现方案（Markdown 文件，可直接编辑）
4. 等待你审批后再执行 [11]

**Step 3：执行与审查**

Plan 默认保存到用户全局目录 `~/.cursor/plans/`；点击"Save to Workspace"可保存到项目级 `.cursor/plans/`（可纳入 Git 版本控制）。每次增量变更通过 diff view 审查，而非审查千行代码。

**Cursor 的特色功能**：

- **Worktree 并行**（Cursor 3+）：多个 Agent 在独立 git worktree 中并行工作，互不干扰 [11]
- **多模型对比**（Cursor 3+）：`/best-of-n` 命令将同一任务并行发给多个模型，各跑在独立 worktree，比较结果选最优方案 [11]
- **AI Code Review**：完成后 Agent 对生成的代码做逐行审查，识别 bug、性能问题和安全漏洞（changelog 2.1+ 引入）[11]

### 6.4 通用最佳实践（跨工具适用）

无论使用哪个 AI Agent，以下实践普遍适用：

**1. 永远先规划再编码**

芝加哥大学的研究发现，经验丰富的开发者更倾向于在 AI 生成代码之前先做规划 [11]。

**2. 让 Agent 自己找上下文**

不需要手动 tag 每个文件。现代 Agent 拥有强大的搜索工具（语义搜索、grep），能按需获取上下文 [11]。

**3. 提供可验证的目标**

使用类型化语言、配置 linter、编写测试——给 Agent 明确的信号来判断变更是否正确。Agent 在有清晰迭代目标时表现最好 [11]。

**4. 小步迭代，及时开新 session**

长对话会导致 Agent 失焦。完成一个逻辑单元后，开新 session 并通过 `@Past Chats` 或规范文件传递上下文 [11]。

**5. 将规范纳入版本控制**

规范文件和代码一起进入 Git。规范变更没有对应代码变更（或反之）在 Code Review 中立即可见 [1]。

---

## 7. 实战案例：issue2md 项目的 SDD 实践

本项目 [issue2md](https://github.com/johnqtcg/issue2md) 是 SDD + TDD 的直接产物，整体实现由 Codex 5.3 完成。以下展示 SDD 在实际 Go 项目中的完整落地过程。

### 7.1 项目简介

`issue2md` 是一个 CLI + Web 工具，将 GitHub Issue / PR / Discussion URL 转换为可归档的 Markdown 文档，支持可选的 AI 摘要。

### 7.2 SDD 制品追溯链

issue2md 实现了一条 **完整可追溯的 SDD 制品链**：

```
constitution.md（治理）
    │
    ▼
specs/001-core-functionality/spec.md（需求规范）
    │
    ▼
specs/001-core-functionality/plan.md（技术方案 + 宪法合规审查）
    │
    ▼
specs/001-core-functionality/tasks.md（可执行任务列表）
    │
    ▼
internal/（实现代码）+ AGENTS.md（AI Agent 指令）
```

### 7.3 宪法实例

issue2md 的 `constitution.md` 定义了四条核心法案。下表中"实际效果"列基于对代码库的静态观察（`go.mod` 依赖数、包职责划分），而非 commit 级别的 TDD 过程追踪：

| 法案 | 核心约束 | 代码库中的可观察效果 |
|------|---------|---------|
| Article 1: Simplicity First | YAGNI + 标准库优先 + 不过度工程化 | `go.mod` 仅 2 个直接依赖；Web 服务用 `net/http`，无 Gin/Echo |
| Article 2: Test-First | Red-Green-Refactor + Table-Driven Tests + 少用 Mock | `tasks.md` 中测试任务 ID 在实现任务之前；集成测试用 `httptest.Server` |
| Article 3: Clarity | 显式错误处理 + 无全局状态 + 有意义的注释 | 错误用 `fmt.Errorf("...: %w", err)` 包装；依赖通过构造函数注入 |
| Article 4: Single Responsibility | 包高内聚低耦合 + 接口隔离 | `internal/github` 只做 API 交互；`internal/converter` 只做 Markdown 转换 |

### 7.4 九条法案在本项目中的映射

spec-kit 的宪法有九条，issue2md 并没有照单全收，而是做了**合并与裁剪**，最终形成四条项目宪法。下表展示每条 spec-kit 法案在本项目中的落地情况：

| spec-kit 法案 | 对应的 issue2md 宪法条款 | 代码层的具体体现 |
|--------------|------------------------|----------------|
| **Article I**：Library-First | Article 4: Single Responsibility（合并） | `internal/` 下每个包都是独立库，对外只暴露 interface（`Fetcher`、`Renderer`、`URLParser`、`Runner`），实现细节全部私有 |
| **Article II**：CLI Interface Mandate | Article 4 + Article 3（合并） | `--stdout` 输出纯 Markdown 文本；状态行采用 `key=value` 格式（机器可解析）；`/convert` API 返回 `text/plain`；`/openapi.json` 提供 JSON 契约 |
| **Article III**：Test-First（不可协商） | Article 2: Test-First（直接映射） | `tasks.md` 中每个功能的 `_test.go` 任务 ID 严格排在对应 `.go` 之前；Table-Driven Tests 是项目唯一测试风格 |
| **Article IV–VI**：（项目自定义条款） | Article 3: Clarity（对应） | 本项目将"可见性和可解释性"收敛为 Clarity 法案：显式错误包装（`fmt.Errorf("...: %w")`）、无全局状态、构造函数注入依赖 |
| **Article VII**：Simplicity | Article 1: Simplicity First（直接映射） | `go.mod` 仅 2 个直接依赖；`main.go` 只有 17 行，零业务逻辑；Web 服务选 `net/http` 标准库，不引入 Gin/Echo |
| **Article VIII**：Anti-Abstraction | Article 1: Simplicity First（合并入 I） | 直接调用 `go-github` REST 客户端和 `graphql` 客户端，没有在上层再包一层"API 门面" |
| **Article IX**：Integration-First Testing | Article 2: Test-First（合并） | 集成测试用 `httptest.NewServer` 起真实 HTTP 服务，`github` 包的 fake server 完整模拟 GitHub API 响应，不用 mock 框架 |

**关键观察**：issue2md 将 spec-kit 的九条压缩成四条，核心逻辑是**按关注点合并**——Article I + II + VIII 都在讲"边界清晰、可观测"，统一进 Article 4（Single Responsibility）；Article III + IX 都在讲"测试先行、真实环境"，统一进 Article 2（Test-First）。这说明宪法不是模板，是可以根据项目性质取舍和重组的。

### 7.5 规范实例

`spec.md` 定义了完整的产品需求，包含 15 个章节：

- **目标与成功标准**：95% 的常见 Issue/PR/Discussion 可以一条命令导出
- **CLI 契约**：精确定义命令形状、Flag、参数验证规则
- **输出规范**：固定的文件命名模式、YAML Front Matter 必填字段、Body 各 Section 固定顺序
- **错误处理与重试策略**：4 次尝试（初始 + 3 次重试），退避序列 2s → 4s → 8s
- **退出码**：精确定义 0-5 六种退出码及其含义
- **验收标准**：功能、可靠性、可用性三个维度的具体可测试条件
- **测试计划**：单元测试（Table-Driven）、集成测试（in-memory fake server）、Golden Tests

### 7.6 技术方案中的宪法合规审查

`plan.md` 的第 2 章是 **Constitutional Compliance Review**——逐条对照宪法审查技术方案的合规性：

```markdown
## 2. Constitutional Compliance Review

### 2.1 Article 1: Simplicity First
- 1.1 YAGNI: implement only spec-defined MVP features
- 1.2 Standard Library First: net/http for web server; stdlib for markdown
- 1.3 No Over-Engineering: small packages and focused interfaces

### 2.2 Article 2: Test-First Imperative
- 2.1 TDD: every feature starts with failing tests
- 2.2 Table-Driven Tests: URL parsing, CLI validation, retry decision...
- 2.3 No Mock Abuse: prefer httptest.Server fake integration flows

Conclusion: this plan complies with all constitutional requirements.
```

这正是 spec-kit 所倡导的宪法门禁——AI Agent 在生成 Plan 之前，必须先进行宪法合规检查 [3]。

### 7.7 任务列表与 TDD 执行

`tasks.md` 将 Plan 拆解为 4 个阶段、30+ 个具体任务，每个任务包含：

| 字段 | 说明 |
|------|------|
| ID | 唯一标识（T001, T101...） |
| Parallel | 是否可并行 `[P]` |
| File | 目标文件路径 |
| Task | 具体任务描述 |
| Depends On | 依赖的前置任务 |

**TDD 循环在任务中的体现**（以 URL Parser 为例）：

```
T006 [P]  internal/parser/parser_test.go  ← Red: 写失败的 table-driven 测试
T007      internal/parser/parser.go       ← Green: 实现使测试通过
```

在 `tasks.md` 的任务排列中，测试文件（`_test.go`）的任务 ID 排在对应实现文件之前，体现了 TDD 的 Red-Green-Refactor 设计意图。需要说明的是，这一顺序来自规范制品，实际的 commit 历史是否严格遵循了此顺序，需要通过 `git log` 逐条验证。

### 7.8 规范到实现的映射

| 规范/方案定义 | 实现包 |
|-------------|--------|
| URL 解析、资源类型识别 | `internal/parser/` |
| Flag/ENV 解析 | `internal/config/` |
| REST + GraphQL 获取、重试、标准化 | `internal/github/` |
| Markdown 渲染、Front Matter、AI 摘要 | `internal/converter/` |
| 单条/批量执行、退出码映射 | `internal/cli/` |
| Web UI/API | `internal/webapp/` + `cmd/issue2mdweb/` |

### 7.9 质量门禁

`Makefile` 编码了 SDD 的质量门禁：

```bash
make test           # go test ./...
make cover-check    # 覆盖率门禁（默认 80%）
make lint           # golangci-lint
make fmt-check      # 格式化检查（CI 可因差异失败）
make ci             # fmt-check + cover-check + lint + build-all
```

这些质量门禁直接对应规范中的测试计划和宪法中的测试先行要求。

---

## 8. SDD 的边界与批评

SDD 并非银弹。来自 Martin Fowler 团队和社区的批评值得重视 [7]：

### 8.1 问题规模适配

Martin Fowler 团队在实际使用 spec-kit 时发现，对于 3-5 Story Point 的功能，SDD 工作流生成的大量 Markdown 文件"感觉像是大材小用"——用同样的时间通过普通 AI 辅助编码就能完成 [7]。

**建议**：SDD 适合中大型功能开发和新项目启动，不适合小 Bug 修复或简单改动。按复杂度选择方法：

| 任务规模 | 建议方法 |
|---------|---------|
| 小 Bug 修复、简单改动 | 直接 AI 辅助编码 |
| 中等功能（跨模块） | 轻量级 Spec（一个 spec.md + tasks.md） |
| 大型功能、新项目 | 完整 SDD 工作流（五阶段） |

### 8.2 Markdown 审查疲劳

spec-kit 生成大量 Markdown 文件，且文件之间存在重复内容。对于审查者来说，"我宁愿审查代码而不是这些 Markdown 文件" [7]。

**建议**：将规范审查纳入 PR 流程，用 diff 视图而非全文阅读。社区扩展 `Plan Review Gate` 要求 spec.md 和 plan.md 通过 MR/PR 合并后才能生成 Tasks。

### 8.3 控制的幻觉

即使有了模板、门禁和检查清单，Agent 仍然可能不遵循所有指令。Martin Fowler 团队观察到 Agent 忽略研究笔记（将已有类的描述当作新规范重新生成），也观察到 Agent 过于热切地遵循指令（过度执行宪法某条款） [7]。

**建议**：小步迭代仍然是保持控制的最佳方式。不要将 SDD 视为"一次性交付完美结果"，而是结构化的迭代框架。

### 8.4 形式化验证的缺失

社区批评指出，自然语言规范"减少模糊性但不消除模糊性"。跨模块接口一致性在模块数量增长时无法通过人工审查扩展 [12]。

**前沿探索**：有社区项目（如 formal-spec-driven-dev）尝试在 spec-kit 的自然语言层下增加形式化验证层（VDM-SL），通过类型系统和前/后置条件自动验证模块边界一致性 [12]。

### 8.5 何时使用 SDD

| 场景 | SDD 价值 |
|------|---------|
| 跨多系统/多团队的功能 | 高——`/constitution` 和 `/clarify` 早期暴露协调问题 |
| AI Agent 自主实现 | 高——结构化规范大幅减少幻觉需求 |
| 合规或安全要求 | 高——宪法编码可强制执行的全项目约束 |
| 新人 onboarding | 高——版本化的规范历史解释了"为什么这样决策" |
| 快速原型 | 低——overhead 不值得 |
| 简单且明确的任务 | 低——直接编码更快 |

---

## 9. 总结与行动建议

### 9.1 核心认知

1. **SDD 的核心主张是权力反转**：规范而非代码作为第一公民。这个主张在理论上清晰，在实践中仍在验证。
2. **宪法是 SDD 的差异化设计**：通过不可变原则约束 AI Agent 行为，目标是跨时间、跨模型的一致性——效果取决于原则的具体性和模型的遵循度。
3. **模板降低了 LLM 的自由度**：七种约束机制通过结构化输入减少输出的不确定性，但不能消除。
4. **SDD 不取代开发者判断力**：每个阶段的人工审查检查点是质量保障的核心。规范生成的质量上限由审查者的投入决定。
5. **SDD 有明确的适用边界**：中大型功能、新项目、需要可追溯性的场景最能发挥价值；对小任务可能是 overhead。

### 9.2 团队落地路径

**第 1 步（1 天）：建立项目宪法**

为现有项目编写 `constitution.md`，定义不可协商的原则（技术选型、测试标准、代码规范）。检入 Git，团队 Review。

**第 2 步（1 周）：选一个中等功能试点**

选择一个跨 2-3 个模块的功能需求，走完 SDD 五阶段流程。重点体验"规范驱动而非 prompt 驱动"的差异。

**第 3 步（2 周）：建立团队规范审查流程**

将 `spec.md` 和 `plan.md` 纳入 PR 流程。规范变更与代码变更同等对待——需要 Review 和 Approve。

**第 4 步（持续）：沉淀和迭代**

根据实际使用经验，迭代宪法和模板。添加团队特有的门禁、检查清单和扩展。

### 9.3 一条检验标准

> **Write specifications that pass one test: the AI agent can implement the feature independently without asking for clarification.**

如果你的规范能通过这个测试，你就真正掌握了 SDD。

---

## 附录 A：术语表

| 术语 | 定义 |
|------|------|
| SDD | Specification-Driven Development，规范驱动开发 |
| Vibe Coding | 基于模糊 prompt 的即兴 AI 编码模式 |
| Constitution | 项目宪法，定义不可变的开发原则 |
| spec-kit | GitHub 开源的 SDD 工具包 |
| Spec-first | SDD 成熟度层级 1：写好规范再生成代码 |
| Spec-anchored | SDD 成熟度层级 2：规范和代码共存且同步演化 |
| Spec-as-source | SDD 成熟度层级 3：规范是唯一制品，人不直接接触代码 |
| Gate | 门禁，阶段间的质量检查点 |
| TDD | Test-Driven Development，测试驱动开发 |
| YAGNI | You Aren't Gonna Need It，不要实现不需要的功能 |

## 附录 B：参考资料

[1] HTek Dev, "GitHub Spec-Kit: Turn English Into Production-Ready Specs," dev.to, 2026. https://dev.to/htekdev/github-spec-kit-turn-english-into-production-ready-specs-27o2 — T4 社区博客

[2] orchestrator.dev, "Spec-Driven Development: Building Production-Ready Software with AI," 2025-12. https://orchestrator.dev/blog/2025-12-16-spec_driven_dev_article/ — T3 独立技术出版物

[3] GitHub, "Spec-Driven Development (SDD) — spec-driven.md," github/spec-kit 仓库核心文档. https://github.com/github/spec-kit/blob/main/spec-driven.md — T1 官方文档

[4] GitHub, "spec-kit README.md," github/spec-kit 仓库. https://github.com/github/spec-kit — T1 官方文档

[5] GitHub Blog (wham), "Spec-driven development: Using Markdown as a programming language when building with AI," 2025-09-30. https://github.blog/ai-and-ml/generative-ai/spec-driven-development-using-markdown-as-a-programming-language-when-building-with-ai/ — T2 官方博客

[6] Product Builder (makr.io), "Spec-Driven Development: The 2026 Guide to Production AI Code," 2026-02. https://product.makr.io/learn/spec-driven-development — T3 独立技术出版物

[7] Thoughtworks (Martin Fowler), "Understanding Spec-Driven-Development: Kiro, spec-kit, and Tessl," martinfowler.com, 2025. https://martinfowler.com/articles/exploring-gen-ai/sdd-3-tools.html — T3 权威独立分析

[8] GitHub Blog (localden), "Spec-driven development with AI: Get started with a new open source toolkit," 2025-09-02. https://github.blog/ai-and-ml/generative-ai/spec-driven-development-with-ai-get-started-with-a-new-open-source-toolkit/ — T2 官方博客

[9] Claude Lab, "The Complete Guide to CLAUDE.md and AGENTS.md," claudelab.net, 2026. https://claudelab.net/en/articles/claude-code/claude-md-agents-md-complete-guide — T4 社区文档

[10] Softcery, "Agentic Coding with Claude Code and Cursor: Context Files, Workflows, and MCP Integration," softcery.com. https://softcery.com/lab/softcerys-guide-agentic-coding-best-practices — T3 独立实践指南

[11] Cursor, "Best practices for coding with agents," cursor.com. https://cursor.com/blog/agent-best-practices — T2 官方文档

[12] Hikaru Ando, "GitHub Spec Kit Is 80% Right — Here's the Missing 20% That Would Make It Transformative," dev.to, 2026. https://dev.to/kotaroyamame/github-spec-kit-is-80-right-heres-the-missing-20-that-would-make-it-transformative-2bi6 — T4 社区深度分析

---

*文档维护建议：季度审查。更新触发条件：spec-kit 发布大版本、支持的 AI Agent 工具有重大变更、团队 SDD 实践流程调整。*
