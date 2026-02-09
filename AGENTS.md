# ==================================
# issue2md Project Context Entry
# ==================================

# --- Import Core Principles (Highest Priority) ---
# Explicitly import the project constitution so the AI loads core principles before reasoning about anything.
@./constitution.md

# --- Core Mission and Role ---
You are a senior Go engineer helping me build a tool called "issue2md".
All your actions must strictly follow the imported project constitution above.

---
## 1. Tech Stack and Environment
- **Language**: Go (version >= 1.25)
- **Build and Test**:
    - Use `Makefile` for standardized operations.
    - Run all tests: `make test`
    - Build web service: `make web`

---
## 2. Git and Version Control
- **Commit message rule**: Strictly follow Conventional Commits.
    - Format: `<type>(<scope>): <subject>`
    - When asked to generate a commit message, you must use this format.

---
## 3. AI Collaboration Instructions
- **When asked to add a new feature**: Your first step should be reading related packages under `internal/` using `@`, then check against the project constitution, and only then propose a plan.
- **When asked to write tests**: Prefer **Table-Driven Tests**.
- **When asked to build the project**: Prefer commands defined in the `Makefile`.
