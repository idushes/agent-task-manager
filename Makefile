# Makefile for Agent Task Manager

# Docker configuration
DOCKER_USERNAME ?= dushes
IMAGE_NAME = agent-task-manager
PLATFORMS = linux/amd64,linux/arm64
BUILDER_NAME = cloud-dushes-builder

# Version detection
GIT_TAG := $(shell git describe --tags --exact-match 2>/dev/null)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
BUILD_DATE := $(shell date +%Y%m%d-%H%M%S)

# Version logic
ifdef VERSION
    TAG = $(VERSION)
else ifdef GIT_TAG
    TAG = $(GIT_TAG)
else
    TAG = $(GIT_BRANCH)-$(GIT_COMMIT)-$(BUILD_DATE)
endif

# Fallback if git is not available
ifeq ($(TAG),)
    TAG = latest
endif

# Full image name
FULL_IMAGE_NAME = $(DOCKER_USERNAME)/$(IMAGE_NAME):$(TAG)
LATEST_IMAGE_NAME = $(DOCKER_USERNAME)/$(IMAGE_NAME):latest

# Go configuration
GO_VERSION = 1.24.3
APP_PORT = 8081

.PHONY: help build build-multi push build-and-push test clean docker-login setup-buildx version-info

# Default target
help: ## Show this help message
	@echo "Agent Task Manager - Makefile Commands"
	@echo "====================================="
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Environment Variables:"
	@echo "  DOCKER_USERNAME  Docker Hub username (default: dushes)"
	@echo "  VERSION         Manual version override (e.g., v1.0.0)"
	@echo "  BUILDER_NAME    Docker buildx builder (default: cloud-dushes-builder)"
	@echo ""
	@echo "Version Detection:"
	@echo "  1. Manual VERSION env var (highest priority)"
	@echo "  2. Git tag if on tagged commit"
	@echo "  3. branch-commit-timestamp for dev builds"
	@echo ""
	@echo "Example usage:"
	@echo "  make build-and-push                    # Auto version"
	@echo "  make VERSION=v1.0.0 build-and-push    # Manual version"
	@echo "  make release-version VERSION=v1.0.0   # Release with version"

version-info: ## Show current version information
	@echo "Version Information:"
	@echo "=================="
	@echo "Current TAG: $(TAG)"
	@echo "Git Tag: $(GIT_TAG)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Git Branch: $(GIT_BRANCH)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo ""
	@echo "Image Names:"
	@echo "  Tagged: $(FULL_IMAGE_NAME)"
	@echo "  Latest: $(LATEST_IMAGE_NAME)"

setup-buildx: ## Setup Docker buildx to use cloud builder
	@echo "Using existing cloud builder: $(BUILDER_NAME)"
	@docker buildx use $(BUILDER_NAME)
	@docker buildx inspect $(BUILDER_NAME)

docker-login: ## Login to Docker Hub
	@echo "Logging in to Docker Hub..."
	@docker login

build: ## Build Docker image for current platform
	@echo "Building Docker image for current platform..."
	@echo "Version: $(TAG)"
	@docker build -t $(FULL_IMAGE_NAME) .
	@echo "✅ Image built: $(FULL_IMAGE_NAME)"

build-multi: setup-buildx docker-login version-info ## Build multi-platform Docker image and push (cloud builder requirement)
	@echo "Building and pushing multi-platform Docker image using cloud builder..."
	@echo "Builder: $(BUILDER_NAME)"
	@echo "Platforms: $(PLATFORMS)"
	@echo "Version: $(TAG)"
	@echo "Note: Cloud builder requires push - image will be built and pushed to registry."
	@docker buildx build \
		--builder $(BUILDER_NAME) \
		--platform $(PLATFORMS) \
		-t $(FULL_IMAGE_NAME) \
		--push \
		.
	@echo "✅ Multi-platform image built and pushed: $(FULL_IMAGE_NAME)"

build-and-push: setup-buildx docker-login version-info ## Build multi-platform image and push to Docker Hub
	@echo "Building and pushing multi-platform Docker image using cloud builder..."
	@echo "Builder: $(BUILDER_NAME)"
	@echo "Platforms: $(PLATFORMS)"
	@echo "Version: $(TAG)"
