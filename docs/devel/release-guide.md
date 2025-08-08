# GoPCA Release Guide

This guide provides detailed instructions for creating consistent, automated releases of the GoPCA toolkit.

## Overview

The release process is fully automated via GitHub Actions. Once you push a version tag, the workflow:
1. Builds all binaries for all platforms
2. Signs and notarizes macOS applications
3. Creates the GitHub release with all artifacts attached
4. Generates release notes automatically

## Prerequisites

- Push access to the repository
- GitHub CLI (`gh`) installed and authenticated
- Be on the `main` branch with all changes synced
- All tests passing

## Release Process

### Step 1: Prepare the Release

Run the release preparation script with your desired version:

```bash
./scripts/prepare-release.sh v0.9.1
```

**Version format:** `vMAJOR.MINOR.PATCH` (e.g., v0.9.0, v1.0.0, v2.1.3)

This script will:
- Verify you're on main with no uncommitted changes
- Run all tests and linters
- Create a release branch (e.g., `release-v0.9.1`)
- Update version in both `cmd/gopca-desktop/wails.json` and `cmd/gocsv/wails.json`
- Commit the version changes

### Step 2: Create and Merge Pull Request

Push the release branch:
```bash
git push -u origin release-v0.9.1
```

Create the PR:
```bash
gh pr create \
  --title "Release v0.9.1" \
  --body "Preparing release v0.9.1"
```

Then:
1. Wait for CI checks to pass
2. Get PR reviewed if needed
3. Merge the PR

### Step 3: Create the Release

After the PR is merged:

```bash
# Switch to main and pull latest
git checkout main
git pull origin main

# Create and push the release tag
./scripts/release.sh v0.9.1
```

This script will:
- Verify you're on main
- Check versions match in all wails.json files
- Create an annotated git tag
- Push the tag to GitHub

**That's it!** The tag push triggers the automated release workflow.

### Step 4: Monitor the Release

Watch the automated process:
```bash
gh run watch
```

Or view in browser:
```bash
open https://github.com/bitjungle/gopca/actions
```

The workflow will:
1. Build CLI binaries (5 platforms)
2. Build Desktop apps (3 platforms)
3. Build GoCSV apps (3 platforms)
4. Sign and notarize macOS applications
5. Sign Windows binaries (if SignPath configured)
6. Generate SHA-256 checksums
7. Create GitHub release with all artifacts
8. Generate release notes from merged PRs

**Expected duration:** 15-25 minutes

### Step 5: Verify the Release

Once complete, verify at:
```bash
open https://github.com/bitjungle/gopca/releases/tag/v0.9.1
```

Check that:
- [ ] All binaries are attached (11 files + checksums)
- [ ] Release notes are accurate
- [ ] Download links work
- [ ] Checksums file is present

## Artifacts Produced

Each release includes:

### CLI Binaries (5 files)
- `pca-darwin-amd64` - macOS Intel
- `pca-darwin-arm64` - macOS Apple Silicon
- `pca-linux-amd64` - Linux x64
- `pca-linux-arm64` - Linux ARM64
- `pca-windows-amd64.exe` - Windows x64

### Desktop Applications (3 files)
- `GoPCA-macos.zip` - macOS app (signed & notarized)
- `GoPCA-windows.exe` - Windows executable
- `GoPCA-linux` - Linux executable

### GoCSV Editor (3 files)
- `GoCSV-macos.zip` - macOS app (signed & notarized)
- `GoCSV-windows.exe` - Windows executable
- `GoCSV-linux` - Linux executable

### Verification
- `checksums.txt` - SHA-256 checksums for all artifacts

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.0.0): Breaking changes
- **MINOR** (0.1.0): New features, backwards compatible
- **PATCH** (0.0.1): Bug fixes, backwards compatible

### Pre-releases
- Release candidates: `v1.0.0-rc.1`
- Beta releases: `v1.0.0-beta.1`
- Alpha releases: `v1.0.0-alpha.1`

## Hotfix Releases

For urgent fixes to the current release:

