.PHONY: all install test lint build clean help
.PHONY: install-go install-python install-java
.PHONY: test-go test-python test-java test-docker
.PHONY: test-image test-integration-go test-integration-python test-integration-java
.PHONY: lint-go lint-python lint-java
.PHONY: build-go build-python build-java
.PHONY: image build-docker-all

# Configuration
K8S_VERSION ?= 1.35.0
IMAGE_NAME ?= ghcr.io/roma-glushko/testcontainers-envtest

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

all: lint test ## Run lint and test

# =============================================================================
# Aggregate commands
# =============================================================================

install: install-go install-python install-java ## Install all dependencies

test: test-go test-python test-java ## Run all tests

lint: lint-go lint-python lint-java ## Run all linters

build: build-go build-python build-java ## Build all packages

clean: ## Clean all build artifacts
	@$(MAKE) -C go clean
	@$(MAKE) -C python clean
	@$(MAKE) -C java clean

# =============================================================================
# Go
# =============================================================================

install-go: ## Install Go dependencies
	@$(MAKE) -C go install

test-go: ## Run Go tests
	@$(MAKE) -C go test

lint-go: ## Run Go linter
	@$(MAKE) -C go lint

build-go: ## Build Go module
	@$(MAKE) -C go build

# =============================================================================
# Python
# =============================================================================

install-python: ## Install Python dependencies
	@$(MAKE) -C python install

test-python: ## Run Python tests
	@$(MAKE) -C python test

lint-python: ## Run Python linter
	@$(MAKE) -C python lint

build-python: ## Build Python package
	@$(MAKE) -C python build

# =============================================================================
# Java
# =============================================================================

install-java: ## Install Java dependencies
	@$(MAKE) -C java install

test-java: ## Run Java tests
	@$(MAKE) -C java test

lint-java: ## Run Java linter
	@$(MAKE) -C java lint

build-java: ## Build Java package
	@$(MAKE) -C java build

# =============================================================================
# Docker
# =============================================================================

image: ## Build Docker image (K8S_VERSION=x.y.z)
	@echo "==> Building Docker image for Kubernetes $(K8S_VERSION)..."
	docker build -t $(IMAGE_NAME):v$(K8S_VERSION) \
		--build-arg KUBERNETES_VERSION=$(K8S_VERSION) \
		./docker
	@echo "==> Built $(IMAGE_NAME):v$(K8S_VERSION)"

build-docker-all: ## Build Docker images for all K8s versions
	@echo "==> Building Docker images for all versions..."
	@for version in 1.31.0 1.32.0 1.33.0 1.34.1 1.35.0; do \
		echo "Building for Kubernetes $$version..."; \
		docker build -t $(IMAGE_NAME):v$$version \
			--build-arg KUBERNETES_VERSION=$$version \
			./docker; \
	done
	docker tag $(IMAGE_NAME):v1.35.0 $(IMAGE_NAME):latest
	@echo "==> Built all images"

test-docker: ## Build and test Docker image
	@echo "==> Building and testing Docker image..."
	docker build -t envtest:test --build-arg KUBERNETES_VERSION=1.35.0 ./docker
	@echo "==> Starting container..."
	docker run -d --name envtest-test -p 6443:6443 envtest:test
	@echo "==> Waiting for API server..."
	@sleep 30
	@curl -sk https://localhost:6443/healthz && echo " - Health check passed" || (docker logs envtest-test && docker rm -f envtest-test && exit 1)
	@docker rm -f envtest-test
	@echo "==> Docker test passed"

# =============================================================================
# Image Testing (cross-language integration tests)
# =============================================================================

test-image: ## Test envtest image against all language test suites (ENVTEST_IMAGE=...)
	@echo "==> Testing image: $(ENVTEST_IMAGE)"
	@ENVTEST_IMAGE=$(ENVTEST_IMAGE) $(MAKE) test-integration-go
	@ENVTEST_IMAGE=$(ENVTEST_IMAGE) $(MAKE) test-integration-python
	@ENVTEST_IMAGE=$(ENVTEST_IMAGE) $(MAKE) test-integration-java

test-integration-go: ## Run Go integration tests (supports ENVTEST_IMAGE)
	@ENVTEST_IMAGE=$(ENVTEST_IMAGE) $(MAKE) -C go test-integration

test-integration-python: ## Run Python integration tests (supports ENVTEST_IMAGE)
	@ENVTEST_IMAGE=$(ENVTEST_IMAGE) $(MAKE) -C python test-integration

test-integration-java: ## Run Java integration tests (supports ENVTEST_IMAGE)
	@ENVTEST_IMAGE=$(ENVTEST_IMAGE) $(MAKE) -C java test-integration
