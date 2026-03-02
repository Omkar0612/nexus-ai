.PHONY: help build test lint fmt clean install run docker-build docker-run coverage security deps tidy

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := nexus
DOCKER_IMAGE := ghcr.io/omkar0612/nexus-ai
DOCKER_TAG := latest
GO := go
GOFLAGS := -v
CGO_ENABLED := 1

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

##@ General

help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

deps: ## Download Go module dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GO) mod download
	tidy: ## Tidy Go module dependencies
	@echo "$(BLUE)Tidying dependencies...$(NC)"
	$(GO) mod tidy

fmt: ## Format Go code
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GO) fmt ./...
	@echo "$(BLUE)Running goimports...$(NC)"
	@command -v goimports >/dev/null 2>&1 || { echo "Installing goimports..."; $(GO) install golang.org/x/tools/cmd/goimports@latest; }
	goimports -w .

lint: ## Run golangci-lint
	@echo "$(BLUE)Running golangci-lint...$(NC)"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "$(RED)golangci-lint not installed. Install from: https://golangci-lint.run/$(NC)"; exit 1; }
	golangci-lint run --timeout=5m

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GO) vet ./...

##@ Building

build: ## Build the binary
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/nexus
	@echo "$(GREEN)Build complete: $(BINARY_NAME)$(NC)"

build-all: ## Build binaries for all platforms
	@echo "$(BLUE)Building for all platforms...$(NC)"
	@mkdir -p dist
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/nexus
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/nexus
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/nexus
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/nexus
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/nexus
	@echo "$(GREEN)Multi-platform build complete!$(NC)"

install: ## Install the binary to GOPATH/bin
	@echo "$(BLUE)Installing $(BINARY_NAME)...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) install ./cmd/nexus
	@echo "$(GREEN)Installed to: $$(go env GOPATH)/bin/$(BINARY_NAME)$(NC)"

##@ Testing

test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test ./... -v -coverprofile=coverage.txt -covermode=atomic
	@echo "$(GREEN)Tests complete!$(NC)"

test-short: ## Run short tests
	@echo "$(BLUE)Running short tests...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test ./... -short -v

test-verbose: ## Run tests with verbose output
	@echo "$(BLUE)Running tests with verbose output...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test ./... -v -race -coverprofile=coverage.txt -covermode=atomic

coverage: test ## Generate test coverage report
	@echo "$(BLUE)Generating coverage report...$(NC)"
	$(GO) tool cover -html=coverage.txt -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"
	@echo "$(BLUE)Coverage summary:$(NC)"
	@$(GO) tool cover -func=coverage.txt | tail -1

bench: ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./...

##@ Security

security: ## Run security scans
	@echo "$(BLUE)Running security scans...$(NC)"
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..."; $(GO) install github.com/securego/gosec/v2/cmd/gosec@latest; }
	gosec -fmt=text ./...
	@echo "$(BLUE)Running govulncheck...$(NC)"
	@command -v govulncheck >/dev/null 2>&1 || { echo "Installing govulncheck..."; $(GO) install golang.org/x/vuln/cmd/govulncheck@latest; }
	govulncheck ./...

##@ Docker

docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

docker-run: ## Run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run -p 7070:7070 \
		-e NEXUS_LLM_PROVIDER=ollama \
		-e NEXUS_LLM_BASE_URL=http://host.docker.internal:11434/v1 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: docker-build ## Push Docker image to registry
	@echo "$(BLUE)Pushing Docker image...$(NC)"
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "$(GREEN)Docker image pushed!$(NC)"

##@ Running

run: build ## Build and run the application
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	./$(BINARY_NAME) start

dev: ## Run in development mode with hot reload (requires air)
	@command -v air >/dev/null 2>&1 || { echo "$(RED)air not installed. Install with: go install github.com/cosmtrek/air@latest$(NC)"; exit 1; }
	air

##@ Cleanup

clean: ## Remove build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.txt coverage.html
	rm -rf *.log
	@echo "$(GREEN)Cleanup complete!$(NC)"

clean-all: clean ## Remove all generated files including caches
	@echo "$(BLUE)Deep cleaning...$(NC)"
	$(GO) clean -cache -testcache -modcache
	@echo "$(GREEN)Deep cleanup complete!$(NC)"

##@ Documentation

docs: ## Generate documentation
	@echo "$(BLUE)Generating documentation...$(NC)"
	@command -v godoc >/dev/null 2>&1 || { echo "Installing godoc..."; $(GO) install golang.org/x/tools/cmd/godoc@latest; }
	@echo "$(GREEN)Start godoc server with: godoc -http=:6060$(NC)"
	@echo "$(GREEN)View docs at: http://localhost:6060/pkg/github.com/Omkar0612/nexus-ai/$(NC)"

##@ Release

release: ## Create a new release (requires goreleaser)
	@command -v goreleaser >/dev/null 2>&1 || { echo "$(RED)goreleaser not installed. Install from: https://goreleaser.com/$(NC)"; exit 1; }
	@echo "$(BLUE)Creating release...$(NC)"
	goreleaser release --clean

release-snapshot: ## Create a snapshot release (no publish)
	@command -v goreleaser >/dev/null 2>&1 || { echo "$(RED)goreleaser not installed. Install from: https://goreleaser.com/$(NC)"; exit 1; }
	@echo "$(BLUE)Creating snapshot release...$(NC)"
	goreleaser release --snapshot --clean

##@ Pre-commit

pre-commit: fmt lint vet test ## Run pre-commit checks
	@echo "$(GREEN)All pre-commit checks passed!$(NC)"

ci: deps tidy fmt lint vet test security ## Run all CI checks locally
	@echo "$(GREEN)All CI checks passed!$(NC)"

##@ Database

db-reset: ## Reset SQLite database (WARNING: deletes data)
	@echo "$(RED)WARNING: This will delete all data!$(NC)"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		rm -f nexus.db; \
		echo "$(GREEN)Database reset complete$(NC)"; \
	fi

##@ Info

version: ## Show version info
	@echo "Go version: $$($(GO) version)"
	@echo "CGO enabled: $(CGO_ENABLED)"
	@echo "Binary name: $(BINARY_NAME)"
	@echo "Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)"

env: ## Show environment info
	@echo "$(BLUE)Go environment:$(NC)"
	$(GO) env
