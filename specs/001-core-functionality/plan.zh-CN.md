# issue2md 技术实现计划

版本：v1.0  
日期：2026-02-09  
来源规格：`specs/001-core-functionality/spec.md`  
项目宪章：`constitution.md`

## 1. 技术背景概览

### 1.1 约束与技术栈

- 语言：Go `>= 1.25.0`（当前 `go.mod` 为 `go 1.25.4`）
- Web 框架：仅使用 Go 标准库 `net/http`
- GitHub API：`google/go-github`（REST）+ GitHub GraphQL API v4
- Markdown 输出：仅使用标准库（`bytes.Buffer`、`strings.Builder`），不引入第三方 Markdown 依赖
- 存储：MVP 不使用数据库；所有数据都通过 API 实时获取
- AI 摘要：使用 OpenAI Responses API（`OPENAI_API_KEY`），失败时非阻塞降级

### 1.2 最小化外部依赖

- `github.com/google/go-github/v72/github`
- `golang.org/x/oauth2`（Token 传输）

不使用 Gin/Echo，不使用 ORM，不使用 Markdown 渲染库。

## 2. 宪章合规性审查

### 2.1 第 1 条：简单优先

- 1.1 YAGNI：只实现规格中定义的 MVP 功能（Issue/PR/Discussion 导出、Markdown 格式、重试策略、批量模式、摘要降级）。
- 1.2 标准库优先：Web 服务用 `net/http`；Markdown 组装用标准库；仅引入必需的 GitHub 客户端依赖。
- 1.3 不过度设计：保持包小而专一，接口聚焦，避免复杂分层抽象。

### 2.2 第 2 条：测试优先（不可妥协）

- 2.1 TDD：每个功能都从失败测试开始（Red），再实现通过（Green），最后整理重构（Refactor）。
- 2.2 表驱动测试：URL 解析、CLI 校验、重试判断、文件命名、渲染器行为等统一采用表驱动测试。
- 2.3 不滥用 Mock：优先使用 `httptest.Server` 构造伪集成流程，而不是依赖沉重的 Mock 框架。

### 2.3 第 3 条：清晰与显式

- 3.1 错误处理：所有错误都必须显式处理，并通过 `%w` 包装。
- 3.2 禁止全局状态：依赖通过构造函数或结构体字段显式注入。
- 3.3 有意义的注释：公开 API 需提供 GoDoc，并解释设计意图。

### 2.4 第 4 条：单一职责

- 4.1 包内聚性：`internal/github` 只负责 GitHub 数据访问与规范化；`internal/converter` 只负责 Markdown 转换；`internal/parser` 只负责 URL 解析。
- 4.2 接口隔离：使用小接口（如 `Fetcher`、`Renderer`、`URLParser`），避免“大而全”的接口。

结论：本计划符合全部宪章要求。

## 3. 项目结构与依赖关系

### 3.1 包职责

- `cmd/issue2md/`
  - CLI 入口；解析参数、组装依赖、执行单条 / 批量导出。
- `cmd/issue2mdweb/`
  - Web 入口（`net/http`）；提供轻量 UI / API，并复用核心 `internal` 包。
- `internal/config/`
  - 解析命令行参数与环境变量（`GITHUB_TOKEN`、`OPENAI_API_KEY` 等）。
- `internal/parser/`
  - 解析 GitHub URL，识别资源类型（`issue`、`pull_request`、`discussion`）。
- `internal/github/`
  - 获取 GitHub 数据（REST + GraphQL），处理分页、重试和规范化。
- `internal/converter/`
  - 将规范化数据渲染为 Markdown，生成 front matter 和固定章节，并处理摘要降级。
- `internal/cli/`
  - 编排 CLI 用例（单条 / 批量、stdout / 文件输出、最终汇总、退出码映射）。
- `web/templates/`
  - Web 模式的 HTML 模板。
- `web/static/`
  - Web 模式的静态资源。

### 3.2 依赖方向

`cmd/* -> internal/cli -> internal/{config,parser,github,converter}`  
`cmd/issue2mdweb -> internal/{config,parser,github,converter}`  
`internal/converter` 只能依赖 `internal/github` 的文档模型。  
`internal/github` 绝不能依赖 `internal/converter`。

## 4. 核心数据结构

下面的传输模型覆盖了规格要求的字段，并加入了聚合后的 reaction 信息，方便阅读。

