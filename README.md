# issue2md

[![CI](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml/badge.svg)](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml)
![Go Version](https://img.shields.io/badge/go-1.25.8-00ADD8)
![License](https://img.shields.io/badge/license-MIT-blue)

Turn GitHub `Issue`, `Pull Request`, and `Discussion` URLs into clean Markdown for archiving, sharing, and downstream automation.

Language:
- English (primary): `README.md`
- Chinese: [README.zh-CN.md](README.zh-CN.md)

## Contents

- [Overview](#overview)
- [Highlights](#highlights)
- [Install](#install)
- [Quick Start](#quick-start)
- [End-to-End Example](#end-to-end-example)
- [Configuration and Environment](#configuration-and-environment)
- [Common Commands](#common-commands)
- [Project Docs](#project-docs)

<a id="overview"></a>
## Overview

- Dual entrypoints: `CLI tool + backend web service`.
- Go version: `go 1.25.8` (from `go.mod`).
- Module path: `github.com/johnqtcg/issue2md`.

<a id="highlights"></a>
## Highlights

- One tool, two entrypoints: use the CLI for local export or the Web service for browser and API-driven workflows.
- Optional AI summary: when `OPENAI_API_KEY` is configured, output can include a structured `## AI Summary` section with summary, decisions, and action items.
- Structured output by default: rendered markdown preserves metadata, original description, discussion thread, and source reference URL.

<a id="install"></a>
## Install

Directly with Go:

```bash
go install github.com/johnqtcg/issue2md/cmd/issue2md@latest
go install github.com/johnqtcg/issue2md/cmd/issue2mdweb@latest
```

From a local clone:
- `make install-cli`
- `make install-web`

<a id="quick-start"></a>
## Quick Start

### 30-Second CLI

```bash
export GITHUB_TOKEN=<your_github_pat>
issue2md --stdout https://github.com/github/spec-kit/issues/75
```

### 30-Second Web

```bash
export GITHUB_TOKEN=<your_github_pat>
ISSUE2MD_WEB_ADDR=127.0.0.1:18080 issue2mdweb
```

Then call the HTTP endpoint:

```bash
curl -sS -X POST http://127.0.0.1:18080/convert \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'url=https://github.com/github/spec-kit/issues/75'
```

Notes:
- Requires Go `>= 1.25` and an installed binary.
- If you prefer local binaries, use `make build-cli` or `make web`.
- Contributor checks remain in `Common Commands` and `Testing and Quality`.

<a id="end-to-end-example"></a>
## End-to-End Example

Convert one issue URL and let the CLI choose the default output filename:

```bash
issue2md https://github.com/github/spec-kit/issues/75
# OK url=https://github.com/github/spec-kit/issues/75 type=issue output=github-spec-kit-issue-75.md
```

Default filenames follow:

```text
<owner>-<repo>-<issue|pr|discussion>-<number>.md
```

Generated markdown shape, excerpted from [`internal/converter/testdata/issue.golden.md`](internal/converter/testdata/issue.golden.md):

```markdown
---
type: 'issue'
title: 'Issue: Panic on nil config'
number: 123
---

# Issue: Panic on nil config

## Metadata
- type: issue
- number: 123
- state: open

## AI Summary

### Summary
The thread discusses root cause and fix.

## Original Description

App panics when config is nil.
```

The generated file always includes metadata, original description, thread content, and references. The `## AI Summary` section appears only when `OPENAI_API_KEY` is configured.

<a id="project-structure"></a>
## Project Structure

```text
.
├── cmd/
│   ├── issue2md/            # CLI entrypoint
│   └── issue2mdweb/         # Web entrypoint
├── internal/
│   ├── cli/                 # CLI orchestration and output writing
│   ├── config/              # flags/env config loading
│   ├── parser/              # GitHub URL parsing
│   ├── github/              # GitHub API fetching
│   ├── converter/           # Markdown rendering and optional AI summary
│   └── webapp/              # HTTP handlers and template wiring
├── tests/
│   ├── integration/http/    # API integration tests
│   └── e2e/web/             # Web E2E tests
├── web/                     # Embedded templates and static assets
├── docs/                    # OpenAPI artifacts
├── Makefile
└── Dockerfile
```

<a id="architecture-and-data-flow"></a>
## Architecture and Data Flow

CLI path:

```text
cmd/issue2md -> internal/cli -> internal/parser -> internal/github -> internal/converter -> file/stdout
```

Web path:

```text
cmd/issue2mdweb -> internal/webapp -> internal/parser -> internal/github -> internal/converter -> HTTP response
```

<a id="common-commands"></a>
## Common Commands

Command source of truth: root `Makefile`.

| Command | Purpose | Status |
|---|---|---|
| `make help` | List make targets | Makefile |
| `make fmt` | Format Go code with `gofmt` + `goimports-reviser` | Makefile + CI gate |
| `make ci COVER_MIN=80` | Required CI-equivalent local gate (`fmt-check` + coverage + lint + build) | Makefile + CI |
| `make test` | Run all tests | Makefile |
| `make build-cli` | Build CLI binary | Makefile |
| `make web` | Build Web binary | Makefile |
| `make install-cli` | Install CLI into `GOBIN` / `GOPATH/bin` | Makefile |
| `make install-web` | Install Web binary into `GOBIN` / `GOPATH/bin` | Makefile |
| `make swagger-check` | Regenerate and verify OpenAPI artifacts | Makefile |
| `./bin/issue2md --stdout <github-url>` | Real GitHub conversion | Requires `GITHUB_TOKEN` |
| `make ci-api-integration` | API integration gate equivalent | Makefile + CI |
| `make ci-e2e-web` | Web E2E gate equivalent | Makefile + CI (`push`/`schedule`) |
| `make docker-build` | Build Docker image | Makefile + CI (Linux runner) |

<a id="configuration-and-environment"></a>
## Configuration and Environment

### Runtime Environment Variables

| Variable | Purpose | Required |
|---|---|---|
| `GITHUB_TOKEN` | GitHub token (used when `--token` is not passed) | Recommended |
| `OPENAI_API_KEY` | Enable the `## AI Summary` section | Optional |
| `ISSUE2MD_AI_BASE_URL` | Override AI base URL | Optional |
| `ISSUE2MD_AI_MODEL` | Override AI model | Optional |
| `ISSUE2MD_WEB_ADDR` | Web listen address (default `:8080`) | Optional |
| `ISSUE2MD_WEB_WRITE_TIMEOUT` | Web response write timeout for request handling (Go duration, default `120s`) | Optional |

If `OPENAI_API_KEY` is unset, conversion still succeeds and markdown is rendered without AI summary content.

### CLI Flags (`internal/config/loader.go`)

| Flag | Description | Constraints |
|---|---|---|
| `--output` | Output file/directory | Required in batch mode |
| `--format` | Output format | Only `markdown` is supported |
| `--include-comments` | Include comments (`true` by default) | - |
| `--input-file` | Batch input file | Conflicts with `--stdout` |
| `--stdout` | Write markdown to stdout | Conflicts with `--input-file` |
| `--force` | Overwrite existing output files | - |
| `--token` | GitHub token (higher priority than `GITHUB_TOKEN`) | - |
| `--lang` | Summary language override | Only used when AI summary is enabled via `OPENAI_API_KEY` |

Default output filename pattern (`internal/cli/output.go`):

```text
<owner>-<repo>-<issue|pr|discussion>-<number>.md
```

<a id="web-api-example"></a>
## Web API Example

Routes:
- `GET /`
- `POST /convert` (form field: `url`)
- `GET /openapi.json`
- `GET /swagger` (redirects to `/swagger/index.html`)
- `GET /swagger/index.html`
- `GET /swagger/assets/*`

Example request:

```bash
curl -sS -X POST http://127.0.0.1:8080/convert \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'url=https://github.com/<owner>/<repo>/issues/1'
```

<a id="testing-and-quality"></a>
## Testing and Quality

Local required gate:
- `make ci COVER_MIN=80`

Additional local checks:
- `make ci-api-integration`
- `make ci-e2e-web` (opt-in, same as CI e2e job gate behavior)

CI workflow (`.github/workflows/ci.yml`):
- `ci`: `make ci COVER_MIN=80` (`fmt-check` + `cover-check` + `lint` + `build-all`)
- `docker-build`: Docker build validation (web default + CLI variant)
- `api-integration`: `make ci-api-integration`
- `e2e-web`: `make ci-e2e-web` (push/schedule)
- `govulncheck`: dependency vulnerability scan
- `fieldalignment`: struct field alignment check

<a id="troubleshooting"></a>
## Troubleshooting

### GitHub token and permissions

- Symptom: conversion fails with GitHub auth/permission errors.
- Checks:
  - Confirm `GITHUB_TOKEN` is set (or pass `--token`).
  - Confirm the token can read the target repo/discussion.
  - For fine-grained tokens, ensure repository access is explicitly granted.

### Docker daemon unavailable

- Symptom: `make docker-build` fails with daemon connection errors.
- Checks:
  - Start Docker Desktop or Docker daemon.
  - Run `docker info` and confirm it returns successfully.
  - If local Docker is unavailable, rely on CI `docker-build` job for validation.

### Slow upstream or timeout issues on `/convert`

- Symptom: web conversion fails on slow upstream fetch/summarization paths.
- Checks:
  - Increase `ISSUE2MD_WEB_WRITE_TIMEOUT` (for example `120s`, `180s`) based on your SLA.
  - Verify upstream network reachability and API latency.
  - Re-run with a known small public issue URL to isolate environment latency.

### CI formatting gate fails (`fmt-check`)

- Symptom: CI reports formatting diff and blocks merge.
- Fix:
  - Run `make fmt`.
  - Commit formatting changes.
  - Re-run `make ci COVER_MIN=80` locally before pushing.

<a id="docker"></a>
## Docker

These commands are defined by `Dockerfile` and `Makefile`:

```bash
make docker-build
docker run --rm -p 8080:8080 \
  -e ISSUE2MD_WEB_ADDR=:8080 \
  -e GITHUB_TOKEN=<your_github_pat> \
  issue2md:latest
```

`Dockerfile` defaults to `APP=issue2mdweb`. Use `--build-arg APP=issue2md` for CLI image.

<a id="project-docs"></a>
## Project Docs

Core project docs:
- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [CHANGELOG.md](CHANGELOG.md)
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- [LICENSE](LICENSE)

Chinese translations:
- [README.zh-CN.md](README.zh-CN.md)
- [CONTRIBUTING.zh-CN.md](CONTRIBUTING.zh-CN.md)
- [CODE_OF_CONDUCT.zh-CN.md](CODE_OF_CONDUCT.zh-CN.md)
- [SECURITY.zh-CN.md](SECURITY.zh-CN.md)

Update this README when entrypoints, environment variables, Make targets, API routes, or Go version change.
