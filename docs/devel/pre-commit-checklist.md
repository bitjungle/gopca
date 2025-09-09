# Pre-Commit Checklist for GoPCA Development

## Purpose
This checklist helps prevent CI/CD failures by ensuring code is tested locally before pushing. Following this checklist will save time and reduce frustration from failed CI runs.

## Essential Pre-Commit Checks

### 1. Code Quality Checks
```bash
# Format all Go code
go fmt ./...

# Run static analysis
go vet ./...

# Check module dependencies
go mod tidy
```

### 2. Run Core Tests
```bash
# Run all tests with race detection
go test -race ./...

# Run specific test suites that often fail in CI
go test -v -timeout 10m ./internal/core -run "TestValidate|TestNIPALS|TestMath"
```

### 3. Sklearn Validation Tests (if modified)
```bash
# Generate reference data first
cd testdata && source .venv/bin/activate
cd validation
python generate_reference_pca.py
python generate_kernel_pca_reference.py
python generate_temporal_pca_reference.py
cd ../.. && deactivate

# Run validation tests
go test -v -run TestValidateAgainstSklearn ./internal/core/

# Clean up generated files
rm -rf testdata/validation/reference_results
```

### 4. Frontend Checks (if modified)
```bash
# Build shared UI components
npm run build-ui

# Check TypeScript compilation
cd cmd/gopca-desktop/frontend && npm run build
cd cmd/gocsv/frontend && npm run build
```

### 5. Wails Application Tests
```bash
# Create minimal dist for embed directive
mkdir -p cmd/gopca-desktop/frontend/dist
echo '<!DOCTYPE html><html><body>Test</body></html>' > cmd/gopca-desktop/frontend/dist/index.html
mkdir -p cmd/gocsv/frontend/dist
echo '<!DOCTYPE html><html><body>Test</body></html>' > cmd/gocsv/frontend/dist/index.html

# Run Wails app tests
go test -v ./cmd/gopca-desktop/...
go test -v ./cmd/gocsv/...
```

### 6. Integration Tests
```bash
# Build CLI first
make build

# Run integration tests
make test-e2e
```

### 7. Platform-Specific Considerations

#### Windows
- Test with both forward and backward slashes in paths
- Check for case sensitivity issues
- Verify PowerShell vs Bash script compatibility

#### macOS
- Test on both Intel and Apple Silicon if possible
- Verify code signing doesn't break functionality

#### Linux
- Test with different shell environments
- Check for missing system dependencies

## Quick Pre-Push Command

For a comprehensive check before pushing:
```bash
# Run this script before pushing
#!/bin/bash
set -e  # Exit on first error

echo "Running pre-push checks..."

# 1. Code formatting
echo "Checking code format..."
if [ -n "$(gofmt -l .)" ]; then
    echo "Error: Code needs formatting. Run: go fmt ./..."
    exit 1
fi

# 2. Static analysis
echo "Running go vet..."
go vet ./...

# 3. Module tidiness
echo "Checking go.mod..."
go mod tidy
if [ -n "$(git status --porcelain go.mod go.sum)" ]; then
    echo "Error: go.mod or go.sum modified. Commit these changes."
    exit 1
fi

# 4. Core tests
echo "Running core tests..."
go test -race -timeout 10m ./internal/core ./internal/cli ./pkg/...

# 5. If sklearn validation was modified
if git diff --cached --name-only | grep -q "sklearn.*\.go\|validation.*\.py"; then
    echo "Running sklearn validation tests..."
    cd testdata && source .venv/bin/activate
    cd validation
    python generate_reference_pca.py
    cd ../..
    go test -v -run TestValidateAgainstSklearn ./internal/core/
    rm -rf validation/reference_results
    deactivate
fi

echo "✅ All pre-push checks passed!"
```

## Common CI Failure Patterns

### 1. "Cannot open: File exists"
**Cause**: Incorrect test patterns or missing test files
**Solution**: Verify test function names match patterns in CI scripts

### 2. Python Script Failures
**Cause**: Missing directories or dependencies
**Solution**: Ensure scripts create output directories with `os.makedirs(dir, exist_ok=True)`

### 3. Test Pattern Mismatches
**Cause**: CI test patterns don't match actual test function names
**Solution**: Check both `.github/workflows/build.yml` and `scripts/ci/*.sh` for patterns

### 4. Reference File Issues
**Cause**: Reference files in .gitignore must be generated fresh
**Solution**: Ensure Python scripts run successfully in CI before Go tests

### 5. Silent Failures
**Cause**: `continue-on-error: true` hides real problems
**Solution**: Remove continue-on-error to expose failures immediately

## Emergency Recovery

If CI is failing after merge:
1. Don't panic or repeatedly re-run CI
2. Check the actual error messages in GitHub Actions logs
3. Reproduce the exact CI command locally
4. Fix all issues in one comprehensive commit
5. Test the fix locally with the exact CI commands
6. Push once with confidence

## Git Hooks Setup

The repository includes automatic pre-commit hooks:
```bash
# Install hooks (required for all developers)
make install-hooks

# This automatically runs:
# - go fmt
# - go vet  
# - go test -race (selected packages)
# - TypeScript compilation check
```

## Remember

- **Test locally first** - CI time is valuable
- **Read error messages carefully** - They usually point to the exact problem
- **Check related files** - Problems often cascade
- **Use verbose flags** - More information helps debugging
- **Clean test environment** - Remove generated files between test runs
- **One fix, one push** - Avoid trial-and-error in CI

## Related Documentation

- [CI/CD Workflow](./ci-cd-workflow.md)
- [Git Workflow](./git-workflow-simple.md)
- [Testing Strategy](./testing-strategy.md)
- [Sklearn Validation](./validation-methodology.md)