# Image URL to use for building/pushing image targets
REGISTRY ?= ghcr.io
REPOSITORY ?= naomauss
IMAGE_NAME ?= homecraft-operator
TAG ?= latest
IMG ?= $(REGISTRY)/$(REPOSITORY)/$(IMAGE_NAME):$(TAG)

# Get the currently used golang install path
GOBIN := $(shell go env GOPATH)/bin

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: fmt vet ## Build manager binary
	go build -o bin/manager cmd/main.go

.PHONY: run
run: fmt vet ## Run the controller from your host
	go run ./cmd/main.go

.PHONY: docker-build
docker-build: ## Build docker image
	docker build -t $(IMG) -f Dockerfile ..

.PHONY: docker-push
docker-push: ## Push docker image
	docker push $(IMG)

.PHONY: docker-build-push
docker-build-push: docker-build docker-push ## Build and push docker image

##@ Deployment

.PHONY: deploy
deploy: ## Deploy controller to the K8s cluster
	helm upgrade --install homecraft-operator ./chart \
		--namespace homecraft-system \
		--create-namespace \
		--set image.repository=$(REGISTRY)/$(REPOSITORY)/$(IMAGE_NAME) \
		--set image.tag=$(TAG)

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster
	helm uninstall homecraft-operator -n homecraft-system

##@ Dependencies

.PHONY: deps
deps: ## Download go dependencies
	go mod download
	go mod tidy

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf bin/
	go clean
