# MUNO - Multi-repository UNified Orchestration
.PHONY: build build-dev build-local clean test install install-dev install-local uninstall-dev uninstall-local lint release status help

# Variables
BINARY_NAME := muno
BINARY_NAME_DEV := muno-dev
BINARY_NAME_LOCAL := muno-local
BUILD_DIR := bin
CMD_DIR := cmd/muno
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Extract GitHub owner and repo from git remote origin URL
# Works with both HTTPS and SSH URLs
GIT_REMOTE_URL := $(shell git remote get-url origin 2>/dev/null || echo "")
GITHUB_OWNER := $(shell echo $(GIT_REMOTE_URL) | sed -E 's|.*github\.com[:/]([^/]+)/.*|\1|' || echo "taokim")
GITHUB_REPO := $(shell echo $(GIT_REMOTE_URL) | sed -E 's|.*github\.com[:/][^/]+/([^/.]+)(\.git)?.*|\1|' || echo "muno")

LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.gitBranch=$(GIT_BRANCH) -X main.buildTime=$(BUILD_TIME) -X main.GitHubOwner=$(GITHUB_OWNER) -X main.GitHubRepo=$(GITHUB_REPO)"

# Default target
all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## build-dev: Build development binary with -dev suffix
build-dev:
	@echo "Building $(BINARY_NAME_DEV)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_DEV) ./$(CMD_DIR)
	@echo "Development binary built: $(BUILD_DIR)/$(BINARY_NAME_DEV)"

## build-local: Build local binary with -local suffix
build-local:
	@echo "Building $(BINARY_NAME_LOCAL)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME_LOCAL) ./$(CMD_DIR)
	@echo "Local binary built: $(BUILD_DIR)/$(BINARY_NAME_LOCAL)"

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./internal/...
	@echo "Tests complete"

## test-all: Run all tests including integration and regression
test-all: test-master

## test-basic: Run basic regression tests
test-basic:
	@echo "Running basic regression tests..."
	@./test/regression/regression_test.sh

## test-extended: Run extended regression tests  
test-extended:
	@echo "Running extended regression tests..."
	@./test/regression/extended_regression_test.sh

## test-master: Run complete test suite
test-master:
	@echo "Running complete test suite..."
	@./test/regression/master_test.sh

## test-go: Run Go unit tests
test-go:
	@echo "Running Go unit tests..."
	@go test ./... -cover

## test-coverage: Generate test coverage report
test-coverage:
	@echo "Generating coverage report..."
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## validate: Quick validation before commit
validate: build test-basic test-go
	@echo "✅ Validation complete - ready to commit"

## release-check: Full validation before release
release-check: clean build test-master test-coverage
	@echo "🚀 Release validation complete"
	@echo "Check TEST_SUITE_DOCUMENTATION.md for detailed results"

## install: Install production version
install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install ./$(CMD_DIR)
	@echo "Installed as '$(BINARY_NAME)'"

## install-dev: Install development version
install-dev: build-dev
	@echo "Installing $(BINARY_NAME_DEV)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME_DEV) $(GOPATH)/bin/$(BINARY_NAME_DEV) 2>/dev/null || cp $(BUILD_DIR)/$(BINARY_NAME_DEV) ~/go/bin/$(BINARY_NAME_DEV)
	@chmod +x $(GOPATH)/bin/$(BINARY_NAME_DEV) 2>/dev/null || chmod +x ~/go/bin/$(BINARY_NAME_DEV)
	@echo "Installed as '$(BINARY_NAME_DEV)'"

## install-local: Install local test version
install-local: build-local
	@echo "Installing $(BINARY_NAME_LOCAL)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME_LOCAL) $(GOPATH)/bin/$(BINARY_NAME_LOCAL) 2>/dev/null || cp $(BUILD_DIR)/$(BINARY_NAME_LOCAL) ~/go/bin/$(BINARY_NAME_LOCAL)
	@chmod +x $(GOPATH)/bin/$(BINARY_NAME_LOCAL) 2>/dev/null || chmod +x ~/go/bin/$(BINARY_NAME_LOCAL)
	@echo "Installed as '$(BINARY_NAME_LOCAL)'"

## uninstall-dev: Remove development version
uninstall-dev:
	@echo "Removing $(BINARY_NAME_DEV)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME_DEV) 2>/dev/null || rm -f ~/go/bin/$(BINARY_NAME_DEV)
	@rm -f /usr/local/bin/$(BINARY_NAME_DEV)
	@echo "$(BINARY_NAME_DEV) removed"

## uninstall-local: Remove local version  
uninstall-local:
	@echo "Removing $(BINARY_NAME_LOCAL)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME_LOCAL) 2>/dev/null || rm -f ~/go/bin/$(BINARY_NAME_LOCAL)
	@rm -f /usr/local/bin/$(BINARY_NAME_LOCAL)
	@echo "$(BINARY_NAME_LOCAL) removed"

## lint: Run linters
lint:
	@echo "Running linters..."
	@golangci-lint run ./... || true
	@go vet ./...
	@echo "Linting complete"

## release: Build release binaries for multiple platforms
release:
	@echo "Building release binaries..."
	@mkdir -p $(BUILD_DIR)/release
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/release/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "Release binaries built in $(BUILD_DIR)/release"

## run: Build and run
run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

## status: Show installation status of all versions
status:
	@./scripts/muno-versions.sh status

## help: Show this help message
help:
	@echo "MUNO Build System"
	@echo ""
	@echo "Installation targets:"
	@echo "  make install        - Install production version (muno)"
	@echo "  make install-dev    - Install development version (muno-dev)"
	@echo "  make install-local  - Install local test version (muno-local)"
	@echo "  make status         - Show all installed versions"
	@echo ""
	@echo "Build targets:"
	@echo "  make build          - Build production binary"
	@echo "  make build-dev      - Build development binary"
	@echo "  make build-local    - Build local test binary"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Test targets:"
	@echo "  make test           - Run Go unit tests"
	@echo "  make test-basic     - Run basic regression tests (36 tests)"
	@echo "  make test-extended  - Run extended regression tests (120 tests)"
	@echo "  make test-master    - Run complete test suite (150+ tests)"
	@echo "  make test-all       - Alias for test-master"
	@echo "  make test-go        - Run Go unit tests with coverage"
	@echo "  make test-coverage  - Generate HTML coverage report"
	@echo "  make validate       - Quick validation (build + basic tests)"
	@echo "  make release-check  - Full release validation"
	@echo ""
	@echo "Other targets:"
	@echo "  make lint           - Run linters"
	@echo "  make release        - Build release binaries for all platforms"

.DEFAULT_GOAL := help