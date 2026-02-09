.DEFAULT_GOAL := help

GO ?= go
BIN_DIR ?= bin

COVER_FILE ?= coverage.out
COVER_HTML ?= coverage.html
COVER_MIN ?= 80

GOLANGCI_LINT ?= golangci-lint
GOCACHE_DIR ?= /tmp/gocache
GOLANGCI_LINT_CACHE_DIR ?= /tmp/golangci-lint-cache
SWAG ?= swag
SWAG_INSTALL ?= github.com/swaggo/swag/cmd/swag@latest
SWAG_DIR ?= cmd/issue2mdweb
SWAG_ENTRY ?= main.go
SWAG_OUTPUT ?= docs

.PHONY: help check-tools fmt test test-api-integration test-e2e-web cover cover-check lint install-swag swagger swagger-check \
	build-all build-cli build-web build-issue2md build-issue2mdweb \
	install-cli install-web \
	run-cli run-web run-issue2md run-issue2mdweb ci clean web

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-22s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check-tools: ## Check required and optional tools
	@command -v $(GO) >/dev/null || { echo "missing tool: $(GO)"; exit 1; }
	@command -v gofmt >/dev/null || { echo "missing tool: gofmt"; exit 1; }
	@command -v $(GOLANGCI_LINT) >/dev/null || echo "optional tool missing: $(GOLANGCI_LINT) (required for make lint)"
	@command -v $(SWAG) >/dev/null || echo "optional tool missing: $(SWAG) (required for make swagger)"

fmt: ## Format all tracked Go files
	@files=$$(git ls-files '*.go'); \
	if [ -n "$$files" ]; then gofmt -w $$files; else echo "no tracked Go files"; fi

test: ## Run all tests
	GOCACHE=$(GOCACHE_DIR) $(GO) test ./...

test-api-integration: ## Run opt-in API integration tests for web endpoints
	GOCACHE=$(GOCACHE_DIR) ISSUE2MD_API_INTEGRATION=1 $(GO) test ./cmd/issue2mdweb -run Integration -v

test-e2e-web: ## Run opt-in web E2E journey tests
	GOCACHE=$(GOCACHE_DIR) ISSUE2MD_E2E=1 ISSUE2MD_WEB_ADDR=127.0.0.1:18080 $(GO) test ./cmd/issue2mdweb -run E2E -v

cover: ## Run tests with coverage report
	GOCACHE=$(GOCACHE_DIR) $(GO) test ./... -coverprofile=$(COVER_FILE)
	$(GO) tool cover -func=$(COVER_FILE) | tail -n 1

cover-check: cover ## Enforce minimum coverage with COVER_MIN
	@pct=$$($(GO) tool cover -func=$(COVER_FILE) | awk '/total:/ {gsub("%","",$$3); print $$3}'); \
	awk -v p="$$pct" -v m="$(COVER_MIN)" 'BEGIN { if (p+0 < m+0) { printf("coverage %.2f%% < %.2f%%\n", p, m); exit 1 } else { printf("coverage %.2f%% >= %.2f%%\n", p, m); } }'

lint: ## Run golangci-lint
	@command -v $(GOLANGCI_LINT) >/dev/null || { echo "missing tool: $(GOLANGCI_LINT)"; exit 1; }
	GOCACHE=$(GOCACHE_DIR) GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE_DIR) $(GOLANGCI_LINT) run --config .golangci.yaml ./...

install-swag: ## Install swag CLI
	$(GO) install $(SWAG_INSTALL)

swagger: ## Generate OpenAPI docs to docs/
	@command -v $(SWAG) >/dev/null || { echo "missing tool: $(SWAG)"; exit 1; }
	$(SWAG) init --dir $(SWAG_DIR) --generalInfo $(SWAG_ENTRY) --output $(SWAG_OUTPUT) --parseInternal --outputTypes json,yaml
	@rm -f $(SWAG_OUTPUT)/docs.go

swagger-check: swagger ## Verify generated OpenAPI spec file exists
	@test -f $(SWAG_OUTPUT)/swagger.json || { echo "missing generated file: $(SWAG_OUTPUT)/swagger.json"; exit 1; }

build-all: build-issue2md build-issue2mdweb ## Build all binaries

build-cli: build-issue2md ## Build CLI binary

build-web: build-issue2mdweb ## Build web binary

web: build-web ## Build web service binary (compat alias)

build-issue2md: ## Build cmd/issue2md
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/issue2md ./cmd/issue2md

build-issue2mdweb: ## Build cmd/issue2mdweb
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/issue2mdweb ./cmd/issue2mdweb

install-cli: ## Install CLI binary into GOBIN (or GOPATH/bin)
	$(GO) install ./cmd/issue2md

install-web: ## Install web binary into GOBIN (or GOPATH/bin)
	$(GO) install ./cmd/issue2mdweb

run-cli: run-issue2md ## Run CLI app (pass ARGS='...')

run-web: run-issue2mdweb ## Run web app

run-issue2md: ## Run cmd/issue2md (pass ARGS='...')
	$(GO) run ./cmd/issue2md $(ARGS)

run-issue2mdweb: ## Run cmd/issue2mdweb
	$(GO) run ./cmd/issue2mdweb

ci: fmt lint test ## Local CI parity checks

clean: ## Remove build and coverage artifacts
	rm -rf $(BIN_DIR) $(COVER_FILE) $(COVER_HTML)
