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

deploy-crd: ## Deploy the MinecraftServer CRD to Kubernetes
	@echo "Deploying CRD to Kubernetes..."
	@kubectl apply -f config/crd/minecraftserver-crd.yaml
	@echo "CRD deployed successfully"

delete-crd: ## Delete the MinecraftServer CRD from Kubernetes
	@echo "Deleting CRD from Kubernetes..."
	@kubectl delete -f config/crd/minecraftserver-crd.yaml
	@echo "CRD deleted successfully"

check-crd: ## Check if CRD is installed
	@echo "Checking CRD status..."
	@kubectl get crd minecraftservers.homecraft.io || echo "CRD not found"

local-dev: deploy-crd ## Setup for local development
	@echo "Setting up local development environment..."
	@echo "CRD deployed. You can now run 'make run' to start the API"

verify: ## Verify the installation
	@echo "Verifying installation..."
	@go version
	@kubectl version --client
	@echo "Checking if CRD is installed..."
	@kubectl get crd minecraftservers.homecraft.io 2>/dev/null && echo "✓ CRD installed" || echo "✗ CRD not installed - run 'make deploy-crd'"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

all: clean lint build test ## Run all checks and build
	@echo "All tasks completed successfully"
