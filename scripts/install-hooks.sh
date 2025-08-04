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

# Check if there are any frontend files being committed
FRONTEND_CHANGED=false
if git diff --cached --name-only | grep -qE '\.(ts|tsx|js|jsx)$' | grep -qE '(cmd/gopca-desktop/frontend|cmd/gocsv/frontend|packages/ui-components)'; then
    FRONTEND_CHANGED=true
fi

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

# Check frontend if TypeScript/JavaScript files are being committed
if [ "$FRONTEND_CHANGED" = true ]; then
    echo ""
    echo "→ Checking frontend code..."
    
    # Function to check a frontend directory
    check_frontend() {
        local dir=$1
        local name=$2
        
        if [ -d "$dir" ] && [ -f "$dir/package.json" ]; then
            echo "  Checking $name..."
            
            # Check if node_modules exists (could be local or at workspace root)
            if [ ! -d "$dir/node_modules" ] && [ ! -d "./node_modules" ]; then
                echo "❌ $name: node_modules not found. Run 'npm install' from project root"
                return 1
            fi
            
            # Run TypeScript compilation check
            if [ -f "$dir/tsconfig.json" ]; then
                # Use tsconfig.check.json if it exists for type checking only
                if [ -f "$dir/tsconfig.check.json" ]; then
                    if ! (cd "$dir" && npx tsc --project tsconfig.check.json 2>&1); then
                        echo "❌ $name: TypeScript compilation failed"
                        echo "Run 'npm run build' in $dir to see detailed errors."
                        return 1
                    fi
                else
                    if ! (cd "$dir" && npx tsc --noEmit 2>&1); then
                        echo "❌ $name: TypeScript compilation failed"
                        echo "Run 'npm run build' in $dir to see detailed errors."
                        return 1
                    fi
                fi
                echo "  ✓ $name TypeScript OK"
            fi
            
            # Run ESLint if configured
            if [ -f "$dir/.eslintrc.json" ] || [ -f "$dir/.eslintrc.js" ]; then
                if ! (cd "$dir" && npx eslint src --ext .ts,.tsx,.js,.jsx --max-warnings 0 2>&1); then
                    echo "❌ $name: ESLint found issues"
                    return 1
                fi
                echo "  ✓ $name ESLint OK"
            fi
        fi
        
        return 0
    }
    
    # Check each frontend project
    FRONTEND_OK=true
    
    # Check shared UI components first
    if git diff --cached --name-only | grep -q 'packages/ui-components'; then
        if ! check_frontend "packages/ui-components" "UI Components"; then
            FRONTEND_OK=false
        fi
    fi
    
    # Check GoPCA Desktop frontend
    if git diff --cached --name-only | grep -q 'cmd/gopca-desktop/frontend'; then
        if ! check_frontend "cmd/gopca-desktop/frontend" "GoPCA Desktop"; then
            FRONTEND_OK=false
        fi
    fi
    
    # Check GoCSV frontend
    if git diff --cached --name-only | grep -q 'cmd/gocsv/frontend'; then
        if ! check_frontend "cmd/gocsv/frontend" "GoCSV"; then
            FRONTEND_OK=false
        fi
    fi
    
    if [ "$FRONTEND_OK" = false ]; then
        echo ""
        echo "Frontend checks failed. Please fix the issues before committing."
        exit 1
    fi
    
    echo "✓ Frontend checks passed"
fi

echo ""
echo "✅ All pre-commit checks passed!"
EOF

# Make the hook executable
chmod +x "$PRE_COMMIT_HOOK"

echo "✅ Git hooks installed successfully!"
echo ""
echo "The pre-commit hook will run before each commit to ensure:"
echo "  • Go code is properly formatted (go fmt)"
echo "  • Go code passes static analysis (go vet)"
echo "  • Go tests pass"
echo "  • go.mod is tidy"
echo "  • TypeScript compiles without errors"
echo "  • ESLint checks pass (if configured)"
echo ""
echo "To skip the hook temporarily, use: git commit --no-verify"
echo "To uninstall, run: rm $PRE_COMMIT_HOOK"