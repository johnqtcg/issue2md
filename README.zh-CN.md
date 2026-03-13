# issue2md

[![CI](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml/badge.svg)](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml)
![Go Version](https://img.shields.io/badge/go-1.25.8-00ADD8)
![License](https://img.shields.io/badge/license-MIT-blue)

本项目是规范驱动开发（SDD）和测试驱动开发（TDD）的直接产物。SDD 工作流参考了开源项目 [github/spec-kit](https://github.com/github/spec-kit)，项目规范沉淀在 [`specs/`](specs) 目录下，整体实现由 Codex 5.3 完成。

把 GitHub `Issue`、`Pull Request` 和 `Discussion` URL 转成干净、可归档、可分享、可继续处理的 Markdown。

## 目录

- [项目概览](#cn-overview)
- [功能亮点](#cn-highlights)
- [安装](#cn-install)
- [快速开始](#cn-quick-start)
- [端到端示例](#cn-end-to-end-example)
- [配置与环境变量](#cn-configuration-and-env)
- [常用命令](#cn-common-commands)
- [项目文档](#cn-project-docs)

<a id="cn-overview"></a>
## 项目概览

- 双入口形态：`CLI tool + backend web service`。
- Go 版本：`go 1.25.8`（见 `go.mod`）。
- 模块路径：`github.com/johnqtcg/issue2md`。

<a id="cn-highlights"></a>
## 功能亮点

- 一个工具，两种入口：本地导出可用 CLI，浏览器或接口工作流可用 Web 服务。
- 可选 AI 摘要：配置 `OPENAI_API_KEY` 后，输出可包含结构化的 `## AI Summary` 区块，带 summary、decisions 和 action items。
- 默认结构化输出：生成的 Markdown 会保留 metadata、original description、discussion thread 和原始 URL 引用。

<a id="cn-install"></a>
## 安装

直接通过 Go 安装：

```bash
go install github.com/johnqtcg/issue2md/cmd/issue2md@latest
go install github.com/johnqtcg/issue2md/cmd/issue2mdweb@latest
```

如果你是从本地仓库工作：
- `make install-cli`
- `make install-web`

<a id="cn-quick-start"></a>
## 快速开始

### 30 秒 CLI

```bash
export GITHUB_TOKEN=<your_github_pat>
issue2md --stdout https://github.com/github/spec-kit/issues/75
```

### 30 秒 Web

```bash
export GITHUB_TOKEN=<your_github_pat>
ISSUE2MD_WEB_ADDR=127.0.0.1:18080 issue2mdweb
```

然后调用 HTTP 接口：

```bash
curl -sS -X POST http://127.0.0.1:18080/convert \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'url=https://github.com/github/spec-kit/issues/75'
```

说明：
- 需要 Go `>= 1.25` 和已安装的可执行文件。
- 如果你更偏向本地二进制，可使用 `make build-cli` 或 `make web`。
- 贡献者门禁命令保留在后面的 `常用命令` 和 `测试与质量检查` 章节。

<a id="cn-end-to-end-example"></a>
## 端到端示例

把一个 issue URL 转成 Markdown，并让 CLI 自动生成默认文件名：

```bash
issue2md https://github.com/github/spec-kit/issues/75
# OK url=https://github.com/github/spec-kit/issues/75 type=issue output=github-spec-kit-issue-75.md
```

默认文件名规则：

```text
<owner>-<repo>-<issue|pr|discussion>-<number>.md
```

生成文件的结构示例摘自 [`internal/converter/testdata/issue.golden.md`](internal/converter/testdata/issue.golden.md)：

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

最终生成的文件固定会包含 metadata、original description、thread content 和 references。只有在配置了 `OPENAI_API_KEY` 时，才会出现 `## AI Summary` 区块。

<a id="cn-project-structure"></a>
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
├── specs/                   # SDD 规范源：规格、计划与任务拆解
├── tests/
│   ├── integration/http/    # API 集成测试
│   └── e2e/web/             # Web E2E 测试
├── web/                     # 嵌入式模板与静态资源（含 Swagger UI）
├── docs/                    # OpenAPI 产物（swagger.json / swagger.yaml）
├── Makefile
└── Dockerfile
```

<a id="cn-architecture-and-flow"></a>
## 架构与数据流

CLI 路径：

```text
cmd/issue2md -> internal/cli -> internal/parser -> internal/github -> internal/converter -> 文件或stdout
```

Web 路径：

```text
cmd/issue2mdweb -> internal/webapp -> internal/parser -> internal/github -> internal/converter -> HTTP响应
```

<a id="cn-common-commands"></a>
## 常用命令

命令来源优先级：`Makefile`（本仓库主命令入口）。

| 命令 | 用途 | 状态 |
|---|---|---|
| `make help` | 查看所有目标 | Makefile |
| `make fmt` | 使用 `gofmt` + `goimports-reviser` 格式化 Go 代码 | Makefile + CI 门禁 |
| `make ci COVER_MIN=80` | 必需本地门禁（等价 CI：`fmt-check` + 覆盖率 + lint + build） | Makefile + CI |
| `make test` | 运行全部测试 | Makefile |
| `make build-cli` | 构建 CLI | Makefile |
| `make web` | 构建 Web 二进制 | Makefile |
| `make install-cli` | 将 CLI 安装到 `GOBIN` / `GOPATH/bin` | Makefile |
| `make install-web` | 将 Web 二进制安装到 `GOBIN` / `GOPATH/bin` | Makefile |
| `make swagger-check` | 重新生成并校验 OpenAPI 文档 | Makefile |
| `./bin/issue2md --stdout <github-url>` | 真实 GitHub URL 转换 | 需要 `GITHUB_TOKEN` |
| `make ci-api-integration` | API 集成门禁等价命令 | Makefile + CI |
| `make ci-e2e-web` | Web E2E 门禁等价命令 | Makefile + CI（`push`/`schedule`） |
| `make docker-build` | 构建容器镜像 | Makefile + CI（Linux runner） |

<a id="cn-configuration-and-env"></a>
## 配置与环境变量

### 运行时环境变量

| 变量名 | 用途 | 是否必需 |
|---|---|---|
| `GITHUB_TOKEN` | GitHub API token（`--token` 未传时读取） | 推荐 |
| `OPENAI_API_KEY` | 启用 `## AI Summary` 区块 | 可选 |
| `ISSUE2MD_AI_BASE_URL` | AI 接口 base URL 覆盖 | 可选 |
| `ISSUE2MD_AI_MODEL` | AI 模型名覆盖 | 可选 |
| `ISSUE2MD_WEB_ADDR` | Web 服务监听地址（默认 `:8080`） | 可选 |
| `ISSUE2MD_WEB_WRITE_TIMEOUT` | Web 请求处理阶段的响应写超时（Go duration，默认 `120s`） | 可选 |

如果未设置 `OPENAI_API_KEY`，转换仍会成功，只是输出中不会包含 AI 摘要内容。

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
| `--lang` | AI 摘要语言 | 仅在通过 `OPENAI_API_KEY` 启用 AI 摘要时生效 |

默认文件名规则（`internal/cli/output.go`）：

```text
<owner>-<repo>-<issue|pr|discussion>-<number>.md
```

<a id="cn-web-api-example"></a>
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

<a id="cn-testing-and-quality"></a>
## 测试与质量检查

### 本地质量门禁

- 必需门禁：`make ci COVER_MIN=80`
- 附加检查：`make ci-api-integration`
- 附加检查：`make ci-e2e-web`（可选，对齐 CI e2e job 触发语义）

### CI 工作流（`.github/workflows/ci.yml`）

- `ci`: `make ci COVER_MIN=80`（`fmt-check` + `cover-check` + `lint` + `build-all`）
- `docker-build`: Docker 镜像构建校验（默认 web 入口 + CLI 入口）
- `api-integration`: `make ci-api-integration`
- `e2e-web`: `make ci-e2e-web`（仅 `push` 或定时）
- `govulncheck`: 依赖漏洞检查
- `fieldalignment`: 结构体字段对齐检查

<a id="cn-troubleshooting"></a>
## 故障排查（Troubleshooting）

### GitHub Token 与权限问题

- 现象：转换请求返回 GitHub 鉴权/权限错误。
- 排查：
  - 确认已设置 `GITHUB_TOKEN`（或显式传入 `--token`）。
  - 确认 token 对目标仓库/讨论有读取权限。
  - 若是 fine-grained token，确认目标仓库已被明确授权。

### Docker daemon 不可用

- 现象：`make docker-build` 报 daemon 连接失败。
- 排查：
  - 确认 Docker Desktop 或 Docker daemon 已启动。
  - 执行 `docker info`，确认命令可成功返回。
  - 若本地无 Docker，可依赖 CI 的 `docker-build` job 做构建校验。

### `/convert` 慢请求或超时问题

- 现象：上游抓取/摘要较慢时，Web 转换失败。
- 排查：
  - 按 SLA 调整 `ISSUE2MD_WEB_WRITE_TIMEOUT`（例如 `120s`、`180s`）。
  - 检查上游网络连通性与 API 延迟。
  - 用已知的小型公开 issue URL 复测，以隔离环境延迟因素。

### CI 格式化门禁失败（`fmt-check`）

- 现象：CI 报格式化 diff，阻止合并。
- 处理：
  - 先执行 `make fmt`。
  - 提交格式化改动。
  - 推送前本地执行 `make ci COVER_MIN=80` 复核。

<a id="cn-docker"></a>
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

<a id="cn-project-docs"></a>
## 项目文档

核心文档：
- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [CHANGELOG.md](CHANGELOG.md)
- [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- [LICENSE](LICENSE)

中文文档：
- [README.zh-CN.md](README.zh-CN.md)
- [CONTRIBUTING.zh-CN.md](CONTRIBUTING.zh-CN.md)
- [CODE_OF_CONDUCT.zh-CN.md](CODE_OF_CONDUCT.zh-CN.md)
- [SECURITY.zh-CN.md](SECURITY.zh-CN.md)

当入口、环境变量、Make target、API 路由或 Go 版本变化时，应同步更新本 README。
