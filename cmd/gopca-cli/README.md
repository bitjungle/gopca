# GoPCA CLI

Command-line interface for PCA analysis. Fast, scriptable, and suitable for automation.

## Architecture

- **Entry Point**: `main.go` - CLI initialization and command routing
- **Commands**: Uses `internal/cli` package for all command implementations
- **Core**: Reuses PCA engine from `internal/core`
- **Framework**: urfave/cli v2 for command parsing

## Development

```bash
# From repository root
make build           # Build for current platform
make build-all       # Build for all platforms
make build-cross     # Cross-compile for specific platforms

# Or directly with Go
cd cmd/gopca-cli
go build -o gopca-cli
```

## Key Files

- `main.go` - CLI entry point, version injection via ldflags
- `internal/cli/analyze.go` - Main PCA analysis command
- `internal/cli/validate.go` - Data validation command
- `internal/cli/transform.go` - Data transformation utilities
- `internal/cli/output.go` - Result formatting (table, JSON)

## Commands

- `analyze` - Perform PCA analysis with various methods (SVD, NIPALS, Kernel)
- `validate` - Check CSV data for PCA compatibility
- `transform` - Apply preprocessing transformations
- `version` - Display version information

## Features

- Multiple PCA algorithms (SVD, NIPALS, Kernel PCA)
- Flexible preprocessing pipeline (SNV, scaling, centering)
- Missing value handling strategies
- Eigencorrelation analysis
- Multiple output formats (table, JSON)
- Row/column exclusion for data filtering

## Testing

```bash
cd internal/cli
go test -v           # CLI package tests
```

## Build Output

```
build/
├── gopca-cli                    # Native platform build
├── gopca-cli-darwin-arm64       # macOS ARM64
├── gopca-cli-linux-amd64        # Linux x64
└── gopca-cli-windows-amd64.exe  # Windows x64
```

## Version Management

Version is injected at build time via ldflags:
```bash
go build -ldflags "-X github.com/bitjungle/gopca/internal/version.Version=v1.0.0"
```