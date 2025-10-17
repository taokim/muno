# MUNO Test Suite Makefile
# Convenient targets for running different test suites

.PHONY: all test test-unit test-integration test-e2e test-regression test-all test-coverage clean help

# Default target
all: test-all

# Run all Go tests (unit + integration)
test:
	@echo "Running all Go tests..."
	@go test ./... -timeout 60s

# Run unit tests only (with short flag)
test-unit:
	@echo "Running unit tests..."
	@go test ./... -short -timeout 30s

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test ./test -run TestIntegration -timeout 60s
	@go test ./test -run TestClonePull -timeout 60s

# Run E2E tests
test-e2e:
	@echo "Running E2E tests..."
	@./run_e2e_tests.sh

# Run regression tests
test-regression:
	@echo "Running regression tests..."
	@./run_regression_tests.sh

# Run regression tests in quick mode
test-regression-quick:
	@echo "Running quick regression tests..."
	@./run_regression_tests.sh --quick

# Run all test suites
test-all: test-unit test-integration test-e2e test-regression-quick
	@echo "✓ All test suites completed"

# Generate test coverage report
test-coverage:
	@echo "Generating test coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v ./... -timeout 60s

# Run tests in parallel
test-parallel:
	@echo "Running tests in parallel..."
	@go test -parallel 4 ./... -timeout 60s

# Run specific test by name (usage: make test-run TEST=TestCloneRepos)
test-run:
	@echo "Running test: $(TEST)"
	@go test -v -run $(TEST) ./...

# Benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	@go test -bench=. -benchmem ./...

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	@rm -f coverage.out coverage.html
	@rm -rf /tmp/muno-test-*
	@rm -rf /tmp/muno-regression-*
	@rm -rf /tmp/muno-e2e-*
	@echo "✓ Test artifacts cleaned"

# Help target
help:
	@echo "MUNO Test Suite Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  make test              - Run all Go tests"
	@echo "  make test-unit         - Run unit tests only"
	@echo "  make test-integration  - Run integration tests"
	@echo "  make test-e2e          - Run end-to-end tests"
	@echo "  make test-regression   - Run full regression tests"
	@echo "  make test-regression-quick - Run quick regression tests"
	@echo "  make test-all          - Run all test suites"
	@echo "  make test-coverage     - Generate coverage report"
	@echo "  make test-verbose      - Run tests with verbose output"
	@echo "  make test-parallel     - Run tests in parallel"
	@echo "  make test-run TEST=name - Run specific test by name"
	@echo "  make test-bench        - Run benchmark tests"
	@echo "  make clean             - Clean test artifacts"
	@echo "  make help              - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make test-run TEST=TestCloneRepos"
	@echo "  make test-regression --quick"