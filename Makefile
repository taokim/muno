# Makefile for repo-claude

# Variables
BINARY_NAME=rc
# For local builds, use YYMMDDHHMM format. For tagged releases, use git tag
VERSION=$(shell git describe --tags --exact-match 2>/dev/null || date '+%y%m%d%H%M')
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"
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
	@cp $(GOBIN)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

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
	@echo "  make build-all   - Build for all platforms"
	@echo "  make release-dry - Dry run release"
	@echo "  make release     - Create release"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make deps        - Download dependencies"
	@echo "  make update-deps - Update dependencies"
	@echo "  make help        - Show this help"

.DEFAULT_GOAL := help