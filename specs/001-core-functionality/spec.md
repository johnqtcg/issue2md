# issue2md Product & Technical Specification

Version: v1.0 (MVP)  
Date: 2026-02-09  
Status: Approved for implementation

## 1. Overview

`issue2md` is a CLI tool for open-source users.  
It converts a GitHub Issue / Pull Request / Discussion URL into a readable Markdown archive document.

## 2. Goals and Success Criteria

### 2.1 Goals

- Support GitHub public repository URLs for:
  - Issue
  - Pull Request
  - Discussion
- Export structured, readable Markdown for archival and knowledge capture.
- Support both single-URL and batch export workflows.
- Provide optional AI summary in output.

### 2.2 Success Criteria

- At least 95% of common Issue/PR/Discussion pages can be exported in one command.
- Exported Markdown is readable and keeps key discussion context.

## 3. Non-Goals (MVP)

- Private repository support.
- Downloading or localizing images/attachments.
- PR diff / changed files / commit history export.
- Output formats other than Markdown.

## 4. Input and Authentication

### 4.1 URL Support

- Accept one GitHub URL as positional argument in single mode.
- Accept batch URLs from `--input-file` (one URL per line) in batch mode.

### 4.2 Authentication

- Public repos should work without token when rate limits allow.
- Token source priority:
  1. `--token`
  2. `GITHUB_TOKEN` environment variable
- Token type: GitHub Personal Access Token.

## 5. CLI Contract

### 5.1 Command Shape

```bash
issue2md <url> [flags]
issue2md --input-file urls.txt [flags]
```

### 5.2 Flags

- `--output <dir|file>`
  - Single mode: optional output file path or directory.
  - Batch mode: required directory path.
- `--format <value>`
  - MVP allowed value: `markdown` only.
- `--include-comments <bool>`
  - Default: `true`.
  - Controls inclusion of comments/replies/reviews where applicable.
- `--token <pat>`
  - Optional; overrides `GITHUB_TOKEN`.
- `--stdout`
  - Print Markdown to stdout.
  - Not allowed in batch mode.
- `--force`
  - Overwrite existing files.
  - Default behavior without `--force`: do not overwrite.
- `--lang <code>`
  - Summary language override (e.g. `zh`, `en`).
  - Default: auto-detect from source content.

### 5.3 Argument Validation Rules

- Single mode requires exactly one URL positional argument.
- Batch mode requires `--input-file` and must not include positional URL.
- `--stdout` and `--input-file` cannot be used together.
- `--format` must be `markdown`.

## 6. Output Specification

### 6.1 File Naming

- Default filename pattern:
  - Issue: `<owner>-<repo>-issue-<number>.md`
  - PR: `<owner>-<repo>-pr-<number>.md`
  - Discussion: `<owner>-<repo>-discussion-<number>.md`
- If target exists and `--force` is not set, return conflict error for that item.

### 6.2 Front Matter (Required)

Each output file starts with YAML front matter and preserves GitHub original datetime strings.

Required fields:

- `type` (`issue` | `pull_request` | `discussion`)
- `title`
- `number`
- `state`
- `author`
- `created_at`
- `updated_at`
- `url`
- `labels` (array)

Optional fields by type:

- PR: `merged`, `merged_at`, `review_count`
- Discussion: `category`, `is_answered`, `accepted_answer_author`

### 6.3 Body Structure

Document structure (section order fixed):

1. `# <Title>`
2. `## Metadata`
3. `## AI Summary` (omit this section if summary unavailable)
4. `## Original Description`
5. `## Timeline` (Issue only, key events only)
6. `## Reviews` (PR only)
7. `## Discussion Thread` (comments/replies)
8. `## References`

### 6.4 Media Handling

- Preserve original image/attachment links in Markdown.
- No local download or rewriting.

## 7. Resource-Specific Content Rules

### 7.1 Issue

Include:

- Issue title + body
- Comments (when `--include-comments=true`)
- Key timeline events only:
  - opened
  - closed
  - reopened
  - labeled
  - assigned
  - milestoned
  - locked

### 7.2 Pull Request

Include:

- PR title + body
- Review comments scope:
  - Review summary comments (approve/request changes/comment)
  - Review thread inline comments text
- Exclude:
  - Diff patches
  - Commit history details

### 7.3 Discussion

Include:

- Discussion title + body
- Comments and nested replies
- Accepted answer and accepted status

