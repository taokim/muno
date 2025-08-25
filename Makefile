# Makefile for repo-claude

# Variables
BINARY_NAME=rc

# Version determination:
# 1. If on a tagged commit, use the tag (e.g., v0.4.0)
# 2. Otherwise, use git describe with commit info (e.g., v0.4.0-5-gabcd123)
# 3. Add -dirty suffix if there are uncommitted changes
# 4. Fallback to "dev" if not a git repo
GIT_TAG=$(shell git describe --tags --exact-match 2>/dev/null)
GIT_DESCRIBE=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
VERSION=$(or $(GIT_TAG),$(GIT_DESCRIBE))

# Additional build metadata
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Build flags include version, commit, branch, and build time
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT} -X main.gitBranch=${GIT_BRANCH}"
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(shell find . -name "*.go" -type f -not -path "./vendor/*")

# Build targets
.PHONY: all build clean test coverage lint fmt vet install release-dry release

all: clean test build

build:
	@echo "Building ${BINARY_NAME}..."
	@mkdir -p $(GOBIN)
	@go build $(LDFLAGS) -o $(GOBIN)/$(BINARY_NAME) ./cmd/repo-claude
	@echo "Build complete: $(GOBIN)/$(BINARY_NAME)"

clean:
	@echo "Cleaning..."
	@rm -rf $(GOBIN)
	@rm -rf dist/
	@go clean -cache -testcache
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

test-short:
	@echo "Running short tests..."
	@go test -v -short ./...
	@echo "Short tests complete"

coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Running linter..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@golangci-lint run ./...
	@echo "Lint complete"

fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete"

install: build
	@echo "Installing ${BINARY_NAME}..."
	@if [ -z "$(GOPATH)" ]; then \
		echo "GOPATH not set, installing to /usr/local/bin"; \
		sudo cp $(GOBIN)/$(BINARY_NAME) /usr/local/bin/ || cp $(GOBIN)/$(BINARY_NAME) ~/bin/; \
	else \
		mkdir -p $(GOPATH)/bin; \
		cp $(GOBIN)/$(BINARY_NAME) $(GOPATH)/bin/; \
		echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"; \
	fi

install-local: build
	@echo "Installing $(BINARY_NAME) to local user directory..."
	@LOCAL_BIN_DIR=""; \
	for dir in "$$HOME/.local/bin" "$$HOME/bin" "$$HOME/.bin"; do \
		shortdir=$$(echo "$$dir" | sed "s|$$HOME|~|"); \
		if echo "$$PATH" | grep -q "$$dir" || echo "$$PATH" | grep -q "$$shortdir"; then \
			LOCAL_BIN_DIR="$$dir"; \
			break; \
		fi; \
	done; \
	if [ -z "$$LOCAL_BIN_DIR" ]; then \
		echo "❌ No local bin directory found in PATH."; \
		echo ""; \
		echo "Please choose one of these options:"; \
		echo "1. Add ~/.local/bin to PATH (recommended):"; \
		echo "   mkdir -p ~/.local/bin"; \
		echo "   echo 'export PATH=\"$$HOME/.local/bin:$$PATH\"' >> ~/.zshrc"; \
		echo "   source ~/.zshrc"; \
		echo ""; \
		echo "2. Add ~/bin to PATH:"; \
		echo "   mkdir -p ~/bin"; \
		echo "   echo 'export PATH=\"$$HOME/bin:$$PATH\"' >> ~/.zshrc"; \
		echo "   source ~/.zshrc"; \
		echo ""; \
		echo "Then run 'make install-local' again."; \
		exit 1; \
	else \
		mkdir -p "$$LOCAL_BIN_DIR"; \
		cp $(GOBIN)/$(BINARY_NAME) "$$LOCAL_BIN_DIR/"; \
		echo "✅ Installed to $$LOCAL_BIN_DIR/$(BINARY_NAME)"; \
	fi

