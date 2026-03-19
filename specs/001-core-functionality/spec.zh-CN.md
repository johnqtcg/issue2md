# issue2md 产品与技术规格说明

版本：v1.0（MVP）  
日期：2026-02-09  
状态：已批准实施

## 1. 概述

`issue2md` 是一个面向开源用户的 CLI 工具。  
它可以把 GitHub Issue / Pull Request / Discussion 的 URL 转换为可读的 Markdown 归档文档。

## 2. 目标与成功标准

### 2.1 目标

- 支持以下 GitHub 公共仓库 URL：
  - Issue
  - Pull Request
  - Discussion
- 导出结构化、易读的 Markdown，便于归档和沉淀知识。
- 同时支持单个 URL 导出和批量导出。
- 输出中可选附带 AI 摘要。

### 2.2 成功标准

- 至少 95% 的常见 Issue/PR/Discussion 页面可以通过一条命令导出。
- 导出的 Markdown 易于阅读，并保留关键讨论上下文。

## 3. 非目标（MVP）

- 支持私有仓库。
- 下载或本地化图片/附件。
- 导出 PR diff / 变更文件 / 提交历史。
- 支持 Markdown 以外的输出格式。

## 4. 输入与认证

### 4.1 URL 支持

- 单条模式下，接受一个 GitHub URL 作为位置参数。
- 批量模式下，通过 `--input-file` 接收 URL 列表（每行一个 URL）。

### 4.2 认证

- 在速率限制允许的情况下，公共仓库应支持无 Token 使用。
- Token 来源优先级：
  1. `--token`
  2. `GITHUB_TOKEN` 环境变量
- Token 类型：GitHub Personal Access Token。

## 5. CLI 约定

### 5.1 命令形式

```bash
issue2md <url> [flags]
issue2md --input-file urls.txt [flags]
```

### 5.2 参数

- `--output <dir|file>`
  - 单条模式：可选，输出文件路径或目录。
  - 批量模式：必填，且必须是目录路径。
- `--format <value>`
  - MVP 仅允许 `markdown`。
- `--include-comments <bool>`
  - 默认值：`true`。
  - 控制是否包含评论 / 回复 / Review 内容（按资源类型适用）。
- `--token <pat>`
  - 可选；优先级高于 `GITHUB_TOKEN`。
- `--stdout`
  - 将 Markdown 输出到标准输出。
  - 批量模式下不允许使用。
- `--force`
  - 覆盖已存在的文件。
  - 未指定 `--force` 时，默认不覆盖。
- `--lang <code>`
  - 指定摘要语言（例如 `zh`、`en`）。
  - 默认：根据源内容自动检测。

### 5.3 参数校验规则

- 单条模式必须且只能提供一个 URL 位置参数。
- 批量模式必须提供 `--input-file`，且不能再附带位置参数 URL。
- `--stdout` 与 `--input-file` 不能同时使用。
- `--format` 必须为 `markdown`。

## 6. 输出规范

### 6.1 文件命名

- 默认文件名格式：
  - Issue：`<owner>-<repo>-issue-<number>.md`
  - PR：`<owner>-<repo>-pr-<number>.md`
  - Discussion：`<owner>-<repo>-discussion-<number>.md`
- 如果目标文件已存在且未设置 `--force`，则该项返回冲突错误。

### 6.2 Front Matter（必需）

每个输出文件都必须以 YAML front matter 开头，并保留 GitHub 原始时间字符串。

必填字段：

- `type`（`issue` | `pull_request` | `discussion`）
- `title`
- `number`
- `state`
- `author`
- `created_at`
- `updated_at`
- `url`
- `labels`（数组）

按资源类型可选字段：

- PR：`merged`、`merged_at`、`review_count`
- Discussion：`category`、`is_answered`、`accepted_answer_author`

### 6.3 正文结构

文档结构（章节顺序固定）：

1. `# <Title>`
2. `## Metadata`
3. `## AI Summary`（如果没有摘要则省略）
4. `## Original Description`
5. `## Timeline`（仅 Issue，且只包含关键事件）
6. `## Reviews`（仅 PR）
7. `## Discussion Thread`（评论 / 回复）
8. `## References`

### 6.4 媒体处理

- 保留原始图片 / 附件链接的 Markdown 形式。
- 不做本地下载，也不重写链接。

## 7. 按资源类型的内容规则

### 7.1 Issue

应包含：

- Issue 标题和正文
- 评论（当 `--include-comments=true` 时）
- 仅包含以下关键时间线事件：
  - opened
  - closed
  - reopened
  - labeled
  - assigned
  - milestoned
  - locked

### 7.2 Pull Request

应包含：

- PR 标题和正文
- Review 评论范围：
  - Review 摘要评论（approve / request changes / comment）
  - Review 线程中的行内评论文本
- 不包含：
  - Diff patch
  - Commit 历史详情

### 7.3 Discussion

应包含：

- Discussion 标题和正文
- 评论及其嵌套回复
- 已采纳答案及采纳状态

