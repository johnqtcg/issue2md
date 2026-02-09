# issue2md Technical Implementation Plan

Version: v1.0  
Date: 2026-02-09  
Source Spec: `specs/001-core-functionality/spec.md`  
Constitution: `constitution.md`

## 1. Technical Context Summary

### 1.1 Constraints and Stack

- Language: Go `>= 1.25.0` (current `go.mod` is `go 1.25.4`)
- Web framework: Go standard library `net/http` only
- GitHub API: `google/go-github` (REST) plus GitHub GraphQL API v4
- Markdown output: standard library (`bytes.Buffer`, `strings.Builder`) with no Markdown third-party dependency
- Storage: no database in MVP; all data fetched in real time from APIs
- AI summary: OpenAI Responses API (`OPENAI_API_KEY`), with non-blocking degradation on failure

### 1.2 Minimal External Dependencies

- `github.com/google/go-github/v72/github`
- `golang.org/x/oauth2` (token transport)

No Gin/Echo, no ORM, no Markdown rendering library.

## 2. Constitutional Compliance Review

### 2.1 Article 1: Simplicity First

- 1.1 YAGNI: implement only spec-defined MVP features (Issue/PR/Discussion export, markdown format, retry policy, batch mode, summary degradation).
- 1.2 Standard Library First: `net/http` for web server; stdlib for markdown assembly; only essential GitHub client dependencies.
- 1.3 No Over-Engineering: small packages and focused interfaces; avoid complex layered abstractions.

### 2.2 Article 2: Test-First Imperative

- 2.1 TDD: every feature starts with failing tests (Red), then implementation (Green), then cleanup (Refactor).
- 2.2 Table-Driven Tests: URL parsing, CLI validation, retry decision, file naming, and renderer behavior all use table-driven tests.
- 2.3 No Mock Abuse: prefer `httptest.Server` fake integration flows over heavy mocking frameworks.

### 2.3 Article 3: Clarity and Explicitness

- 3.1 Error Handling: every error is explicitly handled and wrapped using `%w`.
- 3.2 No Global State: dependencies injected via constructors/struct fields.
- 3.3 Meaningful Comments: public APIs include GoDoc with rationale-oriented comments.

### 2.4 Article 4: Single Responsibility

- 4.1 Package Cohesion: `internal/github` handles only GitHub data access/normalization; `internal/converter` handles markdown conversion only; `internal/parser` handles URL parsing only.
- 4.2 Interface Segregation: small interfaces (`Fetcher`, `Renderer`, `URLParser`) instead of god interfaces.

Conclusion: this plan complies with all constitutional requirements.

## 3. Project Structure and Dependency Graph

### 3.1 Package Responsibilities

- `cmd/issue2md/`
  - CLI entrypoint; parses args, wires dependencies, executes single/batch export.
- `cmd/issue2mdweb/`
  - Web entrypoint (`net/http`); lightweight UI/API that reuses core `internal` packages.
- `internal/config/`
  - Parses flags + environment variables (`GITHUB_TOKEN`, `OPENAI_API_KEY`, etc.).
- `internal/parser/`
  - Parses GitHub URLs and identifies resource type (`issue`, `pull_request`, `discussion`).
- `internal/github/`
  - Fetches GitHub data (REST + GraphQL), handles pagination, retry, and normalization.
- `internal/converter/`
  - Renders normalized data into markdown with front matter and fixed sections; applies summary degradation behavior.
- `internal/cli/`
  - Orchestrates CLI use cases (single/batch, stdout/file output, final summary, exit code mapping).
- `web/templates/`
  - HTML templates for web mode.
- `web/static/`
  - Static assets for web mode.

### 3.2 Dependency Direction

`cmd/* -> internal/cli -> internal/{config,parser,github,converter}`  
`cmd/issue2mdweb -> internal/{config,parser,github,converter}`  
`internal/converter` depends on `internal/github` document model only.  
`internal/github` must never depend on `internal/converter`.

## 4. Core Data Structures

The following transport model covers spec-required fields and includes reaction aggregates for readability.

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
    CreatedAt string // keep original GitHub datetime string
    UpdatedAt string // keep original GitHub datetime string
    URL       string
    Labels    []Label

    // PR optional fields
    Merged      bool
    MergedAt    string
    ReviewCount int

    // Discussion optional fields
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
    Comments  []CommentNode // review thread inline comment bodies (no diff)
}