## 8. AI Summary Specification (MVP)

### 8.1 Summary Sections

When enabled and available, `## AI Summary` must contain:

- `### Summary`
- `### Key Decisions`
- `### Action Items`

### 8.2 Failure Degradation

- If AI capability is unavailable (missing key, provider error, timeout, etc.):
  - Continue export successfully.
  - Omit `## AI Summary`.
  - Add a short note under `## Metadata`:
    - `summary_status: skipped (<reason>)`

### 8.3 Language

- If `--lang` provided, force that language.
- Otherwise auto-detect language from source discussion text.

## 9. Error Handling and Retry Policy

### 9.1 Retry Cases

Auto-retry for:

- transient network errors
- GitHub rate limit related temporary failures

No retry for:

- permission/auth errors (401/403 due to access scope)
- invalid URL
- not found (404)

### 9.2 Retry Strategy

- Max attempts: 4 total (initial + 3 retries)
- Backoff sequence after failures:
  - 2s
  - 4s
  - 8s

### 9.3 Batch Failure Behavior

- Continue processing remaining URLs when one URL fails.
- Exit non-zero if any URL failed.
- Print final failure summary containing:
  - input URL
  - normalized resource type (if resolvable)
  - short error reason

## 10. Exit Codes

- `0`: all items succeeded
- `1`: generic/runtime error
- `2`: invalid CLI arguments
- `3`: authentication/authorization error
- `4`: partial success in batch mode (at least one failed)
- `5`: output conflict (file exists without `--force`)

## 11. Logging and UX

- Human-readable logs by default.
- For each item print status line: `OK` / `FAILED`.
- Final summary includes counts:
  - total
  - succeeded
  - failed
- Error messages must include actionable suggestion when possible.

## 12. Architecture Constraints (Constitution Alignment)

- Standard library first.
- No global mutable state.
- Explicit error wrapping with `%w`.
- Keep responsibilities separated under `internal/`:
  - GitHub fetch layer
  - domain transformation layer
  - markdown rendering layer
  - CLI orchestration layer

## 13. Acceptance Criteria

### 13.1 Functional

- Single URL export works for Issue, PR, Discussion.
- Batch export via `--input-file` works and continues on per-item failures.
- `--stdout` works only for single mode and is rejected in batch mode.
- `--force` controls overwrite behavior correctly.
- Front matter required fields exist and are correct.
- PR output contains review summary + thread comments but no diff/commits.
- Issue output includes only specified key timeline events.
- Discussion output includes accepted answer and reply hierarchy.
- AI summary section appears with required 3 subsections when available.
- AI summary failure does not fail overall export.

### 13.2 Reliability

- Retry attempts and backoff durations follow 2s/4s/8s.
- Auth/permission failures are not retried.

### 13.3 Usability

- Output Markdown remains readable for long threads.
- Batch final report clearly lists failures.

## 14. Test Plan (for implementation phase)

### 14.1 Unit Tests (Table-Driven)

- URL parsing and resource type detection.
- CLI argument validation matrix.
- filename generation and conflict logic.
- retry decision and backoff schedule.
- markdown section rendering by resource type.
- AI summary inclusion/omission conditions.

### 14.2 Integration Tests

- Use in-memory fake GitHub HTTP server.
- Cover pagination and mixed success/failure batch run.
- Verify front matter fields and key section presence.

### 14.3 Golden Tests

- Golden Markdown snapshots for:
  - typical Issue
  - typical PR review-heavy thread
  - typical Discussion with accepted answer

## 15. Finalized Decisions

### 15.1 AI Provider and Credentials

- MVP summary provider is OpenAI Responses API.
- Credential variable:
  - `OPENAI_API_KEY` (required for summary generation).
- Optional provider settings:
  - `ISSUE2MD_AI_BASE_URL` (optional, for compatible gateway/proxy).
  - `ISSUE2MD_AI_MODEL` (optional; if unset, use built-in default model).
- If `OPENAI_API_KEY` is missing or provider call fails, export still succeeds and summary is skipped as defined in section 8.2.

### 15.2 `--input-file` Size Policy

- No hard cap in MVP.
- The tool must process input file line-by-line (streaming style) to avoid loading all URLs into memory.
- Empty lines should be ignored.

### 15.3 Machine-Readable Logging

- `--json` log mode is not included in MVP.
- MVP provides human-readable logs only (section 11).