install-dev: build
	@echo "Installing development version as $(BINARY_NAME)-dev..."
	@LOCAL_BIN_DIR=""; \
	for dir in "$$HOME/.local/bin" "$$HOME/bin" "$$HOME/.bin"; do \
		shortdir=$$(echo "$$dir" | sed "s|$$HOME|~|"); \
		if echo "$$PATH" | grep -q "$$dir" || echo "$$PATH" | grep -q "$$shortdir"; then \
			LOCAL_BIN_DIR="$$dir"; \
			break; \
		fi; \
	done; \
	if [ -z "$$LOCAL_BIN_DIR" ]; then \
		echo "❌ No local bin directory found in PATH."; \
		echo ""; \
		echo "Please choose one of these options:"; \
		echo "1. Add ~/.local/bin to PATH (recommended):"; \
		echo "   mkdir -p ~/.local/bin"; \
		echo "   echo 'export PATH=\"$$HOME/.local/bin:$$PATH\"' >> ~/.zshrc"; \
		echo "   source ~/.zshrc"; \
		echo ""; \
		echo "2. Add ~/bin to PATH:"; \
		echo "   mkdir -p ~/bin"; \
		echo "   echo 'export PATH=\"$$HOME/bin:$$PATH\"' >> ~/.zshrc"; \
		echo "   source ~/.zshrc"; \
		echo ""; \
		echo "Then run 'make install-dev' again."; \
		exit 1; \
	else \
		mkdir -p "$$LOCAL_BIN_DIR"; \
		ln -sf $(GOBIN)/$(BINARY_NAME) "$$LOCAL_BIN_DIR/$(BINARY_NAME)-dev"; \
		echo "✅ Installed to $$LOCAL_BIN_DIR/$(BINARY_NAME)-dev"; \
		echo "Use 'rc' for production, 'rc-dev' for development"; \
	fi

dev: build
	@echo "Running development version..."
	@$(GOBIN)/$(BINARY_NAME) $(filter-out $@,$(MAKECMDGOALS))

# Catch all target for dev command arguments
%:
	@:

# Cross-compilation targets
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(GOBIN)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(GOBIN)/$(BINARY_NAME)-darwin-amd64 ./cmd/repo-claude
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(GOBIN)/$(BINARY_NAME)-darwin-arm64 ./cmd/repo-claude
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(GOBIN)/$(BINARY_NAME)-linux-amd64 ./cmd/repo-claude
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(GOBIN)/$(BINARY_NAME)-linux-arm64 ./cmd/repo-claude
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(GOBIN)/$(BINARY_NAME)-windows-amd64.exe ./cmd/repo-claude
	@echo "Multi-platform build complete"

# Release with goreleaser
release-dry:
	@echo "Running release dry-run..."
	@if ! command -v goreleaser &> /dev/null; then \
		echo "Installing goreleaser..."; \
		go install github.com/goreleaser/goreleaser@latest; \
	fi
	@goreleaser release --snapshot --clean
	@echo "Release dry-run complete"

release:
	@echo "Creating release..."
	@if ! command -v goreleaser &> /dev/null; then \
		echo "Installing goreleaser..."; \
		go install github.com/goreleaser/goreleaser@latest; \
	fi
	@goreleaser release --clean
	@echo "Release complete"

# Development helpers
run: build
	@echo "Running ${BINARY_NAME}..."
	@$(GOBIN)/$(BINARY_NAME)

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies complete"

update-deps:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies updated"

# Help target
help:
	@echo "Available targets:"
	@echo "  make build       - Build the binary (rc)"
	@echo "  make test        - Run tests"
	@echo "  make test-short  - Run short tests"
	@echo "  make coverage    - Generate coverage report"
	@echo "  make lint        - Run linter"
	@echo "  make fmt         - Format code"
	@echo "  make vet         - Run go vet"
	@echo "  make install     - Install binary to GOPATH/bin"
	@echo "  make install-dev - Install as rc-dev (for development)"
	@echo "  make dev [args]  - Run development version directly"
	@echo "  make build-all   - Build for all platforms"
	@echo "  make release-dry - Dry run release"
	@echo "  make release     - Create release"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make deps        - Download dependencies"
	@echo "  make update-deps - Update dependencies"
	@echo "  make help        - Show this help"

.DEFAULT_GOAL := help