// IssueData is the unified normalized payload consumed by converter.
type IssueData struct {
    Meta        Metadata
    Description string
    Reactions   ReactionSummary // top-level body reactions

    Timeline []TimelineEvent // used by Issue
    Reviews  []ReviewData    // used by PR
    Thread   []CommentNode   // used by Issue/PR/Discussion
}
```

## 5. Public Interfaces of Internal Packages

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
    MaxRetries     int           // default 3 retries
    InitialBackoff time.Duration // default 2s
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
    Reason       string // skip reason when Status=skipped
}

type Summarizer interface {
    Summarize(ctx context.Context, data gh.IssueData, lang string) (Summary, error)
}

type RenderOptions struct {
    IncludeComments bool
    IncludeSummary  bool
    Lang            string // empty means auto-detect
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
    Token           string // --token first, then GITHUB_TOKEN
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
    Run(ctx context.Context, args []string) int // exit code
}
```

## 6. GitHub Data Fetching Strategy (REST + GraphQL)

### 6.1 Issue

- REST (`go-github`)
  - get issue metadata/body
  - list issue comments (paginated)
- GraphQL v4
  - fetch key timeline events (opened/closed/reopened/labeled/assigned/milestoned/locked)
  - supplement reactions aggregation when needed

### 6.2 Pull Request

- REST (`go-github`)
  - get PR metadata/body (including merged fields)
  - list reviews
  - list review comments (text only for thread context)
- GraphQL v4
  - supplement review thread structure where REST shape is insufficient
  - reactions aggregation

### 6.3 Discussion

- GraphQL v4 (primary channel)
  - fetch discussion metadata, category, answered state, accepted answer
  - fetch comments + nested replies (paginated)
  - reactions aggregation

## 7. Markdown Conversion Strategy

### 7.1 Fixed Section Order

Render sections strictly in spec order:

1. `# Title`
2. `## Metadata`
3. `## AI Summary` (optional)
4. `## Original Description`
5. `## Timeline` (Issue only)
6. `## Reviews` (PR only)
7. `## Discussion Thread`
8. `## References`

### 7.2 Front Matter

- YAML front matter strictly follows required/optional fields in spec.
- Datetime values are emitted as original GitHub strings (no timezone transformation).

### 7.3 Summary Degradation

If summary generation fails:

- do not fail the whole export
- omit `## AI Summary`
- add `summary_status: skipped (<reason>)` under metadata section

## 8. Retry, Error Handling, and Exit Codes

### 8.1 Retry Policy

Use a shared `doWithRetry(ctx, fn)`:

- total attempts: 4 (initial + 3 retries)
- backoff: `2s -> 4s -> 8s`
- retryable: transient network errors, 429/rate limit, 5xx
- non-retryable: 401/403 auth/permission, 404, invalid URL

### 8.2 Exit Codes

- `0`: all succeeded
- `2`: invalid CLI arguments
- `3`: auth/permission errors
- `4`: partial success in batch mode
- `5`: output conflict without `--force`
- `1`: generic runtime failure

## 9. TDD Execution Plan

### Phase 1: parser + config

- Red: table-driven tests for URL parsing and argument validation matrix
- Green: implement parser, validation, env override precedence
- Refactor: shared error types and GoDoc cleanup

### Phase 2: github fetcher

- Red: `httptest.Server` integration tests for issue/pr/discussion success and failures, pagination, retry behavior
- Green: implement REST + GraphQL fetching and normalization into `IssueData`
- Refactor: split query builder/mappers to keep function size small

### Phase 3: converter

- Red: golden tests for Issue/PR/Discussion + summary degradation tests
- Green: implement front matter + fixed section renderers
- Refactor: isolate section render functions

### Phase 4: cli runner + batch mode

- Red: tests for batch continue-on-error, failure summary, stdout conflict rules, and `--force` behavior
- Green: implement runner and output dispatch (stdout/file)
- Refactor: unify result aggregation/reporting structures

### Phase 5: issue2mdweb

- Red: `net/http/httptest` tests for form submit and error handling
- Green: implement minimal web entrypoint reusing core pipeline
- Refactor: extract common request handling helpers

## 10. Risks and Mitigations

- Risk: GraphQL schema/field evolution causes response parse breakage  
  Mitigation: request minimum required fields + tolerant parsing + regression integration tests.
- Risk: large thread pagination pressure and rate limits  
  Mitigation: paginated fetch + centralized retry + per-item failure isolation in batch mode.
- Risk: summary provider unavailability  
  Mitigation: strict degradation path that preserves export success.

## 11. Deliverables

- This plan: `specs/001-core-functionality/plan.md`
- Implementation will follow this plan and `spec.md`, with strict constitutional compliance (especially TDD and package cohesion).
