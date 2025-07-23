# GitHub Actions Workflows

This directory contains automated workflows for building and releasing GoPCA.

## Workflows

### build.yml - Continuous Integration
- **Triggers**: On every push to main and on pull requests
- **Actions**:
  - Runs tests on multiple platforms (Linux, macOS, Windows)
  - Runs linter
  - Builds CLI binaries for all platforms
  - Builds desktop applications
  - Uploads artifacts for inspection

### release.yml - Automated Releases
- **Triggers**: On pushing version tags (e.g., `v1.0.0`)
- **Actions**:
  - Creates a GitHub Release
  - Builds CLI binaries for all platforms:
    - macOS Intel (darwin-amd64)
    - macOS Apple Silicon (darwin-arm64)
    - Linux x64 (linux-amd64)
    - Linux ARM64 (linux-arm64)
    - Windows x64 (windows-amd64)
  - Builds desktop applications for macOS, Windows, and Linux
  - Generates checksums for all artifacts
  - Uploads all binaries to the release

## Creating a Release

To create a new release:

```bash
# Update version in code if needed
# Commit all changes
git add .
git commit -m "Prepare for release v1.0.0"

# Create and push a version tag
git tag -a v1.0.0 -m "Release version 1.0.0"
git push origin v1.0.0
```

The release workflow will automatically:
1. Create a GitHub Release
2. Build all binaries with the version embedded
3. Upload all artifacts
4. Generate checksums

## Version Information

Version information is automatically embedded in binaries:
- CLI: Available via `gopca-cli --version`
- Desktop: Available in the application

The version is determined by Git tags. Without tags, it uses the commit hash.

## Manual Builds

You can also build locally using the Makefile:
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build for specific platform
make build-linux-amd64
```