# 001 核心功能任务清单

日期：2026-02-09  
输入：`specs/001-core-functionality/spec.md`、`specs/001-core-functionality/plan.md`、`constitution.md`

说明：`[P]` 表示在依赖满足后，该任务可并行执行。

## Phase 1：基础层（数据结构与契约）

| ID | 并行 | 文件 | 任务 | 依赖 |
|---|---|---|---|---|
| T001 |  | `go.mod` | 添加必需依赖（`go-github/v72`、`oauth2`），并保持模块整洁。 | - |
| T002 | [P] | `internal/github/types_contract_test.go` | 添加失败中的契约测试，断言 Issue/PR/Discussion 规范化载荷中存在必需模型字段（包含 reactions）。 | - |
| T003 |  | `internal/github/types.go` | 实现核心传输结构体（`ResourceType`、`ResourceRef`、`ReactionSummary`、`Metadata`、`TimelineEvent`、`CommentNode`、`ReviewData`、`IssueData`）。 | T002 |
| T004 | [P] | `internal/github/interfaces_contract_test.go` | 添加失败测试，约束 `Fetcher` 接口行为预期和配置默认值契约。 | T003 |
| T005 |  | `internal/github/interfaces.go` | 定义 `FetchOptions`、`Fetcher` 以及 fetcher 装配所需的包级构造签名。 | T004 |
| T006 | [P] | `internal/parser/parser_test.go` | 为 Issue/PR/Discussion URL 解析和非法 URL 场景添加表驱动失败测试。 | T003 |
| T007 |  | `internal/parser/parser.go` | 实现 URL 解析器，为支持的 URL 类型返回规范化后的 `ResourceRef`。 | T006 |
| T008 | [P] | `internal/config/loader_test.go` | 为 flag / env 优先级（`--token` 高于 `GITHUB_TOKEN`）、`--format=markdown` 限制、`--lang` 读取以及 stdout / input-file 冲突添加表驱动失败测试。 | - |
| T009 |  | `internal/config/loader.go` | 实现配置加载器，按规格合并 CLI 参数与环境变量。 | T008 |
| T010 | [P] | `internal/cli/args_test.go` | 为单条 / 批量模式下的参数组合校验添加表驱动失败测试。 | T009 |
| T011 |  | `internal/cli/args.go` | 实现 runner 层使用的 CLI 参数校验辅助函数。 | T010, T007 |
| T012 | [P] | `internal/config/errors_test.go` | 为类型化配置 / 校验错误及其面向用户的提示信息添加失败测试。 | T009 |
| T013 |  | `internal/config/errors.go` | 实现显式的配置错误类型与包装辅助函数。 | T012 |

## Phase 2：GitHub Fetcher（API 交互，TDD）

| ID | 并行 | 文件 | 任务 | 依赖 |
|---|---|---|---|---|
| T101 | [P] | `internal/github/retry_test.go` | 为可重试错误分类和 `2s/4s/8s` 退避序列（最多重试 3 次）添加表驱动失败测试。 | T005 |
| T102 |  | `internal/github/retry.go` | 实现重试执行器（`doWithRetry`）以及可重试错误分类辅助函数。 | T101 |
| T103 | [P] | `internal/github/rest_client_test.go` | 为 REST 客户端认证头行为和请求错误包装添加失败测试。 | T005 |
| T104 |  | `internal/github/rest_client.go` | 基于 `go-github` 实现 REST 客户端封装，并注入支持 Token 的传输层。 | T103, T001 |
| T105 | [P] | `internal/github/graphql_client_test.go` | 为 GraphQL 请求结构、认证使用、分页游标处理和解码错误添加失败测试。 | T005 |
| T106 |  | `internal/github/graphql_client.go` | 实现 GraphQL v4 HTTP 客户端和分页辅助基础能力。 | T105 |
| T107 | [P] | `internal/github/fetch_issue_test.go` | 使用 `httptest.Server` 添加失败中的集成风格测试，覆盖 issue 元数据 / 正文 / 评论 / 关键时间线事件 / reactions 映射。 | T102, T104, T106 |
| T108 |  | `internal/github/fetch_issue.go` | 实现 Issue 抓取并规范化为 `IssueData`（包含关键时间线事件过滤）。 | T107 |
| T109 | [P] | `internal/github/fetch_pr_test.go` | 为 PR 元数据、reviews、review-thread comments（不含 diff / commit 内容）及 reactions 映射添加失败测试。 | T102, T104, T106 |
| T110 |  | `internal/github/fetch_pr.go` | 实现 PR 抓取与规范化，只保留 review 摘要和 review-thread comments。 | T109 |
| T111 | [P] | `internal/github/fetch_discussion_test.go` | 为 Discussion 元数据、已采纳答案、嵌套回复及 reactions 映射添加失败测试。 | T102, T106 |
| T112 |  | `internal/github/fetch_discussion.go` | 通过 GraphQL 实现 Discussion 抓取，并规范化为 `IssueData`。 | T111 |
| T113 | [P] | `internal/github/fetcher_test.go` | 为顶层 fetch 分发逻辑和 `IncludeComments` 选项行为添加失败测试。 | T108, T110, T112 |
| T114 |  | `internal/github/fetcher.go` | 实现顶层 fetcher 编排与分发逻辑。 | T113 |

## Phase 3：Markdown Converter（TDD）

