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

### Automated Execution (AI Assistant/Claude Code)

When using Claude Code or another AI assistant to create a release, the assistant can automate most steps:

**What the AI CAN do automatically:**
1. ✅ Check current version and determine next version
2. ✅ Verify main branch and clean working directory
3. ✅ Run tests and linters
4. ✅ Execute prepare-release.sh script
5. ✅ Push release branch
6. ✅ Create pull request
7. ✅ Monitor CI checks status
8. ✅ After PR merge: checkout main, pull, and run release.sh
9. ✅ Monitor release workflow progress

**What requires MANUAL action:**
- ❌ **Merge the PR** - Must be done via GitHub web interface due to branch protection rules

**AI Assistant Instructions:**
```bash
# The AI will execute these automatically:
./scripts/prepare-release.sh v0.9.1
git push -u origin release-v0.9.1
gh pr create --title "Release v0.9.1" --body "Preparing release v0.9.1"

# AI will monitor CI and notify when ready to merge
# After you merge, AI continues automatically:
git checkout main && git pull origin main
./scripts/release.sh v0.9.1
gh run watch  # AI monitors until complete
```

### Manual Execution (Human Developer)

If executing manually, follow these steps:

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
2. Build Desktop apps (3 platforms + Linux AppImage)
3. Build GoCSV apps (3 platforms + Linux AppImage)
4. Sign and notarize macOS applications (fully automated, no Gatekeeper issues)
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
- [ ] All binaries are attached (13 files + checksums)
- [ ] Release notes are accurate
- [ ] Download links work
- [ ] Checksums file is present
- [ ] Linux AppImages are included for both GoPCA and GoCSV
- [ ] macOS apps run without Gatekeeper warnings

## Artifacts Produced

Each release includes:

### CLI Binaries (5 files)
- `pca-darwin-amd64` - macOS Intel
- `pca-darwin-arm64` - macOS Apple Silicon
- `pca-linux-amd64` - Linux x64
- `pca-linux-arm64` - Linux ARM64
- `pca-windows-amd64.exe` - Windows x64

### Desktop Applications (4 files)
- `GoPCA-macos.zip` - macOS app (fully signed & notarized, no Gatekeeper warnings)
- `GoPCA-windows.exe` - Windows executable
- `GoPCA-linux` - Linux executable
- `GoPCA-linux.AppImage` - Linux AppImage (portable, works across distributions)

### GoCSV Editor (4 files)
- `GoCSV-macos.zip` - macOS app (fully signed & notarized, no Gatekeeper warnings)
- `GoCSV-windows.exe` - Windows executable
- `GoCSV-linux` - Linux executable
- `GoCSV-linux.AppImage` - Linux AppImage (portable, works across distributions)

### Windows Installer
- `GoPCA-Setup-vX.X.X.exe` - Windows installer containing all components
  - Includes GoPCA Desktop, GoCSV, and PCA CLI  
  - Automated installation to Program Files
  - Start Menu shortcuts and PATH configuration
  - Built automatically in CI/CD when NSIS is available
  - Uses signed binaries when SignPath is configured

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

## Testing the Release Workflow

You can test the release workflow without creating actual releases using the workflow_dispatch trigger:

