# Contributing to GoPCA

Thank you for your interest in contributing to GoPCA! This guide will help you get started with development.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Node.js 18 or later (for GUI development)
- Git

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/bitjungle/gopca.git
   cd gopca
   ```

2. Install dependencies:
   ```bash
   make deps       # Go dependencies
   make gui-deps   # GUI dependencies (if working on desktop app)
   ```

3. **Install Git hooks** (IMPORTANT):
   ```bash
   make install-hooks
   ```
   
   This installs a pre-commit hook that automatically:
   - Formats your code (`go fmt`)
   - Runs static analysis (`go vet`)
   - Runs all tests
   - Ensures `go.mod` is tidy

## Development Workflow

### Before Committing

The pre-commit hook will automatically run checks, but you can run them manually:

```bash
make fmt    # Format code
make lint   # Run linter (if golangci-lint is installed)
make test   # Run tests
```

### Building

```bash
make build      # Build CLI
make gui-build  # Build desktop app
make build-all  # Build CLI for all platforms
```

### Testing

```bash
make test           # Run all tests
make test-verbose   # Run tests with detailed output
make test-coverage  # Generate coverage report
```

### Creating a Pull Request

1. Create a feature branch:
   ```bash
   git checkout -b 42-feature-description
   ```

2. Make your changes and commit:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```
   
   The pre-commit hook will ensure your code is properly formatted and tested.

3. Push to your fork and create a PR:
   ```bash
   git push origin 42-feature-description
   ```

## Code Standards

### Commit Messages

Follow conventional commits format:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Test additions or fixes
- `refactor:` Code refactoring
- `chore:` Maintenance tasks

### Code Style

- Follow Go conventions and idioms
- Use meaningful variable names
- Add comments for complex logic
- Keep functions focused and small

### Testing

- Write tests for all new functionality
- Maintain test coverage above 70%
- Use table-driven tests where appropriate
- Test edge cases and error conditions

## Pre-commit Hook Details

The installed pre-commit hook checks:

1. **Code Formatting**: Ensures all Go code is properly formatted
2. **Static Analysis**: Runs `go vet` to catch common issues
3. **Tests**: Runs all tests to ensure nothing is broken
4. **Dependencies**: Ensures `go.mod` is tidy when modified

To skip the hook temporarily (not recommended):
```bash
git commit --no-verify
```

## Troubleshooting

### Pre-commit Hook Issues

If the pre-commit hook is causing issues:

1. Check the hook is installed:
   ```bash
   ls -la .git/hooks/pre-commit
   ```

2. Run checks manually to see detailed output:
   ```bash
   make fmt
   make test
   go mod tidy
   ```

3. To uninstall the hook:
   ```bash
   rm .git/hooks/pre-commit
   ```

### Build Issues

- Ensure Go version is 1.21+: `go version`
- Clear module cache: `go clean -modcache`
- Re-download dependencies: `make deps`

## Need Help?

- Check existing issues on GitHub
- Read the developer guide in `CLAUDE.md`
- Create a new issue with detailed information

Happy coding!