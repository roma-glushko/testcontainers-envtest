.PHONY: install test lint build clean help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "==> Installing Java dependencies..."
	@mvn -B dependency:resolve

test: ## Run tests
	@echo "==> Running Java tests..."
	@mvn -B test

lint: ## Run linters (checkstyle)
	@echo "==> Running Java linter..."
	@mvn -B validate

build: ## Build package
	@echo "==> Building Java package..."
	@mvn -B package -DskipTests

clean: ## Clean build artifacts
	@echo "==> Cleaning Java artifacts..."
	@mvn -B clean
	@rm -rf target
