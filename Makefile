.PHONY: help build test test-race test-coverage lint fmt clean install run

# Variables
BINARY_NAME=nexus-ai
MAIN_PATH=./cmd/nexus
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Colors
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

help: ## Show this help message
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "  ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${RESET}%s\n", substr($$1,4)} \
	}' $(MAKEFILE_LIST)

## Build commands

build: ## Build the binary
	@echo "🔨 Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: bin/$(BINARY_NAME)"

build-all: ## Build for all platforms
	@echo "🔨 Building for all platforms..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@GOOS=linux GOARCH=arm64 go build -o bin/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ All builds complete"

## Test commands

test: ## Run all tests
	@echo "🧪 Running tests..."
	@go test -v ./internal/...

test-race: ## Run tests with race detector
	@echo "🏃 Running tests with race detector..."
	@go test -race -timeout 30s ./internal/...

test-coverage: ## Run tests with coverage report
	@echo "📊 Running tests with coverage..."
	@go test -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./internal/...
	@go tool cover -func=$(COVERAGE_FILE)
	@go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "✅ Coverage report generated: $(COVERAGE_HTML)"

test-integration: ## Run integration tests
	@echo "🔗 Running integration tests..."
	@go test -v -tags=integration ./test/integration/...

test-all: test-race test-coverage ## Run all test suites
	@echo "✅ All tests complete"

## Code quality commands

lint: ## Run linters
	@echo "🔍 Running linters..."
	@go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "Code needs formatting. Run 'make fmt'" && exit 1)
	@echo "✅ Linting complete"

fmt: ## Format code
	@echo "📐 Formatting code..."
	@gofmt -w .
	@echo "✅ Formatting complete"

vet: ## Run go vet
	@go vet ./...

## Dependency commands

install: ## Install dependencies
	@echo "📦 Installing dependencies..."
	@go mod download
	@go mod verify
	@echo "✅ Dependencies installed"

tidy: ## Tidy dependencies
	@echo "🧹 Tidying dependencies..."
	@go mod tidy
	@echo "✅ Dependencies tidied"

## Run commands

run: ## Run the application
	@go run $(MAIN_PATH)

run-mesh: ## Run with mesh network enabled
	@NEXUS_MESH_PORT=5353 go run $(MAIN_PATH)

run-dev: ## Run in development mode with debug logging
	@NEXUS_LOG_LEVEL=debug go run $(MAIN_PATH)

## Utility commands

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	@rm -rf bin/
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@rm -f test_output.log build_output.log
	@echo "✅ Cleaned"

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t nexus-ai:latest .
	@echo "✅ Docker image built"

docker-run: ## Run Docker container
	@docker run -it --rm \
		-e NEXUS_MESH_PORT=5353 \
		-e NEXUS_PREDICTIVE_CONFIDENCE=0.7 \
		-p 8080:8080 \
		nexus-ai:latest

## CI/CD commands

ci: lint test-race test-coverage ## Run CI checks
	@echo "✅ CI checks passed"

release: clean build-all test-all ## Prepare release
	@echo "✅ Release ready"

.DEFAULT_GOAL := help
