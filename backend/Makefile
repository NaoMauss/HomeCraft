# HomeCraft Backend Makefile

# Variables
BINARY_NAME=api
DOCKER_IMAGE=homecraft/backend
DOCKER_TAG?=latest
NAMESPACE?=default

.PHONY: help build run test clean docker-build docker-push deploy-crd deploy-api local-dev

help: ## Display this help message
	@echo "HomeCraft Backend - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the backend binary
	@echo "Building backend..."
	@go build -o bin/$(BINARY_NAME) ./cmd/api
	@echo "Build complete: bin/$(BINARY_NAME)"

run: build ## Build and run the backend locally
	@echo "Running backend..."
	@./bin/$(BINARY_NAME)

test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-short: ## Run tests without verbose output
	@echo "Running tests (short)..."
	@go test ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-func: ## Show coverage by function
	@echo "Running tests with function coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -race -v ./...

test-bench: ## Run benchmarks
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

test-unit: ## Run only unit tests (fast)
	@echo "Running unit tests..."
	@go test -v -short ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean
	@echo "Clean complete"

tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	@go mod tidy

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

lint: fmt vet ## Run linters

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-push: docker-build ## Push Docker image to registry
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Docker image pushed: $(DOCKER_IMAGE):$(DOCKER_TAG)"

check-namespace: ## Check if namespace is created
	@echo "Checking namespace status..."
	@kubectl get namespace minecraft-servers 2>/dev/null || echo "Namespace not found - ensure Flux has synced the infra/homecraft/base resources"

check-crd: ## Check if CRD is installed
	@echo "Checking CRD status..."
	@kubectl get crd minecraftservers.homecraft.io 2>/dev/null || echo "CRD not found - ensure Flux has synced the infra/homecraft/base resources"

check-flux: ## Check if Flux is syncing resources
	@echo "Checking Flux system Kustomization status..."
	@kubectl get kustomization flux-system -n flux-system 2>/dev/null || echo "Flux system Kustomization not found"
	@echo "\nForcing reconciliation..."
	@flux reconcile kustomization flux-system --with-source 2>/dev/null || echo "flux CLI not installed - use: kubectl reconcile kustomization flux-system -n flux-system"

verify: ## Verify the installation
	@echo "Verifying installation..."
	@go version
	@kubectl version --client
	@echo "\nChecking if namespace is created..."
	@kubectl get namespace minecraft-servers 2>/dev/null && echo "✓ Namespace created" || echo "✗ Namespace not created - check Flux sync status"
	@echo "\nChecking if CRD is installed..."
	@kubectl get crd minecraftservers.homecraft.io 2>/dev/null && echo "✓ CRD installed" || echo "✗ CRD not installed - check Flux sync status"
	@echo "\nChecking Flux system..."
	@kubectl get kustomization flux-system -n flux-system 2>/dev/null && echo "✓ Flux is running" || echo "✗ Flux not found - ensure Flux is installed"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

all: clean lint build test ## Run all checks and build
	@echo "All tasks completed successfully"
