# Contributing Guide

Thanks for contributing to `issue2md`.

## 1. Principles

Please follow these repository policies in order:

- `constitution.md` (highest priority)
- `AGENTS.md`

Core expectations:
- Keep solutions simple and avoid over-engineering.
- Prefer TDD (Red-Green-Refactor) for features and bug fixes.
- Prefer table-driven tests for unit tests.
- Handle errors explicitly and wrap errors with context.

## 2. Development Setup

Prerequisites:
- Go `>= 1.25`
- Optional tools: `golangci-lint`, `swag`

Common commands:

```bash
make help
make test
make lint
make cover-check COVER_MIN=80
make build-all
```

## 3. Branch and Commit Rules

Suggested branch names:
- `feature/<topic>`
- `fix/<topic>`
- `chore/<topic>`

Commit messages must follow Conventional Commits:

```text
<type>(<scope>): <subject>
```

Examples:

```text
feat(cli): support batch output force overwrite
fix(web): handle openapi file missing as 503
docs(readme): refresh command verifiability section
```

## 4. Code and Test Requirements

Before opening a PR, ensure at least:

1. Code builds successfully.
2. Relevant tests pass: `make test`.
3. Lint passes: `make lint`.
4. Coverage gate passes: `make cover-check COVER_MIN=80`.
5. If Web API docs changed: run `make swagger-check`.

Testing recommendations:
- Add a failing regression test first, then fix code.
- Cover boundaries, error paths, and argument conflicts.
- Avoid low-value tests that only exercise happy paths.

## 5. Pull Request Requirements

Include in your PR description:
- Background and goal
- Key design decisions and tradeoffs
- Test evidence (commands + result summary)
- Compatibility impact (CLI flags / HTTP routes / output format)

Suggested PR checklist:
- [ ] I followed `constitution.md`
- [ ] I added/updated necessary tests (including edge/error cases)
- [ ] `make test` passed
- [ ] `make lint` passed
- [ ] Docs were updated (for example `README.md`, `docs/swagger.json`)

## 6. Security and Conduct

- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- Vulnerability reporting: [SECURITY.md](SECURITY.md)

Do not disclose exploitable vulnerability details in public Issue/PR threads.
