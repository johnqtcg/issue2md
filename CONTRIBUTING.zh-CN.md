# 贡献指南

感谢你参与 `issue2md` 项目。本文档说明如何提交高质量变更，并与现有工程规范保持一致。

## 1. 基本原则

请优先遵循以下仓库规范文件：

- `constitution.md`（最高优先级）
- `AGENTS.md`

核心要求：

- 保持实现简单（Simplicity First），避免过度设计。
- 新功能/缺陷修复优先采用 TDD（Red-Green-Refactor）。
- 单元测试优先使用 Table-Driven Tests。
- 错误处理必须显式且可追踪（error wrapping）。

## 2. 开发环境

前置条件：

- Go `>= 1.25`
- 可选工具：`golangci-lint`、`swag`

常用命令：

```bash
make help
make test
make lint
make cover-check COVER_MIN=80
make build-all
```

## 3. 分支与提交

- 分支建议：`feature/<topic>`、`fix/<topic>`、`chore/<topic>`
- 提交信息必须使用 Conventional Commits：

```text
<type>(<scope>): <subject>
```

示例：

```text
feat(cli): support batch output force overwrite
fix(web): handle openapi file missing as 503
docs(readme): refresh command verifiability section
```

## 4. 代码与测试要求

提交前请至少满足：

1. 代码可编译，关键路径可运行。
2. 相关测试通过：`make test`
3. 静态检查通过：`make lint`
4. 覆盖率门禁满足：`make cover-check COVER_MIN=80`
5. 若涉及 Web API 文档变更，执行：`make swagger-check`

测试建议：

- 优先补“缺陷复现测试”再修复代码。
- 对边界条件、异常路径、参数冲突增加测试用例。
- 避免无价值测试（仅覆盖 happy path 且不验证行为）。

## 5. Pull Request 要求

请在 PR 描述中包含：

- 变更背景与目标
- 关键设计点与取舍
- 测试证据（命令 + 结果摘要）
- 兼容性影响（如 CLI 参数/HTTP 路由/输出格式变化）

推荐 PR 自检清单：

- [ ] 已阅读并遵守 `constitution.md`
- [ ] 已补充必要测试（含边界/异常）
- [ ] `make test` 通过
- [ ] `make lint` 通过
- [ ] 文档已同步（如 `README.md`、`docs/swagger.json`）

## 6. 安全与行为规范

- 行为规范：见 `CODE_OF_CONDUCT.md`
- 漏洞上报：见 `SECURITY.md`

请不要在公开 Issue/PR 中披露可利用的漏洞细节。

