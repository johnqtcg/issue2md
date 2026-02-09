# 001 Core Functionality Task List

Date: 2026-02-09  
Inputs: `specs/001-core-functionality/spec.md`, `specs/001-core-functionality/plan.md`, `constitution.md`

Legend: `[P]` means the task can run in parallel once its dependencies are satisfied.

## Phase 1: Foundation (Data Structures and Contracts)

| ID | Parallel | File | Task | Depends On |
|---|---|---|---|---|
| T001 |  | `go.mod` | Add required dependencies (`go-github/v72`, `oauth2`) and keep module tidy. | - |
| T002 | [P] | `internal/github/types_contract_test.go` | Add failing contract tests asserting required normalized model fields exist for Issue/PR/Discussion payloads (including reactions). | - |
| T003 |  | `internal/github/types.go` | Implement core transport structs (`ResourceType`, `ResourceRef`, `ReactionSummary`, `Metadata`, `TimelineEvent`, `CommentNode`, `ReviewData`, `IssueData`). | T002 |
| T004 | [P] | `internal/github/interfaces_contract_test.go` | Add failing tests for `Fetcher` interface behavior expectations and config defaults contract. | T003 |
| T005 |  | `internal/github/interfaces.go` | Define `FetchOptions`, `Fetcher`, and package-level constructor signatures for fetcher wiring. | T004 |
| T006 | [P] | `internal/parser/parser_test.go` | Add table-driven failing tests for Issue/PR/Discussion URL parsing and invalid URL cases. | T003 |
| T007 |  | `internal/parser/parser.go` | Implement URL parser returning normalized `ResourceRef` for supported URL types. | T006 |
| T008 | [P] | `internal/config/loader_test.go` | Add table-driven failing tests for flag/env precedence (`--token` over `GITHUB_TOKEN`), `--format=markdown` enforcement, `--lang` loading, and stdout/input-file conflict. | - |
| T009 |  | `internal/config/loader.go` | Implement config loader that merges CLI args and env vars per spec. | T008 |
| T010 | [P] | `internal/cli/args_test.go` | Add table-driven failing tests for argument matrix validation in single vs batch mode. | T009 |
| T011 |  | `internal/cli/args.go` | Implement CLI argument validation helpers used by runner layer. | T010, T007 |
| T012 | [P] | `internal/config/errors_test.go` | Add failing tests for typed config/validation errors and user-facing messages. | T009 |
| T013 |  | `internal/config/errors.go` | Implement explicit config error types and wrapping helpers. | T012 |

## Phase 2: GitHub Fetcher (API Interaction, TDD)

| ID | Parallel | File | Task | Depends On |
|---|---|---|---|---|
| T101 | [P] | `internal/github/retry_test.go` | Add table-driven failing tests for retryable classification and backoff sequence `2s/4s/8s` with max 3 retries. | T005 |
| T102 |  | `internal/github/retry.go` | Implement retry executor (`doWithRetry`) and retryable error classification helpers. | T101 |
| T103 | [P] | `internal/github/rest_client_test.go` | Add failing tests for REST client auth header behavior and request error wrapping. | T005 |
| T104 |  | `internal/github/rest_client.go` | Implement REST client wrapper on top of `go-github` with token-aware transport injection. | T103, T001 |
| T105 | [P] | `internal/github/graphql_client_test.go` | Add failing tests for GraphQL request shape, auth usage, pagination cursor handling, and decode errors. | T005 |
| T106 |  | `internal/github/graphql_client.go` | Implement GraphQL v4 HTTP client and pagination helper primitives. | T105 |
| T107 | [P] | `internal/github/fetch_issue_test.go` | Add failing `httptest.Server` integration-style tests for issue metadata/body/comments/key timeline events/reactions mapping. | T102, T104, T106 |
| T108 |  | `internal/github/fetch_issue.go` | Implement Issue fetching and normalization into `IssueData` (including key timeline event filtering). | T107 |
| T109 | [P] | `internal/github/fetch_pr_test.go` | Add failing tests for PR metadata, reviews, review-thread comments (no diff/commit content), and reactions mapping. | T102, T104, T106 |
| T110 |  | `internal/github/fetch_pr.go` | Implement PR fetch + normalization with review summaries and review-thread comments only. | T109 |
| T111 | [P] | `internal/github/fetch_discussion_test.go` | Add failing tests for Discussion metadata, accepted answer, nested replies, and reactions mapping. | T102, T106 |
| T112 |  | `internal/github/fetch_discussion.go` | Implement Discussion fetch + normalization via GraphQL into `IssueData`. | T111 |
| T113 | [P] | `internal/github/fetcher_test.go` | Add failing tests for top-level fetch dispatcher by `ResourceType` and `IncludeComments` option behavior. | T108, T110, T112 |
| T114 |  | `internal/github/fetcher.go` | Implement top-level fetcher orchestration and dispatch logic. | T113 |

## Phase 3: Markdown Converter (TDD)

