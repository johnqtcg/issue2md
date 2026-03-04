<a name="unreleased"></a>
## [Unreleased]


<a name="v0.1.0"></a>
## v0.1.0 - 2026-03-04
### Build
- **docker:** add Dockerfile and build helper configs
- **makefile:** add docker-build target for issue2md image

### Chore
- refresh docs and ignore local artifacts
- add install targets and update setup docs
- **deps:** sync module files for go 1.25.7

### Ci
- enforce coverage gate in test workflow step
- add GitHub Actions workflow for test, lint, and build

### Docs
- **readme:** add docker quick start instructions
- **readme:** update project structure and CI documentation
- **readme:** refresh project overview and usage examples

### Feat
- bootstrap issue2md CLI, web, and test suites
- **web:** support offline swagger ui docs

### Fix
- enforce safe output perms and pin ci analysis tools
- correct stdout export, PR thread, and invalid URL exit code
- embed web assets and align assigned event mapping
- **ci:** stabilize goconst lint and print linter version
- **ci:** install golangci-lint v2 in workflow
- **web:** add graceful shutdown and ci integration gates

### Refactor
- apply fieldalignment across structs
- **web:** migrate tests and preserve swagger docs

### Test
- **web:** avoid redirect auto-follow in swagger e2e case
- **webapp:** add handler and template unit tests


[Unreleased]: https://github.com/johnqtcg/issue2md/compare/v0.1.0...HEAD
