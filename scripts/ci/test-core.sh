#!/bin/bash
#
# test-core.sh - Run tests for core GoPCA packages
#
# This script runs tests excluding the desktop package which requires
# frontend build artifacts that may not be available during testing.

set -e

echo "=== Running Core GoPCA Tests ==="

# Show Go version
echo "Go version: $(go version)"

# Download dependencies if needed
echo "Downloading dependencies..."
go mod download

# Run tests with proper exclusions
echo "Running tests (excluding desktop package)..."

# Method 1: Explicitly list packages to test
# This is the most reliable method as it doesn't require Go to parse the desktop package
if go test -v -cover ./internal/... ./pkg/... ./cmd/gopca-cli/...; then
    echo "✓ All core tests passed"
else
    echo "✗ Some tests failed"
    exit 1
fi

# Show coverage summary
echo ""
echo "=== Coverage Summary ==="
go test -cover ./internal/... ./pkg/... ./cmd/gopca-cli/... 2>/dev/null | grep -E "coverage:|ok" || true

echo ""
echo "=== Core tests completed successfully ===" 