ifdef GIT_TAG
	@echo "🏷️  Building release version from git tag: $(GIT_TAG)"
	@docker buildx build \
		--builder $(BUILDER_NAME) \
		--platform $(PLATFORMS) \
		-t $(FULL_IMAGE_NAME) \
		-t $(LATEST_IMAGE_NAME) \
		--push \
		.
	@echo "✅ Release image built and pushed:"
	@echo "   📦 $(FULL_IMAGE_NAME)"
	@echo "   📦 $(LATEST_IMAGE_NAME)"
else
	@echo "🚧 Building development version"
	@docker buildx build \
		--builder $(BUILDER_NAME) \
		--platform $(PLATFORMS) \
		-t $(FULL_IMAGE_NAME) \
		--push \
		.
	@echo "✅ Development image built and pushed:"
	@echo "   📦 $(FULL_IMAGE_NAME)"
endif

push: ## Push Docker image to Docker Hub
	@echo "Pushing image to Docker Hub..."
	@docker push $(FULL_IMAGE_NAME)
	@echo "✅ Image pushed: $(FULL_IMAGE_NAME)"

test-local: build ## Test Docker image locally
	@echo "Testing Docker image locally..."
	@echo "Starting container on port $(APP_PORT)..."
	@docker run -d --name $(IMAGE_NAME)-test -p $(APP_PORT):$(APP_PORT) $(FULL_IMAGE_NAME)
	@sleep 3
	@echo "Testing health endpoint..."
	@curl -f http://localhost:$(APP_PORT)/health || (echo "❌ Health check failed" && exit 1)
	@echo ""
	@echo "Testing ready endpoint..."
	@curl -f http://localhost:$(APP_PORT)/ready || (echo "❌ Ready check failed" && exit 1)
	@echo ""
	@echo "✅ All tests passed!"
	@docker stop $(IMAGE_NAME)-test
	@docker rm $(IMAGE_NAME)-test

run: ## Run the application locally with Go
	@echo "Running application locally..."
	@go run .

dev: ## Run application in development mode with hot reload
	@echo "Starting development server..."
	@go run .

clean: ## Clean up Docker images and containers
	@echo "Cleaning up..."
	@docker stop $(IMAGE_NAME)-test 2>/dev/null || true
	@docker rm $(IMAGE_NAME)-test 2>/dev/null || true
	@docker rmi $(FULL_IMAGE_NAME) 2>/dev/null || true
	@echo "✅ Cleanup completed"

inspect-image: ## Inspect multi-platform image
	@echo "Inspecting image platforms..."
	@docker buildx imagetools inspect $(FULL_IMAGE_NAME)

# Development targets
deps: ## Download Go dependencies
	@echo "Downloading Go dependencies..."
	@go mod download
	@echo "✅ Dependencies downloaded"

tidy: ## Tidy Go modules
	@echo "Tidying Go modules..."
	@go mod tidy
	@echo "✅ Go modules tidied"

fmt: ## Format Go code
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "✅ Code formatted"

vet: ## Run Go vet
	@echo "Running Go vet..."
	@go vet ./...
	@echo "✅ Go vet passed"

# Release targets
tag-version: ## Create and push git tag (use: make tag-version VERSION=v1.0.0)
ifndef VERSION
	@echo "❌ VERSION is required. Usage: make tag-version VERSION=v1.0.0"
	@exit 1
endif
	@echo "Creating git tag: $(VERSION)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "✅ Tag $(VERSION) created and pushed"

release-version: tag-version build-and-push ## Create git tag and release (use: make release-version VERSION=v1.0.0)
	@echo "🚀 Release $(VERSION) completed!"
	@echo "   📦 Image: $(DOCKER_USERNAME)/$(IMAGE_NAME):$(VERSION)"
	@echo "   📦 Latest: $(DOCKER_USERNAME)/$(IMAGE_NAME):latest"

# Quick commands
quick-build: build test-local ## Quick build and test
	@echo "✅ Quick build and test completed"

release: fmt vet build-and-push ## Full release: format, vet, build and push
	@echo "🚀 Release completed: $(FULL_IMAGE_NAME)" 