```bash
# 1. Create hotfix branch from tag
git checkout v0.9.0
git checkout -b hotfix-v0.9.1

# 2. Make fixes and commit
# ... make changes ...
git commit -m "fix: critical bug in ..."

# 3. Prepare release (creates release branch)
./scripts/prepare-release.sh v0.9.1

# 4. Continue normal release process from Step 2
```

## Troubleshooting

### Release Workflow Fails

If the workflow fails:
1. Check the error in GitHub Actions logs
2. Fix the issue in a new PR
3. After merging, delete the failed release and tag:
   ```bash
   gh release delete v0.9.1 --yes
   git push origin :refs/tags/v0.9.1
   git tag -d v0.9.1
   ```
4. Start over with `./scripts/release.sh v0.9.1`

### Version Mismatch

If release.sh reports version mismatch:
- Ensure the release PR was merged
- Check both wails.json files have correct version
- Pull latest changes: `git pull origin main`

### Tag Already Exists

If tag exists locally but not remotely:
```bash
git tag -d v0.9.1
./scripts/release.sh v0.9.1
```

If tag exists remotely (be careful!):
```bash
git push origin :refs/tags/v0.9.1
git tag -d v0.9.1
./scripts/release.sh v0.9.1
```

### Self-Hosted Runner Issues

If self-hosted runner is offline:
- CLI builds for Linux/Windows will fail
- Check runner status: Settings → Actions → Runners
- The workflow will wait for runner to come online

## Version Information

### CLI

Check the version using:
```bash
pca --version  # Shows version number only (e.g., "0.9.0")
pca version    # Shows detailed version information
```

Example output:
```
$ pca version
GoPCA 0.9.0 (abc123) built on 2025-01-01T00:00:00Z with go1.24.5 for darwin/arm64
```

### Desktop Applications

- **GoPCA Desktop**: Version displayed next to the logo in the application header (e.g., "v0.9.0")
- **GoCSV**: Version displayed in the application header

## Best Practices

1. **Always test locally first**: Run `make test` and `make lint`
2. **Use descriptive PR titles**: They become release notes
3. **Don't skip CI checks**: Let them complete before merging
4. **One release at a time**: Don't start a new release until previous completes
5. **Document breaking changes**: Clearly mark in PR descriptions

## How It Works (Technical Details)

### Release Scripts

1. **`scripts/prepare-release.sh`**:
   - Creates release branch
   - Updates versions in wails.json files
   - Commits changes
   - Ready for PR

2. **`scripts/release.sh`**:
   - Verifies main branch
   - Checks version consistency
   - Creates and pushes tag
   - Tag push triggers workflow

### GitHub Actions Workflow

The `.github/workflows/release.yml` workflow:

1. **Triggered by**: Push of tags matching `v*`

2. **Build Jobs** (run in parallel):
   - `build-cli-binaries`: Builds CLI for 5 platforms
     - Self-hosted runner: Linux x64, Linux ARM64, Windows x64
     - GitHub runner: macOS Intel, macOS ARM
   - `build-desktop`: Builds Desktop app for 3 platforms
     - GitHub runners: ubuntu-latest, windows-latest, macos-latest
   - `build-gocsv`: Builds GoCSV for 3 platforms
     - GitHub runners: ubuntu-latest, windows-latest, macos-latest

3. **Release Job**:
   - Downloads all artifacts
   - Organizes and packages them
   - Generates checksums
   - Creates GitHub release using `softprops/action-gh-release`
   - Attaches all artifacts
   - Generates release notes

### Infrastructure

- **Self-hosted runner**: Used for Linux and Windows CLI builds to reduce costs
- **GitHub-hosted runners**: Used for all macOS builds and desktop/GoCSV applications (all platforms)
- **Code signing**:
  - **macOS**: Automated signing and notarization for all binaries
  - **Windows**: Optional SignPath.io integration for digital signatures (when configured)

## Questions?

For issues with the release process:
- Check GitHub Actions logs for errors
- Open an issue with the error details
- Contact maintainers if urgent