| ID | Parallel | File | Task | Depends On |
|---|---|---|---|---|
| T201 | [P] | `internal/converter/testdata/issue.golden.md` | Create Issue golden markdown fixture matching section order and front matter requirements. | T003 |
| T202 | [P] | `internal/converter/testdata/pr.golden.md` | Create PR golden markdown fixture with review sections and no diff/commit details. | T003 |
| T203 | [P] | `internal/converter/testdata/discussion.golden.md` | Create Discussion golden markdown fixture with accepted answer and nested replies. | T003 |
| T204 | [P] | `internal/converter/frontmatter_test.go` | Add failing tests for required/optional front matter fields and original datetime string preservation. | T003 |
| T205 |  | `internal/converter/frontmatter.go` | Implement YAML front matter rendering helpers from normalized metadata. | T204 |
| T206 | [P] | `internal/converter/section_issue_test.go` | Add failing tests for Issue sections: description, key timeline events, discussion thread rendering. | T205, T201 |
| T207 |  | `internal/converter/section_issue.go` | Implement Issue-specific markdown section renderers. | T206 |
| T208 | [P] | `internal/converter/section_pr_test.go` | Add failing tests for PR sections: description, reviews, review-thread comments, excluded diff/commit info. | T205, T202 |
| T209 |  | `internal/converter/section_pr.go` | Implement PR-specific markdown section renderers. | T208 |
| T210 | [P] | `internal/converter/section_discussion_test.go` | Add failing tests for Discussion sections: accepted answer and reply hierarchy rendering. | T205, T203 |
| T211 |  | `internal/converter/section_discussion.go` | Implement Discussion-specific markdown section renderers. | T210 |
| T212 | [P] | `internal/converter/summary_test.go` | Add failing tests for summary success path, degrade path (skip section, include `summary_status`), and language behavior (`--lang` override vs auto-detect). | T003 |
| T213 |  | `internal/converter/summary_openai.go` | Implement OpenAI Responses API summarizer and failure-to-skip mapping. | T212 |
| T214 | [P] | `internal/converter/renderer_test.go` | Add failing tests for global section order and `--include-comments` behavior across all resource types. | T207, T209, T211, T213 |
| T215 |  | `internal/converter/renderer.go` | Implement top-level renderer that assembles full markdown document. | T214 |

## Phase 4: CLI Assembly (Entry Integration)

| ID | Parallel | File | Task | Depends On |
|---|---|---|---|---|
| T301 | [P] | `internal/cli/output_test.go` | Add failing tests for filename pattern, output path behavior, stdout mode, and `--force` overwrite rule. | T011 |
| T302 |  | `internal/cli/output.go` | Implement output writer for stdout/file modes and conflict checks. | T301 |
| T303 | [P] | `internal/cli/exitcode_test.go` | Add failing tests mapping runtime outcomes to spec exit codes (0/1/2/3/4/5). | T011 |
| T304 |  | `internal/cli/exitcode.go` | Implement exit code resolver functions. | T303 |
| T305 | [P] | `internal/cli/report_test.go` | Add failing tests for final human-readable run summary (`OK`/`FAILED`, counts, and failure entries with URL/resource-type/reason). | T011 |
| T306 |  | `internal/cli/report.go` | Implement batch/single report formatter. | T305 |
| T307 | [P] | `internal/cli/input_reader_test.go` | Add failing tests for `--input-file` streaming line-by-line processing and empty-line skipping. | T011 |
| T308 |  | `internal/cli/input_reader.go` | Implement streaming input-file reader that ignores empty lines. | T307 |
| T309 | [P] | `internal/cli/runner_single_test.go` | Add failing tests for single URL flow: parse -> fetch -> convert -> write/stdout -> exit code mapping. | T011, T114, T215, T302, T304 |
| T310 |  | `internal/cli/runner_single.go` | Implement single-item runner orchestration and error wrapping. | T309 |
| T311 | [P] | `internal/cli/runner_batch_test.go` | Add failing tests for batch mode: input-file streaming, continue on per-item failure, and final summary counts. | T011, T114, T215, T302, T304, T306, T308 |
| T312 |  | `internal/cli/runner_batch.go` | Implement batch runner with per-item isolation and failure aggregation. | T311 |
| T313 | [P] | `cmd/issue2md/main_test.go` | Add failing smoke tests for CLI main wiring and argument passthrough to runner. | T310, T312 |
| T314 |  | `cmd/issue2md/main.go` | Implement CLI main entrypoint and dependency wiring. | T313 |
| T315 | [P] | `cmd/issue2mdweb/main_test.go` | Add failing `httptest` smoke tests for web endpoint wiring using shared parser/fetcher/converter pipeline. | T114, T215 |
| T316 |  | `cmd/issue2mdweb/main.go` | Implement `net/http` web entrypoint (minimal MVP form/API flow). | T315 |
| T317 | [P] | `web/templates/index.html` | Create minimal web template for URL input and markdown result display. | T316 |
| T318 | [P] | `web/static/style.css` | Create minimal static stylesheet for web page readability. | T316 |
