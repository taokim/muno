# MUNO - Multi-repository UNified Orchestration
.PHONY: build clean test install lint release help

# Variables
BINARY_NAME := muno
BUILD_DIR := bin
CMD_DIR := cmd/muno
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.gitBranch=$(GIT_BRANCH) -X main.buildTime=$(BUILD_TIME)"

# Default target
all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

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

## test-all: Run all tests including integration
test-all:
	@echo "Running all tests..."
	@go test -v ./...
	@echo "All tests complete"

## install: Install the binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install ./$(CMD_DIR)
	@echo "Installation complete"

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

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^##' Makefile | sed 's/## /  /'

.DEFAULT_GOAL := help