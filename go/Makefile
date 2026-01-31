.PHONY: install tools test test-integration lint build clean help

TMP_DIR := $(PWD)/../tmp
BIN_DIR := $(TMP_DIR)/bin
GOBIN ?= $(BIN_DIR)

export GOBIN
export PATH := $(BIN_DIR):$(PATH)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: tools ## Install dependencies
	@echo "==> Installing Go dependencies..."
	@go mod download

tools: ## Install development tools
	@echo "==> Installing Go tools..."
	@mkdir -p $(BIN_DIR)
	@command -v $(BIN_DIR)/golangci-lint > /dev/null || curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(BIN_DIR) v2.8.0

test: ## Run tests
	@echo "==> Running Go tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-integration: ## Run integration tests (requires Docker, supports ENVTEST_IMAGE)
	@echo "==> Running Go integration tests..."
	@go test -v -race ./...

lint: tools ## Run linters
	@echo "==> Tidying go.mod..."
	@go mod tidy
	@echo "==> Formatting files..."
	@go fmt ./...
	@echo "==> Vetting..."
	@go vet ./...
	@echo "==> Running golangci-lint..."
	@$(BIN_DIR)/golangci-lint run ./...

build: ## Build the module
	@echo "==> Building Go module..."
	@go build ./...

clean: ## Clean build artifacts
	@echo "==> Cleaning Go artifacts..."
	@rm -rf coverage.out $(BIN_DIR)
