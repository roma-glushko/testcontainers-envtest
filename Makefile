.PHONY: all test test-go test-python test-java test-docker image build-docker-all lint lint-go lint-python clean help

# Configuration
K8S_VERSION ?= 1.31.4
IMAGE_NAME ?= ghcr.io/roma-glushko/testcontainers-envtest

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: lint test # Default target

# ============================================================================
# Dependencies
# ============================================================================

install: install-go install-python install-java


install-go: ## Install Go dependencies
	cd go && go mod download

install-python: ## Install Python dependencies
	cd python && uv sync --group dev

install-java: ## Install Java dependencies
	cd java && mvn -B dependency:resolve

test: test-go test-python test-java ## Run all tests

test-go: ## Run Go tests
	@echo "==> Running Go tests..."
	cd go && go test -v -race -coverprofile=coverage.out ./...

test-python: ## Run Python tests
	@echo "==> Running Python tests..."
	cd python && uv run --group test pytest -v --cov=testcontainers_envtest --cov-report=term-missing


test-java: ## Run Java tests
	@echo "==> Running Java tests..."
	cd java && mvn -B test

test-docker: ## Run Docker build test
	@echo "==> Building and testing Docker image..."
	cd docker && docker build -t envtest:test --build-arg KUBERNETES_VERSION=1.31.0 .
	@echo "==> Starting container..."
	docker run -d --name envtest-test -p 6443:6443 envtest:test
	@echo "==> Waiting for API server..."
	@sleep 30
	@curl -sk https://localhost:6443/healthz && echo " - Health check passed" || (docker logs envtest-test && docker rm -f envtest-test && exit 1)
	@docker rm -f envtest-test
	@echo "==> Docker test passed"

lint: lint-go lint-python ## Run all linters

lint-go: ## Run Go linter
	@echo "==> Running Go linter..."
	cd go && go vet ./...
	@which golangci-lint > /dev/null && (cd go && golangci-lint run) || echo "golangci-lint not installed, skipping"

lint-python: ## Run Python linter
	@echo "==> Running Python linter..."
	cd python && uv run --group dev ruff check . && uv run --group dev mypy testcontainers_envtest --ignore-missing-imports

image: ## Build Docker image (use K8S_VERSION=x.y.z to specify version)
	@echo "==> Building Docker image for Kubernetes $(K8S_VERSION)..."
	docker build -t $(IMAGE_NAME):v$(K8S_VERSION) \
		--build-arg KUBERNETES_VERSION=$(K8S_VERSION) \
		./docker
	@echo "==> Built $(IMAGE_NAME):v$(K8S_VERSION)"

build-docker-all: ## Build Docker image for all Kubernetes versions
	@echo "==> Building Docker images for all versions..."
	@for version in 1.27.16 1.28.15 1.29.12 1.30.8 1.31.4; do \
		echo "Building for Kubernetes $$version..."; \
		docker build -t $(IMAGE_NAME):v$$version \
			--build-arg KUBERNETES_VERSION=$$version \
			./docker; \
	done
	docker tag $(IMAGE_NAME):v1.31.4 $(IMAGE_NAME):latest
	@echo "==> Built all images"

build-go: ## Build Go module
	@echo "==> Building Go module..."
	cd go && go build ./...

build-python: ## Build Python package
	@echo "==> Building Python package..."
	cd python && uv build

build-java: ## Build Java package
	@echo "==> Building Java package..."
	cd java && mvn -B package -DskipTests

clean: ## Clean build artifacts
	@echo "==> Cleaning..."
	rm -rf go/coverage.out
	rm -rf python/dist python/build python/*.egg-info python/.pytest_cache python/.coverage python/.mypy_cache python/.ruff_cache
	rm -rf python/testcontainers_envtest/__pycache__ python/tests/__pycache__
	rm -rf python/.venv python/uv.lock
	rm -rf java/target