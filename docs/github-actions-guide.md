# GitHub Actions and Release Guide

This guide explains how GoPCA's automated build system works on GitHub, how to manually trigger builds, and how to create releases.

## Table of Contents
- [Overview](#overview)
- [Automated Build Workflow](#automated-build-workflow)
- [Release Workflow](#release-workflow)
- [Manual Build Triggers](#manual-build-triggers)
- [Creating a New Release](#creating-a-new-release)
- [Troubleshooting](#troubleshooting)

## Overview

GoPCA uses GitHub Actions for continuous integration and deployment. We have two main workflows:

1. **Build and Test** (`build.yml`) - Runs tests and builds on every push and PR
2. **Release** (`release.yml`) - Creates releases with binaries for all platforms

## Automated Build Workflow

The build workflow automatically runs when:
- Code is pushed to the `main` branch
- A pull request is opened or updated targeting `main`
- Manually triggered via workflow dispatch

### What it does:

1. **Tests** (all platforms: Ubuntu, macOS, Windows)
   - Runs all Go tests
   - Checks code coverage
   - Runs golangci-lint

2. **Builds CLI** (Ubuntu)
   - Builds binaries for all platforms:
     - macOS Intel (`darwin-amd64`)
     - macOS Apple Silicon (`darwin-arm64`)
     - Linux x64 (`linux-amd64`)
     - Linux ARM64 (`linux-arm64`)
     - Windows x64 (`windows-amd64`)

3. **Builds Desktop App** (platform-specific)
   - Builds Wails desktop application for each OS
   - Creates platform-specific packages

### Build Matrix

| Platform | CLI Binary | Desktop App |
|----------|------------|-------------|
| macOS Intel | ✅ | ✅ (.app) |
| macOS ARM | ✅ | ✅ (.app) |
| Linux x64 | ✅ | ✅ (AppImage) |
| Linux ARM64 | ✅ | ❌ |
| Windows x64 | ✅ | ✅ (.exe) |

## Release Workflow

The release workflow creates official releases when:
- A version tag is pushed (e.g., `v1.0.0`)
- Manually triggered with a version number

### What it does:

1. **Creates GitHub Release**
   - Generates release notes
   - Creates a release page

2. **Builds All Binaries**
   - CLI binaries for all platforms
   - Desktop applications for supported platforms
   - Embeds version information in binaries

3. **Packages Artifacts**
   - Creates `.tar.gz` archives for Unix platforms
   - Creates `.zip` archives for Windows
   - Generates checksums for all files

4. **Uploads to Release**
   - Attaches all binaries to the release
   - Includes checksum file

## Manual Build Triggers

### Triggering a Build Workflow

1. Go to the [Actions tab](https://github.com/bitjungle/gopca/actions)
2. Select "Build and Test" workflow
3. Click "Run workflow"
4. Select the branch to build
5. Click "Run workflow" button

### Triggering a Release Workflow

1. Go to the [Actions tab](https://github.com/bitjungle/gopca/actions)
2. Select "Release" workflow
3. Click "Run workflow"
4. Enter a version (e.g., `v0.1.3-test`)
5. Click "Run workflow" button

## Creating a New Release

### Prerequisites

- Ensure all tests pass on `main` branch
- Update version numbers in code if needed
- Update CHANGELOG.md with release notes

### Method 1: Using Git Tags (Recommended)

```bash
# 1. Ensure you're on main and up to date
git checkout main
git pull origin main

# 2. Create and push a version tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

The release workflow will automatically:
- Create a GitHub release
- Build all binaries with the version embedded
- Upload artifacts to the release page

### Method 2: Using GitHub UI

1. Go to [Releases page](https://github.com/bitjungle/gopca/releases)
2. Click "Draft a new release"
3. Click "Choose a tag" and create a new tag (e.g., `v1.0.0`)
4. Fill in release title and notes
5. Click "Publish release"

### Method 3: Using GitHub CLI

```bash
# Create a release with the gh command
gh release create v1.0.0 \
  --title "GoPCA v1.0.0" \
  --notes "Release notes here" \
  --draft
```

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):
- Format: `vMAJOR.MINOR.PATCH`
- Examples: `v1.0.0`, `v1.2.3`, `v2.0.0-beta.1`

**Version increments:**
- **MAJOR**: Breaking API changes
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)

## Workflow Configuration

### Build Workflow Triggers
```yaml
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]
  workflow_dispatch:  # Manual trigger
```

### Release Workflow Triggers
```yaml
on:
  push:
    tags:
      - 'v*'  # Triggers on version tags
  workflow_dispatch:  # Manual trigger
    inputs:
      version:
        description: 'Version to release'
        required: true
```

## Troubleshooting

### Common Issues

**Build fails on Ubuntu with webkit error**
- The workflow handles Ubuntu 24.04's different package names
- Uses `libwebkit2gtk-4.1-dev` on newer Ubuntu versions

**macOS build fails with "No such file or directory"**
- Ensure the build directory is created before moving artifacts
- Check the workflow creates `mkdir -p build`

**Version not showing in binary**
- Version is injected at build time via ldflags
- Check the format: `-X github.com/bitjungle/gopca/internal/cli.Version=v1.0.0`

**Workflows not visible in Actions tab**
- Workflows must exist in the default branch (`main`) to appear
- Create a PR to trigger workflows before merging

### Checking Workflow Runs

1. Go to [Actions tab](https://github.com/bitjungle/gopca/actions)
2. Click on a workflow run to see details
3. Click on individual jobs to see logs
4. Download artifacts from successful runs

### Re-running Failed Workflows

1. Navigate to the failed workflow run
2. Click "Re-run all jobs" or "Re-run failed jobs"
3. Check logs for specific error messages

## Security Notes

- Never commit sensitive data or credentials
- Use GitHub Secrets for sensitive values
- The `GITHUB_TOKEN` is automatically provided by GitHub Actions
- Binary signing is not currently implemented

## Further Reading

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Wails Build Documentation](https://wails.io/docs/reference/cli#build)
- [Go Cross Compilation](https://go.dev/doc/install/source#environment)