```go
package github

type ResourceType string

const (
    ResourceIssue       ResourceType = "issue"
    ResourcePullRequest ResourceType = "pull_request"
    ResourceDiscussion  ResourceType = "discussion"
)

type ResourceRef struct {
    Owner  string
    Repo   string
    Number int
    Type   ResourceType
    URL    string
}

type ReactionSummary struct {
    PlusOne  int // +1
    MinusOne int // -1
    Laugh    int
    Hooray   int
    Confused int
    Heart    int
    Rocket   int
    Eyes     int
    Total    int
}

type Label struct {
    Name string
}

type Metadata struct {
    Type      ResourceType
    Title     string
    Number    int
    State     string
    Author    string
    CreatedAt string // 保留 GitHub 原始时间字符串
    UpdatedAt string // 保留 GitHub 原始时间字符串
    URL       string
    Labels    []Label

    // PR 可选字段
    Merged      bool
    MergedAt    string
    ReviewCount int

    // Discussion 可选字段
    Category             string
    IsAnswered           bool
    AcceptedAnswerAuthor string
}

type TimelineEvent struct {
    EventType string // opened/closed/reopened/labeled/assigned/milestoned/locked
    Actor     string
    CreatedAt string
    Details   string
}

type CommentNode struct {
    ID        string
    Author    string
    Body      string
    CreatedAt string
    UpdatedAt string
    URL       string
    Reactions ReactionSummary
    Replies   []CommentNode
}

type ReviewData struct {
    ID        string
    State     string // APPROVED/CHANGES_REQUESTED/COMMENTED
    Author    string
    Body      string
    CreatedAt string
    Reactions ReactionSummary
    Comments  []CommentNode // review 线程中的行内评论正文（不含 diff）
}

// IssueData 是提供给 converter 使用的统一规范化载荷。
type IssueData struct {
    Meta        Metadata
    Description string
    Reactions   ReactionSummary // 顶层正文的 reactions

    Timeline []TimelineEvent // Issue 使用
    Reviews  []ReviewData    // PR 使用
    Thread   []CommentNode   // Issue/PR/Discussion 使用
}
```

## 5. internal 包的公开接口

### 5.1 `internal/parser`

```go
package parser

import gh "github.com/johnqtcg/issue2md/internal/github"

type URLParser interface {
    Parse(rawURL string) (gh.ResourceRef, error)
}

func New() URLParser
```

### 5.2 `internal/github`

```go
package github

import (
    "context"
    "net/http"
    "time"
)

type FetchOptions struct {
    IncludeComments bool
}

type Fetcher interface {
    Fetch(ctx context.Context, ref ResourceRef, opts FetchOptions) (IssueData, error)
}

type Config struct {
    Token          string
    HTTPClient     *http.Client
    MaxRetries     int           // 默认重试 3 次
    InitialBackoff time.Duration // 默认 2s
}

func NewFetcher(cfg Config) (Fetcher, error)
```

### 5.3 `internal/converter`

```go
package converter

import (
    "context"

    gh "github.com/johnqtcg/issue2md/internal/github"
)

type Summary struct {
    Summary      string
    KeyDecisions []string
    ActionItems  []string
    Language     string
    Status       string // ok | skipped
    Reason       string // 当 Status=skipped 时记录跳过原因
}

type Summarizer interface {
    Summarize(ctx context.Context, data gh.IssueData, lang string) (Summary, error)
}

type RenderOptions struct {
    IncludeComments bool
    IncludeSummary  bool
    Lang            string // 为空表示自动检测
}

type Renderer interface {
    Render(ctx context.Context, data gh.IssueData, opts RenderOptions) ([]byte, error)
}

func NewRenderer(summarizer Summarizer) Renderer
```

### 5.4 `internal/config`

```go
package config

type Config struct {
    OutputPath      string
    Format          string // markdown
    IncludeComments bool
    Stdout          bool
    Force           bool
    InputFile       string
    Token           string // 优先 --token，其次 GITHUB_TOKEN
    SummaryLang     string
    OpenAIAPIKey    string
    OpenAIBaseURL   string
    OpenAIModel     string
}

type Loader interface {
    Load(args []string) (Config, error)
}

func NewLoader() Loader
```

### 5.5 `internal/cli`