## 8. AI 摘要规范（MVP）

### 8.1 摘要章节

启用且可用时，`## AI Summary` 必须包含：

- `### Summary`
- `### Key Decisions`
- `### Action Items`

### 8.2 失败降级

- 如果 AI 能力不可用（缺少 Key、提供方报错、超时等）：
  - 导出流程仍应成功完成。
  - 省略 `## AI Summary`。
  - 在 `## Metadata` 下增加一条简短说明：
    - `summary_status: skipped (<reason>)`

### 8.3 语言

- 如果提供了 `--lang`，则强制使用该语言。
- 否则根据原始讨论文本自动检测语言。

## 9. 错误处理与重试策略

### 9.1 需要重试的场景

自动重试以下情况：

- 临时性网络错误
- 与 GitHub 限流相关的临时失败

以下情况不重试：

- 权限 / 认证错误（由于访问范围导致的 401/403）
- 非法 URL
- 资源不存在（404）

### 9.2 重试策略

- 最大尝试次数：共 4 次（首次 + 3 次重试）
- 失败后的退避序列：
  - 2s
  - 4s
  - 8s

### 9.3 批量失败行为

- 某个 URL 失败时，继续处理剩余 URL。
- 只要有任意 URL 失败，进程就以非零状态退出。
- 最终失败汇总必须包含：
  - 输入 URL
  - 规范化后的资源类型（如果可识别）
  - 简短错误原因

## 10. 退出码

- `0`：全部成功
- `1`：通用 / 运行时错误
- `2`：CLI 参数无效
- `3`：认证 / 授权错误
- `4`：批量模式部分成功（至少一项失败）
- `5`：输出冲突（文件已存在且未指定 `--force`）

## 11. 日志与用户体验

- 默认输出人类可读日志。
- 每个条目打印状态行：`OK` / `FAILED`。
- 最终汇总包含以下计数：
  - total
  - succeeded
  - failed
- 错误信息在可能的情况下必须包含可执行的建议。

## 12. 架构约束（与 Constitution 对齐）

- 优先使用标准库。
- 不使用全局可变状态。
- 所有错误必须显式用 `%w` 包装。
- `internal/` 下职责清晰分离：
  - GitHub 抓取层
  - 领域转换层
  - Markdown 渲染层
  - CLI 编排层

## 13. 验收标准

### 13.1 功能性

- 单个 URL 导出对 Issue、PR、Discussion 都可正常工作。
- 通过 `--input-file` 的批量导出可正常工作，并能在单项失败时继续执行。
- `--stdout` 仅在单条模式可用，在批量模式下必须拒绝。
- `--force` 能正确控制覆盖行为。
- Front matter 必填字段完整且正确。
- PR 输出包含 Review 摘要和线程评论，但不包含 diff / commits。
- Issue 输出只包含规定的关键时间线事件。
- Discussion 输出包含已采纳答案和回复层级结构。
- 当 AI 摘要可用时，会输出包含 3 个必需子章节的摘要部分。
- AI 摘要失败不会导致整体导出失败。

### 13.2 可靠性

- 重试次数和退避时长符合 `2s/4s/8s`。
- 认证 / 权限失败不应重试。

### 13.3 易用性

- 对于长讨论串，输出的 Markdown 仍应保持良好可读性。
- 批量执行的最终报告能清晰列出失败项。

## 14. 测试计划（实施阶段）

### 14.1 单元测试（表驱动）

- URL 解析与资源类型识别。
- CLI 参数校验矩阵。
- 文件名生成与冲突逻辑。
- 重试判定与退避计划。
- 按资源类型渲染 Markdown 章节。
- AI 摘要的包含 / 省略条件。

### 14.2 集成测试

- 使用内存中的伪 GitHub HTTP 服务。
- 覆盖分页，以及批量执行中成功 / 失败混合场景。
- 校验 front matter 字段和关键章节是否存在。

### 14.3 Golden Tests

- 为以下场景准备 Golden Markdown 快照：
  - 典型 Issue
  - 典型且 Review 较多的 PR 讨论串
  - 带有已采纳答案的典型 Discussion

## 15. 最终确认的决策

### 15.1 AI 提供方与凭证

- MVP 的摘要提供方为 OpenAI Responses API。
- 凭证变量：
  - `OPENAI_API_KEY`（生成摘要时必需）。
- 可选提供方配置：
  - `ISSUE2MD_AI_BASE_URL`（可选，用于兼容网关 / 代理）。
  - `ISSUE2MD_AI_MODEL`（可选；未设置时使用内置默认模型）。
- 如果缺少 `OPENAI_API_KEY` 或提供方调用失败，导出仍应成功，摘要按 8.2 节定义跳过。

### 15.2 `--input-file` 大小策略

- MVP 不设硬性上限。
- 工具必须按行流式处理输入文件，避免一次性将所有 URL 读入内存。
- 空行应被忽略。

### 15.3 机器可读日志

- MVP 不包含 `--json` 日志模式。
- MVP 仅提供人类可读日志（见第 11 节）。
