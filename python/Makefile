.PHONY: install test test-integration lint build clean help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies
	@echo "==> Installing Python dependencies..."
	@uv sync --group dev

test: ## Run tests
	@echo "==> Running Python tests..."
	@uv run --group test pytest -v --cov=testcontainers_envtest --cov-report=term-missing

test-integration: ## Run integration tests (requires Docker, supports ENVTEST_IMAGE)
	@echo "==> Running Python integration tests..."
	@uv run --group test pytest -v -m integration

lint: ## Run linters
	@echo "==> Running ruff..."
	@uv run --group dev ruff check .
	@echo "==> Running mypy..."
	@uv run --group dev mypy testcontainers_envtest --ignore-missing-imports

lint-fix: ## Run linters and fix issues
	@echo "==> Running ruff with fixes..."
	@uv run --group dev ruff check --fix .
	@uv run --group dev ruff format .

build: ## Build package
	@echo "==> Building Python package..."
	@uv build

clean: ## Clean build artifacts
	@echo "==> Cleaning Python artifacts..."
	@rm -rf dist build *.egg-info .pytest_cache .coverage coverage.xml .mypy_cache .ruff_cache
	@rm -rf testcontainers_envtest/__pycache__ tests/__pycache__
	@rm -rf .venv
