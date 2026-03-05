# issue2md

[![CI](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml/badge.svg)](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml)
![Go Version](https://img.shields.io/badge/go-1.25.7-00ADD8)
![License](https://img.shields.io/badge/license-MIT-blue)

Convert GitHub `Issue` / `Pull Request` / `Discussion` URLs into Markdown with both CLI and Web entrypoints.

Language:
- English (primary): `README.md`
- Chinese: [README.zh-CN.md](README.zh-CN.md)

## Overview

- Project type: `CLI tool + backend web service` (dual entrypoints).
- Go version: `go 1.25.7` (from `go.mod`).
- Module path: `github.com/johnqtcg/issue2md`.

## Quick Start

### Prerequisites

- Go `>= 1.25` (current: `1.25.7`)
- `goimports-reviser` for `make fmt`
- `golangci-lint` for `make lint`
- `swag` for `make swagger` / `make swagger-check`
- Docker for `make docker-build`

### Command Verifiability

`Verified` below means executed in this agent session on `2026-03-04`.

Verified in this session:

```bash
make help
make fmt
make test
make lint
make cover-check COVER_MIN=80
make build-cli
make web
make swagger-check
./bin/issue2md --stdout https://github.com/github/spec-kit/issues/75
```

Local note:
- `make docker-build` failed locally in this session because Docker daemon was unreachable (`Cannot connect to the Docker daemon`).
- Docker build validation is enforced in GitHub Actions on Linux runner (`docker-build` job).

### Build Locally

```bash
make build-cli
make web
```

Generated binaries:
- `bin/issue2md`
- `bin/issue2mdweb`

### Run CLI

Single URL:

```bash
./bin/issue2md https://github.com/<owner>/<repo>/issues/<number>
```

Batch mode (one URL per line):

```bash
./bin/issue2md --input-file urls.txt --output out
```

### Run Web Service

```bash
./bin/issue2mdweb
```

Default listen address is `:8080`. Override with `ISSUE2MD_WEB_ADDR`:

```bash
ISSUE2MD_WEB_ADDR=127.0.0.1:18080 ./bin/issue2mdweb
```

`/convert` write deadline can be tuned via `ISSUE2MD_WEB_WRITE_TIMEOUT` (Go duration format, default `120s`):

```bash
ISSUE2MD_WEB_WRITE_TIMEOUT=90s ./bin/issue2mdweb
```

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

## Architecture and Data Flow

CLI path:

```text
cmd/issue2md -> internal/cli -> internal/parser -> internal/github -> internal/converter -> file/stdout
```

Web path:

```text
cmd/issue2mdweb -> internal/webapp -> internal/parser -> internal/github -> internal/converter -> HTTP response
```

## Common Commands

Command source of truth: root `Makefile`.

| Command | Purpose | Status |
|---|---|---|
| `make help` | List make targets | Verified (local session) |
| `make fmt` | Format Go code with `gofmt` + `goimports-reviser` | Defined in Makefile + CI |
| `make test` | Run all tests | Verified (local session) |
| `make lint` | Run `golangci-lint` | Verified (local session) |
| `make cover-check COVER_MIN=80` | Coverage gate | Verified (local session) |
| `make build-cli` | Build CLI binary | Verified (local session) |
| `make web` | Build Web binary | Verified (local session) |
| `make swagger-check` | Regenerate and verify OpenAPI artifacts | Verified (local session) |
| `./bin/issue2md --stdout <github-url>` | Real GitHub conversion | Verified (local session; requires `GITHUB_TOKEN`) |
| `make test-api-integration` | API integration tests | Defined in Makefile + CI |
| `make test-e2e-web` | Web E2E tests | Defined in Makefile + CI |
| `make docker-build` | Build Docker image | Defined in Makefile + CI (Linux runner) |

## Configuration and Environment

### Runtime Environment Variables

| Variable | Purpose | Required |
|---|---|---|
| `GITHUB_TOKEN` | GitHub token (used when `--token` is not passed) | Recommended |
| `OPENAI_API_KEY` | Enable AI summary | Optional |
| `ISSUE2MD_AI_BASE_URL` | Override AI base URL | Optional |
| `ISSUE2MD_AI_MODEL` | Override AI model | Optional |
| `ISSUE2MD_WEB_ADDR` | Web listen address (default `:8080`) | Optional |
| `ISSUE2MD_WEB_WRITE_TIMEOUT` | Web response write timeout for request handling (Go duration, default `120s`) | Optional |

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
| `--lang` | Summary language override | - |

Default output filename pattern (`internal/cli/output.go`):

```text
<owner>-<repo>-<issue|pr|discussion>-<number>.md
```

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

## Testing and Quality

Local quality gates:
- `make fmt`
- `make test`
- `make lint`
- `make cover-check COVER_MIN=80` (session result: `81.5%`)

CI workflow (`.github/workflows/ci.yml`):
- `ci`: `fmt` (gofmt + goimports-reviser, diff check) + `cover-check` + `lint` + `build-all`
- `docker-build`: Docker build validation (web default + CLI variant)
- `api-integration`: `make test-api-integration`
- `e2e-web`: `make test-e2e-web` (push/schedule)
- `govulncheck`: dependency vulnerability scan
- `fieldalignment`: struct field alignment check

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

## Documentation Maintenance

Update this README when these repository changes happen:

| Repository Change | README Sections to Update |
|---|---|
| New `cmd/*/main.go` entrypoint | Overview, Quick Start, Structure, Commands |
| Added/changed environment variables | Configuration and Environment |
| Makefile target added/renamed | Common Commands |
| CI workflow changed | Badges, Testing and Quality |
| New module directory (`internal/*`/`tests/*`) | Structure, Architecture |
| API route changed | Web API Example |
| Go version changed | Badges, Quick Start prerequisites |

## Governance Files

- `LICENSE`: present (`MIT`) -> [LICENSE](LICENSE)
- `CONTRIBUTING.md`: present -> [CONTRIBUTING.md](CONTRIBUTING.md)
- `CODE_OF_CONDUCT.md`: present -> [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- `SECURITY.md`: present -> [SECURITY.md](SECURITY.md)
- `CHANGELOG.md`: present -> [CHANGELOG.md](CHANGELOG.md)

Chinese translations:
- [README.zh-CN.md](README.zh-CN.md)
- [CONTRIBUTING.zh-CN.md](CONTRIBUTING.zh-CN.md)
- [CODE_OF_CONDUCT.zh-CN.md](CODE_OF_CONDUCT.zh-CN.md)
- [SECURITY.zh-CN.md](SECURITY.zh-CN.md)
