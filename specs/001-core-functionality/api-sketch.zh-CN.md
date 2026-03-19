# API 草图：核心功能

日期：2026-02-09  
范围：`internal/github` 与 `internal/converter`

## 1. 设计意图

- 保持包职责内聚：
  - `internal/github`：只负责 GitHub 数据抓取与规范化。
  - `internal/converter`：只负责把规范化数据渲染为 Markdown。
- 明确边界，便于编写表驱动单元测试和基于 fake 的集成测试。

## 2. `internal/github` 包草图

### 2.1 资源类型

```go
package github

type ResourceType string

const (
    ResourceIssue      ResourceType = "issue"
    ResourcePullRequest ResourceType = "pull_request"
    ResourceDiscussion  ResourceType = "discussion"
)
```

### 2.2 核心模型（规范化后，面向 converter）

```go
package github

type ResourceRef struct {
    Owner  string
    Repo   string
    Number int
    Type   ResourceType
    URL    string
}

type Metadata struct {
    Title     string
    Number    int
    State     string
    Author    string
    CreatedAt string // 保留 GitHub 原始字符串格式
    UpdatedAt string // 保留 GitHub 原始字符串格式
    URL       string
    Labels    []string
}

type TimelineEvent struct {
    Type      string // opened/closed/reopened/labeled/assigned/milestoned/locked
    Actor     string
    CreatedAt string
    Details   string
}

type Comment struct {
    ID        string
    Author    string
    Body      string
    CreatedAt string
    Replies   []Comment
}

type Review struct {
    ID        string
    State     string // APPROVED / CHANGES_REQUESTED / COMMENTED
    Author    string
    Body      string
    CreatedAt string
    Comments  []Comment // 仅保留 review 线程中的行内评论正文
}

type Document struct {
    Type ResourceType
    Meta Metadata

    Description string

    Timeline []TimelineEvent // 仅 Issue 使用
    Reviews  []Review        // 仅 PR 使用
    Thread   []Comment       // 评论 / 回复

    // 仅 Discussion 使用
    Category             string
    IsAnswered           bool
    AcceptedAnswerAuthor string
}
```

### 2.3 客户端接口

```go
package github

import "context"

type Client interface {
    FetchDocument(ctx context.Context, ref ResourceRef, opts FetchOptions) (Document, error)
}

type FetchOptions struct {
    IncludeComments bool
}
```

### 2.4 构造函数

```go
package github

import (
    "net/http"
    "time"
)

type Config struct {
    Token       string
    HTTPClient  *http.Client
    MaxRetries  int           // 默认 3 次
    BackoffBase time.Duration // 默认 2s
}

func NewClient(cfg Config) (Client, error)
```

## 3. `internal/converter` 包草图

### 3.1 摘要模型

```go
package converter

type Summary struct {
    Summary      string
    KeyDecisions []string
    ActionItems  []string
    Language     string
    Status       string // ok | skipped
    Reason       string // skipped 时填写
}
```

### 3.2 摘要提供方接口

```go
package converter

import (
    "context"

    gh "issue2md/internal/github"
)

type Summarizer interface {
    Summarize(ctx context.Context, doc gh.Document, lang string) (Summary, error)
}
```

### 3.3 渲染器接口与构造函数

```go
package converter

import (
    "context"

    gh "issue2md/internal/github"
)

type Renderer interface {
    Render(ctx context.Context, doc gh.Document, opts RenderOptions) ([]byte, error)
}

type RenderOptions struct {
    IncludeSummary bool
    Lang           string // 为空时自动检测
}

func NewRenderer(summarizer Summarizer) Renderer
```

### 3.4 可选的顶层辅助函数

```go
package converter

import (
    "context"

    gh "issue2md/internal/github"
)

func ConvertToMarkdown(ctx context.Context, doc gh.Document, opts RenderOptions, r Renderer) ([]byte, error)
```

## 4. 边界与错误处理说明

- `internal/github` 返回规范化后的 `github.Document`；它不负责生成 Markdown。
- `internal/converter` 不直接调用 GitHub API。
- 如果摘要生成失败，渲染器仍必须返回 Markdown（不包含 AI 摘要章节），并在 metadata 章节中写入摘要跳过状态。
- 所有包装错误都应遵循 `%w`，方便上层继续处理。
