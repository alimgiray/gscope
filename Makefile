# GScope Makefile
# A comprehensive build system for the GScope GitHub repository analytics tool

# Variables
BINARY_NAME=gscope
BUILD_DIR=build
MAIN_PATH=cmd/server/main.go
TEST_DIRS=./internal/... ./pkg/...
COVERAGE_DIR=coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html

# Go build flags
LDFLAGS=-ldflags "-X main.version=$(shell git describe --tags --always --dirty)"
BUILD_FLAGS=-trimpath

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo "GScope - GitHub Repository Analytics Tool"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make test        # Run all tests"
	@echo "  make build       # Build the application"
	@echo "  make run         # Run the application"
	@echo "  make clean       # Clean build artifacts"

# Build targets
.PHONY: build
build: ## Build the application
	@echo "Building GScope..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "Building GScope for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux"

.PHONY: build-darwin
build-darwin: ## Build for macOS
	@echo "Building GScope for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-darwin"

.PHONY: build-windows
build-windows: ## Build for Windows
	@echo "Building GScope for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)-windows.exe"

.PHONY: build-all
build-all: build build-linux build-darwin build-windows ## Build for all platforms

# Test targets
.PHONY: test
test: ## Run all tests
	@echo "Running all tests..."
	go test $(TEST_DIRS) -v

.PHONY: test-short
test-short: ## Run tests with short flag
	@echo "Running tests (short mode)..."
	go test $(TEST_DIRS) -v -short

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	go test $(TEST_DIRS) -v -race

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	go test $(TEST_DIRS) -v -coverprofile=$(COVERAGE_FILE)
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

.PHONY: test-coverage-func
test-coverage-func: ## Show function coverage
	@echo "Function coverage:"
	@go tool cover -func=$(COVERAGE_FILE)

.PHONY: test-benchmark
test-benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	go test $(TEST_DIRS) -v -bench=.

.PHONY: test-benchmark-mem
test-benchmark-mem: ## Run benchmark tests with memory profiling
	@echo "Running benchmark tests with memory profiling..."
	go test $(TEST_DIRS) -v -bench=. -benchmem

# Specific test targets
.PHONY: test-services
test-services: ## Run service layer tests
	@echo "Running service layer tests..."
	go test ./internal/services -v

.PHONY: test-handlers
test-handlers: ## Run handler tests
	@echo "Running handler tests..."
	go test ./internal/handlers -v

.PHONY: test-middleware
test-middleware: ## Run middleware tests
	@echo "Running middleware tests..."
	go test ./internal/middleware -v

.PHONY: test-models
test-models: ## Run model tests
	@echo "Running model tests..."
	go test ./internal/models -v

.PHONY: test-repositories
test-repositories: ## Run repository tests
	@echo "Running repository tests..."
	go test ./internal/repositories -v

.PHONY: test-pkg
test-pkg: ## Run package tests
	@echo "Running package tests..."
	go test ./pkg/... -v

# Linting and formatting
.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

.PHONY: fmt-check
fmt-check: ## Check code formatting
	@echo "Checking code formatting..."
	@if [ "$$(gofmt -l . | wc -l)" -gt 0 ]; then \
		echo "Code is not formatted. Run 'make fmt' to format."; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "Code is properly formatted."; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "Tidying go modules..."
	go mod tidy
	go mod verify

# Database targets
.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@if [ -f "./gscope.db" ]; then \
		echo "Database exists, running migrations..."; \
	else \
		echo "No database found. Starting fresh..."; \
	fi
	@go run $(MAIN_PATH) migrate

.PHONY: db-reset
db-reset: ## Reset database (delete and recreate)
	@echo "Resetting database..."
	@rm -f ./gscope.db
	@make db-migrate

# Development targets
.PHONY: run
run: ## Run the application
	@echo "Running GScope..."
	go run $(MAIN_PATH)

.PHONY: run-dev
run-dev: ## Run in development mode
	@echo "Running GScope in development mode..."
	GIN_MODE=debug go run $(MAIN_PATH)

.PHONY: dev
dev: fmt mod-tidy test run-dev ## Development workflow (format, tidy, test, run)

# Dependencies
.PHONY: deps
deps: ## Install dependencies
	@echo "Installing dependencies..."
	go mod download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Clean targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@go clean -cache -testcache

.PHONY: clean-db
clean-db: ## Clean database
	@echo "Cleaning database..."
	@rm -f ./gscope.db

.PHONY: clean-all
clean-all: clean clean-db ## Clean everything

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t gscope .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 gscope

.PHONY: docker-clean
docker-clean: ## Clean Docker images
	@echo "Cleaning Docker images..."
	docker rmi gscope 2>/dev/null || true

# Security targets
.PHONY: security-scan
security-scan: ## Run security scan
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
		exit 1; \
	fi

# Documentation targets
.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060; \
	else \
		echo "godoc not found. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
		exit 1; \
	fi

# Release targets
.PHONY: release
release: clean build-all ## Create release builds
	@echo "Creating release builds..."
	@mkdir -p release
	@cp $(BUILD_DIR)/* release/
	@echo "Release builds created in release/ directory"

.PHONY: release-tag
release-tag: ## Create a new release tag
	@echo "Creating release tag..."
	@read -p "Enter version (e.g., v1.0.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	git push origin $$version

# Monitoring and profiling
.PHONY: profile-cpu
profile-cpu: ## Generate CPU profile
	@echo "Generating CPU profile..."
	go test -cpuprofile=cpu.prof -bench=. ./internal/services

.PHONY: profile-mem
profile-mem: ## Generate memory profile
	@echo "Generating memory profile..."
	go test -memprofile=mem.prof -bench=. ./internal/services

.PHONY: profile-web
profile-web: ## View profiles in web interface
	@echo "Starting profile web interface..."
	go tool pprof -http=:8081 cpu.prof

# Git hooks
.PHONY: pre-commit
pre-commit: fmt-check vet test ## Pre-commit checks
	@echo "Pre-commit checks passed!"

# CI/CD targets
.PHONY: ci
ci: deps fmt-check vet lint test-coverage ## CI pipeline
	@echo "CI pipeline completed successfully!"

.PHONY: ci-full
ci-full: ci security-scan ## Full CI pipeline with security scan
	@echo "Full CI pipeline completed successfully!"

# Utility targets
.PHONY: version
version: ## Show version information
	@echo "GScope Version Information:"
	@echo "  Go version: $(shell go version)"
	@echo "  Git commit: $(shell git describe --tags --always --dirty)"
	@echo "  Build time: $(shell date)"

.PHONY: check
check: ## Check if all tools are available
	@echo "Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "Go is not installed"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "Git is not installed"; exit 1; }
	@echo "All required tools are available"

.PHONY: setup
setup: check deps ## Setup development environment
	@echo "Development environment setup complete!"

# Default target for quick development
.PHONY: quick
quick: fmt test run ## Quick development cycle (format, test, run) 