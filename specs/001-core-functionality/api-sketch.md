# API Sketch: Core Functionality

Date: 2026-02-09  
Scope: `internal/github` and `internal/converter`

## 1. Design Intent

- Keep package responsibilities cohesive:
  - `internal/github`: only GitHub data fetching and normalization.
  - `internal/converter`: only rendering normalized data to Markdown.
- Keep boundaries explicit to support table-driven unit tests and fake-based integration tests.

## 2. `internal/github` Package Sketch

### 2.1 Resource Type

```go
package github

type ResourceType string

const (
    ResourceIssue      ResourceType = "issue"
    ResourcePullRequest ResourceType = "pull_request"
    ResourceDiscussion  ResourceType = "discussion"
)
```

### 2.2 Core Models (normalized, converter-facing)

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
    CreatedAt string // keep GitHub original string format
    UpdatedAt string // keep GitHub original string format
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
    Comments  []Comment // review thread inline comments body only
}

type Document struct {
    Type ResourceType
    Meta Metadata

    Description string

    Timeline []TimelineEvent // issue only
    Reviews  []Review        // pr only
    Thread   []Comment       // comments/replies

    // discussion only
    Category             string
    IsAnswered           bool
    AcceptedAnswerAuthor string
}
```

### 2.3 Client Interface

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

### 2.4 Constructor

```go
package github

import (
    "net/http"
    "time"
)

type Config struct {
    Token       string
    HTTPClient  *http.Client
    MaxRetries  int           // default 3
    BackoffBase time.Duration // default 2s
}

func NewClient(cfg Config) (Client, error)
```

## 3. `internal/converter` Package Sketch

### 3.1 Summary Models

```go
package converter

type Summary struct {
    Summary      string
    KeyDecisions []string
    ActionItems  []string
    Language     string
    Status       string // ok | skipped
    Reason       string // filled when skipped
}
```

### 3.2 Summary Provider Interface

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

### 3.3 Renderer Interface and Constructor

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
    Lang           string // auto when empty
}

func NewRenderer(summarizer Summarizer) Renderer
```

### 3.4 Optional Top-Level Helper

```go
package converter

import (
    "context"

    gh "issue2md/internal/github"
)

func ConvertToMarkdown(ctx context.Context, doc gh.Document, opts RenderOptions, r Renderer) ([]byte, error)
```

## 4. Boundary and Error Notes

- `internal/github` returns normalized `github.Document`; it does not format Markdown.
- `internal/converter` does not call GitHub APIs directly.
- If summary generation fails, renderer must still return Markdown (without AI section) and embed summary skip status in metadata section.
- All wrapped errors should follow `%w` for upstream handling.
