# issue2md

[![CI](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml/badge.svg)](https://github.com/johnqtcg/issue2md/actions/workflows/ci.yml)
![Go Version](https://img.shields.io/badge/go-1.25.7-00ADD8)
![License](https://img.shields.io/badge/license-MIT-blue)

将 GitHub `Issue` / `Pull Request` / `Discussion` URL 转换为 Markdown 的 Go 工具，提供 CLI 与 Web 两种入口。

## 目录

- [项目概览](#cn-overview)
- [快速开始](#cn-quick-start)
- [项目结构](#cn-project-structure)
- [架构与数据流](#cn-architecture-and-flow)
- [常用命令](#cn-common-commands)
- [配置与环境变量](#cn-configuration-and-env)
- [API 示例（Web）](#cn-web-api-example)
- [测试与质量检查](#cn-testing-and-quality)
- [故障排查（Troubleshooting）](#cn-troubleshooting)
- [部署与运行（Docker）](#cn-docker)
- [文档维护说明](#cn-documentation-maintenance)
- [治理文件状态](#cn-governance-files)

<a id="cn-overview"></a>
## 项目概览

- 语言策略：中文说明 + 英文技术术语（命令、路径、环境变量名保持英文）。
- 项目类型（routing）：`CLI tool + backend web service`（双入口仓库）。
- Go 版本：`go 1.25.7`（见 `go.mod`）。
- 模块路径：`github.com/johnqtcg/issue2md`。

<a id="cn-quick-start"></a>
## 快速开始

### 前置条件

- Go `>= 1.25`（`go.mod` 当前为 `1.25.7`）
- `goimports-reviser`（`make fmt` 需要）
- `golangci-lint`（`make lint` 需要）
- `swag`（`make swagger` / `make swagger-check` 需要）
- Docker（`make docker-build` 需要）

### 命令可验证性（Command Verifiability Gate）

在本地执行质量/构建命令前，请先安装依赖工具并运行：

```bash
make help
make ci COVER_MIN=80
make build-cli
make web
make swagger-check
./bin/issue2md --stdout https://github.com/github/spec-kit/issues/75
```

验证策略：
- 以你本地实际执行结果为准。
- CI（`.github/workflows/ci.yml`）会在 Linux runner 上执行并校验必需门禁。

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

<a id="cn-documentation-maintenance"></a>
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

维护建议：在提交前至少执行 `make ci COVER_MIN=80`，必要时补充执行 `make ci-api-integration` 与 `make ci-e2e-web`。

<a id="cn-governance-files"></a>
## 治理文件状态

- `LICENSE`: present (`MIT`)
- `CONTRIBUTING.md`: present (`./CONTRIBUTING.md`)
- `CODE_OF_CONDUCT.md`: present (`./CODE_OF_CONDUCT.md`)
- `SECURITY.md`: present (`./SECURITY.md`)
- `CHANGELOG.md`: present -> (`./CHANGELOG.md`)
