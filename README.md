# issue2md

将 GitHub Issue / Pull Request / Discussion URL 转换为结构化 Markdown 归档文档的 Go 工具，支持 CLI 批量处理与 Web 转换接口。

- 语言与版本: Go `1.25.4`（见 `go.mod`）
- 核心依赖: `google/go-github/v72`、`oauth2`（见 `go.mod`）
- 入口程序: `cmd/issue2md`（CLI）、`cmd/issue2mdweb`（Web）

## 目录

- [项目代码结构](#项目代码结构)
- [快速开始](#快速开始)
- [架构与数据流](#架构与数据流)
- [运行模式](#运行模式)
- [常用命令（Makefile）](#常用命令makefile)
- [文档索引](#文档索引)
- [文档维护规则](#文档维护规则)
- [联系信息](#联系信息)

## 项目代码结构

```text
.
├── AGENTS.md
├── constitution.md
├── Makefile
├── go.mod
├── go.sum
├── cmd
│   ├── issue2md
│   │   ├── main.go              # CLI 入口
│   │   └── main_test.go
│   └── issue2mdweb
│       ├── main.go              # Web 服务入口（默认 :8080）
│       ├── handler.go           # HTTP 路由与处理器
│       ├── templates.go         # HTML 模板与 Swagger 页面
│       └── main_test.go
├── internal
│   ├── cli                      # CLI 编排、输出、批处理、退出码
│   ├── config                   # flags + env 配置加载
│   ├── parser                   # GitHub URL 解析与规范化
│   ├── github                   # REST/GraphQL 抓取、重试、错误分类
│   └── converter                # Markdown 渲染与 AI Summary
├── web
│   ├── templates
│   │   └── index.html
│   └── static
│       └── style.css
├── docs
│   ├── swagger.json             # 生成产物（OpenAPI）
│   └── swagger.yaml             # 生成产物（OpenAPI）
└── specs
    └── 001-core-functionality
        ├── spec.md
        ├── plan.md
        ├── tasks.md
        └── api-sketch.md
```

## 快速开始

### 1) 环境准备

- Go: `>= 1.25`（当前仓库为 `1.25.4`）
- 可选工具:
  - `golangci-lint`（用于 `make lint`）
  - `swag`（用于 `make swagger`，也可先执行 `make install-swag`）

### 2) 安装 CLI 到 GOBIN（或 GOPATH/bin）

远程安装（基于当前模块路径）:

```bash
go install github.com/johnqtcg/issue2md/cmd/issue2md@latest
```

本地安装（当前工作副本）:

```bash
make install-cli
# 或
go install ./cmd/issue2md
```

提示:
- 若仓库是私有仓库，请配置 `GOPRIVATE` 与 GitHub 访问凭据。
- 安装后请确认 `$GOBIN`（或 `$GOPATH/bin`）已在 `PATH` 中。

### 3) 拉起 CLI（单条 URL）

```bash
make build-cli
./bin/issue2md https://github.com/<owner>/<repo>/issues/<number>
```

### 4) 批量导出

```bash
cat > urls.txt <<'EOF'
https://github.com/<owner>/<repo>/issues/1
https://github.com/<owner>/<repo>/pull/2
https://github.com/<owner>/<repo>/discussions/3
EOF

./bin/issue2md --input-file urls.txt --output out
```

### 5) 拉起 Web 服务并验证

```bash
make web
./bin/issue2mdweb
```

服务默认监听 `:8080`（见 `cmd/issue2mdweb/main.go`）。

```bash
curl -i http://127.0.0.1:8080/
curl -i http://127.0.0.1:8080/swagger
```

### 6) 认证与 AI 总结（可选）

```bash
export GITHUB_TOKEN=<your_github_pat>
export OPENAI_API_KEY=<your_openai_key>
export ISSUE2MD_AI_BASE_URL=<optional_base_url>
export ISSUE2MD_AI_MODEL=<optional_model>
```

说明:
- `GITHUB_TOKEN` 可提升 GitHub API 限额（见 `internal/config/loader.go`）。
- 配置 `OPENAI_API_KEY` 后会启用 AI Summary（见 `internal/converter/summary_openai.go`）。

## 架构与数据流

### 组件职责

- `internal/parser`: 将 GitHub URL 规范化为 `ResourceRef`。
- `internal/github`: 按资源类型抓取 Issue/PR/Discussion，处理分页、重试与错误分类。
- `internal/converter`: 渲染 front matter、metadata、timeline/thread/reviews、references；可选 AI Summary。
- `internal/cli`: 负责单条/批量流程、输出落盘、状态行与退出码。
- `cmd/issue2mdweb`: 复用 parser/fetcher/renderer，通过 HTTP 暴露转换能力。

### 调用路径

```text
CLI:
cmd/issue2md -> internal/cli -> parser -> github(fetch) -> converter(render) -> output(file/stdout)

Web:
cmd/issue2mdweb -> parser -> github(fetch) -> converter(render) -> HTTP response
```

### 错误与退出码（CLI）

来自 `internal/cli/exitcode.go`:

- `0`: 全部成功
- `1`: 通用运行时错误
- `2`: 参数错误
- `3`: 鉴权/鉴权范围错误
- `4`: 批处理部分失败
- `5`: 输出文件冲突（未使用 `--force`）

## 运行模式

### CLI 模式

单条 URL:

```bash
issue2md <url> [flags]
```

批处理:

```bash
issue2md --input-file urls.txt --output out [flags]
```

主要 flags（见 `internal/config/loader.go`）:

- `--output`: 输出文件或目录
- `--format`: 仅支持 `markdown`
- `--include-comments`: 是否包含评论（默认 `true`）
- `--input-file`: 批量输入文件（每行一个 URL）
- `--stdout`: 输出到标准输出（不能与 `--input-file` 同时使用）
- `--force`: 覆盖同名输出文件
- `--token`: GitHub token（优先于 `GITHUB_TOKEN`）
- `--lang`: AI 总结语言覆盖（如 `zh` / `en`）

### Web 模式

来自 `cmd/issue2mdweb/handler.go`:

- `GET /` 页面入口
- `POST /convert` 转换接口（表单字段 `url`）
- `GET /openapi.json` 返回本地生成的 OpenAPI JSON
- `GET /swagger` 文档入口页

OpenAPI 文档由 `make swagger` 生成到 `docs/`。

## 常用命令（Makefile）

```bash
make help           # 查看所有目标
make check-tools    # 检查 Go/gofmt，提示可选工具 golangci-lint/swag
make fmt            # 格式化所有受 Git 跟踪的 .go 文件
make test           # 运行全部测试
make cover          # 覆盖率报告（coverage.out）
make cover-check    # 覆盖率门禁（默认 >= 80，可用 COVER_MIN 调整）
make lint           # golangci-lint run --config .golangci.yaml ./...
make install-swag   # 安装 swag CLI
make swagger        # 生成 docs/swagger.json 与 docs/swagger.yaml
make swagger-check  # 校验 swagger 生成文件存在
make build-all      # 构建全部二进制
make build-cli      # 构建 issue2md
make build-web      # 构建 issue2mdweb
make install-cli    # 安装 issue2md 到 GOBIN（或 GOPATH/bin）
make install-web    # 安装 issue2mdweb 到 GOBIN（或 GOPATH/bin）
make run-cli ARGS="https://github.com/<owner>/<repo>/issues/1"
make run-web
make ci             # fmt + lint + test
make clean          # 清理 bin 与覆盖率产物
```

## 文档索引

- 核心规范: `constitution.md`
- 产品/技术规格: `specs/001-core-functionality/spec.md`
- 实施计划: `specs/001-core-functionality/plan.md`
- 开发任务拆分: `specs/001-core-functionality/tasks.md`
- API 草案: `specs/001-core-functionality/api-sketch.md`
- 生成的 OpenAPI:
  - `docs/swagger.json`
  - `docs/swagger.yaml`
- Lint 配置: `.golangci.yaml`

## 文档维护规则

- 维护责任人: `Not found in repo`
- 更新时机:
  - 新增/变更 CLI flags 后
  - 新增/变更 HTTP 路由后
  - Makefile 新增/删除目标后
  - `docs/swagger.json` / `docs/swagger.yaml` 重新生成后
- 提交前最小检查:
  - `make fmt`
  - `make lint`
  - `make test`
  - `make swagger-check`（涉及 API 文档时）
- CI 状态: `Not found in repo`（未发现 `.github/workflows/*`）

## 联系信息

- desperateslope@gmail.com
