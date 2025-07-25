#!/bin/bash
#
# install-hooks.sh - Install Git hooks for GoPCA development
#
# This script installs a pre-commit hook that runs formatting and tests
# before allowing commits.

set -e

HOOKS_DIR=".git/hooks"
PRE_COMMIT_HOOK="$HOOKS_DIR/pre-commit"

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Create pre-commit hook
cat > "$PRE_COMMIT_HOOK" << 'EOF'
#!/bin/bash
#
# GoPCA pre-commit hook
# Runs formatting, linting, and tests before commit

set -e

echo "Running pre-commit checks..."

# Check if there are any Go files being committed
if git diff --cached --name-only | grep -q '\.go$'; then
    echo "→ Running go fmt..."
    UNFORMATTED=$(gofmt -l internal pkg cmd/gopca-cli 2>/dev/null | grep -v vendor || true)
    if [ -n "$UNFORMATTED" ]; then
        echo "❌ The following files need formatting:"
        echo "$UNFORMATTED"
        echo ""
        echo "Run 'make fmt' to fix formatting."
        exit 1
    fi
    echo "✓ Code formatting OK"

    echo ""
    echo "→ Running go vet..."
    if ! go vet ./internal/... ./pkg/... ./cmd/gopca-cli/... 2>&1; then
        echo "❌ go vet found issues"
        exit 1
    fi
    echo "✓ go vet OK"

    echo ""
    echo "→ Running tests..."
    if ! go test -race -timeout 5m ./internal/... ./pkg/... ./cmd/gopca-cli/... >/dev/null 2>&1; then
        echo "❌ Tests failed"
        echo "Run 'make test' to see detailed test output."
        exit 1
    fi
    echo "✓ Tests passed"

    # Check for common issues
    echo ""
    echo "→ Checking for common issues..."
    
    # Check for fmt.Println in non-test files
    PRINTLN_FILES=$(git diff --cached --name-only | grep '\.go$' | grep -v '_test\.go$' | xargs grep -l 'fmt\.Println' 2>/dev/null || true)
    if [ -n "$PRINTLN_FILES" ]; then
        echo "⚠️  Warning: fmt.Println found in non-test files:"
        echo "$PRINTLN_FILES"
        echo "Consider using proper logging instead."
    fi

    # Check for TODO/FIXME in new code
    NEW_TODOS=$(git diff --cached | grep '^+' | grep -E 'TODO|FIXME' || true)
    if [ -n "$NEW_TODOS" ]; then
        echo "⚠️  Warning: New TODO/FIXME comments added:"
        echo "$NEW_TODOS"
        echo "Consider creating GitHub issues for these items."
    fi
fi

# Check go.mod if it's being committed
if git diff --cached --name-only | grep -q 'go\.mod\|go\.sum'; then
    echo ""
    echo "→ Checking go.mod tidiness..."
    go mod tidy
    if ! git diff --exit-code go.mod go.sum >/dev/null 2>&1; then
        echo "❌ go.mod or go.sum is not tidy"
        echo "Run 'go mod tidy' and stage the changes."
        exit 1
    fi
    echo "✓ go.mod is tidy"
fi

echo ""
echo "✅ All pre-commit checks passed!"
EOF

# Make the hook executable
chmod +x "$PRE_COMMIT_HOOK"

echo "✅ Git hooks installed successfully!"
echo ""
echo "The pre-commit hook will run before each commit to ensure:"
echo "  • Code is properly formatted (go fmt)"
echo "  • Code passes static analysis (go vet)"
echo "  • All tests pass"
echo "  • go.mod is tidy"
echo ""
echo "To skip the hook temporarily, use: git commit --no-verify"
echo "To uninstall, run: rm $PRE_COMMIT_HOOK"