# issue2md Project Development Constitution
# Version: 1.0, Ratified: 2026-02-09
This document defines the non-negotiable core development principles of this project.
All AI agents must follow these rules unconditionally when planning and implementing technology.
---

## Article 1: Simplicity First
**Core idea:** Follow Go’s “less is more” philosophy. Never add unnecessary abstraction. Never introduce non-essential dependencies.

- 1.1 (YAGNI): You Aren’t Gonna Need It. Only implement features explicitly required in spec.md.
- 1.2 (Standard Library First): Unless there is a very strong reason, always prefer Go’s standard library.
For example, use net/http for web services instead of Gin or Echo.
- 1.3 (No Over-Engineering): Avoid complex design patterns. Simple functions and data structures are better than complex interfaces and inheritance-like structures.

---

## Article 2: Test-First Imperative (Non-Negotiable)
**Core idea:** Every new feature or bug fix must start with one or more failing tests.

- 2.1 (TDD Cycle): Strictly follow Red-Green-Refactor
(write a failing test → make it pass → refactor).
- 2.2 (Table-Driven Tests): Unit tests should use table-driven style first, to cover multiple inputs and edge cases.
- 2.3 (No Mock Abuse): Prefer integration tests with real dependencies or fake objects
(such as an in-memory GitHub API test server), instead of relying too much on mocks.

---

## Article 3: Clarity and Explicitness
**Core idea:** Code should first be easy for humans to understand, and only then for machines to run.

- 3.1 (Error Handling): Non-negotiable: Every error must be handled explicitly.
Never discard errors with _.
When passing errors upward, always wrap them with fmt.Errorf("...: %w", err).
- 3.2 (No Global State): Never use global variables to pass state.
All dependencies must be injected explicitly through function parameters or struct fields.
- 3.3 (Meaningful Comments): Comments should explain why, not what.
All public APIs must have clear GoDoc comments.

---

## Article 4: Single Responsibility
**Core idea:** Every package, file, and function should do one thing well.

- 4.1 (Package Cohesion): Packages under internal must stay highly cohesive and loosely coupled.
For example, the github package should only handle GitHub API interaction and must not include Markdown conversion logic.
- 4.2 (Interface Segregation): Define small, focused interfaces, not large “god interfaces.”

---

## Governance

This constitution has the highest priority.
It overrides any CLAUDE.md file or single-session instruction.
Before generating any plan (plan.md), a constitutional compliance check must be performed first.