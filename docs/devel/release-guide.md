# GoPCA Release Guide

This guide describes how to create a new release of GoPCA.

## Prerequisites

- You must have push access to the repository
- GitHub CLI (`gh`) must be installed and authenticated
- You must be on the `main` branch with all changes synced

## Release Process

### 1. Prepare the Release

From the repository root, run the release preparation script:

```bash
./scripts/prepare-release.sh v0.9.0
```

Replace `v0.9.0` with your desired version number. The version must follow semantic versioning format: `vMAJOR.MINOR.PATCH`.

This script will:
- Verify you're on the main branch with no uncommitted changes
- Run all tests and linters
- Create a new release branch (e.g., `release-v0.9.0`)
- Update the version in `cmd/gopca-desktop/wails.json`
- Commit the version change

**Note:** GoCSV version in `cmd/gocsv/wails.json` should be updated manually if needed before the release.

### 2. Create Pull Request

Push the release branch to GitHub:

```bash
git push -u origin release-v0.9.0
```

Create a pull request using GitHub CLI:

```bash
gh pr create \
  --title "Release v0.9.0" \
  --body "Preparing release v0.9.0"
```

Or create the PR manually through the GitHub web interface.

### 3. Review and Merge

- Have the PR reviewed by another team member
- Ensure all CI checks pass
- Merge the PR into main

### 4. Create the Release

After the PR is merged, switch back to main and pull the latest changes:

```bash
git checkout main
git pull origin main
```

Run the release script:

```bash
./scripts/release.sh v0.9.0
```

This script will:
- Verify you're on the main branch
- Check that the version was properly updated in `wails.json`
- Create and push a git tag
- Create a GitHub release with auto-generated release notes

### 5. Post-Release

After the release is created:
1. Review the auto-generated release notes on GitHub
2. Edit the release notes if needed to highlight important changes
3. Monitor the automated release workflow which builds and uploads binary artifacts for all platforms

## Version Numbering

GoPCA follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (1.x.x): Incompatible API changes
- **MINOR** version (x.1.x): New functionality, backwards compatible
- **PATCH** version (x.x.1): Bug fixes, backwards compatible

### Pre-release Versions

For pre-release versions, append a suffix:
- Release candidates: `v0.9.0-rc.1`, `v0.9.0-rc.2`
- Beta releases: `v0.9.0-beta.1`
- Alpha releases: `v0.9.0-alpha.1`

## Hotfix Releases

For critical fixes that need to be released immediately:

1. Check out the tag of the current release:
   ```bash
   git checkout v0.9.0
   ```

2. Create a hotfix branch:
   ```bash
   git checkout -b hotfix-v0.9.1
   ```

3. Make your fixes and commit them

4. Run the prepare-release script:
   ```bash
   ./scripts/prepare-release.sh v0.9.1
   ```

5. Continue with the normal release process from step 2

## Version Information

### CLI

Check the version using:
```bash
gopca-cli --version  # Shows version number only (e.g., "0.9.0")
gopca-cli version    # Shows detailed version information
```

Example output:
```
$ gopca-cli version
GoPCA 0.9.0 (abc123) built on 2025-01-01T00:00:00Z with go1.24.5 for darwin/arm64
```

### Desktop Applications

- **GoPCA Desktop**: Version displayed next to the logo in the application header (e.g., "v0.9.0")
- **GoCSV**: Version displayed in the application header

## Troubleshooting

### Script Errors

If the prepare-release script fails:
- Ensure you have no uncommitted changes: `git status`
- Make sure all tests pass: `make test`
- Check that the linter passes: `make lint`

### Permission Issues

If you get permission errors:
- Ensure the scripts are executable: `chmod +x scripts/*.sh`
- Check your GitHub permissions for creating releases

### Version Conflicts

If a tag already exists:
- Check existing tags: `git tag -l`
- Delete local tag if needed: `git tag -d v0.9.0`
- Delete remote tag if needed: `git push origin :refs/tags/v0.9.0` (use with caution)

## Questions?

For questions about the release process, please open an issue on GitHub or contact the maintainers.