| ID | 并行 | 文件 | 任务 | 依赖 |
|---|---|---|---|---|
| T201 | [P] | `internal/converter/testdata/issue.golden.md` | 创建符合章节顺序和 front matter 要求的 Issue golden markdown 基准文件。 | T003 |
| T202 | [P] | `internal/converter/testdata/pr.golden.md` | 创建包含 review 章节且不含 diff / commit 详情的 PR golden markdown 基准文件。 | T003 |
| T203 | [P] | `internal/converter/testdata/discussion.golden.md` | 创建包含已采纳答案和嵌套回复的 Discussion golden markdown 基准文件。 | T003 |
| T204 | [P] | `internal/converter/frontmatter_test.go` | 为必填 / 可选 front matter 字段以及原始时间字符串保留行为添加失败测试。 | T003 |
| T205 |  | `internal/converter/frontmatter.go` | 根据规范化元数据实现 YAML front matter 渲染辅助函数。 | T204 |
| T206 | [P] | `internal/converter/section_issue_test.go` | 为 Issue 章节添加失败测试：描述、关键时间线事件、讨论串渲染。 | T205, T201 |
| T207 |  | `internal/converter/section_issue.go` | 实现 Issue 专属的 Markdown 章节渲染器。 | T206 |
| T208 | [P] | `internal/converter/section_pr_test.go` | 为 PR 章节添加失败测试：描述、reviews、review-thread comments，以及排除 diff / commit 信息。 | T205, T202 |
| T209 |  | `internal/converter/section_pr.go` | 实现 PR 专属的 Markdown 章节渲染器。 | T208 |
| T210 | [P] | `internal/converter/section_discussion_test.go` | 为 Discussion 章节添加失败测试：已采纳答案和回复层级渲染。 | T205, T203 |
| T211 |  | `internal/converter/section_discussion.go` | 实现 Discussion 专属的 Markdown 章节渲染器。 | T210 |
| T212 | [P] | `internal/converter/summary_test.go` | 为摘要成功路径、降级路径（跳过章节并写入 `summary_status`）以及语言行为（`--lang` 覆盖与自动检测）添加失败测试。 | T003 |
| T213 |  | `internal/converter/summary_openai.go` | 实现基于 OpenAI Responses API 的摘要器，以及失败到跳过的映射逻辑。 | T212 |
| T214 | [P] | `internal/converter/renderer_test.go` | 为全局章节顺序和 `--include-comments` 在所有资源类型中的行为添加失败测试。 | T207, T209, T211, T213 |
| T215 |  | `internal/converter/renderer.go` | 实现负责组装完整 Markdown 文档的顶层渲染器。 | T214 |

## Phase 4：CLI 装配（入口集成）

| ID | 并行 | 文件 | 任务 | 依赖 |
|---|---|---|---|---|
| T301 | [P] | `internal/cli/output_test.go` | 为文件名模式、输出路径行为、stdout 模式和 `--force` 覆盖规则添加失败测试。 | T011 |
| T302 |  | `internal/cli/output.go` | 实现 stdout / 文件模式的输出写入器和冲突检查。 | T301 |
| T303 | [P] | `internal/cli/exitcode_test.go` | 为运行结果到规格退出码（0/1/2/3/4/5）的映射添加失败测试。 | T011 |
| T304 |  | `internal/cli/exitcode.go` | 实现退出码解析函数。 | T303 |
| T305 | [P] | `internal/cli/report_test.go` | 为最终的人类可读运行汇总添加失败测试（`OK` / `FAILED`、计数，以及包含 URL / 资源类型 / 原因的失败条目）。 | T011 |
| T306 |  | `internal/cli/report.go` | 实现单条 / 批量报告格式化器。 | T305 |
| T307 | [P] | `internal/cli/input_reader_test.go` | 为 `--input-file` 按行流式处理和空行跳过行为添加失败测试。 | T011 |
| T308 |  | `internal/cli/input_reader.go` | 实现按流读取输入文件的逻辑，并忽略空行。 | T307 |
| T309 | [P] | `internal/cli/runner_single_test.go` | 为单条 URL 流程添加失败测试：parse -> fetch -> convert -> write/stdout -> exit code mapping。 | T011, T114, T215, T302, T304 |
| T310 |  | `internal/cli/runner_single.go` | 实现单项 runner 编排逻辑和错误包装。 | T309 |
| T311 | [P] | `internal/cli/runner_batch_test.go` | 为批量模式添加失败测试：输入文件流式处理、单项失败后继续，以及最终汇总计数。 | T011, T114, T215, T302, T304, T306, T308 |
| T312 |  | `internal/cli/runner_batch.go` | 实现批量 runner，支持逐项隔离和失败聚合。 | T311 |
| T313 | [P] | `cmd/issue2md/main_test.go` | 为 CLI 主入口装配和参数透传到 runner 添加失败 smoke tests。 | T310, T312 |
| T314 |  | `cmd/issue2md/main.go` | 实现 CLI 主入口及依赖装配。 | T313 |
| T315 | [P] | `cmd/issue2mdweb/main_test.go` | 使用 `httptest` 为 Web 端点装配添加失败 smoke tests，复用 parser / fetcher / converter 管线。 | T114, T215 |
| T316 |  | `cmd/issue2mdweb/main.go` | 实现 `net/http` Web 入口（最小可用 MVP 表单 / API 流程）。 | T315 |
| T317 | [P] | `web/templates/index.html` | 创建一个最小可用的 Web 模板，用于输入 URL 和展示 Markdown 结果。 | T316 |
| T318 | [P] | `web/static/style.css` | 创建一个最小静态样式表，提升页面可读性。 | T316 |
