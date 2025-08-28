# GitHub Actions Workflows

## Overview

This directory contains the CI/CD workflows for the GoPCA project, optimized for efficiency and cost-effectiveness.

## Workflows

### build.yml - Continuous Integration

**Purpose**: Run tests and validate code quality

**Triggers** (optimized to prevent duplicate runs):
- Push events: Only on direct pushes to `main` branch
- Pull requests: All PRs targeting `main`
- Manual: Via workflow_dispatch
- Ignores: Documentation changes (*.md, docs/*)

**Actions**:
- Runs tests on multiple platforms (Linux, macOS, Windows)
- Validates code compilation
- Checks for race conditions
- Caches dependencies for faster builds

**Cost Optimization**: 
- No duplicate runs when pushing to feature branches with open PRs
- Path filtering prevents unnecessary runs for documentation changes

### release.yml - Automated Release Creation

**Purpose**: Create production releases with platform-specific bundles

**Triggers**: 
- Push of version tags (e.g., `v0.9.1`)

**Actions**:
1. **Build Phase**: Creates binaries for all platforms
   - pca CLI: macOS (Intel/ARM), Linux (x64/ARM64), Windows (x64)
   - GoPCA Desktop: macOS, Windows, Linux
   - GoCSV Desktop: macOS, Windows, Linux

2. **Signing Phase**:
   - macOS: Automatic signing and notarization
   - Windows: SignPath.io integration (when configured)

3. **Bundling Phase**: Creates platform-specific bundles
   - `gopca-macos-universal.zip`: All macOS binaries
   - `gopca-windows-x64.zip`: All Windows binaries
   - `gopca-linux-x64.tar.gz`: All Linux binaries

4. **Release Phase**: 
   - Creates GitHub release with bundled artifacts
   - Generates checksums
   - Auto-generates release notes from PRs

## Platform Bundles

Each platform bundle is a complete package containing all three GoPCA Suite tools:

### macOS Bundle (`gopca-macos-universal.zip`)
```
├── pca-intel          # pca CLI for Intel Macs
├── pca-arm64          # pca CLI for Apple Silicon
├── GoPCA.app/         # GoPCA Desktop (signed & notarized)
└── GoCSV.app/         # GoCSV Desktop (signed & notarized)
```

### Windows Bundle (`gopca-windows-x64.zip`)
```
├── pca.exe            # pca CLI
├── GoPCA.exe          # GoPCA Desktop
└── GoCSV.exe          # GoCSV Desktop
```
*All signed when SignPath is configured*

### Linux Bundle (`gopca-linux-x64.tar.gz`)
```
├── pca-x64            # pca CLI for x64
├── pca-arm64          # pca CLI for ARM64
├── GoPCA              # GoPCA Desktop
└── GoCSV              # GoCSV Desktop
```

## Creating a Release

### Automated Process (Recommended)

1. **Prepare the release**:
   ```bash
   ./scripts/prepare-release.sh v0.9.1
   ```
   This creates a release branch and updates version files.

2. **Create and merge PR**:
   ```bash
   git push -u origin release-v0.9.1
   gh pr create --title "Release v0.9.1" --body "Preparing release v0.9.1"
   # After CI passes and review, merge the PR
   ```

3. **Create and push tag**:
   ```bash
   git checkout main
   git pull origin main
   ./scripts/release.sh v0.9.1
   ```

4. **Monitor**: The release workflow automatically creates the GitHub release

### Manual Process

```bash
# Update versions in wails.json files
# Commit changes
git add .
git commit -m "chore: prepare release v0.9.1"

# Create and push tag
git tag -a v0.9.1 -m "Release v0.9.1"
git push origin v0.9.1
```

## Infrastructure

### Runners
- **GitHub-hosted**: Used for all macOS builds and GoPCA Desktop/GoCSV Desktop applications
- **Self-hosted**: Used for Linux/Windows pca CLI builds (cost optimization)

### Code Signing
- **macOS**: Automated via GitHub secrets (Apple Developer certificates)
- **Windows**: Optional SignPath.io integration
- **Linux**: Not currently implemented

## Cost Optimization Strategies

1. **Intelligent Triggers**: Prevents ~50% of redundant workflow runs
2. **Path Filtering**: Skips builds for non-code changes
3. **Self-hosted Runners**: Reduces costs for high-volume builds
4. **Artifact Retention**: 1-day retention for temporary artifacts
5. **Platform Bundling**: Reduces release asset count and API calls

## Windows Code Signing Setup

To enable Windows code signing:

1. Create SignPath.io account
2. Configure project with slug: `gopca`
3. Add signing policy with slug: `test-signing` or `release-signing`
4. Add GitHub secrets:
   - `SIGNPATH_API_TOKEN`
   - `SIGNPATH_ORG_ID`

The workflow automatically detects and uses SignPath when configured.

## Troubleshooting

### Release Workflow Fails
1. Check GitHub Actions logs for specific error
2. Common issues:
   - Self-hosted runner offline
   - SignPath authentication failure (non-blocking)
   - Missing artifacts

### Duplicate CI Runs
- This should not happen with current configuration
- If it does, check that feature branches don't have PRs to branches other than `main`

### Missing Binaries in Release
- Check artifact organization logs
- Ensure all build jobs completed successfully
- Verify artifact names match expected patterns

## Version Information

Versions are embedded at build time:
- **Source**: Git tags (e.g., `v0.9.1`)
- **pca CLI**: `pca version` or `pca --version`
- **GoPCA Desktop/GoCSV Desktop**: Displayed in application header

## Local Development

Build locally using the Makefile:
```bash
# Current platform
make build

# All platforms
make build-all

# Specific platform
make build-linux-amd64

# GoPCA Desktop and GoCSV Desktop
make pca-build
make csv-build
```