```go
package cli

import "context"

type Runner interface {
    Run(ctx context.Context, args []string) int // 返回退出码
}
```

## 6. GitHub 数据抓取策略（REST + GraphQL）

### 6.1 Issue

- REST（`go-github`）
  - 获取 Issue 元数据和正文
  - 获取 Issue 评论列表（分页）
- GraphQL v4
  - 获取关键时间线事件（opened/closed/reopened/labeled/assigned/milestoned/locked）
  - 在需要时补充 reactions 聚合信息

### 6.2 Pull Request

- REST（`go-github`）
  - 获取 PR 元数据和正文（含 merged 字段）
  - 获取 reviews 列表
  - 获取 review comments（仅保留线程上下文所需文本）
- GraphQL v4
  - 当 REST 返回结构不足时，补充 review thread 结构
  - 获取 reactions 聚合信息

### 6.3 Discussion

- GraphQL v4（主要通道）
  - 获取 Discussion 元数据、分类、是否已回答、已采纳答案
  - 获取评论及嵌套回复（分页）
  - 获取 reactions 聚合信息

## 7. Markdown 转换策略

### 7.1 固定章节顺序

必须严格按规格中的顺序渲染章节：

1. `# Title`
2. `## Metadata`
3. `## AI Summary`（可选）
4. `## Original Description`
5. `## Timeline`（仅 Issue）
6. `## Reviews`（仅 PR）
7. `## Discussion Thread`
8. `## References`

### 7.2 Front Matter

- YAML front matter 必须严格遵循规格中要求的必填 / 可选字段。
- 时间字段直接输出 GitHub 原始字符串，不做时区转换。

### 7.3 摘要降级

如果摘要生成失败：

- 不影响整体导出
- 省略 `## AI Summary`
- 在 metadata 章节下增加 `summary_status: skipped (<reason>)`

## 8. 重试、错误处理与退出码

### 8.1 重试策略

使用统一的 `doWithRetry(ctx, fn)`：

- 总尝试次数：4 次（首次 + 3 次重试）
- 退避：`2s -> 4s -> 8s`
- 可重试：临时网络错误、429 / 限流、5xx
- 不可重试：401/403 认证或权限问题、404、非法 URL

### 8.2 退出码

- `0`：全部成功
- `2`：CLI 参数无效
- `3`：认证 / 权限错误
- `4`：批量模式部分成功
- `5`：未指定 `--force` 时输出冲突
- `1`：通用运行时失败

## 9. TDD 执行计划

### Phase 1：parser + config

- Red：为 URL 解析和参数校验矩阵编写表驱动测试
- Green：实现解析器、参数校验、环境变量覆盖优先级
- Refactor：整理共享错误类型并补充 GoDoc

### Phase 2：github fetcher

- Red：为 issue/pr/discussion 的成功与失败、分页、重试行为编写 `httptest.Server` 集成测试
- Green：实现 REST + GraphQL 抓取，并规范化为 `IssueData`
- Refactor：拆分查询构造和映射逻辑，保持函数短小清晰

### Phase 3：converter

- Red：为 Issue/PR/Discussion 编写 golden tests，并覆盖摘要降级场景
- Green：实现 front matter 和固定章节渲染
- Refactor：抽离各章节渲染函数

### Phase 4：cli runner + batch mode

- Red：为批量继续执行、失败汇总、stdout 冲突规则和 `--force` 行为编写测试
- Green：实现 runner 和输出分发（stdout / 文件）
- Refactor：统一结果聚合与汇报结构

### Phase 5：issue2mdweb

- Red：为表单提交和错误处理编写 `net/http/httptest` 测试
- Green：实现最小可用的 Web 入口，并复用核心处理链路
- Refactor：抽取公共请求处理辅助函数

## 10. 风险与缓解措施

- 风险：GraphQL schema / 字段演进导致响应解析失败  
  缓解：只请求最小必需字段 + 容错解析 + 回归集成测试。
- 风险：长讨论串带来的分页压力与限流  
  缓解：分页抓取 + 统一重试 + 批量模式下按项隔离失败。
- 风险：摘要服务不可用  
  缓解：严格执行降级路径，确保导出仍然成功。

## 11. 交付物

- 本计划文件：`specs/001-core-functionality/plan.md`
- 后续实现将遵循本计划与 `spec.md`，并严格遵守项目宪章（尤其是 TDD 和包职责内聚）。