### Manual Testing Steps
1. Go to [GitHub Actions](https://github.com/bitjungle/gopca/actions)
2. Select "Release" workflow from the left sidebar
3. Click "Run workflow" button
4. Configure the test run:
   - **Branch**: Select branch to test from (e.g., main or feature branch)
   - **Test version**: Use default `v0.9.5-test` or enter custom version
5. Click the green "Run workflow" button

### What Happens in Test Mode
- ✅ All binaries and installers are built
- ✅ Artifacts are uploaded and downloadable from the workflow run
- ✅ Clear "TEST MODE" notification appears in logs
- ❌ No GitHub release is created (intentional)
- ❌ No tag is created or pushed

This is useful for:
- Testing workflow changes before merging
- Verifying Windows installer builds correctly
- Checking cross-platform builds without creating releases

## Maintenance Releases

For bug fixes to the current stable version while v1.1 development continues:

### When to Create Maintenance Releases

Create a maintenance release (e.g., v1.0.2, v1.0.3) when:
- Bug affects current production users
- Fix cannot wait for next minor release (v1.1.0)
- Change is backward compatible (no breaking changes)
- Fix has been tested thoroughly

### Maintenance Release Process

```bash
# 1. Work from maintenance branch (not main!)
git checkout maintenance/v1.0.x
git pull origin maintenance/v1.0.x

# 2. Apply fixes (usually cherry-picked from develop)
# Option A: Cherry-pick existing fix
git cherry-pick <commit-hash-from-develop>

# Option B: Create fix directly
git checkout -b bugfix-critical-issue
# ... make changes ...
git commit -m "fix: critical bug in ..."
git push -u origin bugfix-critical-issue
gh pr create --base maintenance/v1.0.x

# 3. After fixes are merged to maintenance branch
git checkout maintenance/v1.0.x
git pull origin maintenance/v1.0.x

# 4. Prepare release
./scripts/prepare-release.sh v1.0.2

# 5. Create PR to main (not maintenance!)
git push -u origin release-v1.0.2
gh pr create --base main --title "Release v1.0.2"

# 6. After PR is merged, create release
git checkout main
git pull origin main
./scripts/release.sh v1.0.2

# 7. Update maintenance branch with release
git checkout maintenance/v1.0.x
git merge main
git push origin maintenance/v1.0.x
```

### Cherry-Picking Guidelines

When cherry-picking fixes between branches:

```bash
# Cherry-pick with commit reference
git cherry-pick -x <commit-hash>  # -x adds reference to original commit

# If conflicts occur
git status  # Check conflicted files
# ... resolve conflicts ...
git add .
git cherry-pick --continue

# Verify the fix works in this branch
make test
```

## Hotfix Releases

For emergency fixes when no maintenance branch exists or for critical security issues:

```bash
# 1. Create hotfix branch from main
git checkout main
git pull origin main
git checkout -b hotfix-v1.0.2-security

# 2. Make fixes and commit
# ... make changes ...
git commit -m "fix(security): critical vulnerability"

# 3. Prepare release
./scripts/prepare-release.sh v1.0.2

# 4. Continue normal release process from Step 2
```

After hotfix is released, merge back to develop and maintenance branches:

```bash
# Merge to develop
git checkout develop
git merge main
git push origin develop

# Merge to maintenance (if exists)
git checkout maintenance/v1.0.x
git merge main
git push origin maintenance/v1.0.x
```

## Multi-Version Support

When supporting multiple versions simultaneously (e.g., v1.0.x and v1.1.x):

### Branch Structure
- `main`: Latest stable release
- `develop`: Next minor/major version (v1.1.0)
- `maintenance/v1.0.x`: Bug fixes for v1.0 series
- `maintenance/v1.1.x`: Created after v1.1.0 release

### Version Decision Tree

| Scenario | Target Branch | Release Type |
|----------|--------------|--------------|
| New feature | develop | Next minor (v1.1.0) |
| Bug in v1.0.x only | maintenance/v1.0.x | Patch (v1.0.2) |
| Bug in both versions | maintenance/v1.0.x + cherry-pick to develop | Patch + include in next minor |
| Security issue | hotfix from main | Immediate patch |
| Breaking change | develop | Next major (v2.0.0) |

For comprehensive Git workflow documentation, see [git-workflow.md](git-workflow.md).

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

## Windows Code Signing (Manual Process)

After the automated release workflow completes, the Windows installer needs to be manually signed to avoid security warnings.

### Prerequisites

- SignPath.io account (free tier works)
- Access to the project's SignPath organization
- Signing certificate configured in SignPath

### Signing Process

1. **Download the unsigned installer**
   ```bash
   # From the release page, download:
   GoPCA-Setup-vX.X.X.exe
   ```

2. **Sign with SignPath**
   - Go to [SignPath.io](https://app.signpath.io)
   - Navigate to your organization
   - Click "Upload and Sign"
   - Select the downloaded installer file
   - Choose the appropriate signing certificate
   - Wait for signing to complete (usually ~1 minute)
   - Download the signed file

3. **Update the GitHub release**
   ```bash
   # Rename the signed file
   mv GoPCA-Setup-vX.X.X.exe GoPCA-Setup-vX.X.X-signed.exe
   
   # Upload to the existing release
   gh release upload vX.X.X GoPCA-Setup-vX.X.X-signed.exe
   ```

4. **Update release notes**
   - Edit the release on GitHub
   - Change the Windows installer download link from:
     ```
     GoPCA-Setup-vX.X.X.exe
     ```
     to:
     ```
     GoPCA-Setup-vX.X.X-signed.exe
     ```
   - Add a note: "(digitally signed)" next to the Windows installer link

### Why Manual Signing?

The free SignPath tier doesn't support automated GitHub Actions integration with repository-based policies. Manual signing through the web interface works reliably and only adds ~5 minutes to the release process.

### Future Improvements

If we upgrade to a paid SignPath plan, we can re-enable the automated signing in `.github/workflows/release.yml` (currently commented out).

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
   - `build-cli-binaries`: Builds pca CLI for 5 platforms
     - Self-hosted runner: Linux x64, Linux ARM64, Windows x64
     - GitHub runner: macOS Intel, macOS ARM
   - `build-desktop`: Builds GoPCA Desktop for 3 platforms
     - GitHub runners: ubuntu-latest, windows-latest, macos-latest
   - `build-gocsv`: Builds GoCSV Desktop for 3 platforms
     - GitHub runners: ubuntu-latest, windows-latest, macos-latest

3. **Release Job**:
   - Downloads all artifacts
   - Organizes and packages them
   - Generates checksums
   - Creates GitHub release using `softprops/action-gh-release`
   - Attaches all artifacts
   - Generates release notes

### Infrastructure

- **Self-hosted runner**: Used for binary builds to reduce costs
  - Linux runner with NSIS installed for Windows installer creation
- **GitHub-hosted runners**: Used for all testing
- **Code signing**:
  - **macOS**: Fully automated signing and notarization for all binaries (no Gatekeeper warnings)
  - **Windows**: Optional SignPath.io integration for digital signatures (when configured)
  - **Linux AppImages**: Automatically generated for both GoPCA and GoCSV Desktop apps
- **Windows Installer**: Built on self-hosted Linux runner using NSIS

## Questions?

For issues with the release process:
- Check GitHub Actions logs for errors
- Open an issue with the error details
- Contact maintainers if urgent