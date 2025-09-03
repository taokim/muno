#!/bin/bash

echo "Running test coverage for all packages..."
echo "========================================="

# Run tests with coverage
go test -coverprofile=coverage.out ./... 2>/dev/null

# Get coverage by package
echo -e "\nCoverage by package:"
go test -cover ./... 2>&1 | grep "coverage:" | while read line; do
    pkg=$(echo "$line" | awk '{print $2}')
    coverage=$(echo "$line" | grep -oE '[0-9]+\.[0-9]%')
    if [ ! -z "$coverage" ]; then
        printf "%-50s %s\n" "$pkg" "$coverage"
    fi
done

# Get total coverage
echo -e "\n========================================="
go tool cover -func=coverage.out 2>/dev/null | grep "^total:" | awk '{print "Total Coverage: " $3}'
echo "========================================="

# Clean threshold indicator
total=$(go tool cover -func=coverage.out 2>/dev/null | grep "^total:" | awk '{print $3}' | sed 's/%//')
if [ ! -z "$total" ]; then
    if (( $(echo "$total >= 70" | bc -l) )); then
        echo "✅ Coverage is above 70% threshold!"
    else
        echo "⚠️  Coverage is below 70% threshold"
    fi
fi