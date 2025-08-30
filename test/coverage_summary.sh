#!/bin/bash

echo "=== Test Coverage Summary for muno-go ==="
echo ""

# Run tests with coverage
go test -coverprofile=coverage.out ./... 2>&1 | grep -E "(PASS|FAIL|coverage:|ok)" > test_results.txt

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html 2>/dev/null

# Function-level coverage
echo "Function-level coverage for key files:"
echo "----------------------------------------"
go tool cover -func=coverage.out 2>/dev/null | grep -E "(total:|manager\.go|git\.go|config\.go|main\.go|commands\.go|agent\.go|process|init)" | sort -k3 -n

echo ""
echo "Package-level summary:"
echo "---------------------"
cat test_results.txt | grep -E "(ok|FAIL).*coverage:" | sort

echo ""
echo "Overall coverage:"
echo "-----------------"
go tool cover -func=coverage.out 2>/dev/null | grep "total:" | awk '{print "Total: " $3}'

# Count untested functions
echo ""
echo "Functions with 0% coverage:"
echo "--------------------------"
go tool cover -func=coverage.out 2>/dev/null | grep " 0.0%" | wc -l | awk '{print "Count: " $1}'

echo ""
echo "Top uncovered functions:"
echo "-----------------------"
go tool cover -func=coverage.out 2>/dev/null | grep " 0.0%" | head -10

# Clean up
rm -f test_results.txt