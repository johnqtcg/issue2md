# issue2md

[![CI](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml/badge.svg)](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml)
![Go Version](https://img.shields.io/badge/go-1.25.7-00ADD8)
![License](https://img.shields.io/badge/license-MIT-blue)

将 GitHub `Issue` / `Pull Request` / `Discussion` URL 转换为 Markdown 的 Go 工具，提供 CLI 与 Web 两种入口。

## 项目概览

- 语言策略：中文说明 + 英文技术术语（命令、路径、环境变量名保持英文）。
- 项目类型（routing）：`CLI tool + backend web service`（双入口仓库）。
- Go 版本：`go 1.25.7`（见 `go.mod`）。
- 模块路径：`github.com/johnqtcg/issue2md`。

## 快速开始

### 前置条件

- Go `>= 1.25`（`go.mod` 当前为 `1.25.7`）
- `golangci-lint`（`make lint` 需要）
- `swag`（`make swagger` / `make swagger-check` 需要）
- Docker（`make docker-build` 需要）

### 命令可验证性（Command Verifiability Gate）

说明：这里的 “Verified” 指本次 agent 会话在 `2026-03-04` 的实际执行结果，不等同于你本机历史执行状态。

本次会话已执行并通过：

```bash
make help
make test
make lint
make cover-check COVER_MIN=80
make build-cli
make web
make swagger-check
./bin/issue2md --stdout https://github.com/github/spec-kit/issues/75
```

### 本地构建

```bash
make build-cli
make web
```

生成二进制：

- `bin/issue2md`
- `bin/issue2mdweb`

### 运行 CLI

单条 URL：

```bash
./bin/issue2md https://github.com/<owner>/<repo>/issues/<number>
```

批量 URL（每行一个 URL）：

```bash
./bin/issue2md --input-file urls.txt --output out
```

### 运行 Web 服务

```bash
./bin/issue2mdweb
```

默认监听地址为 `:8080`，可用 `ISSUE2MD_WEB_ADDR` 覆盖：

```bash
ISSUE2MD_WEB_ADDR=127.0.0.1:18080 ./bin/issue2mdweb
```

`/convert` 的写超时可通过 `ISSUE2MD_WEB_WRITE_TIMEOUT` 调整（Go duration 格式，默认 `120s`）：

```bash
ISSUE2MD_WEB_WRITE_TIMEOUT=90s ./bin/issue2mdweb
```

## 项目结构

```text
.
├── cmd/
│   ├── issue2md/            # CLI 入口
│   └── issue2mdweb/         # Web 入口
├── internal/
│   ├── cli/                 # CLI 编排与输出落盘
│   ├── config/              # flags/env 配置解析
│   ├── parser/              # GitHub URL 解析
│   ├── github/              # GitHub API 抓取
│   ├── converter/           # Markdown 渲染与可选 AI 摘要
│   └── webapp/              # HTTP handler 与页面模板装配
├── tests/
│   ├── integration/http/    # API 集成测试
│   └── e2e/web/             # Web E2E 测试
├── web/                     # 嵌入式模板与静态资源（含 Swagger UI）
├── docs/                    # OpenAPI 产物（swagger.json / swagger.yaml）
├── Makefile
└── Dockerfile
```

## 架构与数据流

CLI 路径：

```text
cmd/issue2md -> internal/cli -> internal/parser -> internal/github -> internal/converter -> 文件或stdout
```

Web 路径：

```text
cmd/issue2mdweb -> internal/webapp -> internal/parser -> internal/github -> internal/converter -> HTTP响应
```

## 常用命令

命令来源优先级：`Makefile`（本仓库主命令入口）。

| 命令 | 用途 | 状态 |
|---|---|---|
| `make help` | 查看所有目标 | Verified（local session） |
| `make test` | 运行全部测试 | Verified（local session） |
| `make lint` | 执行 `golangci-lint` | Verified（local session） |
| `make cover-check COVER_MIN=80` | 覆盖率门禁 | Verified（local session） |
| `make build-cli` | 构建 CLI | Verified（local session） |
| `make web` | 构建 Web 二进制 | Verified（local session） |
| `make swagger-check` | 重新生成并校验 OpenAPI 文档 | Verified（local session） |
| `./bin/issue2md --stdout <github-url>` | 真实 GitHub URL 转换 | Verified（local session, requires `GITHUB_TOKEN`） |
| `make test-api-integration` | 运行 API 集成测试（target 内设置 `ISSUE2MD_API_INTEGRATION=1`） | Defined in Makefile + CI |
| `make test-e2e-web` | 运行 Web E2E（target 内设置 `ISSUE2MD_E2E=1`） | Defined in Makefile + CI |
| `make docker-build` | 构建容器镜像 | Defined in Makefile + CI（Linux runner） |

## 配置与环境变量

### 运行时环境变量

| 变量名 | 用途 | 是否必需 |
|---|---|---|
| `GITHUB_TOKEN` | GitHub API token（`--token` 未传时读取） | 推荐 |
| `OPENAI_API_KEY` | 启用 AI 摘要 | 可选 |
| `ISSUE2MD_AI_BASE_URL` | AI 接口 base URL 覆盖 | 可选 |
| `ISSUE2MD_AI_MODEL` | AI 模型名覆盖 | 可选 |
| `ISSUE2MD_WEB_ADDR` | Web 服务监听地址（默认 `:8080`） | 可选 |
| `ISSUE2MD_WEB_WRITE_TIMEOUT` | Web 请求处理阶段的响应写超时（Go duration，默认 `120s`） | 可选 |

### CLI flags（`internal/config/loader.go`）

| 参数 | 说明 | 约束 |
|---|---|---|
| `--output` | 输出文件或目录 | 批处理模式必填 |
| `--format` | 输出格式 | 仅支持 `markdown` |
| `--include-comments` | 是否包含评论（默认 `true`） | - |
| `--input-file` | 批量输入文件（每行一个 URL） | 与 `--stdout` 冲突 |
| `--stdout` | 将 markdown 打印到 stdout | 与 `--input-file` 冲突 |
| `--force` | 覆盖已存在输出文件 | - |
| `--token` | GitHub token（优先级高于 `GITHUB_TOKEN`） | - |
| `--lang` | AI 摘要语言 | - |

默认文件名规则（`internal/cli/output.go`）：

```text
<owner>-<repo>-<issue|pr|discussion>-<number>.md
```

## API 示例（Web）

### 路由

- `GET /`
- `POST /convert`（form 字段 `url`）
- `GET /openapi.json`
- `GET /swagger`（重定向到 `/swagger/index.html`）
- `GET /swagger/index.html`
- `GET /swagger/assets/*`

### 示例请求

```bash
curl -sS -X POST http://127.0.0.1:8080/convert \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'url=https://github.com/<owner>/<repo>/issues/1'
```

## 测试与质量检查

### 本地质量门禁

- 单元/集成测试：`make test`
- 静态检查：`make lint`
- 覆盖率门禁：`make cover-check COVER_MIN=80`（当前环境结果：`81.5%`）

### CI 工作流（`.github/workflows/ci.yml`）

- `ci`: `cover-check` + `lint` + `build-all`
- `docker-build`: Docker 镜像构建校验（默认 web 入口 + CLI 入口）
- `api-integration`: `make test-api-integration`
- `e2e-web`: `make test-e2e-web`（仅 `push` 或定时）
- `govulncheck`: 依赖漏洞检查
- `fieldalignment`: 结构体字段对齐检查

## 部署与运行（Docker）

以下命令来自 `Dockerfile` 与 `Makefile`。本地执行依赖 Docker daemon；CI 中已由 `docker-build` job 在 Linux runner 验证构建。

```bash
make docker-build
docker run --rm -p 8080:8080 \
  -e ISSUE2MD_WEB_ADDR=:8080 \
  -e GITHUB_TOKEN=<your_github_pat> \
  issue2md:latest
```

`Dockerfile` 默认 `APP=issue2mdweb`，可通过 `--build-arg APP=issue2md` 构建 CLI 入口。

## 文档维护说明

以下变更发生时，应同步更新本 README：

| 仓库变更 | 需要更新的 README 部分 |
|---|---|
| 新增 `cmd/*/main.go` 入口 | 项目概览、快速开始、项目结构、常用命令 |
| 新增或修改环境变量 | 配置与环境变量 |
| Makefile target 新增/重命名 | 常用命令 |
| CI workflow 变化 | 徽章、测试与质量检查 |
| 新增模块目录（`internal/*`/`tests/*`） | 项目结构、架构与数据流 |
| API 路由变更 | API 示例（Web） |
| Go 版本变更 | 徽章、快速开始 |

维护建议：在提交前至少执行 `make test` 与 `make lint`，涉及覆盖率门禁时执行 `make cover-check COVER_MIN=80`。

## 治理文件状态

- `LICENSE`: present (`MIT`)
- `CONTRIBUTING.md`: present (`./CONTRIBUTING.md`)
- `CODE_OF_CONDUCT.md`: present (`./CODE_OF_CONDUCT.md`)
- `SECURITY.md`: present (`./SECURITY.md`)
- `CHANGELOG.md`: present -> (`./CHANGELOG.md`)
